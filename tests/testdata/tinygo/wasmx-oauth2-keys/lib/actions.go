package lib

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/math"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
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
		IdentityUserID: msg.IdentityUserID,
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

	// Fund the ephemeral account first (before registration)
	// The account needs to exist and have funds to pay for gas
	initAccountResp := InitAccount(&MsgInitAccount{Address: address})
	fmt.Println("InitAccount response", []string{"response", string(initAccountResp)})

	// Link the ephemeral address to the existing identity user via internal forward
	identityUserID := msg.IdentityUserID
	if identityUserID == "" {
		identityUserID = msg.UserID
	}
	if err := broadcastAddAddressInternalForward(identityUserID, address, publicKey, msg.ServiceDomain, msg.ExpiresAt, privateKey); err != nil {
		LoggerError("Failed to link ephemeral address to identity", []string{"error", err.Error(), "address", address})
		fmt.Println("ERROR: Failed to link ephemeral address to identity:", err.Error())
	}

	// Return key info (identity linking happens via AddAddressInternalForward)
	return MarshalJSON(MsgGenerateEphemeralKeyResponse{
		PublicKey:  publicKey,
		PrivateKey: privateKey, // Return plaintext for browser storage
		Address:    address,    // Derived bech32 address
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

// broadcastIdentityRegistration prepares and broadcasts a register_user transaction to wasmx-identity
// This uses the ephemeral key to sign the transaction (self-registration).
// For EXISTING users, this will fail because wasmx-identity requires self-registration.
// Existing users must sign add_address transactions themselves from their existing address.
func broadcastIdentityRegistration(address string, publicKey []byte, privateKey []byte, serviceDomain string, expiresAt int64) error {
	// Get wasmx-identity contract address
	identityAddr := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	if identityAddr == "" {
		return errors.New("identity contract not found")
	}

	fmt.Println("broadcastIdentityRegistration", []string{"identity_addr", string(identityAddr), "address", address})

	// Prepare register_user message (self-registration)
	registerMsg := map[string]interface{}{
		"register_user": map[string]interface{}{
			"address":        address,
			"public_key":     publicKey,
			"service_domain": serviceDomain,
			"expires_at":     expiresAt,
		},
	}
	msgBz, err := json.Marshal(registerMsg)
	if err != nil {
		return errors.New("failed to marshal register message: " + err.Error())
	}

	// Load gas price
	gasPrice := LoadGasPrice()
	if gasPrice == nil {
		gasPrice = &wasmx.Coin{Denom: "amyt", Amount: math.NewInt(10)}
	}

	// Prepare transaction: execute wasmx-identity.register_user
	// The ephemeral key signs this - it's self-registration (sender == address being registered)
	txBytes, err := wasmxcore.PrepareTx(
		string(identityAddr), // target contract
		msgBz,                // message data
		nil,                  // no coins to send
		nil,                  // no memo
		5000000,              // gas limit
		*gasPrice,            // gas price
		privateKey,           // sign with ephemeral key (self-registration)
	)
	if err != nil {
		return errors.New("failed to prepare registration tx: " + err.Error())
	}

	fmt.Println("broadcastIdentityRegistration: transaction prepared", []string{"tx_bytes_len", fmt.Sprintf("%d", len(txBytes))})

	// Broadcast transaction asynchronously
	broadcastResp, err := wasmxcore.BroadcastTxAsync(txBytes)
	if err != nil {
		return errors.New("failed to broadcast registration tx: " + err.Error())
	}

	if broadcastResp.Error != "" {
		return errors.New("registration broadcast error: " + broadcastResp.Error)
	}

	fmt.Println("broadcastIdentityRegistration: tx broadcasted", []string{"tx_hash", hex.EncodeToString(broadcastResp.TxHash)})
	return nil
}

// broadcastAddAddressInternalForward calls this contract's AddAddressInternalForward using the ephemeral key.
func broadcastAddAddressInternalForward(userID, address string, publicKey []byte, serviceDomain string, expiresAt int64, privateKey []byte) error {
	keysAddr := wasmx.GetAddressByRole(wasmx.ROLE_OAUTH2_KEYS)
	if keysAddr == "" {
		return errors.New("oauth2-keys contract not found")
	}

	callData := map[string]interface{}{
		"add_address_internal_forward": map[string]interface{}{
			"user_id":        userID,
			"address":        address,
			"public_key":     publicKey,
			"service_domain": serviceDomain,
			"expires_at":     expiresAt,
		},
	}
	msgBz, err := json.Marshal(callData)
	if err != nil {
		return errors.New("failed to marshal forward message: " + err.Error())
	}

	gasPrice := LoadGasPrice()
	if gasPrice == nil {
		gasPrice = &wasmx.Coin{Denom: "amyt", Amount: math.NewInt(10)}
	}

	txBytes, err := wasmxcore.PrepareTx(
		string(keysAddr),
		msgBz,
		nil,
		nil,
		5000000,
		*gasPrice,
		privateKey,
	)
	if err != nil {
		return errors.New("failed to prepare forward tx: " + err.Error())
	}

	broadcastResp, err := wasmxcore.BroadcastTxAsync(txBytes)
	if err != nil {
		return errors.New("failed to broadcast forward tx: " + err.Error())
	}
	if broadcastResp.Error != "" {
		return errors.New("forward broadcast error: " + broadcastResp.Error)
	}
	fmt.Println("broadcastAddAddressInternalForward: tx broadcasted", []string{"tx_hash", hex.EncodeToString(broadcastResp.TxHash)})
	return nil
}

// AddAddressInternalForward validates the caller and forwards an internal add_address to wasmx-identity.
func AddAddressInternalForward(msg *MsgAddAddressInternalForward) []byte {
	if msg.UserID == "" || msg.Address == "" || len(msg.PublicKey) == 0 {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "user_id, address, and public_key are required"})
	}

	derivedAddr, err := DeriveAddressFromPublicKey(msg.PublicKey)
	if err != nil {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "failed to derive address"})
	}
	if derivedAddr != msg.Address {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "address does not match public_key"})
	}

	keyPair, err := LoadKeyPair(msg.PublicKey)
	if err != nil || keyPair == nil {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "key pair not found"})
	}
	if keyPair.IdentityUserID == "" {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "user_id missing for key"})
	}
	if keyPair.IdentityUserID != msg.UserID {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "user_id does not match key"})
	}

	caller := string(wasmx.GetCaller())
	if caller != msg.Address {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "caller does not match key address"})
	}

	identityAddr := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	if identityAddr == "" {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "identity contract not found"})
	}

	forwardMsg := map[string]interface{}{
		"add_address_internal": map[string]interface{}{
			"user_id":        msg.UserID,
			"address":        msg.Address,
			"public_key":     msg.PublicKey,
			"service_domain": msg.ServiceDomain,
			"expires_at":     msg.ExpiresAt,
			"permissions":    []interface{}{},
		},
	}
	msgBz, err := json.Marshal(forwardMsg)
	if err != nil {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: "failed to marshal forward msg"})
	}

	ok, resp := wasmx.CallSimple(identityAddr, msgBz, false, MODULE_NAME)
	if !ok {
		return MarshalJSON(MsgAddAddressInternalForwardResponse{Success: false, Error: string(resp)})
	}
	return resp
}

