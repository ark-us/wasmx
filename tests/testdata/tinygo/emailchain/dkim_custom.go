package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"regexp"
	"strings"
)

// CustomDKIMVerifier implements DKIM verification without net package dependencies
type CustomDKIMVerifier struct {
	dnsResolver *DNSResolver
}

// NewCustomDKIMVerifier creates a new custom DKIM verifier
func NewCustomDKIMVerifier() *CustomDKIMVerifier {
	return &CustomDKIMVerifier{
		dnsResolver: NewDNSResolver(),
	}
}

// DKIMVerification represents the result of DKIM verification
type DKIMVerification struct {
	Domain     string `json:"domain"`
	Selector   string `json:"selector"`
	Valid      bool   `json:"valid"`
	Error      string `json:"error,omitempty"`
	Algorithm  string `json:"algorithm"`
	HeaderKeys []string `json:"header_keys"`
}

// VerifyDKIMSignatures verifies all DKIM signatures in an email
func (v *CustomDKIMVerifier) VerifyDKIMSignatures(emailRaw string) ([]DKIMVerification, error) {
	lines := strings.Split(emailRaw, "\n")
	var verifications []DKIMVerification

	// Find DKIM-Signature headers
	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(line), "dkim-signature:") {
			verification := v.verifyDKIMSignature(line, emailRaw)
			verifications = append(verifications, verification)
		}
	}

	if len(verifications) == 0 {
		return nil, errors.New("no DKIM signatures found")
	}

	return verifications, nil
}

// verifyDKIMSignature verifies a single DKIM signature
func (v *CustomDKIMVerifier) verifyDKIMSignature(dkimHeader, emailRaw string) DKIMVerification {
	verification := DKIMVerification{}

	// Parse DKIM header parameters
	params, err := v.parseDKIMHeader(dkimHeader)
	if err != nil {
		verification.Error = fmt.Sprintf("failed to parse DKIM header: %v", err)
		return verification
	}

	verification.Domain = params["d"]
	verification.Selector = params["s"]
	verification.Algorithm = params["a"]
	
	if params["h"] != "" {
		verification.HeaderKeys = strings.Split(params["h"], ":")
	}

	// Get public key from DNS
	publicKey, err := v.getPublicKey(verification.Domain, verification.Selector)
	if err != nil {
		verification.Error = fmt.Sprintf("failed to get public key: %v", err)
		return verification
	}

	// Verify signature
	err = v.verifySignature(emailRaw, params, publicKey)
	if err != nil {
		verification.Error = fmt.Sprintf("signature verification failed: %v", err)
		return verification
	}

	verification.Valid = true
	return verification
}

// parseDKIMHeader parses DKIM-Signature header parameters
func (v *CustomDKIMVerifier) parseDKIMHeader(header string) (map[string]string, error) {
	params := make(map[string]string)
	
	// Remove "DKIM-Signature:" prefix and normalize whitespace
	content := strings.TrimPrefix(header, "DKIM-Signature:")
	content = strings.TrimPrefix(content, "dkim-signature:")
	content = strings.TrimSpace(content)
	
	// Split by semicolon and parse key=value pairs
	re := regexp.MustCompile(`([a-z]+)\s*=\s*([^;]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) == 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			params[key] = value
		}
	}

	// Validate required parameters
	required := []string{"v", "a", "d", "s", "h", "b"}
	for _, req := range required {
		if params[req] == "" {
			return nil, fmt.Errorf("missing required parameter: %s", req)
		}
	}

	return params, nil
}

// getPublicKey retrieves the DKIM public key from DNS
func (v *CustomDKIMVerifier) getPublicKey(domain, selector string) (*rsa.PublicKey, error) {
	// Construct DNS query for DKIM public key
	dnsName := fmt.Sprintf("%s._domainkey.%s", selector, domain)
	
	txtRecords, err := v.dnsResolver.LookupTXT(dnsName)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed: %v", err)
	}

	if len(txtRecords) == 0 {
		return nil, errors.New("no TXT records found")
	}

	// Parse DKIM public key from TXT record
	// Concatenate all TXT records (they might be split)
	fullRecord := strings.Join(txtRecords, "")
	
	// Extract public key
	return v.parsePublicKey(fullRecord)
}

// parsePublicKey parses DKIM public key from TXT record
func (v *CustomDKIMVerifier) parsePublicKey(txtRecord string) (*rsa.PublicKey, error) {
	// Parse key-value pairs from TXT record
	params := make(map[string]string)
	pairs := strings.Split(txtRecord, ";")
	
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			params[key] = value
		}
	}

	// Extract public key data
	pubKeyData := params["p"]
	if pubKeyData == "" {
		return nil, errors.New("no public key data found in TXT record")
	}

	// Decode base64 public key
	keyBytes, err := base64.StdEncoding.DecodeString(pubKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}

	// Parse RSA public key
	pubKey, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		// Try parsing as PEM format
		block, _ := pem.Decode(keyBytes)
		if block != nil {
			pubKey, err = x509.ParsePKIXPublicKey(block.Bytes)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %v", err)
		}
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPubKey, nil
}

// verifySignature verifies the DKIM signature
func (v *CustomDKIMVerifier) verifySignature(emailRaw string, params map[string]string, publicKey *rsa.PublicKey) error {
	// Get hash algorithm
	var hasher hash.Hash
	var hashType crypto.Hash
	
	switch params["a"] {
	case "rsa-sha1":
		hasher = sha1.New()
		hashType = crypto.SHA1
	case "rsa-sha256":
		hasher = sha256.New()
		hashType = crypto.SHA256
	default:
		return fmt.Errorf("unsupported algorithm: %s", params["a"])
	}

	// Build canonical headers and body for verification
	// This is a simplified implementation - full DKIM canonicalization is complex
	canonicalData := v.buildCanonicalData(emailRaw, params)
	
	// Hash the canonical data
	hasher.Write([]byte(canonicalData))
	hashed := hasher.Sum(nil)

	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(params["b"])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %v", err)
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, hashType, hashed, signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}

	return nil
}

// buildCanonicalData builds the canonical representation for signature verification
// This is a simplified implementation - full DKIM canonicalization is more complex
func (v *CustomDKIMVerifier) buildCanonicalData(emailRaw string, params map[string]string) string {
	lines := strings.Split(emailRaw, "\n")
	
	// Simple canonicalization - just concatenate specified headers
	// In practice, DKIM has complex canonicalization rules
	var canonicalData strings.Builder
	
	headerNames := strings.Split(params["h"], ":")
	for _, headerName := range headerNames {
		headerName = strings.TrimSpace(strings.ToLower(headerName))
		
		// Find matching header in email
		for _, line := range lines {
			if strings.HasPrefix(strings.ToLower(line), headerName+":") {
				canonicalData.WriteString(line)
				canonicalData.WriteString("\n")
				break
			}
		}
	}

	return canonicalData.String()
}