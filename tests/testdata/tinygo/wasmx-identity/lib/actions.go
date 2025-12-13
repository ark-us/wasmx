package lib

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// InitGenesis initializes the contract with genesis state
func InitGenesis(msg *MsgInitGenesis) []byte {
	LoggerInfo("InitGenesis called", nil)

	if msg != nil && len(msg.InitialUsers) > 0 {
		for _, initialUser := range msg.InitialUsers {
			if len(initialUser.Addresses) == 0 {
				continue
			}

			// Register first address to create user
			firstAddr := initialUser.Addresses[0]
			resp := RegisterUser(&MsgRegisterUser{
				Address:       firstAddr.Address,
				PublicKey:     firstAddr.PublicKey,
				ServiceDomain: firstAddr.ServiceDomain,
				Permissions:   firstAddr.Permissions,
				ExpiresAt:     firstAddr.ExpiresAt,
			})

			var regResp MsgRegisterUserResponse
			json.Unmarshal(resp, &regResp)
			userID := regResp.UserID

			// Add remaining addresses
			for i := 1; i < len(initialUser.Addresses); i++ {
				addr := initialUser.Addresses[i]
				AddAddress(&MsgAddAddress{
					UserID:        userID,
					Address:       addr.Address,
					PublicKey:     addr.PublicKey,
					ServiceDomain: addr.ServiceDomain,
					Permissions:   addr.Permissions,
					ExpiresAt:     addr.ExpiresAt,
				})
			}
		}
	}

	return MarshalJSON(map[string]bool{"success": true})
}

