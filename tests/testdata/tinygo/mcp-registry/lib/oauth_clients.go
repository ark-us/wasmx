package lib

import (
	"encoding/json"

	oauth2server "github.com/loredanacirstea/wasmx-env-oauth2server/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// RegisterOAuthClient registers a new OAuth client and generates credentials
func RegisterOAuthClient(req RegisterOAuthClientRequest) []byte {
	// Only governance can register OAuth clients
	caller := wasmx.GetCaller()
	if !hasGovRole(caller) {
		Revert("unauthorized: only governance can register OAuth clients")
	}

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

	// Register client with OAuth2 server - it will generate client_id and client_secret
	oauth2Resp := oauth2server.RegisterClient(&oauth2server.RegisterClientRequest{
		InstanceID:   OAUTH2_INSTANCE_ID,
		ClientID:     "", // Let OAuth2 server generate it
		ClientSecret: "", // Let OAuth2 server generate it
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		GrantTypes:   []string{"authorization_code", "refresh_token"},
	})

	if oauth2Resp.Error != "" {
		Revert("failed to register OAuth client: " + oauth2Resp.Error)
	}

	// Get current block height
	currentBlock := wasmx.GetCurrentBlock()

	// Store client metadata
	clientInfo := OAuthClientInfo{
		ClientID:     oauth2Resp.ClientID,
		Name:         req.Name,
		Description:  req.Description,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		WebsiteURL:   req.WebsiteURL,
		LogoURL:      req.LogoURL,
		CreatedAt:    int64(currentBlock.Height),
		Active:       true,
	}

	storeOAuthClient(clientInfo)
	addToOAuthClientList(oauth2Resp.ClientID)

	LoggerInfo("OAuth client registered", []string{
		"client_id", oauth2Resp.ClientID,
		"name", req.Name,
	})

	// Return response with credentials
	response := RegisterOAuthClientResponse{
		ClientID:     oauth2Resp.ClientID,
		ClientSecret: oauth2Resp.ClientSecret,
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
	}

	data, _ := json.Marshal(response)
	return data
}

// UpdateOAuthClient updates an existing OAuth client
func UpdateOAuthClient(req UpdateOAuthClientRequest) []byte {
	// Only governance can update OAuth clients
	caller := wasmx.GetCaller()
	if !hasGovRole(caller) {
		Revert("unauthorized: only governance can update OAuth clients")
	}

	// Get existing client
	clientInfo := getOAuthClient(req.ClientID)
	if clientInfo == nil {
		Revert("OAuth client not found")
	}

	// Update fields if provided
	if req.Name != "" {
		clientInfo.Name = req.Name
	}
	if req.Description != "" {
		clientInfo.Description = req.Description
	}
	if len(req.RedirectURIs) > 0 {
		clientInfo.RedirectURIs = req.RedirectURIs
	}
	if len(req.Scopes) > 0 {
		clientInfo.Scopes = req.Scopes
	}
	if req.WebsiteURL != "" {
		clientInfo.WebsiteURL = req.WebsiteURL
	}
	if req.LogoURL != "" {
		clientInfo.LogoURL = req.LogoURL
	}

	// Update in OAuth2 server if redirect URIs or scopes changed
	if len(req.RedirectURIs) > 0 || len(req.Scopes) > 0 {
		// Note: OAuth2 server might not support updating, in which case we just update metadata
		LoggerInfo("OAuth client metadata updated", []string{
			"client_id", req.ClientID,
		})
	}

	storeOAuthClient(*clientInfo)

	LoggerInfo("OAuth client updated", []string{
		"client_id", req.ClientID,
		"name", clientInfo.Name,
	})

	data, _ := json.Marshal(clientInfo)
	return data
}

// RevokeOAuthClient revokes an OAuth client
func RevokeOAuthClient(req RevokeOAuthClientRequest) []byte {
	// Only governance can revoke OAuth clients
	caller := wasmx.GetCaller()
	if !hasGovRole(caller) {
		Revert("unauthorized: only governance can revoke OAuth clients")
	}

	// Get existing client
	clientInfo := getOAuthClient(req.ClientID)
	if clientInfo == nil {
		Revert("OAuth client not found")
	}

	// Mark as inactive
	clientInfo.Active = false
	storeOAuthClient(*clientInfo)

	// Note: OAuth2 server might not support revoking clients
	// We just mark it as inactive in our metadata

	LoggerInfo("OAuth client revoked", []string{
		"client_id", req.ClientID,
	})

	return []byte(`{"success": true}`)
}

// GetOAuthClient gets info about an OAuth client
func GetOAuthClient(req GetOAuthClientRequest) []byte {
	clientInfo := getOAuthClient(req.ClientID)
	if clientInfo == nil {
		return []byte(`{"error": "client not found"}`)
	}

	data, _ := json.Marshal(clientInfo)
	return data
}

// ListOAuthClients lists all registered OAuth clients
func ListOAuthClients(req ListOAuthClientsRequest) []byte {
	clientIDs := getOAuthClientList()
	clients := []OAuthClientInfo{}

	for _, clientID := range clientIDs {
		clientInfo := getOAuthClient(clientID)
		if clientInfo != nil {
			clients = append(clients, *clientInfo)
		}
	}

	data, _ := json.Marshal(clients)
	return data
}

// Storage helpers for OAuth clients

func storeOAuthClient(client OAuthClientInfo) {
	key := []byte(STORAGE_OAUTH_CLIENT_PREFIX + client.ClientID)
	data, _ := json.Marshal(client)
	wasmx.StorageStore(key, data)
}

func getOAuthClient(clientID string) *OAuthClientInfo {
	key := []byte(STORAGE_OAUTH_CLIENT_PREFIX + clientID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var client OAuthClientInfo
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
