package lib

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// InitGenesis initializes the contract with genesis state
func InitGenesis(msg *MsgInitGenesis) []byte {
	fmt.Println("InitGenesis called", nil)

	// Generate or use provided server secret
	serverSecret := msg.ServerSecret
	if serverSecret == "" {
		var err error
		serverSecret, err = GenerateRandomSecret()
		if err != nil {
			LoggerError("Failed to generate server secret", []string{"error", err.Error()})
			return MarshalJSON(map[string]string{"error": "failed to generate server secret"})
		}
	}

	// Save server secret
	SaveServerSecret(serverSecret)

	// Save funder private key if provided
	if msg.FunderPrivKey != "" {
		SaveFunderPrivateKey(msg.FunderPrivKey)
		fmt.Println("Funder private key initialized", nil)
	}

	// Save init account amount if provided
	if msg.InitAccountAmt.Amount == math.ZeroInt() {
		return MarshalJSON(map[string]string{"error": "init account amount must be greater than zero"})
	}

	if err := SaveInitAccountAmount(msg.InitAccountAmt); err != nil {
		return MarshalJSON(map[string]string{"error": "Failed to save init account amount"})
	}

	// Gas price is required
	if msg.GasPrice.Amount.IsNil() {
		return MarshalJSON(map[string]string{"error": "gas price is required"})
	}
	if msg.GasPrice.Amount.IsZero() {
		return MarshalJSON(map[string]string{"error": "gas price must be greater than zero"})
	}
	if err := SaveGasPrice(msg.GasPrice); err != nil {
		return MarshalJSON(map[string]string{"error": "Failed to save gas price"})
	}

	// Save route prefix if provided (for HTTP route registration later)
	if msg.RoutePrefix != "" {
		SaveRoutePrefix(msg.RoutePrefix)
	}

	fmt.Println("Server secret initialized", nil)
	return MarshalJSON(map[string]bool{"success": true})
}

