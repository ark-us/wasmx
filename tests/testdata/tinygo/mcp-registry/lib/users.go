package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const (
	SESSION_DURATION_SECONDS = 86400 // 24 hours
)

// RegisterUser creates a new user account
func RegisterUser(req RegisterUserRequest) []byte {
	// Validate request
	if req.Email == "" {
		Revert("email is required")
	}
	if req.Password == "" {
		Revert("password is required")
	}
	if len(req.Password) < 8 {
		Revert("password must be at least 8 characters")
	}

	// Check if email already exists
	if getUserByEmail(req.Email) != nil {
		Revert("email already registered")
	}

	// Generate user ID
	userID := generateUserID(req.Email)

	// Hash password
	passwordHash := hashPassword(req.Password)

	// Get current block height
	currentBlock := wasmx.GetCurrentBlock()

	// Create user account
	user := UserAccount{
		UserID:       userID,
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		CreatedAt:    int64(currentBlock.Height),
		Active:       true,
	}

	// Store user
	storeUser(user)
	storeEmailMapping(req.Email, userID)
	addToUserList(userID)

	LoggerInfo("User registered", []string{
		"user_id", userID,
		"email", req.Email,
	})

	// Return response (without password hash)
	response := RegisterUserResponse{
		UserID:   userID,
		Email:    req.Email,
		Username: req.Username,
	}

	data, _ := json.Marshal(response)
	return data
}

// Login authenticates a user and creates a session
func Login(req LoginRequest) []byte {
	// Validate request
	if req.Email == "" {
		Revert("email is required")
	}
	if req.Password == "" {
		Revert("password is required")
	}

	// Get user by email
	user := getUserByEmail(req.Email)
	if user == nil {
		Revert("invalid email or password")
	}

	// Check if user is active
	if !user.Active {
		Revert("user account is disabled")
	}

	// Verify password
	passwordHash := hashPassword(req.Password)
	if passwordHash != user.PasswordHash {
		Revert("invalid email or password")
	}

	// Generate session ID
	sessionID := generateSessionID(user.UserID)

	// Get current block height for timestamps
	currentBlock := wasmx.GetCurrentBlock()
	createdAt := int64(currentBlock.Height)
	expiresAt := createdAt + SESSION_DURATION_SECONDS

	// Create session
	session := SessionInfo{
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
	})

	// Return response
	response := LoginResponse{
		SessionID: sessionID,
		UserID:    user.UserID,
		Email:     user.Email,
		ExpiresAt: expiresAt,
	}

	data, _ := json.Marshal(response)
	return data
}

// Logout invalidates a session
func Logout(req LogoutRequest) []byte {
	// Validate request
	if req.SessionID == "" {
		Revert("session_id is required")
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

// Helper functions

func generateUserID(email string) string {
	// Generate random bytes for user ID
	randomBytes := wasmxcore.Random(16)
	hash := sha256.Sum256(append([]byte(email), randomBytes...))
	return "user_" + hex.EncodeToString(hash[:8])
}

func generateSessionID(userID string) string {
	// Generate random session ID
	randomBytes := wasmxcore.Random(32)
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

func storeUser(user UserAccount) {
	key := []byte(STORAGE_USER_PREFIX + user.UserID)
	data, _ := json.Marshal(user)
	wasmx.StorageStore(key, data)
}

func getUser(userID string) *UserAccount {
	key := []byte(STORAGE_USER_PREFIX + userID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var user UserAccount
	if err := json.Unmarshal(data, &user); err != nil {
		return nil
	}
	return &user
}

func storeEmailMapping(email, userID string) {
	key := []byte(STORAGE_USER_EMAIL_PREFIX + email)
	wasmx.StorageStore(key, []byte(userID))
}

func getUserByEmail(email string) *UserAccount {
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

func storeSession(session SessionInfo) {
	key := []byte(STORAGE_SESSION_PREFIX + session.SessionID)
	data, _ := json.Marshal(session)
	wasmx.StorageStore(key, data)
}

func getSession(sessionID string) *SessionInfo {
	key := []byte(STORAGE_SESSION_PREFIX + sessionID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var session SessionInfo
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

// validateSession checks if a session is valid and returns the user_id
func validateSession(sessionID string) (string, error) {
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
