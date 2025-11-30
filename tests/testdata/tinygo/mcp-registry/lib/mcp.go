package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	httpserver "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	oauth2server "github.com/loredanacirstea/wasmx-env-oauth2server/lib"
	postgresql "github.com/loredanacirstea/wasmx-env-postgresql/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const (
	OAUTH2_INSTANCE_ID       = "mcp-oauth"
	POSTGRESQL_CONNECTION_ID = "mcp-db"
)

func InitGenesis(req InitGenesisRequest) []byte {
	// Store the OAuth client configuration
	paramsBz, _ := json.Marshal(req.Params)
	wasmx.StorageStore([]byte(STORAGE_PARAMS), paramsBz)

	LoggerInfo("MCP Registry initialized - params stored", []string{
		"client_id", req.Params.ClientID,
	})

	// Register initial contracts if provided
	if len(req.InitialContracts) > 0 {
		for _, contract := range req.InitialContracts {
			// Parse and validate tools JSON
			var tools []MCPToolDefinition
			if err := json.Unmarshal([]byte(contract.ToolsJSON), &tools); err != nil {
				LoggerError("Failed to parse tools for initial contract", []string{
					"address", contract.ContractAddress,
					"error", err.Error(),
				})
				continue
			}

			// Get current block height
			currentBlock := wasmx.GetCurrentBlock()

			// Create registration
			registration := ContractRegistration{
				Address:       contract.ContractAddress,
				RoutePrefix:   contract.RoutePrefix,
				Tools:         tools,
				RegisteredAt:  int64(currentBlock.Height),
				LastUpdatedAt: int64(currentBlock.Height),
				Active:        true,
			}

			// Store registration
			storeContractRegistration(registration)
			addToRegisteredList(contract.ContractAddress)
			storeRouteMapping(contract.RoutePrefix, contract.ContractAddress)

			// Update HTTP routing
			updateHttpRoute(contract.RoutePrefix, contract.ContractAddress)

			LoggerInfo("Initial MCP contract registered", []string{
				"contract_address", contract.ContractAddress,
				"route_prefix", contract.RoutePrefix,
				"tool_count", fmt.Sprintf("%d", len(tools)),
			})
		}
	}

	return []byte(`{"success": true}`)
}

func ConnectDatabase(req *ConnectDatabaseRequest) []byte {
	// Connect to PostgreSQL
	pgResp := postgresql.Connect(&postgresql.SqlConnectionRequest{
		Connection: req.Connection,
		DbName:     req.DbName,
		Id:         POSTGRESQL_CONNECTION_ID,
	})

	if pgResp.Error != "" {
		LoggerError("Failed to connect to PostgreSQL", []string{"error", pgResp.Error})
		return []byte(fmt.Sprintf(`{"error": "%s"}`, pgResp.Error))
	}

	LoggerInfo("Database connected", nil)
	return []byte(`{"success": true}`)
}

// InitializeTables connects to PostgreSQL and creates tables
func InitializeTables() []byte {
	// Get stored params to get database connection info
	params := getParams()
	if params == nil {
		LoggerError("Failed to get params", nil)
		return []byte(`{"error": "params not initialized"}`)
	}

	// Connect to PostgreSQL if not already connected
	pgResp := postgresql.Connect(&postgresql.SqlConnectionRequest{
		Connection: params.DbConnection,
		DbName:     params.DbName,
		Id:         POSTGRESQL_CONNECTION_ID,
	})

	if pgResp.Error != "" {
		LoggerError("Failed to connect to PostgreSQL", []string{"error", pgResp.Error})
		return []byte(fmt.Sprintf(`{"error": "%s"}`, pgResp.Error))
	}

	LoggerInfo("Database connected for registry", nil)

	// Initialize OAuth2 server
	return InitializeOAuth2()
}

// InitializeOAuth2 sets up OAuth2 server and registers client
func InitializeOAuth2() []byte {
	// Initialize OAuth2 server
	oauth2Resp := oauth2server.Initialize(&oauth2server.InitializeRequest{
		InstanceID:   OAUTH2_INSTANCE_ID,
		ConnectionID: POSTGRESQL_CONNECTION_ID,
	})

	if oauth2Resp.Error != "" {
		LoggerError("Failed to initialize OAuth2 server", []string{"error", oauth2Resp.Error})
		return []byte(fmt.Sprintf(`{"error": "%s"}`, oauth2Resp.Error))
	}

	// Register the OAuth client from params
	params := getParams()
	if params != nil {
		registerResp := oauth2server.RegisterClient(&oauth2server.RegisterClientRequest{
			InstanceID:   OAUTH2_INSTANCE_ID,
			ClientID:     params.ClientID,
			ClientSecret: params.ClientSecret,
			RedirectURIs: params.RedirectURIs,
			Scopes:       params.Scopes,
			GrantTypes:   []string{"authorization_code", "refresh_token"},
		})

		if registerResp.Error != "" {
			LoggerError("Failed to register OAuth client", []string{"error", registerResp.Error})
			// Not fatal - client might already exist
		} else {
			LoggerInfo("OAuth client registered", []string{"client_id", params.ClientID})
		}
	}

	LoggerInfo("OAuth2 initialized", nil)
	return []byte(`{"success": true}`)
}

