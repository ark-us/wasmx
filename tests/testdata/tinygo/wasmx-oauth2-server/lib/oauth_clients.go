package lib

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// RegisterOAuthClient registers a new OAuth client and generates credentials
func RegisterOAuthClient(req RegisterOAuthClientRequest) []byte {
	// Validate request
	if req.Name == "" {
		Revert("client name is required")
	}
	if len(req.RedirectURIs) == 0 {
		Revert("at least one redirect URI is required")
	}
	if len(req.Scopes) == 0 {
		Revert("at least one scope is required")
	}

	// Generate client ID and secret
	clientID := generateClientID(req.Name)
	clientSecret := generateClientSecret()

	// Get current block height
	currentBlock := wasmx.GetCurrentBlock()

	// Store client
	client := OAuthClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Name:         req.Name,
		Description:  req.Description,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		WebsiteURL:   req.WebsiteURL,
		LogoURL:      req.LogoURL,
		CreatedAt:    int64(currentBlock.Height),
		Active:       true,
	}

	storeOAuthClient(client)
	addToOAuthClientList(clientID)

	LoggerInfo("OAuth client registered", []string{
		"client_id", clientID,
		"name", req.Name,
	})

	// Return response with credentials
	response := RegisterOAuthClientResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
	}

	data, _ := json.Marshal(response)
	return data
}

// UpdateOAuthClient updates an existing OAuth client
func UpdateOAuthClient(req UpdateOAuthClientRequest) []byte {
	// Get existing client
	client := getOAuthClient(req.ClientID)
	if client == nil {
		Revert("OAuth client not found")
	}

	// Update fields if provided
	if req.Name != "" {
		client.Name = req.Name
	}
	if req.Description != "" {
		client.Description = req.Description
	}
	if len(req.RedirectURIs) > 0 {
		client.RedirectURIs = req.RedirectURIs
	}
	if len(req.Scopes) > 0 {
		client.Scopes = req.Scopes
	}
	if req.WebsiteURL != "" {
		client.WebsiteURL = req.WebsiteURL
	}
	if req.LogoURL != "" {
		client.LogoURL = req.LogoURL
	}

	storeOAuthClient(*client)

	LoggerInfo("OAuth client updated", []string{
		"client_id", req.ClientID,
		"name", client.Name,
	})

	data, _ := json.Marshal(client)
	return data
}

// RevokeOAuthClient revokes an OAuth client
func RevokeOAuthClient(req RevokeOAuthClientRequest) []byte {
	// Get existing client
	client := getOAuthClient(req.ClientID)
	if client == nil {
		Revert("OAuth client not found")
	}

	// Mark as inactive
	client.Active = false
	storeOAuthClient(*client)

	LoggerInfo("OAuth client revoked", []string{
		"client_id", req.ClientID,
	})

	return []byte(`{"success": true}`)
}

// GetOAuthClient gets info about an OAuth client
func GetOAuthClient(req GetOAuthClientRequest) []byte {
	client := getOAuthClient(req.ClientID)
	if client == nil {
		return []byte(`{"error": "client not found"}`)
	}

	// Don't return client secret
	clientResponse := struct {
		ClientID     string   `json:"client_id"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		RedirectURIs []string `json:"redirect_uris"`
		Scopes       []string `json:"scopes"`
		WebsiteURL   string   `json:"website_url,omitempty"`
		LogoURL      string   `json:"logo_url,omitempty"`
		CreatedAt    int64    `json:"created_at"`
		Active       bool     `json:"active"`
	}{
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		RedirectURIs: client.RedirectURIs,
		Scopes:       client.Scopes,
		WebsiteURL:   client.WebsiteURL,
		LogoURL:      client.LogoURL,
		CreatedAt:    client.CreatedAt,
		Active:       client.Active,
	}

	data, _ := json.Marshal(clientResponse)
	return data
}

// ListOAuthClients lists all registered OAuth clients
func ListOAuthClients(req ListOAuthClientsRequest) []byte {
	clientIDs := getOAuthClientList()
	clients := []OAuthClient{}

	for _, clientID := range clientIDs {
		client := getOAuthClient(clientID)
		if client != nil {
			// Don't include client secret in list
			clientCopy := *client
			clientCopy.ClientSecret = ""
			clients = append(clients, clientCopy)
		}
	}

	data, _ := json.Marshal(clients)
	return data
}

// Helper functions

func generateClientID(name string) string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	hash := sha256.Sum256(append([]byte(name), randomBytes...))
	return "client_" + hex.EncodeToString(hash[:8])
}

func generateClientSecret() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	return "secret_" + hex.EncodeToString(hash[:])
}

// Storage helpers for OAuth clients

func storeOAuthClient(client OAuthClient) {
	key := []byte(STORAGE_OAUTH_CLIENT_PREFIX + client.ClientID)
	data, _ := json.Marshal(client)
	wasmx.StorageStore(key, data)
}

func getOAuthClient(clientID string) *OAuthClient {
	key := []byte(STORAGE_OAUTH_CLIENT_PREFIX + clientID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var client OAuthClient
	if err := json.Unmarshal(data, &client); err != nil {
		return nil
	}
	return &client
}

func getOAuthClientList() []string {
	data := wasmx.StorageLoad([]byte(STORAGE_OAUTH_CLIENTS))
	if len(data) == 0 {
		return []string{}
	}
	var list []string
	json.Unmarshal(data, &list)
	return list
}

func addToOAuthClientList(clientID string) {
	list := getOAuthClientList()
	// Check if already exists
	for _, id := range list {
		if id == clientID {
			return
		}
	}
	list = append(list, clientID)
	data, _ := json.Marshal(list)
	wasmx.StorageStore([]byte(STORAGE_OAUTH_CLIENTS), data)
}

func removeFromOAuthClientList(clientID string) {
	list := getOAuthClientList()
	newList := []string{}
	for _, id := range list {
		if id != clientID {
			newList = append(newList, id)
		}
	}
	data, _ := json.Marshal(newList)
	wasmx.StorageStore([]byte(STORAGE_OAUTH_CLIENTS), data)
}
