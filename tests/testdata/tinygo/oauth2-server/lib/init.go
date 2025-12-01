package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// InitGenesis initializes the OAuth2 server with initial clients
func InitGenesis(req InitGenesisRequest) []byte {
	// Register initial clients if provided
	if len(req.InitialClients) > 0 {
		for _, clientReq := range req.InitialClients {
			// Generate client ID and secret
			clientID := generateClientID(clientReq.Name)
			clientSecret := generateClientSecret()

			// Get current block height
			currentBlock := wasmx.GetCurrentBlock()

			// Create client
			client := OAuthClient{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Name:         clientReq.Name,
				Description:  clientReq.Description,
				RedirectURIs: clientReq.RedirectURIs,
				Scopes:       clientReq.Scopes,
				WebsiteURL:   clientReq.WebsiteURL,
				LogoURL:      clientReq.LogoURL,
				CreatedAt:    int64(currentBlock.Height),
				Active:       true,
			}

			storeOAuthClient(client)
			addToOAuthClientList(clientID)

			LoggerInfo("Initial OAuth client registered", []string{
				"client_id", clientID,
				"name", clientReq.Name,
			})
		}
	}

	return []byte(`{"success": true}`)
}
