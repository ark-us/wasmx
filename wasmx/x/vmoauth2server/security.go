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
// The length parameter specifies the desired length of the final base64-encoded token
func GenerateSecureToken(length int) (string, error) {
	// Calculate bytes needed to produce desired base64 length without truncation
	// Base64 encoding: 3 bytes -> 4 characters
	// For RawURLEncoding (no padding), we need ceil(length * 3 / 4) bytes
	bytesNeeded := (length*3 + 3) / 4 // Round up

	bytes := make([]byte, bytesNeeded)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Use RawURLEncoding (no padding) for consistent, URL-safe output
	token := base64.RawURLEncoding.EncodeToString(bytes)

	// Ensure we return exactly the requested length (trim excess if any)
	if len(token) > length {
		return token[:length], nil
	}
	return token, nil
}

// ValidatePKCE validates the code verifier against the code challenge per RFC 7636
// Supports both S256 (recommended) and plain methods
func ValidatePKCE(codeVerifier, codeChallenge, method string) bool {
	// If a code_challenge was provided, method and verifier are REQUIRED
	if codeChallenge != "" {
		if method == "" || codeVerifier == "" {
			// Challenge provided but missing method or verifier - REJECT
			return false
		}
	} else {
		// No challenge provided - only valid if method is also empty
		// This allows non-PKCE flows (though PKCE is recommended)
		return method == ""
	}

	// Validate code_verifier format per RFC 7636 section 4.1
	// Must be 43-128 characters from [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return false
	}

	switch method {
	case "plain":
		// Plain method: code_challenge equals code_verifier
		return codeVerifier == codeChallenge

	case "S256":
		// S256 method: code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
		// Per RFC 7636, use base64url encoding WITHOUT padding
		hash := sha256.Sum256([]byte(codeVerifier))
		encoded := base64.RawURLEncoding.EncodeToString(hash[:])
		return encoded == codeChallenge

	default:
		// Unknown method - reject
		return false
	}
}
