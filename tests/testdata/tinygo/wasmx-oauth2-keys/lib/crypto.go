package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
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

// GenerateKeyPair generates a new secp256k1 key pair using the host function
func GenerateKeyPair() (publicKey []byte, privateKey []byte, err error) {
	// Use the host function to generate secp256k1 key pair
	publicKey, privateKey, err = wasmxcore.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
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

// EncryptPrivateKey encrypts a private key using AES-256-GCM
func EncryptPrivateKey(privKeyBytes []byte, encryptionKey []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, privKeyBytes, nil)

	return ciphertext, nil
}

// DecryptPrivateKey decrypts a private key using AES-256-GCM
func DecryptPrivateKey(encrypted []byte, encryptionKey []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// DeriveAddressFromPublicKey derives a bech32 address from a secp256k1 public key
func DeriveAddressFromPublicKey(pubKeyBytes []byte) (string, error) {
	// Verify public key is valid secp256k1 format (33 bytes compressed)
	if len(pubKeyBytes) != 33 {
		return "", errors.New("invalid public key length, expected 33 bytes")
	}

	// Hash the public key with SHA256
	hash := sha256.Sum256(pubKeyBytes)

	// Take first 20 bytes as the address bytes
	addrBytes := hash[:20]

	// Use wasmx-env SDK to convert to bech32 with "wasmx" prefix
	bech32Addr := wasmx.AddrHumanize(addrBytes)

	return bech32Addr, nil
}
