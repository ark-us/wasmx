package lib

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// matchRedirectURI checks if a redirect URI matches a pattern (supports wildcard *)
func matchRedirectURI(pattern, uri string) bool {
	// Exact match
	if pattern == uri {
		return true
	}

	// Wildcard match - pattern must end with /*
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(uri, prefix)
	}

	return false
}

// CreateAuthorizationCode creates an authorization code for OAuth2 flow
func CreateAuthorizationCode(req CreateAuthorizationCodeRequest) []byte {
	// Validate request
	if req.ClientID == "" {
		Revert("client_id is required")
	}
	if req.UserID == "" {
		Revert("user_id is required")
	}
	if req.RedirectURI == "" {
		Revert("redirect_uri is required")
	}

	// Verify client exists and is active
	client := getOAuthClient(req.ClientID)
	if client == nil {
		Revert("invalid client_id")
	}
	if !client.Active {
		Revert("client is inactive")
	}

	// Verify redirect URI matches one of the registered URIs (supports wildcards)
	validRedirect := false
	for _, pattern := range client.RedirectURIs {
		if matchRedirectURI(pattern, req.RedirectURI) {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		Revert("redirect_uri not registered for this client")
	}

	// Generate authorization code
	code := generateAuthorizationCode()

	// Use actual timestamps
	now := time.Now().Unix()
	createdAt := now
	expiresAt := now + AUTH_CODE_DURATION_SECONDS

	// Create authorization code record
	authCode := AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		UserID:              req.UserID,
		RedirectURI:         req.RedirectURI,
		Scopes:              req.Scopes,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		CreatedAt:           createdAt,
		ExpiresAt:           expiresAt,
		Used:                false,
	}

	// Store authorization code
	storeAuthorizationCode(authCode)
	addToAuthCodeList(code)

	LoggerInfo("Authorization code created", []string{
		"client_id", req.ClientID,
		"user_id", req.UserID,
	})

	// Return response
	response := CreateAuthorizationCodeResponse{
		Code:      code,
		ExpiresAt: expiresAt,
	}

	data, _ := json.Marshal(response)
	return data
}

// ExchangeCodeForToken exchanges an authorization code for access and refresh tokens
func ExchangeCodeForToken(req ExchangeCodeForTokenRequest) []byte {
	// Validate request
	if req.Code == "" {
		Revert("code is required")
	}
	if req.ClientID == "" {
		Revert("client_id is required")
	}
	if req.ClientSecret == "" {
		Revert("client_secret is required")
	}

	// Get authorization code
	authCode := getAuthorizationCode(req.Code)
	if authCode == nil {
		Revert("invalid authorization code")
	}

	// Check if code is expired
	now := time.Now().Unix()
	if now > authCode.ExpiresAt {
		Revert("authorization code expired")
	}

	// Check if code has been used
	if authCode.Used {
		Revert("authorization code already used")
	}

	// Verify client ID matches
	if authCode.ClientID != req.ClientID {
		Revert("client_id mismatch")
	}

	// Verify client secret
	client := getOAuthClient(req.ClientID)
	if client == nil || client.ClientSecret != req.ClientSecret {
		Revert("invalid client credentials")
	}

	// Verify redirect URI matches
	if authCode.RedirectURI != req.RedirectURI {
		Revert("redirect_uri mismatch")
	}

	// Verify PKCE if code_challenge was provided
	if authCode.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			Revert("code_verifier required")
		}

		// Verify code challenge
		if !verifyPKCE(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			Revert("invalid code_verifier")
		}
	}

	// Mark code as used
	authCode.Used = true
	storeAuthorizationCode(*authCode)

	// Generate access token
	accessToken := generateAccessToken()
	refreshToken := generateRefreshToken()

	createdAt := now
	accessExpiresAt := now + ACCESS_TOKEN_DURATION_SECONDS
	refreshExpiresAt := now + (REFRESH_TOKEN_DURATION_DAYS * 86400)

	// Store access token
	accessTokenRecord := AccessToken{
		Token:     accessToken,
		UserID:    authCode.UserID,
		ClientID:  authCode.ClientID,
		Scopes:    authCode.Scopes,
		CreatedAt: createdAt,
		ExpiresAt: accessExpiresAt,
	}
	storeAccessToken(accessTokenRecord)
	addToAccessTokenList(accessToken)

	// Store refresh token
	refreshTokenRecord := RefreshToken{
		Token:     refreshToken,
		UserID:    authCode.UserID,
		ClientID:  authCode.ClientID,
		Scopes:    authCode.Scopes,
		CreatedAt: createdAt,
		ExpiresAt: refreshExpiresAt,
		Revoked:   false,
	}
	storeRefreshToken(refreshTokenRecord)
	addToRefreshTokenList(refreshToken)

	LoggerInfo("Tokens issued", []string{
		"user_id", authCode.UserID,
		"client_id", authCode.ClientID,
	})

	// Return response
	response := ExchangeCodeForTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    ACCESS_TOKEN_DURATION_SECONDS,
		RefreshToken: refreshToken,
	}

	data, _ := json.Marshal(response)
	return data
}

