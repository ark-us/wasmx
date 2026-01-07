package lib

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// RegisterUser creates a new user account
func RegisterUser(req RegisterUserRequest) []byte {
	fmt.Println("--wasmx.oauth2server.RegisterUser--")
	// Validate request
	if req.Email == "" {
		return []byte(`{"error":"email is required"}`)
	}

	// Check if email already exists
	if getUserByEmail(req.Email) != nil {
		return []byte(`{"error":"email already registered"}`)
	}
	fmt.Println("--wasmx.oauth2server.RegisterUser2--")

	// Generate user ID
	userID := generateUserID(req.Email)

	// Hash password
	passwordHash := hashPassword(generateSessionID(userID))

	// Register in identity contract if public key provided
	var identityUserID string
	var blockchainAddress string

	fmt.Println("--wasmx.oauth2server.RegisterUser3--")

	if req.PublicKey != "" && req.Address != "" {
		// Query identity contract to get user_id by address
		// (user should have already registered via /register_tx)
		fmt.Println("--wasmx.oauth2server.RegisterUser: Querying identity contract by address--")
		identityUserID, blockchainAddress = queryIdentityUserByAddress(req.Address)
		if identityUserID == "" {
			fmt.Println("WARNING: No identity user found for address:", req.Address)
			LoggerError("Identity user not found for address", []string{"address", req.Address})
		} else {
			LoggerInfo("Found identity user for address", []string{
				"identity_user_id", identityUserID,
				"address", blockchainAddress,
			})
		}
	}
	fmt.Println("--wasmx.oauth2server.RegisterUser4--")

	// Create user account
	user := User{
		UserID:         userID,
		Email:          req.Email,
		Username:       req.Username,
		PasswordHash:   passwordHash,
		CreatedAt:      time.Now().Unix(),
		Active:         true,
		IdentityUserID: identityUserID,
		Address:        blockchainAddress,
	}

	// Store user
	storeUser(user)
	fmt.Println("--wasmx.oauth2server.RegisterUser5--")
	storeEmailMapping(req.Email, userID)
	fmt.Println("--wasmx.oauth2server.RegisterUser6--")
	addToUserList(userID)
	fmt.Println("--wasmx.oauth2server.RegisterUser7--")

	LoggerInfo("User registered", []string{
		"user_id", userID,
		"email", req.Email,
		"identity_user_id", identityUserID,
	})

	// Return response (without password hash)
	response := RegisterUserResponse{
		UserID:            userID,
		Email:             req.Email,
		Username:          req.Username,
		IdentityUserID:    identityUserID,
		BlockchainAddress: blockchainAddress,
	}

	data, _ := json.Marshal(response)
	return data
}

// Login authenticates a user and creates a session
func Login(req LoginRequest) []byte {
	// Validate request
	if req.Email == "" {
		return []byte(`{"error":"email is required"}`)
	}
	if req.Password == "" {
		return []byte(`{"error":"password is required"}`)
	}

	// Get user by email
	user := getUserByEmail(req.Email)
	if user == nil {
		return []byte(`{"error":"invalid email or password"}`)
	}

	// Check if user is active
	if !user.Active {
		return []byte(`{"error":"user account is disabled"}`)
	}

	// Verify password
	passwordHash := hashPassword(req.Password)
	if passwordHash != user.PasswordHash {
		return []byte(`{"error":"invalid email or password"}`)
	}

	response, err := createSessionForUser(user)
	if err != nil {
		return []byte(`{"error":"failed to create session"}`)
	}

	data, _ := json.Marshal(response)
	return data
}

func createSessionForUser(user *User) (LoginResponse, error) {
	if user == nil {
		return LoginResponse{}, fmt.Errorf("user is required")
	}

	// Generate session ID
	sessionID := generateSessionID(user.UserID)

	// Get current block height for timestamps
	currentBlock := wasmx.GetCurrentBlock()
	createdAt := int64(currentBlock.Height)
	expiresAt := createdAt + SESSION_DURATION_SECONDS

	// Create session
	session := Session{
		SessionID: sessionID,
		UserID:    user.UserID,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}

	// Store session
	storeSession(session)
	addToSessionList(sessionID)

	LoggerInfo("User logged in", []string{
		"user_id", user.UserID,
		"email", user.Email,
		"session_id", sessionID,
	})

	// SERVER-SIDE LOG: Print OAuth session token
	fmt.Println("=== SERVER-SIDE LOGIN ===")
	fmt.Println("Session ID (OAuth token):", sessionID)
	fmt.Println("User ID:", user.UserID)
	fmt.Println("Identity User ID:", user.IdentityUserID)
	fmt.Println("Blockchain Address:", user.Address)
	fmt.Println("========================")

	// Note: Ephemeral key generation moved to ExchangeCodeForToken
	// where the actual OAuth access token is created

	// Return response
	response := LoginResponse{
		SessionID:         sessionID,
		UserID:            user.UserID,
		Email:             user.Email,
		ExpiresAt:         expiresAt,
		IdentityUserID:    user.IdentityUserID,
		BlockchainAddress: user.Address,
	}

	return response, nil
}

