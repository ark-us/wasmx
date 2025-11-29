package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	httpserver "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	oauth2server "github.com/loredanacirstea/wasmx-env-oauth2server/lib"
)

func handleOAuthAuthorize(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	params := getParams()
	if params == nil {
		return sendJSONResponse(map[string]string{"error": "Server not configured"}, 500)
	}

	// Parse query parameters from URL
	parsedURL, err := url.Parse(req.Url)
	if err != nil {
		return sendTextResponse("Invalid URL", 400)
	}
	query := parsedURL.Query()

	responseType := query.Get("response_type")
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	state := query.Get("state")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")

	// Validate response_type
	if responseType != "code" {
		return sendTextResponse("Invalid response_type, must be 'code'", 400)
	}

	// Check if client exists (don't validate secret in authorize endpoint)
	getClientResp := oauth2server.GetClient(&oauth2server.GetClientRequest{
		InstanceID: OAUTH2_INSTANCE_ID,
		ClientID:   clientID,
	})

	if getClientResp.Error != "" {
		return sendTextResponse("Invalid client_id", 400)
	}

	// Create authorization code (using OAuth2 server with PKCE support)
	codeResp := oauth2server.CreateAuthorizationCode(&oauth2server.CreateAuthorizationCodeRequest{
		InstanceID:          OAUTH2_INSTANCE_ID,
		ClientID:            clientID,
		UserID:              "default-user", // In production, get from authenticated session
		RedirectURI:         redirectURI,
		Scopes:              params.Scopes,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresInSeconds:    600, // 10 minutes
	})

	if codeResp.Error != "" {
		return sendTextResponse("Failed to create authorization code: "+codeResp.Error, 500)
	}

	code := codeResp.Code

	// Build redirect URL
	if redirectURI == "" {
		// If no redirect_uri, show the code to the user
		html := fmt.Sprintf(`
			<html>
			<body>
				<h1>Authorization Successful</h1>
				<p>Your authorization code is:</p>
				<code style="background: #f0f0f0; padding: 10px; display: block;">%s</code>
				<p>This code expires in 10 minutes.</p>
				<p>PKCE: %s</p>
			</body>
			</html>
		`, code, codeChallenge != "")
		return sendHTMLResponse(html, 200)
	}

	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	return httpserver.HttpResponseWrap{
		Error: "",
		Data: httpserver.HttpResponse{
			StatusCode:  302,
			Status:      "302 Found",
			Header:      http.Header{"Location": []string{redirectURL.String()}},
			RedirectUrl: redirectURL.String(),
		},
	}
}

func handleOAuthToken(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	if req.Method != "POST" {
		return sendTextResponse("Method not allowed", 405)
	}

	params := getParams()
	if params == nil {
		return sendJSONResponse(map[string]string{"error": "Server not configured"}, 500)
	}

	// Parse form data from request body
	formData, err := url.ParseQuery(string(req.Data))
	if err != nil {
		return sendJSONResponse(map[string]string{"error": "invalid_request"}, 400)
	}

	grantType := formData.Get("grant_type")
	code := formData.Get("code")
	redirectURI := formData.Get("redirect_uri")
	clientID := formData.Get("client_id")
	clientSecret := formData.Get("client_secret")
	codeVerifier := formData.Get("code_verifier") // PKCE
	refreshToken := formData.Get("refresh_token")

	// Handle different grant types
	if grantType == "authorization_code" {
		return handleAuthorizationCodeGrant(code, redirectURI, clientID, clientSecret, codeVerifier)
	} else if grantType == "refresh_token" {
		return handleRefreshTokenGrant(refreshToken, clientID, clientSecret)
	}

	return sendJSONResponse(map[string]string{"error": "unsupported_grant_type"}, 400)
}

