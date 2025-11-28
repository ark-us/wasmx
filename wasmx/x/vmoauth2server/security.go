package vmoauth2server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// Bcrypt cost for hashing client secrets
	BcryptCost = 12

	// Token lengths
	AuthorizationCodeLength = 32
	AccessTokenLength       = 48
	RefreshTokenLength      = 64
)

// HashSecret hashes a client secret using bcrypt
func HashSecret(secret string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}
	return string(hashed), nil
}

// ValidateSecret validates a client secret against its hash
func ValidateSecret(secret, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret))
	return err == nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(bytes)
	if len(token) > length {
		return token[:length], nil
	}
	return token, nil
}

// ValidatePKCE validates the code verifier against the code challenge
// Supports both S256 and plain methods
func ValidatePKCE(codeVerifier, codeChallenge, method string) bool {
	if method == "" {
		// No PKCE used
		return codeChallenge == ""
	}

	if codeChallenge == "" {
		// PKCE required but challenge not provided
		return false
	}

	switch method {
	case "plain":
		return codeVerifier == codeChallenge
	case "S256":
		// code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
		hash := sha256.Sum256([]byte(codeVerifier))
		encoded := base64.URLEncoding.EncodeToString(hash[:])
		// Remove padding
		for len(encoded) > 0 && encoded[len(encoded)-1] == '=' {
			encoded = encoded[:len(encoded)-1]
		}
		return encoded == codeChallenge
	default:
		return false
	}
}