// RegisterUser registers a new user with an initial address
func RegisterUser(msg *MsgRegisterUser) []byte {
	LoggerInfo("RegisterUser called", []string{"address", msg.Address})

	// Check if address is already associated with a user
	existingUserID := LoadUserIDByAddress(msg.Address)
	if existingUserID != "" {
		LoggerError("Address already registered", []string{"address", msg.Address, "user_id", existingUserID})
		return MarshalJSON(map[string]string{"error": "address already registered"})
	}

	// Generate new user ID
	userID := GetNextUserID()
	blockTime := GetBlockTime()

	// Create user identity
	user := &UserIdentity{
		UserID:         userID,
		PrimaryAddress: msg.Address, // First address is primary
		Addresses:      []string{msg.Address},
		CreatedAt:      blockTime,
		UpdatedAt:      blockTime,
	}

	// Save user
	if err := SaveUser(user); err != nil {
		LoggerError("Failed to save user", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save user"})
	}

	// Create address info
	addrInfo := &AddressInfo{
		Address:       msg.Address,
		PublicKey:     msg.PublicKey,
		ServiceDomain: msg.ServiceDomain,
		Permissions:   msg.Permissions,
		ExpiresAt:     msg.ExpiresAt,
		CreatedAt:     blockTime,
	}

	// Save address info
	if err := SaveAddressInfo(userID, addrInfo); err != nil {
		LoggerError("Failed to save address info", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save address info"})
	}

	// Save address to user ID mapping
	SaveUserIDByAddress(msg.Address, userID)

	LoggerInfo("User registered", []string{"user_id", userID, "address", msg.Address})

	return MarshalJSON(MsgRegisterUserResponse{UserID: userID})
}

// AddAddress adds a new address to an existing user
func AddAddress(msg *MsgAddAddress) []byte {
	LoggerInfo("AddAddress called", []string{"user_id", msg.UserID, "address", msg.Address})

	// Validate transaction sender is associated with this user
	sender := wasmx.GetCaller()
	senderUserID := LoadUserIDByAddress(string(sender))
	if senderUserID != msg.UserID {
		LoggerError("Sender not authorized", []string{"sender", string(sender), "user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "sender not authorized for this user"})
	}

	// Load user
	user, err := LoadUser(msg.UserID)
	if err != nil || user == nil {
		LoggerError("User not found", []string{"user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "user not found"})
	}

	// Check if address is already associated with a user
	existingUserID := LoadUserIDByAddress(msg.Address)
	if existingUserID != "" {
		LoggerError("Address already registered", []string{"address", msg.Address, "existing_user_id", existingUserID})
		return MarshalJSON(map[string]string{"error": "address already registered to another user"})
	}

	// Check if address already in user's list
	if AddressInList(msg.Address, user.Addresses) {
		LoggerError("Address already in user list", []string{"address", msg.Address})
		return MarshalJSON(map[string]string{"error": "address already in user list"})
	}

	// Add address to user's list
	user.Addresses = append(user.Addresses, msg.Address)
	user.UpdatedAt = GetBlockTime()

	// Save updated user
	if err := SaveUser(user); err != nil {
		LoggerError("Failed to save user", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save user"})
	}

	// Create address info
	addrInfo := &AddressInfo{
		Address:       msg.Address,
		PublicKey:     msg.PublicKey,
		ServiceDomain: msg.ServiceDomain,
		Permissions:   msg.Permissions,
		ExpiresAt:     msg.ExpiresAt,
		CreatedAt:     GetBlockTime(),
	}

	// Save address info
	if err := SaveAddressInfo(msg.UserID, addrInfo); err != nil {
		LoggerError("Failed to save address info", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save address info"})
	}

	// Save address to user ID mapping
	SaveUserIDByAddress(msg.Address, msg.UserID)

	LoggerInfo("Address added", []string{"user_id", msg.UserID, "address", msg.Address})

	return MarshalJSON(MsgAddAddressResponse{Success: true})
}

// RemoveAddress removes an address from a user
func RemoveAddress(msg *MsgRemoveAddress) []byte {
	LoggerInfo("RemoveAddress called", []string{"user_id", msg.UserID, "address", msg.Address})

	// Validate transaction sender is associated with this user
	sender := wasmx.GetCaller()
	senderUserID := LoadUserIDByAddress(string(sender))
	if senderUserID != msg.UserID {
		LoggerError("Sender not authorized", []string{"sender", string(sender), "user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "sender not authorized for this user"})
	}

	// Load user
	user, err := LoadUser(msg.UserID)
	if err != nil || user == nil {
		LoggerError("User not found", []string{"user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "user not found"})
	}

	// Check if address exists in user's list
	if !AddressInList(msg.Address, user.Addresses) {
		LoggerError("Address not in user list", []string{"address", msg.Address})
		return MarshalJSON(map[string]string{"error": "address not in user list"})
	}

	// Don't allow removing the last address
	if len(user.Addresses) == 1 {
		LoggerError("Cannot remove last address", []string{"user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "cannot remove last address from user"})
	}

	// Remove address from user's list
	user.Addresses = RemoveAddressFromList(msg.Address, user.Addresses)
	user.UpdatedAt = GetBlockTime()

	// Save updated user
	if err := SaveUser(user); err != nil {
		LoggerError("Failed to save user", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save user"})
	}

	// Delete address info
	DeleteAddressInfo(msg.UserID, msg.Address)

	// Delete address to user ID mapping
	DeleteUserIDByAddress(msg.Address)

	LoggerInfo("Address removed", []string{"user_id", msg.UserID, "address", msg.Address})

	return MarshalJSON(MsgRemoveAddressResponse{Success: true})
}

// UpdatePermissions updates permissions for an existing address
func UpdatePermissions(msg *MsgUpdatePermissions) []byte {
	LoggerInfo("UpdatePermissions called", []string{"user_id", msg.UserID, "address", msg.Address})

	// Validate transaction sender is associated with this user
	sender := wasmx.GetCaller()
	senderUserID := LoadUserIDByAddress(string(sender))
	if senderUserID != msg.UserID {
		LoggerError("Sender not authorized", []string{"sender", string(sender), "user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "sender not authorized for this user"})
	}

	// Load address info
	addrInfo, err := LoadAddressInfo(msg.UserID, msg.Address)
	if err != nil || addrInfo == nil {
		LoggerError("Address info not found", []string{"user_id", msg.UserID, "address", msg.Address})
		return MarshalJSON(map[string]string{"error": "address not found"})
	}

	// Update permissions
	addrInfo.Permissions = msg.Permissions

	// Save updated address info
	if err := SaveAddressInfo(msg.UserID, addrInfo); err != nil {
		LoggerError("Failed to save address info", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save address info"})
	}

	LoggerInfo("Permissions updated", []string{"user_id", msg.UserID, "address", msg.Address})

	return MarshalJSON(MsgUpdatePermissionsResponse{Success: true})
}

// SetPrimaryAddress sets the primary address for a user
func SetPrimaryAddress(msg *MsgSetPrimaryAddress) []byte {
	LoggerInfo("SetPrimaryAddress called", []string{"user_id", msg.UserID, "address", msg.Address})

	// Validate transaction sender is associated with this user
	sender := wasmx.GetCaller()
	senderUserID := LoadUserIDByAddress(string(sender))
	if senderUserID != msg.UserID {
		LoggerError("Sender not authorized", []string{"sender", string(sender), "user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "sender not authorized for this user"})
	}

	// Load user
	user, err := LoadUser(msg.UserID)
	if err != nil || user == nil {
		LoggerError("User not found", []string{"user_id", msg.UserID})
		return MarshalJSON(map[string]string{"error": "user not found"})
	}

	// Check if address exists in user's list
	if !AddressInList(msg.Address, user.Addresses) {
		LoggerError("Address not in user list", []string{"address", msg.Address})
		return MarshalJSON(map[string]string{"error": "address not in user list"})
	}

	// Set primary address
	user.PrimaryAddress = msg.Address
	user.UpdatedAt = GetBlockTime()

	// Save updated user
	if err := SaveUser(user); err != nil {
		LoggerError("Failed to save user", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save user"})
	}

	LoggerInfo("Primary address set", []string{"user_id", msg.UserID, "address", msg.Address})

	return MarshalJSON(MsgSetPrimaryAddressResponse{Success: true})
}

// QueryUserByID retrieves user information by user ID
func QueryUserByID(msg *MsgQueryUserByID) []byte {
	user, err := LoadUser(msg.UserID)
	if err != nil {
		return MarshalJSON(map[string]string{"error": "failed to load user"})
	}
	if user == nil {
		return MarshalJSON(map[string]string{"error": "user not found"})
	}

	return MarshalJSON(QueryUserByIDResponse{User: *user})
}

// QueryUserByAddress retrieves user information by address
func QueryUserByAddress(msg *MsgQueryUserByAddress) []byte {
	userID := LoadUserIDByAddress(msg.Address)
	if userID == "" {
		return MarshalJSON(map[string]string{"error": "address not registered"})
	}

	user, err := LoadUser(userID)
	if err != nil {
		return MarshalJSON(map[string]string{"error": "failed to load user"})
	}
	if user == nil {
		return MarshalJSON(map[string]string{"error": "user not found"})
	}

	return MarshalJSON(QueryUserByAddressResponse{UserID: userID, User: *user})
}

// QueryAddressInfo retrieves address information
func QueryAddressInfo(msg *MsgQueryAddressInfo) []byte {
	addrInfo, err := LoadAddressInfo(msg.UserID, msg.Address)
	if err != nil {
		return MarshalJSON(map[string]string{"error": "failed to load address info"})
	}
	if addrInfo == nil {
		return MarshalJSON(map[string]string{"error": "address not found"})
	}

	return MarshalJSON(QueryAddressInfoResponse{AddressInfo: *addrInfo})
}

// QueryValidatePermission validates if an address has permission for an operation
func QueryValidatePermission(msg *MsgQueryValidatePermission) []byte {
	// Get user ID for address
	userID := LoadUserIDByAddress(msg.Address)
	if userID == "" {
		return MarshalJSON(QueryValidatePermissionResponse{
			Valid:  false,
			Reason: "address not registered",
		})
	}

	// Load address info
	addrInfo, err := LoadAddressInfo(userID, msg.Address)
	if err != nil || addrInfo == nil {
		return MarshalJSON(QueryValidatePermissionResponse{
			Valid:  false,
			Reason: "address info not found",
		})
	}

	// Check expiry
	if addrInfo.ExpiresAt > 0 && GetBlockTime() > addrInfo.ExpiresAt {
		return MarshalJSON(QueryValidatePermissionResponse{
			Valid:  false,
			Reason: "address expired",
		})
	}

	// Check permissions
	hasPermission := false
	for _, perm := range addrInfo.Permissions {
		if perm.Type == msg.OperationType {
			// TODO: Add more sophisticated permission validation based on OperationData and perm.Params
			// For now, just check if the permission type exists
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return MarshalJSON(QueryValidatePermissionResponse{
			Valid:  false,
			Reason: fmt.Sprintf("no permission for operation: %s", msg.OperationType),
		})
	}

	return MarshalJSON(QueryValidatePermissionResponse{Valid: true})
}