func handleAuthorizationCodeGrant(code, redirectURI, clientID, clientSecret, codeVerifier string) httpserver.HttpResponseWrap {
	// Validate client credentials
	validateClientResp := oauth2server.ValidateClient(&oauth2server.ValidateClientRequest{
		InstanceID:   OAUTH2_INSTANCE_ID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	if validateClientResp.Error != "" || !validateClientResp.Valid {
		return sendJSONResponse(map[string]string{"error": "invalid_client"}, 401)
	}

	// Validate and consume authorization code (using OAuth2 server with PKCE validation)
	validateCodeResp := oauth2server.ValidateAuthorizationCode(&oauth2server.ValidateAuthorizationCodeRequest{
		InstanceID:   OAUTH2_INSTANCE_ID,
		Code:         code,
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		CodeVerifier: codeVerifier, // PKCE
	})

	if validateCodeResp.Error != "" || !validateCodeResp.Valid {
		return sendJSONResponse(map[string]string{
			"error":             "invalid_grant",
			"error_description": validateCodeResp.Error,
		}, 400)
	}

	// Issue access token
	accessTokenResp := oauth2server.IssueAccessToken(&oauth2server.IssueAccessTokenRequest{
		InstanceID:       OAUTH2_INSTANCE_ID,
		ClientID:         clientID,
		UserID:           validateCodeResp.UserID,
		Scopes:           validateCodeResp.Scopes,
		ExpiresInSeconds: 3600, // 1 hour
	})

	if accessTokenResp.Error != "" {
		return sendJSONResponse(map[string]string{"error": "server_error"}, 500)
	}

	// Issue refresh token
	refreshTokenResp := oauth2server.IssueRefreshToken(&oauth2server.IssueRefreshTokenRequest{
		InstanceID:       OAUTH2_INSTANCE_ID,
		ClientID:         clientID,
		UserID:           validateCodeResp.UserID,
		Scopes:           validateCodeResp.Scopes,
		ExpiresInSeconds: 0, // Non-expiring
	})

	if refreshTokenResp.Error != "" {
		LoggerError("Failed to issue refresh token", []string{"error", refreshTokenResp.Error})
	}

	LoggerInfo("Generated access token", []string{
		"user_id", validateCodeResp.UserID,
		"client_id", clientID,
	})

	// Return token response
	response := TokenResponse{
		AccessToken:  accessTokenResp.Token,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshTokenResp.Token,
	}

	return sendJSONResponse(response, 200)
}

func handleRefreshTokenGrant(refreshToken, clientID, clientSecret string) httpserver.HttpResponseWrap {
	// Validate client credentials
	validateClientResp := oauth2server.ValidateClient(&oauth2server.ValidateClientRequest{
		InstanceID:   OAUTH2_INSTANCE_ID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	if validateClientResp.Error != "" || !validateClientResp.Valid {
		return sendJSONResponse(map[string]string{"error": "invalid_client"}, 401)
	}

	// Refresh the access token
	refreshResp := oauth2server.RefreshAccessToken(&oauth2server.RefreshAccessTokenRequest{
		InstanceID:       OAUTH2_INSTANCE_ID,
		RefreshToken:     refreshToken,
		ExpiresInSeconds: 3600, // 1 hour
	})

	if refreshResp.Error != "" {
		return sendJSONResponse(map[string]string{
			"error":             "invalid_grant",
			"error_description": refreshResp.Error,
		}, 400)
	}

	LoggerInfo("Refreshed access token", []string{
		"user_id", refreshResp.UserID,
		"client_id", clientID,
	})

	// Return token response with NEW rotated refresh token
	response := TokenResponse{
		AccessToken:  refreshResp.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshResp.RefreshToken, // New rotated refresh token
	}

	return sendJSONResponse(response, 200)
}

func handleSSE(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	// Handle OPTIONS preflight
	if req.Method == "OPTIONS" {
		return httpserver.HttpResponseWrap{
			Error: "",
			Data: httpserver.HttpResponse{
				StatusCode: 200,
				Status:     "200 OK",
				Header: http.Header{
					"Access-Control-Allow-Origin":  []string{"*"},
					"Access-Control-Allow-Methods": []string{"GET, POST, OPTIONS"},
					"Access-Control-Allow-Headers": []string{"Content-Type, Authorization"},
				},
			},
		}
	}

	// Extract and validate OAuth token
	token := extractToken(req)
	userID, valid := validateToken(token)

	if token != "" {
		if valid {
			LoggerInfo("Token valid for user", []string{"user_id", userID})
		} else {
			LoggerInfo("Token invalid or expired", nil)
			userID = ""
		}
	} else {
		LoggerInfo("No token provided", nil)
	}

	// Handle GET request - send endpoint info
	if req.Method == "GET" {
		scheme := "http"
		if req.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := req.Header.Get("Host")
		baseURL := fmt.Sprintf("%s://%s", scheme, host)

		event := map[string]interface{}{
			"url": "/sse",
			"oauth": map[string]string{
				"authorization_endpoint": baseURL + "/oauth/authorize",
				"token_endpoint":         baseURL + "/oauth/token",
			},
		}

		eventData, _ := json.Marshal(event)
		sseData := fmt.Sprintf("event: endpoint\ndata: %s\n\n", string(eventData))

		return httpserver.HttpResponseWrap{
			Error: "",
			Data: httpserver.HttpResponse{
				StatusCode: 200,
				Status:     "200 OK",
				Header: http.Header{
					"Content-Type":  []string{"text/event-stream"},
					"Cache-Control": []string{"no-cache"},
					"Connection":    []string{"keep-alive"},
				},
				Data: []byte(sseData),
			},
		}
	}

	// Handle POST request - process MCP messages
	if req.Method != "POST" {
		return sendTextResponse("Method not allowed", 405)
	}

	var rpcReq JSONRPCRequest
	if err := json.Unmarshal(req.Data, &rpcReq); err != nil {
		return sendMCPErrorResponse(nil, -32700, "Parse error")
	}

	switch rpcReq.Method {
	case "initialize":
		return handleMCPInitialize(&rpcReq, userID)
	case "tools/list":
		return handleMCPToolsList(&rpcReq, userID)
	case "tools/call":
		return handleMCPToolsCall(&rpcReq, userID)
	default:
		return sendMCPErrorResponse(rpcReq.ID, -32601, "Method not found")
	}
}

func handleMCPInitialize(req *JSONRPCRequest, userID string) httpserver.HttpResponseWrap {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "mcp-auth-server",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		},
	}
	return sendMCPResponse(response)
}

func handleMCPToolsList(req *JSONRPCRequest, userID string) httpserver.HttpResponseWrap {
	tools := []map[string]interface{}{
		{
			"name":        "list_items",
			"description": "Returns a list of basic items (unauthenticated)",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "set_favorite_color",
			"description": "Set your favorite color (requires authentication)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"color": map[string]string{
						"type":        "string",
						"description": "Your favorite color",
					},
				},
				"required": []string{"color"},
			},
		},
		{
			"name":        "get_favorite_color",
			"description": "Get your favorite color (requires authentication)",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
	return sendMCPResponse(response)
}

func handleMCPToolsCall(req *JSONRPCRequest, userID string) httpserver.HttpResponseWrap {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sendMCPErrorResponse(req.ID, -32602, "Invalid params")
	}

	switch params.Name {
	case "list_items":
		return handleListItems(req)
	case "set_favorite_color":
		if userID == "" {
			return sendMCPErrorResponse(req.ID, -32001, "Authentication required")
		}
		return handleSetFavoriteColor(req, userID, params.Arguments)
	case "get_favorite_color":
		if userID == "" {
			return sendMCPErrorResponse(req.ID, -32002, "Authentication required")
		}
		return handleGetFavoriteColor(req, userID)
	default:
		return sendMCPErrorResponse(req.ID, -32601, "Tool not found")
	}
}