// Logout invalidates a session
func Logout(req LogoutRequest) []byte {
	// Validate request
	if req.SessionID == "" {
		return []byte(`{"error":"session_id is required"}`)
	}

	// Get session
	session := getSession(req.SessionID)
	if session != nil {
		// Delete session
		deleteSession(req.SessionID)
		removeFromSessionList(req.SessionID)

		LoggerInfo("User logged out", []string{
			"user_id", session.UserID,
		})
	}

	return []byte(`{"success": true}`)
}

// GetCurrentUser returns the current user based on session
func GetCurrentUser(req GetCurrentUserRequest) []byte {
	// Validate request
	if req.SessionID == "" {
		return []byte(`{"error": "session_id is required"}`)
	}

	// Get and validate session
	session := getSession(req.SessionID)
	if session == nil {
		return []byte(`{"error": "invalid session"}`)
	}

	// Check if session is expired
	currentBlock := wasmx.GetCurrentBlock()
	if int64(currentBlock.Height) > session.ExpiresAt {
		// Clean up expired session
		deleteSession(req.SessionID)
		removeFromSessionList(req.SessionID)
		return []byte(`{"error": "session expired"}`)
	}

	// Get user
	user := getUser(session.UserID)
	if user == nil {
		return []byte(`{"error": "user not found"}`)
	}

	// Don't return password hash
	userResponse := struct {
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username,omitempty"`
	}{
		UserID:   user.UserID,
		Email:    user.Email,
		Username: user.Username,
	}

	data, _ := json.Marshal(userResponse)
	return data
}

// ValidateSession checks if a session is valid and returns user info
func ValidateSession(req ValidateSessionRequest) []byte {
	userID, err := validateSessionInternal(req.SessionID)
	if err != nil {
		response := ValidateSessionResponse{
			Valid: false,
		}
		data, _ := json.Marshal(response)
		return data
	}

	response := ValidateSessionResponse{
		Valid:  true,
		UserID: userID,
	}
	data, _ := json.Marshal(response)
	return data
}

// Helper functions

func generateUserID(email string) string {
	// Generate random bytes for user ID
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	hash := sha256.Sum256(append([]byte(email), randomBytes...))
	return "user_" + hex.EncodeToString(hash[:8])
}

func generateSessionID(userID string) string {
	// Generate random session ID
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	hash := sha256.Sum256(append([]byte(userID), randomBytes...))
	return "sess_" + hex.EncodeToString(hash[:])
}