func GetParams(req GetParamsRequest) []byte {
	paramsBz := wasmx.StorageLoad([]byte(STORAGE_PARAMS))
	if len(paramsBz) == 0 {
		return []byte(`{"error": "params not initialized"}`)
	}
	return paramsBz
}

func StartServer(req *StartServerRequest) {
	config := httpserver.WebsrvConfig{
		Address:            req.Address,
		CORSAllowedOrigins: []string{"*"},
		CORSAllowedMethods: []string{"GET", "POST", "OPTIONS"},
		CORSAllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxOpenConnections: 100,
		RequestBodyMaxSize: 1024 * 1024, // 1MB
		RouteToContractAddress: map[string]string{
			"/.well-known/oauth-authorization-server": string(wasmx.GetAddress()),
			"/oauth/authorize":                        string(wasmx.GetAddress()),
			"/oauth/token":                            string(wasmx.GetAddress()),
			"/sse":                                    string(wasmx.GetAddress()),
			"/":                                       string(wasmx.GetAddress()),
		},
	}

	resp := httpserver.StartWebServer(&httpserver.StartWebServerRequest{
		Config: config,
	})

	if resp.Error != "" {
		LoggerError("Failed to start web server", []string{"error", resp.Error})
		Revert("failed to start server: " + resp.Error)
	}

	LoggerInfo("MCP Registry Server started", []string{
		"address", req.Address,
	})
}

func getParams() *ServerParams {
	paramsBz := wasmx.StorageLoad([]byte(STORAGE_PARAMS))
	if len(paramsBz) == 0 {
		return nil
	}
	var params ServerParams
	if err := json.Unmarshal(paramsBz, &params); err != nil {
		return nil
	}
	return &params
}

func extractToken(req *httpserver.HttpRequestIncoming) string {
	// Try Authorization header first (OAuth Bearer token)
	auth := req.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

func validateToken(token string) (string, bool) {
	if token == "" {
		return "", false
	}

	resp := oauth2server.ValidateAccessToken(&oauth2server.ValidateAccessTokenRequest{
		InstanceID: OAUTH2_INSTANCE_ID,
		Token:      token,
	})

	if resp.Error != "" || !resp.Valid {
		return "", false
	}

	return resp.UserID, true
}

func HandleHttpRequest(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	// Extract path from RequestURI (remove query string)
	path := req.RequestURI
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Route the request based on the URL path
	switch path {
	case "/.well-known/oauth-authorization-server":
		return handleOAuthMetadata(req)
	case "/oauth/authorize":
		return handleOAuthAuthorize(req)
	case "/oauth/token":
		return handleOAuthToken(req)
	case "/sse":
		return handleSSE(req)
	case "/":
		return handleRoot(req)
	default:
		return httpserver.HttpResponseWrap{
			Error: "",
			Data: httpserver.HttpResponse{
				StatusCode: 404,
				Status:     "404 Not Found",
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Data:       []byte("Not Found"),
			},
		}
	}
}

func handleRoot(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	params := getParams()
	if params == nil {
		return sendJSONResponse(map[string]string{
			"error": "Server not configured",
		}, 500)
	}

	response := map[string]interface{}{
		"message":       "MCP Registry Server",
		"version":       "1.0.0",
		"sse_endpoint":  "/sse",
		"authorize_url": "/oauth/authorize",
		"token_url":     "/oauth/token",
		"auth_methods": map[string]interface{}{
			"oauth2": map[string]string{
				"client_id":     params.ClientID,
				"client_secret": params.ClientSecret,
			},
		},
	}

	return sendJSONResponse(response, 200)
}

func handleOAuthMetadata(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	// Parse the base URL from the request
	scheme := "http"
	if req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	// Use X-Forwarded-Host if available (for proxies like ngrok)
	host := req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Header.Get("Host")
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	metadata := map[string]interface{}{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/oauth/authorize",
		"token_endpoint":                        baseURL + "/oauth/token",
		"grant_types_supported":                 []string{"authorization_code"},
		"response_types_supported":              []string{"code"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"code_challenge_methods_supported":      []string{"plain", "S256"},
	}

	return sendJSONResponse(metadata, 200)
}

func sendJSONResponse(data interface{}, statusCode int) httpserver.HttpResponseWrap {
	jsonData, _ := json.Marshal(data)
	return httpserver.HttpResponseWrap{
		Error: "",
		Data: httpserver.HttpResponse{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d OK", statusCode),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

func sendTextResponse(text string, statusCode int) httpserver.HttpResponseWrap {
	return httpserver.HttpResponseWrap{
		Error: "",
		Data: httpserver.HttpResponse{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d", statusCode),
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Data:       []byte(text),
		},
	}
}