// ValidateAccessToken validates an access token and returns user info
func ValidateAccessToken(req ValidateAccessTokenRequest) []byte {
	if req.Token == "" {
		response := ValidateAccessTokenResponse{Valid: false}
		data, _ := json.Marshal(response)
		return data
	}

	token := getAccessToken(req.Token)
	if token == nil {
		response := ValidateAccessTokenResponse{Valid: false}
		data, _ := json.Marshal(response)
		return data
	}

	// Check if token is expired
	now := time.Now().Unix()
	if now > token.ExpiresAt {
		response := ValidateAccessTokenResponse{Valid: false}
		data, _ := json.Marshal(response)
		return data
	}

	response := ValidateAccessTokenResponse{
		Valid:  true,
		UserID: token.UserID,
		Scopes: token.Scopes,
	}

	data, _ := json.Marshal(response)
	return data
}

// RefreshAccessToken refreshes an access token using a refresh token
func RefreshAccessToken(req RefreshAccessTokenRequest) []byte {
	// Validate request
	if req.RefreshToken == "" {
		Revert("refresh_token is required")
	}
	if req.ClientID == "" {
		Revert("client_id is required")
	}
	if req.ClientSecret == "" {
		Revert("client_secret is required")
	}

	// Get refresh token
	refreshToken := getRefreshToken(req.RefreshToken)
	if refreshToken == nil {
		Revert("invalid refresh_token")
	}

	// Check if token is expired
	now := time.Now().Unix()
	if now > refreshToken.ExpiresAt {
		Revert("refresh_token expired")
	}

	// Check if token is revoked
	if refreshToken.Revoked {
		Revert("refresh_token revoked")
	}

	// Verify client ID matches
	if refreshToken.ClientID != req.ClientID {
		Revert("client_id mismatch")
	}

	// Verify client secret
	client := getOAuthClient(req.ClientID)
	if client == nil || client.ClientSecret != req.ClientSecret {
		Revert("invalid client credentials")
	}

	// Generate new access token
	accessToken := generateAccessToken()
	createdAt := now
	expiresAt := now + ACCESS_TOKEN_DURATION_SECONDS

	// Store new access token
	accessTokenRecord := AccessToken{
		Token:     accessToken,
		UserID:    refreshToken.UserID,
		ClientID:  refreshToken.ClientID,
		Scopes:    refreshToken.Scopes,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}
	storeAccessToken(accessTokenRecord)
	addToAccessTokenList(accessToken)

	LoggerInfo("Access token refreshed", []string{
		"user_id", refreshToken.UserID,
		"client_id", refreshToken.ClientID,
	})

	// Return response (optionally issue new refresh token too)
	response := RefreshAccessTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   ACCESS_TOKEN_DURATION_SECONDS,
	}

	data, _ := json.Marshal(response)
	return data
}