func handleListItems(req *JSONRPCRequest) httpserver.HttpResponseWrap {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Items: water, salt, pepper",
				},
			},
		},
	}
	return sendMCPResponse(response)
}

func handleSetFavoriteColor(req *JSONRPCRequest, userID string, args map[string]interface{}) httpserver.HttpResponseWrap {
	color, ok := args["color"].(string)
	if !ok || color == "" {
		return sendMCPErrorResponse(req.ID, -32602, "Color parameter required")
	}

	userData := &UserData{FavoriteColor: color}
	storeUserData(userID, userData)

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Favorite color set to: %s", color),
				},
			},
		},
	}
	return sendMCPResponse(response)
}

func handleGetFavoriteColor(req *JSONRPCRequest, userID string) httpserver.HttpResponseWrap {
	userData := getUserData(userID)

	message := "No favorite color set"
	if userData.FavoriteColor != "" {
		message = fmt.Sprintf("Your favorite color is: %s", userData.FavoriteColor)
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		},
	}
	return sendMCPResponse(response)
}

func sendMCPResponse(response JSONRPCResponse) httpserver.HttpResponseWrap {
	jsonData, _ := json.Marshal(response)
	return httpserver.HttpResponseWrap{
		Error: "",
		Data: httpserver.HttpResponse{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

func sendMCPErrorResponse(id interface{}, code int, message string) httpserver.HttpResponseWrap {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	return sendMCPResponse(response)
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

func sendHTMLResponse(html string, statusCode int) httpserver.HttpResponseWrap {
	return httpserver.HttpResponseWrap{
		Error: "",
		Data: httpserver.HttpResponse{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d OK", statusCode),
			Header:     http.Header{"Content-Type": []string{"text/html"}},
			Data:       []byte(html),
		},
	}
}
