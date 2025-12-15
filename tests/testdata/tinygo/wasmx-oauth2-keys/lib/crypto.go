package lib

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/hkdf"
)

// GenerateRandomSecret generates a random secret for key encryption
func GenerateRandomSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(secret), nil
}

// GenerateKeyPair generates a new secp256k1 key pair
// For now, we generate random bytes - in production, use proper secp256k1 key generation
func GenerateKeyPair() (publicKey []byte, privateKey []byte, err error) {
	// Generate random private key (32 bytes for secp256k1)
	privKeyBytes := make([]byte, 32)
	if _, err := rand.Read(privKeyBytes); err != nil {
		return nil, nil, err
	}

	// In a real implementation, derive the public key from private key using secp256k1
	// For now, we'll just generate random bytes as a placeholder
	pubKeyBytes := make([]byte, 33) // Compressed public key
	if _, err := rand.Read(pubKeyBytes); err != nil {
		return nil, nil, err
	}
	return pubKeyBytes, privKeyBytes, nil
}

// DeriveEncryptionKey derives an encryption key from server secret and OAuth token
// Uses HKDF (HMAC-based Key Derivation Function)
func DeriveEncryptionKey(serverSecret string, oauthToken string) ([]byte, error) {
	// Decode server secret
	secretBytes, err := base64.StdEncoding.DecodeString(serverSecret)
	if err != nil {
		return nil, err
	}

	// Create HKDF reader
	hash := sha256.New
	info := []byte("ephemeral-key-encryption")
	salt := []byte(oauthToken) // Use OAuth token as salt

	kdf := hkdf.New(hash, secretBytes, salt, info)

	// Derive 32-byte encryption key
	key := make([]byte, 32)
	if _, err := kdf.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

// EncryptPrivateKey encrypts a private key using XOR with derived key
// NOTE: This is a simple XOR encryption for demonstration.
// In production, use AES-GCM or ChaCha20-Poly1305
func EncryptPrivateKey(privKeyBytes []byte, encryptionKey []byte) ([]byte, error) {
	encrypted := make([]byte, len(privKeyBytes))

	// Simple XOR encryption (expand key if needed)
	for i := 0; i < len(privKeyBytes); i++ {
		encrypted[i] = privKeyBytes[i] ^ encryptionKey[i%len(encryptionKey)]
	}

	return encrypted, nil
}

// DecryptPrivateKey decrypts a private key using XOR with derived key
func DecryptPrivateKey(encrypted []byte, encryptionKey []byte) ([]byte, error) {
	decrypted := make([]byte, len(encrypted))

	// XOR decryption (same as encryption for XOR)
	for i := 0; i < len(encrypted); i++ {
		decrypted[i] = encrypted[i] ^ encryptionKey[i%len(encryptionKey)]
	}

	return decrypted, nil
}

// DeriveAddressFromPublicKey derives a bech32 address from a public key
// NOTE: This uses a simplified approach. In production, use proper secp256k1 address derivation
func DeriveAddressFromPublicKey(pubKeyBytes []byte) (string, error) {
	// Hash the public key with SHA256
	hash := sha256.Sum256(pubKeyBytes)

	// Take first 20 bytes as the address
	addrBytes := hash[:20]

	// Convert to hex for address
	// In production, this would be converted to bech32 format with proper prefix
	// For now, we'll create a placeholder bech32-like string
	return "wasmx" + hex.EncodeToString(addrBytes), nil
}
