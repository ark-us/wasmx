package lib

import (
	"encoding/hex"
	"fmt"
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
	if msg.InitAccountAmt != "" {
		SaveInitAccountAmount(msg.InitAccountAmt)
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