func hashPassword(password string) string {
	// Simple SHA256 hash for now
	// In production, use bcrypt or argon2
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// Storage helpers for users

func storeUser(user User) {
	key := []byte(STORAGE_USER_PREFIX + user.UserID)
	data, _ := json.Marshal(user)
	wasmx.StorageStore(key, data)
}

func getUser(userID string) *User {
	key := []byte(STORAGE_USER_PREFIX + userID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil
	}
	return &user
}

func storeEmailMapping(email, userID string) {
	key := []byte(STORAGE_USER_EMAIL_PREFIX + email)
	wasmx.StorageStore(key, []byte(userID))
}

func getUserByEmail(email string) *User {
	key := []byte(STORAGE_USER_EMAIL_PREFIX + email)
	userIDBytes := wasmx.StorageLoad(key)
	if len(userIDBytes) == 0 {
		return nil
	}
	userID := string(userIDBytes)
	return getUser(userID)
}

func getUserList() []string {
	data := wasmx.StorageLoad([]byte(STORAGE_USERS))
	if len(data) == 0 {
		return []string{}
	}
	var list []string
	json.Unmarshal(data, &list)
	return list
}

func addToUserList(userID string) {
	list := getUserList()
	// Check if already exists
	for _, id := range list {
		if id == userID {
			return
		}
	}
	list = append(list, userID)
	data, _ := json.Marshal(list)
	wasmx.StorageStore([]byte(STORAGE_USERS), data)
}

// Storage helpers for sessions

func storeSession(session Session) {
	key := []byte(STORAGE_SESSION_PREFIX + session.SessionID)
	data, _ := json.Marshal(session)
	wasmx.StorageStore(key, data)
}

func getSession(sessionID string) *Session {
	key := []byte(STORAGE_SESSION_PREFIX + sessionID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil
	}
	return &session
}

func deleteSession(sessionID string) {
	key := STORAGE_SESSION_PREFIX + sessionID
	wasmx.StorageDelete(key)
}

func getSessionList() []string {
	data := wasmx.StorageLoad([]byte(STORAGE_SESSIONS))
	if len(data) == 0 {
		return []string{}
	}
	var list []string
	json.Unmarshal(data, &list)
	return list
}

func addToSessionList(sessionID string) {
	list := getSessionList()
	// Check if already exists
	for _, id := range list {
		if id == sessionID {
			return
		}
	}
	list = append(list, sessionID)
	data, _ := json.Marshal(list)
	wasmx.StorageStore([]byte(STORAGE_SESSIONS), data)
}

func removeFromSessionList(sessionID string) {
	list := getSessionList()
	newList := []string{}
	for _, id := range list {
		if id != sessionID {
			newList = append(newList, id)
		}
	}
	data, _ := json.Marshal(newList)
	wasmx.StorageStore([]byte(STORAGE_SESSIONS), data)
}

// validateSessionInternal checks if a session is valid and returns the user_id
func validateSessionInternal(sessionID string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session_id is required")
	}

	session := getSession(sessionID)
	if session == nil {
		return "", fmt.Errorf("invalid session")
	}

	// Check if session is expired
	currentBlock := wasmx.GetCurrentBlock()
	if int64(currentBlock.Height) > session.ExpiresAt {
		// Clean up expired session
		deleteSession(sessionID)
		removeFromSessionList(sessionID)
		return "", fmt.Errorf("session expired")
	}

	return session.UserID, nil
}

// queryIdentityUserByAddress queries the identity contract to get user_id by address
func queryIdentityUserByAddress(address string) (string, string) {
	fmt.Println("--wasmx.oauth2server.queryIdentityUserByAddresssssssss--", address)
	fmt.Println("--wasmx.oauth2server.queryIdentityUserByAddress.ROLE_ACCOUNT_IDENTITY--", wasmx.ROLE_ACCOUNT_IDENTITY)

	// Get identity contract address
	identityAddr := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	fmt.Println("--wasmx.oauth2server.queryIdentityUserByAddress.identityAddr--", identityAddr)
	if identityAddr == "" {
		LoggerError("Identity contract not found", nil)
		return "", ""
	}

	// Create query message
	queryMsg := map[string]interface{}{
		"query_user_by_address": map[string]interface{}{
			"address": address,
		},
	}

	queryMsgBytes, err := json.Marshal(queryMsg)
	if err != nil {
		LoggerError("Failed to marshal identity query", []string{"error", err.Error()})
		return "", ""
	}

	fmt.Println("--wasmx.oauth2server.queryIdentityUserByAddress.CallSimple--", identityAddr, string(queryMsgBytes))

	// Query identity contract
	ok, data := wasmx.CallSimple(identityAddr, queryMsgBytes, true, MODULE_NAME)
	fmt.Println("--wasmx.oauth2server.queryIdentityUserByAddress.CallSimple--", ok, string(data))
	if !ok {
		LoggerError("Failed to query identity contract", []string{"error", string(data)})
		fmt.Println("ERROR: Identity contract query failed:", string(data))
		return "", ""
	}

	// Parse response
	var response struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		LoggerError("Failed to parse identity query response", []string{"error", err.Error()})
		return "", ""
	}

	fmt.Println("Identity user_id for address:", response.UserID)
	return response.UserID, address
}