// Helper functions

func generateAuthorizationCode() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	return "code_" + hex.EncodeToString(hash[:])
}

func generateAccessToken() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	return "token_" + hex.EncodeToString(hash[:])
}

func generateRefreshToken() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	return "refresh_" + hex.EncodeToString(hash[:])
}

func verifyPKCE(verifier, challenge, method string) bool {
	if method == "S256" {
		// SHA256 hash of verifier, base64url encoded per RFC 7636
		hash := sha256.Sum256([]byte(verifier))
		computed := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
		return computed == challenge
	} else if method == "plain" {
		return verifier == challenge
	}
	return false
}

// Storage helpers for authorization codes

func storeAuthorizationCode(authCode AuthorizationCode) {
	key := []byte(STORAGE_AUTH_CODE_PREFIX + authCode.Code)
	data, _ := json.Marshal(authCode)
	wasmx.StorageStore(key, data)
}

func getAuthorizationCode(code string) *AuthorizationCode {
	key := []byte(STORAGE_AUTH_CODE_PREFIX + code)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var authCode AuthorizationCode
	if err := json.Unmarshal(data, &authCode); err != nil {
		return nil
	}
	return &authCode
}

func getAuthCodeList() []string {
	data := wasmx.StorageLoad([]byte(STORAGE_AUTH_CODES))
	if len(data) == 0 {
		return []string{}
	}
	var list []string
	json.Unmarshal(data, &list)
	return list
}

func addToAuthCodeList(code string) {
	list := getAuthCodeList()
	for _, c := range list {
		if c == code {
			return
		}
	}
	list = append(list, code)
	data, _ := json.Marshal(list)
	wasmx.StorageStore([]byte(STORAGE_AUTH_CODES), data)
}

// Storage helpers for access tokens

func storeAccessToken(token AccessToken) {
	key := []byte(STORAGE_ACCESS_TOKEN_PREFIX + token.Token)
	data, _ := json.Marshal(token)
	wasmx.StorageStore(key, data)
}

func getAccessToken(tokenStr string) *AccessToken {
	key := []byte(STORAGE_ACCESS_TOKEN_PREFIX + tokenStr)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var token AccessToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil
	}
	return &token
}

func getAccessTokenList() []string {
	data := wasmx.StorageLoad([]byte(STORAGE_ACCESS_TOKENS))
	if len(data) == 0 {
		return []string{}
	}
	var list []string
	json.Unmarshal(data, &list)
	return list
}

func addToAccessTokenList(token string) {
	list := getAccessTokenList()
	for _, t := range list {
		if t == token {
			return
		}
	}
	list = append(list, token)
	data, _ := json.Marshal(list)
	wasmx.StorageStore([]byte(STORAGE_ACCESS_TOKENS), data)
}

// Storage helpers for refresh tokens

func storeRefreshToken(token RefreshToken) {
	key := []byte(STORAGE_REFRESH_TOKEN_PREFIX + token.Token)
	data, _ := json.Marshal(token)
	wasmx.StorageStore(key, data)
}

func getRefreshToken(tokenStr string) *RefreshToken {
	key := []byte(STORAGE_REFRESH_TOKEN_PREFIX + tokenStr)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var token RefreshToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil
	}
	return &token
}

func getRefreshTokenList() []string {
	data := wasmx.StorageLoad([]byte(STORAGE_REFRESH_TOKENS))
	if len(data) == 0 {
		return []string{}
	}
	var list []string
	json.Unmarshal(data, &list)
	return list
}

func addToRefreshTokenList(token string) {
	list := getRefreshTokenList()
	for _, t := range list {
		if t == token {
			return
		}
	}
	list = append(list, token)
	data, _ := json.Marshal(list)
	wasmx.StorageStore([]byte(STORAGE_REFRESH_TOKENS), data)
}
