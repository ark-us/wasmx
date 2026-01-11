package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// InitGenesis initializes the OAuth2 server with initial clients
func InitGenesis(req InitGenesisRequest) []byte {
	// Store the initialization data for later use on RoleChanged
	initDataBz, _ := json.Marshal(req)
	wasmx.StorageStore([]byte(STORAGE_INIT_DATA), initDataBz)

	if req.SupabaseJWT != "" {
		wasmx.StorageStore([]byte(STORAGE_SUPABASE_JWT_SECRET), []byte(req.SupabaseJWT))
	}
	if req.SupabaseKeyEndpoint != "" {
		wasmx.StorageStore([]byte(STORAGE_SUPABASE_KEY_ENDPOINT), []byte(req.SupabaseKeyEndpoint))
	}

	// Don't register clients or routes here - this is called during instantiate
	// when the contract doesn't have roles yet. These will be done in OnRoleChanged.

	return []byte(`{"success": true}`)
}

// OnRoleChanged is called when roles are assigned to this contract
func OnRoleChanged() []byte {
	// Load initialization data
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	var initData InitGenesisRequest
	if len(initDataBz) > 0 {
		if err := json.Unmarshal(initDataBz, &initData); err == nil {
			// Register initial clients if provided
			if len(initData.InitialClients) > 0 {
				for _, clientReq := range initData.InitialClients {
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
		}
	}

	// Register HTTP routes with HTTP registry (including static routes from init data)
	httpRegistryAddr := wasmx.GetAddressByRole(wasmx.ROLE_HTTP_SERVER)
	registerHttpRoutes(httpRegistryAddr, initData.StaticRoutes)

	LoggerInfo("OAuth2 server initialized on RoleChanged", nil)
	return []byte(`{"success": true}`)
}
