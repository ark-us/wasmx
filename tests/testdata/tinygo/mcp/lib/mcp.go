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
	STORAGE_PARAMS = "params"
)

func InitGenesis(req InitGenesisRequest) []byte {
	// Store the OAuth client configuration
	paramsBz, _ := json.Marshal(req.Params)
	wasmx.StorageStore([]byte(STORAGE_PARAMS), paramsBz)

	LoggerInfo("MCP server initialized - params stored", []string{
		"client_id", req.Params.ClientID,
	})

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

// InitializeTables creates PostgreSQL tables (can be called from RoleChanged hook)
func InitializeTables() []byte {
	// Create user data table
	createTableResp := postgresql.Execute(&postgresql.SqlExecuteRequest{
		Id: POSTGRESQL_CONNECTION_ID,
		Query: `
			CREATE TABLE IF NOT EXISTS mcp_user_data (
				user_id VARCHAR(255) PRIMARY KEY,
				favorite_color VARCHAR(100)
			)
		`,
		Params: postgresql.Params{},
	})

	if createTableResp.Error != "" {
		LoggerError("Failed to create user data table", []string{"error", createTableResp.Error})
		return []byte(fmt.Sprintf(`{"error": "%s"}`, createTableResp.Error))
	}

	LoggerInfo("Database tables created", nil)
	return []byte(`{"success": true}`)
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

	LoggerInfo("MCP OAuth Server started", []string{
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

func storeUserData(userID string, data *UserData) {
	query := `
		INSERT INTO mcp_user_data (user_id, favorite_color)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET favorite_color = $2
	`

	colorParam, _ := json.Marshal(postgresql.SqlQueryParam{Value: data.FavoriteColor})
	userIDParam, _ := json.Marshal(postgresql.SqlQueryParam{Value: userID})

	resp := postgresql.Execute(&postgresql.SqlExecuteRequest{
		Id:     POSTGRESQL_CONNECTION_ID,
		Query:  query,
		Params: postgresql.Params{userIDParam, colorParam},
	})

	if resp.Error != "" {
		LoggerError("Failed to store user data", []string{"error", resp.Error})
	}
}

func getUserData(userID string) *UserData {
	query := "SELECT favorite_color FROM mcp_user_data WHERE user_id = $1"

	userIDParam, _ := json.Marshal(postgresql.SqlQueryParam{Value: userID})

	resp := postgresql.Query(&postgresql.SqlQueryRequest{
		Id:     POSTGRESQL_CONNECTION_ID,
		Query:  query,
		Params: postgresql.Params{userIDParam},
	})

	if resp.Error != "" {
		LoggerError("Failed to get user data", []string{"error", resp.Error})
		return &UserData{}
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(resp.Data, &results); err != nil || len(results) == 0 {
		return &UserData{}
	}

	userData := &UserData{}
	if color, ok := results[0]["favorite_color"].(string); ok {
		userData.FavoriteColor = color
	}
	return userData
}

func extractToken(req *httpserver.HttpRequestIncoming) string {
	// Try Authorization header first (OAuth Bearer token)
	auth := req.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Note: API key authentication removed - use OAuth2 only
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
	// Route the request based on the URL path
	switch req.RequestURI {
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
		"message":       "MCP OAuth Server",
		"version":       "1.0.0",
		"sse_endpoint":  "/sse",
		"authorize_url": "/oauth/authorize",
		"token_url":     "/oauth/token",
		"auth_methods": map[string]interface{}{
			"oauth2": map[string]string{
				"client_id":     params.ClientID,
				"client_secret": params.ClientSecret,
			},
			"api_key": map[string]string{
				"header":  "X-API-Key",
				"api_key": "claude-api-key-123",
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
	host := req.Header.Get("Host")
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