// SignAndBroadcastTx signs and broadcasts a transaction using the ephemeral key
// This allows the oauth2-keys contract to handle all transaction signing,
// ensuring consistent address derivation with the funded ephemeral address
func SignAndBroadcastTx(msg *MsgSignAndBroadcastTx) []byte {
	fmt.Println("SignAndBroadcastTx called", []string{
		"target_contract", msg.TargetContract,
		"gas_limit", fmt.Sprintf("%d", msg.GasLimit),
	})

	// Validate OAuth token and get the ephemeral key
	validateResp := QueryValidateAndGetKey(&MsgQueryValidateAndGetKey{
		OAuthToken: msg.OAuthToken,
	})

	var keyResp QueryValidateAndGetKeyResponse
	if err := json.Unmarshal(validateResp, &keyResp); err != nil {
		LoggerError("SignAndBroadcastTx: failed to parse validation response", []string{"error", err.Error()})
		return MarshalJSON(MsgSignAndBroadcastTxResponse{
			Success: false,
			Error:   "failed to validate OAuth token",
		})
	}

	if !keyResp.Valid {
		LoggerError("SignAndBroadcastTx: invalid OAuth token", []string{"reason", keyResp.Reason})
		return MarshalJSON(MsgSignAndBroadcastTxResponse{
			Success: false,
			Error:   "invalid OAuth token: " + keyResp.Reason,
		})
	}

	fmt.Println("SignAndBroadcastTx: key validated", []string{
		"address", keyResp.Address,
		"public_key", hex.EncodeToString(keyResp.PublicKey),
	})

	// Load gas price from storage (if msg doesn't provide one)
	gasPrice := msg.GasPrice
	if gasPrice.Amount.IsNil() || gasPrice.Amount.IsZero() {
		storedGasPrice := LoadGasPrice()
		if storedGasPrice != nil {
			gasPrice = *storedGasPrice
		}
	}

	// Prepare the transaction using PrepareTx
	// This will derive the from address from the private key on the host side
	txBytes, err := wasmxcore.PrepareTx(
		msg.TargetContract,
		msg.Calldata,
		nil, // no coins to send with the call
		nil, // no events
		msg.GasLimit,
		gasPrice,
		keyResp.PrivateKey,
	)
	if err != nil {
		LoggerError("SignAndBroadcastTx: failed to prepare transaction", []string{"error", err.Error()})
		return MarshalJSON(MsgSignAndBroadcastTxResponse{
			Success: false,
			Error:   "failed to prepare transaction: " + err.Error(),
		})
	}

	fmt.Println("SignAndBroadcastTx: transaction prepared", []string{"tx_bytes_len", fmt.Sprintf("%d", len(txBytes))})

	// Broadcast the transaction
	broadcastResp, err := wasmxcore.BroadcastTxAsync(txBytes)
	fmt.Println("SignAndBroadcastTx: tx broadcasted", err, broadcastResp)
	if err != nil {
		LoggerError("SignAndBroadcastTx: failed to broadcast transaction", []string{"error", err.Error()})
		return MarshalJSON(MsgSignAndBroadcastTxResponse{
			Success: false,
			Error:   "failed to broadcast transaction: " + err.Error(),
		})
	}

	if broadcastResp.Error != "" {
		LoggerError("SignAndBroadcastTx: transaction broadcast error", []string{"error", broadcastResp.Error})
		return MarshalJSON(MsgSignAndBroadcastTxResponse{
			Success: false,
			Error:   "transaction broadcast error: " + broadcastResp.Error,
		})
	}

	// Get transaction hash
	txHash := broadcastResp.TxHash

	fmt.Println("SignAndBroadcastTx: transaction broadcasted", []string{
		"tx_hash", hex.EncodeToString(txHash),
		"address", keyResp.Address,
	})

	return MarshalJSON(MsgSignAndBroadcastTxResponse{
		Success: true,
		TxHash:  txHash,
		Address: keyResp.Address,
	})
}