// GenerateEphemeralKey generates a new ephemeral key pair
func GenerateEphemeralKey(msg *MsgGenerateEphemeralKey) []byte {
	fmt.Println("GenerateEphemeralKey called", []string{"user_id", msg.UserID, "service_domain", msg.ServiceDomain})

	// Check if OAuth token already has a key
	existingPubKey := LoadPublicKeyByToken(msg.OAuthToken)
	if len(existingPubKey) > 0 {
		fmt.Println("OAuth token already has a key", "oauth_token", msg.OAuthToken, "existingPubKey", hex.EncodeToString(existingPubKey))
		return MarshalJSON(map[string]string{"error": "OAuth token already has an associated key"})
	}

	// Generate key pair
	publicKey, privateKey, err := GenerateKeyPair()
	if err != nil {
		LoggerError("Failed to generate key pair", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to generate key pair"})
	}

	// Load server secret
	serverSecret := LoadServerSecret()
	if serverSecret == "" {
		LoggerError("Server secret not initialized", nil)
		return MarshalJSON(map[string]string{"error": "server secret not initialized"})
	}

	// Derive encryption key from server secret and OAuth token
	encryptionKey, err := DeriveEncryptionKey(serverSecret, msg.OAuthToken)
	if err != nil {
		LoggerError("Failed to derive encryption key", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to derive encryption key"})
	}

	// Encrypt private key
	encryptedPrivateKey, err := EncryptPrivateKey(privateKey, encryptionKey)
	if err != nil {
		LoggerError("Failed to encrypt private key", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to encrypt private key"})
	}

	// Create key pair record
	keyPair := &EphemeralKeyPair{
		PublicKey:     publicKey,
		PrivateKey:    encryptedPrivateKey,
		UserID:        msg.UserID,
		ServiceDomain: msg.ServiceDomain,
		CreatedAt:     GetBlockTime(),
		ExpiresAt:     msg.ExpiresAt,
		OAuthToken:    msg.OAuthToken,
	}

	// Save key pair
	if err := SaveKeyPair(keyPair); err != nil {
		LoggerError("Failed to save key pair", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save key pair"})
	}

	// Save token mapping
	SaveTokenMapping(msg.OAuthToken, publicKey)

	fmt.Println("Ephemeral key generated", []string{"public_key", hex.EncodeToString(publicKey), "user_id", msg.UserID})

	// Derive address from public key
	address, err := DeriveAddressFromPublicKey(publicKey)
	if err != nil {
		LoggerError("Failed to derive address from public key", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to derive address"})
	}

	fmt.Println("Ephemeral address derived", []string{"address", address})

	// Register address with wasmx-identity
	if err := registerAddressWithIdentity(msg.UserID, address, publicKey, msg.ServiceDomain, msg.ExpiresAt); err != nil {
		LoggerError("Failed to register address with identity", []string{"error", err.Error()})
		// Don't fail the whole operation, just log the error
		// The ephemeral key is still usable, just not linked to identity
	}

	// Fund the ephemeral account
	initAccountResp := InitAccount(&MsgInitAccount{Address: address})
	fmt.Println("InitAccount response", []string{"response", string(initAccountResp)})

	return MarshalJSON(MsgGenerateEphemeralKeyResponse{
		PublicKey:  publicKey,
		PrivateKey: privateKey, // Return plaintext for browser storage
		Success:    true,
	})
}

// RegisterExternalKey registers a key pair generated externally (e.g., in browser)
func RegisterExternalKey(msg *MsgRegisterExternalKey) []byte {
	fmt.Println("RegisterExternalKey called", []string{"user_id", msg.UserID, "public_key", hex.EncodeToString(msg.PublicKey)})

	// Check if OAuth token already has a key
	existingPubKey := LoadPublicKeyByToken(msg.OAuthToken)
	if existingPubKey != nil {
		LoggerError("OAuth token already has a key", []string{"oauth_token", msg.OAuthToken})
		return MarshalJSON(map[string]string{"error": "OAuth token already has an associated key"})
	}

	// Load server secret
	serverSecret := LoadServerSecret()
	if serverSecret == "" {
		LoggerError("Server secret not initialized", nil)
		return MarshalJSON(map[string]string{"error": "server secret not initialized"})
	}

	// Derive encryption key from server secret and OAuth token
	encryptionKey, err := DeriveEncryptionKey(serverSecret, msg.OAuthToken)
	if err != nil {
		LoggerError("Failed to derive encryption key", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to derive encryption key"})
	}

	// Encrypt private key
	encryptedPrivateKey, err := EncryptPrivateKey(msg.PrivateKey, encryptionKey)
	if err != nil {
		LoggerError("Failed to encrypt private key", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to encrypt private key"})
	}

	// Create key pair record
	keyPair := &EphemeralKeyPair{
		PublicKey:     msg.PublicKey,
		PrivateKey:    encryptedPrivateKey,
		UserID:        msg.UserID,
		ServiceDomain: msg.ServiceDomain,
		CreatedAt:     GetBlockTime(),
		ExpiresAt:     msg.ExpiresAt,
		OAuthToken:    msg.OAuthToken,
	}

	// Save key pair
	if err := SaveKeyPair(keyPair); err != nil {
		LoggerError("Failed to save key pair", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "failed to save key pair"})
	}

	// Save token mapping
	SaveTokenMapping(msg.OAuthToken, msg.PublicKey)

	fmt.Println("External key registered", []string{"public_key", hex.EncodeToString(msg.PublicKey), "user_id", msg.UserID})

	return MarshalJSON(MsgRegisterExternalKeyResponse{Success: true})
}

// RevokeKey revokes an ephemeral key
func RevokeKey(msg *MsgRevokeKey) []byte {
	fmt.Println("RevokeKey called", []string{"oauth_token", msg.OAuthToken})

	// Get public key from OAuth token
	publicKey := LoadPublicKeyByToken(msg.OAuthToken)
	if publicKey == nil {
		LoggerError("No key found for OAuth token", []string{"oauth_token", msg.OAuthToken})
		return MarshalJSON(map[string]string{"error": "no key found for OAuth token"})
	}

	// Delete key pair
	DeleteKeyPair(publicKey)

	// Delete token mapping
	DeleteTokenMapping(msg.OAuthToken)

	fmt.Println("Key revoked", []string{"public_key", hex.EncodeToString(publicKey)})

	return MarshalJSON(MsgRevokeKeyResponse{Success: true})
}

// DeleteExpiredKeys cleans up expired keys
func DeleteExpiredKeys(msg *MsgDeleteExpiredKeys) []byte {
	fmt.Println("DeleteExpiredKeys called", nil)

	// Note: In a real implementation, we would iterate through all keys
	// For now, this is a placeholder that returns 0 deleted
	// Proper implementation would require iteration support in storage

	deletedCount := 0

	fmt.Println("Expired keys deleted", []string{"count", fmt.Sprintf("%d", deletedCount)})

	return MarshalJSON(MsgDeleteExpiredKeysResponse{DeletedCount: deletedCount})
}

// QueryGetPublicKey retrieves public key for an OAuth token
func QueryGetPublicKey(msg *MsgQueryGetPublicKey) []byte {
	publicKey := LoadPublicKeyByToken(msg.OAuthToken)
	if publicKey == nil {
		return MarshalJSON(map[string]string{"error": "no key found for OAuth token"})
	}

	keyPair, err := LoadKeyPair(publicKey)
	if err != nil || keyPair == nil {
		return MarshalJSON(map[string]string{"error": "failed to load key pair"})
	}

	return MarshalJSON(QueryGetPublicKeyResponse{
		PublicKey: publicKey,
		UserID:    keyPair.UserID,
		ExpiresAt: keyPair.ExpiresAt,
	})
}

// QueryGetKeyInfo retrieves key information (without private key)
func QueryGetKeyInfo(msg *MsgQueryGetKeyInfo) []byte {
	keyPair, err := LoadKeyPair(msg.PublicKey)
	if err != nil {
		return MarshalJSON(map[string]string{"error": "failed to load key pair"})
	}
	if keyPair == nil {
		return MarshalJSON(map[string]string{"error": "key not found"})
	}

	return MarshalJSON(QueryGetKeyInfoResponse{
		UserID:        keyPair.UserID,
		ServiceDomain: keyPair.ServiceDomain,
		ExpiresAt:     keyPair.ExpiresAt,
		CreatedAt:     keyPair.CreatedAt,
	})
}

// QueryValidateAndGetKey validates OAuth token and returns ephemeral key with private key
func QueryValidateAndGetKey(msg *MsgQueryValidateAndGetKey) []byte {
	// Get public key from OAuth token
	publicKey := LoadPublicKeyByToken(msg.OAuthToken)
	if publicKey == nil {
		return MarshalJSON(QueryValidateAndGetKeyResponse{
			Valid:  false,
			Reason: "no key found for OAuth token",
		})
	}

	// Load key pair
	keyPair, err := LoadKeyPair(publicKey)
	if err != nil || keyPair == nil {
		return MarshalJSON(QueryValidateAndGetKeyResponse{
			Valid:  false,
			Reason: "failed to load key pair",
		})
	}

	// Check if key has expired
	currentTime := GetBlockTime()
	if keyPair.ExpiresAt > 0 && currentTime > keyPair.ExpiresAt {
		return MarshalJSON(QueryValidateAndGetKeyResponse{
			Valid:  false,
			Reason: "key has expired",
		})
	}

	// Load server secret
	serverSecret := LoadServerSecret()
	if serverSecret == "" {
		return MarshalJSON(QueryValidateAndGetKeyResponse{
			Valid:  false,
			Reason: "server secret not initialized",
		})
	}

	// Derive encryption key
	encryptionKey, err := DeriveEncryptionKey(serverSecret, msg.OAuthToken)
	if err != nil {
		return MarshalJSON(QueryValidateAndGetKeyResponse{
			Valid:  false,
			Reason: "failed to derive encryption key",
		})
	}

	// Decrypt private key
	privateKey, err := DecryptPrivateKey(keyPair.PrivateKey, encryptionKey)
	if err != nil {
		return MarshalJSON(QueryValidateAndGetKeyResponse{
			Valid:  false,
			Reason: "failed to decrypt private key",
		})
	}

	// Derive address from public key
	address, err := DeriveAddressFromPublicKey(publicKey)
	if err != nil {
		return MarshalJSON(QueryValidateAndGetKeyResponse{
			Valid:  false,
			Reason: "failed to derive address from public key",
		})
	}

	return MarshalJSON(QueryValidateAndGetKeyResponse{
		Valid:      true,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	})
}

// registerAddressWithIdentity registers the ephemeral address with the wasmx-identity contract
func registerAddressWithIdentity(userID, address string, publicKey []byte, serviceDomain string, expiresAt int64) error {
	// Get wasmx-identity contract address
	identityAddr := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	if identityAddr == "" {
		return errors.New("identity contract not found")
	}

	fmt.Println("registerAddressWithIdentity", []string{"identity_addr", string(identityAddr), "user_id", userID, "address", address})

	// First, check if user exists by querying identity
	queryMsg := map[string]interface{}{
		"query_user_by_id": map[string]string{
			"user_id": userID,
		},
	}
	queryBz, _ := json.Marshal(queryMsg)

	ok, respData := wasmx.CallSimple(identityAddr, queryBz, true, MODULE_NAME)
	fmt.Println("Query user by ID response", []string{"ok", fmt.Sprintf("%v", ok), "response", string(respData)})

	// Check if user exists
	var queryResp struct {
		User  interface{} `json:"user"`
		Error string      `json:"error"`
	}
	json.Unmarshal(respData, &queryResp)

	if queryResp.Error != "" && queryResp.Error == "user not found" {
		// User doesn't exist, create with RegisterUser
		fmt.Println("User not found, registering new user", []string{"user_id", userID})
		registerMsg := map[string]interface{}{
			"register_user": map[string]interface{}{
				"address":        address,
				"public_key":     publicKey,
				"service_domain": serviceDomain,
				"expires_at":     expiresAt,
			},
		}
		registerBz, _ := json.Marshal(registerMsg)

		ok, respData = wasmx.CallSimple(identityAddr, registerBz, false, MODULE_NAME)
		fmt.Println("RegisterUser response", []string{"ok", fmt.Sprintf("%v", ok), "response", string(respData)})

		if !ok {
			return errors.New("failed to register user: " + string(respData))
		}

		// Note: RegisterUser generates its own userID, but we're using the OAuth2 userID
		// This means we need to track the mapping separately or modify identity contract
		// For now, we'll proceed with the identity contract's generated userID
		return nil
	}

	// User exists, add address using AddAddressInternal
	fmt.Println("User exists, adding address internally", []string{"user_id", userID})
	addMsg := map[string]interface{}{
		"add_address_internal": map[string]interface{}{
			"user_id":        userID,
			"address":        address,
			"public_key":     publicKey,
			"service_domain": serviceDomain,
			"expires_at":     expiresAt,
		},
	}
	addBz, _ := json.Marshal(addMsg)

	ok, respData = wasmx.CallSimple(identityAddr, addBz, false, MODULE_NAME)
	fmt.Println("AddAddressInternal response", []string{"ok", fmt.Sprintf("%v", ok), "response", string(respData)})

	if !ok {
		return errors.New("failed to add address: " + string(respData))
	}

	return nil
}