// generateEphemeralKey generates a server-side ephemeral key for the user session
func generateEphemeralKey(oauthToken, userID, identityUserID string, expiresAt int64) {
	// Get oauth2-keys contract address
	keysAddr := wasmx.GetAddressByRole(wasmx.ROLE_OAUTH2_KEYS)
	if keysAddr == "" {
		LoggerError("OAuth2 keys contract not found", nil)
		fmt.Println("ERROR: OAuth2 keys contract not found by role")
		return
	}

	// Create message to generate ephemeral key
	generateMsg := map[string]interface{}{
		"generate_ephemeral_key": map[string]interface{}{
			"user_id":        userID,
			"oauth_token":    oauthToken,
			"service_domain": "", // Empty for general use
			"expires_at":     expiresAt,
		},
	}

	msgBz, err := json.Marshal(generateMsg)
	if err != nil {
		LoggerError("Failed to marshal ephemeral key message", []string{"error", err.Error()})
		fmt.Println("ERROR: Failed to marshal ephemeral key message:", err.Error())
		return
	}

	// Call oauth2-keys contract to generate ephemeral key
	ok, data := wasmx.CallSimple(keysAddr, msgBz, false, MODULE_NAME)
	if !ok {
		LoggerError("Failed to generate ephemeral key", []string{"error", string(data)})
		fmt.Println("ERROR: Failed to generate ephemeral key:", string(data))
		return
	}

	// Log raw response
	fmt.Println("=== OAUTH2-KEYS CONTRACT RESPONSE ===")
	fmt.Println("Raw data:", string(data))
	fmt.Println("Data length:", len(data))
	fmt.Println("=====================================")

	// Parse response
	var response struct {
		Success    bool   `json:"success"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		LoggerError("Failed to parse ephemeral key response", []string{"error", err.Error(), "data", string(data)})
		fmt.Println("ERROR: Failed to parse ephemeral key response:", err.Error())
		fmt.Println("Raw data was:", string(data))
		return
	}

	if !response.Success {
		LoggerError("Ephemeral key generation failed", nil)
		fmt.Println("ERROR: Ephemeral key generation failed")
		return
	}

	LoggerInfo("Ephemeral key generated", []string{
		"user_id", userID,
		"public_key", response.PublicKey,
	})

	// SERVER-SIDE LOG: Print ephemeral key info
	fmt.Println("=== EPHEMERAL KEY GENERATED ===")
	fmt.Println("OAuth Token:", oauthToken)
	fmt.Println("User ID:", userID)
	fmt.Println("Identity User ID:", identityUserID)
	fmt.Println("Public Key (hex):", response.PublicKey)
	fmt.Println("Private Key (hex, first 16 chars):", response.PrivateKey[:16]+"...")
	fmt.Println("Expires At:", expiresAt)
	fmt.Println("===============================")

	// Note: Account initialization for the ephemeral address is now done
	// inside GenerateEphemeralKey in oauth2-keys contract.
	// We don't need to initialize user.Address separately since:
	// 1. The ephemeral address is what needs funding for OAuth2 transactions
	// 2. The user's original address should already be funded if they registered with their own wallet
}

// revokeEphemeralKey revokes an ephemeral key associated with an OAuth token
func revokeEphemeralKey(oauthToken string) {
	// Get oauth2-keys contract address
	keysAddr := wasmx.GetAddressByRole(wasmx.ROLE_OAUTH2_KEYS)
	if keysAddr == "" {
		LoggerError("OAuth2 keys contract not found", nil)
		fmt.Println("ERROR: OAuth2 keys contract not found by role")
		return
	}

	// Create message to revoke ephemeral key
	revokeMsg := map[string]interface{}{
		"revoke_key": map[string]interface{}{
			"oauth_token": oauthToken,
		},
	}

	msgBz, err := json.Marshal(revokeMsg)
	if err != nil {
		LoggerError("Failed to marshal revoke key message", []string{"error", err.Error()})
		fmt.Println("ERROR: Failed to marshal revoke key message:", err.Error())
		return
	}

	// Call oauth2-keys contract to revoke ephemeral key
	ok, data := wasmx.CallSimple(keysAddr, msgBz, false, MODULE_NAME)
	if !ok {
		LoggerError("Failed to revoke ephemeral key", []string{"error", string(data)})
		fmt.Println("ERROR: Failed to revoke ephemeral key:", string(data))
		return
	}

	LoggerInfo("Ephemeral key revoked", []string{"oauth_token", oauthToken})
	fmt.Println("=== EPHEMERAL KEY REVOKED ===")
	fmt.Println("OAuth Token:", oauthToken)
	fmt.Println("=============================")
}
