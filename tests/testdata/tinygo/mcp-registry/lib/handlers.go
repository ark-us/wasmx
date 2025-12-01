package lib

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"

	sdkmath "cosmossdk.io/math"
	httpserver "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const DEFAULT_GAS_TOOL_CALL = 10000000

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

	// Check if user is logged in (session cookie)
	sessionID := getSessionFromCookie(req)
	if sessionID == "" {
		// Redirect to login with return_to parameter
		returnTo := req.Url // Full authorize URL
		loginURL := "/login?return_to=" + url.QueryEscape(returnTo)
		return httpserver.HttpResponseWrap{
			Error: "",
			Data: httpserver.HttpResponse{
				StatusCode: 302,
				Status:     "302 Found",
				Header:     http.Header{"Location": []string{loginURL}},
			},
		}
	}

	// Validate session and get user ID from OAuth2 server
	userID, err := callOAuth2ValidateSession(sessionID)
	if err != nil {
		// Session expired or invalid - redirect to login
		returnTo := req.Url
		loginURL := "/login?return_to=" + url.QueryEscape(returnTo)
		return httpserver.HttpResponseWrap{
			Error: "",
			Data: httpserver.HttpResponse{
				StatusCode: 302,
				Status:     "302 Found",
				Header:     http.Header{"Location": []string{loginURL}},
			},
		}
	}

	// TODO: Show consent screen here instead of auto-approving
	// For now, auto-approve and create authorization code

	// Create authorization code via OAuth2 server contract
	code, err := callOAuth2CreateAuthorizationCode(
		clientID,
		userID,
		redirectURI,
		params.Scopes,
		codeChallenge,
		codeChallengeMethod,
	)

	if err != nil {
		return sendTextResponse("Failed to create authorization code: "+err.Error(), 500)
	}

	// Build redirect URL
	if redirectURI == "" {
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
	codeVerifier := formData.Get("code_verifier")
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
	// Exchange code for tokens via OAuth2 server contract
	tokenResp, err := callOAuth2ExchangeCodeForToken(code, clientID, clientSecret, redirectURI, codeVerifier)
	if err != nil {
		return sendJSONResponse(map[string]string{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		}, 400)
	}

	LoggerInfo("Issued tokens", []string{
		"client_id", clientID,
	})

	return sendJSONResponse(tokenResp, 200)
}

func handleRefreshTokenGrant(refreshToken, clientID, clientSecret string) httpserver.HttpResponseWrap {
	// Refresh token via OAuth2 server contract
	tokenResp, err := callOAuth2RefreshAccessToken(refreshToken, clientID, clientSecret)
	if err != nil {
		return sendJSONResponse(map[string]string{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		}, 400)
	}

	LoggerInfo("Refreshed access token", []string{
		"client_id", clientID,
	})

	return sendJSONResponse(tokenResp, 200)
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
				"name":    "mcp-registry-server",
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
	// Get all tools from all registered contracts
	allToolsReq := GetAllToolsRequest{}
	toolsData := GetAllTools(allToolsReq)

	var toolsList []ToolsListEntry
	if err := json.Unmarshal(toolsData, &toolsList); err != nil {
		return sendMCPErrorResponse(req.ID, -32603, "Internal error")
	}

	// Convert to MCP protocol format
	mcpTools := []map[string]interface{}{}
	for _, tool := range toolsList {
		mcpTools = append(mcpTools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": mcpTools,
		},
	}
	return sendMCPResponse(response)
}

func handleMCPToolsCall(req *JSONRPCRequest, userID string) httpserver.HttpResponseWrap {
	// Parse params
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return sendMCPErrorResponse(req.ID, -32602, "Invalid params")
	}

	var params ToolCallParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return sendMCPErrorResponse(req.ID, -32602, "Invalid params")
	}

	// Parse tool name: "contract.tool" -> contract="contract", tool="tool"
	contractID, toolName, ok := parseToolName(params.Name)
	if !ok {
		return sendMCPErrorResponse(req.ID, -32602, "Invalid tool name format, expected 'contract.tool'")
	}

	// Find contract by route prefix
	routePrefix := "/tools/" + contractID
	contractAddress := getContractByRoute(routePrefix)
	if contractAddress == "" {
		return sendMCPErrorResponse(req.ID, -32601, "Tool contract not found")
	}

	// Validate contract is registered and has this tool
	registration := getContractRegistration(contractAddress)
	if registration == nil || !registration.Active {
		return sendMCPErrorResponse(req.ID, -32603, "Contract not active")
	}

	// Verify tool exists
	toolExists := false
	for _, tool := range registration.Tools {
		if tool.Name == toolName {
			toolExists = true
			break
		}
	}
	if !toolExists {
		return sendMCPErrorResponse(req.ID, -32601, "Tool not found in contract")
	}

	// Forward the call to the contract using wasmx.Call
	executeToolReq := map[string]interface{}{
		"execute_tool": map[string]interface{}{
			"tool_name": toolName,
			"arguments": params.Arguments,
			"user_id":   userID,
		},
	}

	callData, _ := json.Marshal(executeToolReq)

	// Make call to the contract
	value := sdkmath.NewInt(0)
	gas := big.NewInt(DEFAULT_GAS_TOOL_CALL)
	success, responseData := wasmx.Call(
		wasmx.Bech32String(contractAddress),
		&value,
		callData,
		gas,
		MODULE_NAME,
	)

	if !success {
		return sendMCPErrorResponse(req.ID, -32000, "Tool execution failed: "+string(responseData))
	}

	// Parse and return the response from contract
	var toolResponse map[string]interface{}
	if err := json.Unmarshal(responseData, &toolResponse); err != nil {
		return sendMCPErrorResponse(req.ID, -32603, "Invalid response from tool contract")
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolResponse,
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
