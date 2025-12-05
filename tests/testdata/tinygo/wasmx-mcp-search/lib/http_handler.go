package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// HandleHttpRequest handles incoming HTTP requests for this tool contract
func HandleHttpRequest(req HttpRequestIncoming) []byte {
	// Extract path from Url or RequestURI (remove query string)
	path := req.Url
	if path == "" {
		path = req.RequestURI
	}
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Extract token from Authorization header
	token := extractToken(&req)
	if token == "" {
		return marshalHttpResponse(sendErrorResponse("Unauthorized: missing token", 401))
	}

	// Validate token and get user_id
	userID, valid := validateToken(token)
	if !valid {
		return marshalHttpResponse(sendErrorResponse("Unauthorized: invalid token", 401))
	}

	// All requests under /tools/search/* should be MCP JSON-RPC requests
	if req.Method != "POST" {
		return marshalHttpResponse(sendErrorResponse("Method not allowed", 405))
	}

	// Get request body
	body := req.Data

	// Parse JSON-RPC request
	var rpcReq JSONRPCRequest
	if err := json.Unmarshal(body, &rpcReq); err != nil {
		return marshalHttpResponse(sendJSONRPCError(nil, -32700, "Parse error: "+err.Error()))
	}

	// Route based on JSON-RPC method
	switch rpcReq.Method {
	case "tools/list":
		return marshalHttpResponse(handleToolsList(&rpcReq, userID))
	case "tools/call":
		return marshalHttpResponse(handleToolsCall(&rpcReq, userID))
	default:
		return marshalHttpResponse(sendJSONRPCError(rpcReq.ID, -32601, "Method not found: "+rpcReq.Method))
	}
}

// extractToken extracts Bearer token from Authorization header
func extractToken(req *HttpRequestIncoming) string {
	authHeaders := req.Header["Authorization"]
	if len(authHeaders) == 0 {
		return ""
	}
	auth := authHeaders[0]
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// validateToken validates the token via OAuth2 server
func validateToken(token string) (string, bool) {
	if token == "" {
		return "", false
	}

	// Get OAuth2 server address
	oauth2Addr := wasmx.GetAddressByRole(wasmx.ROLE_OAUTH2_SERVER)
	if oauth2Addr == "" {
		LoggerError("OAuth2 server not found", nil)
		return "", false
	}

	// Call OAuth2 server to validate token
	validateMsg := map[string]interface{}{
		"validate_access_token": map[string]interface{}{
			"access_token": token,
		},
	}
	msgBz, _ := json.Marshal(validateMsg)
	ok, data := wasmx.CallSimple(oauth2Addr, msgBz, false, MODULE_NAME)
	if !ok {
		return "", false
	}

	// Parse response
	var resp struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", false
	}

	return resp.UserID, resp.UserID != ""
}

// handleToolsList returns the list of tools provided by this contract
func handleToolsList(req *JSONRPCRequest, userID string) HttpResponseWrap {
	tools := getToolDefinitions()

	result := map[string]interface{}{
		"tools": tools,
	}

	return sendJSONRPCResponse(req.ID, result)
}

// handleToolsCall executes a tool
func handleToolsCall(req *JSONRPCRequest, userID string) HttpResponseWrap {
	LoggerInfo("handleToolsCall started", []string{"user_id", userID})

	// Parse params
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return sendJSONRPCError(req.ID, -32602, "Invalid params: expected object")
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return sendJSONRPCError(req.ID, -32602, "Missing or invalid 'name' parameter")
	}

	LoggerInfo("Tool requested", []string{"tool_name", toolName})

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = map[string]interface{}{}
	}

	// Execute the tool
	executeReq := ExecuteToolRequest{
		ToolName:  toolName,
		Arguments: arguments,
		UserID:    userID,
	}

	LoggerInfo("Calling ExecuteTool", []string{"tool_name", toolName})
	resultBz := ExecuteTool(executeReq)
	LoggerInfo("ExecuteTool returned", []string{"result_length", fmt.Sprintf("%d", len(resultBz))})

	// Parse result
	var toolResult ExecuteToolResponse
	if err := json.Unmarshal(resultBz, &toolResult); err != nil {
		return sendJSONRPCError(req.ID, -32603, "Failed to parse tool result: "+err.Error())
	}

	// Return result
	result := map[string]interface{}{
		"content": toolResult.Content,
	}

	if toolResult.IsError {
		result["isError"] = true
	}

	return sendJSONRPCResponse(req.ID, result)
}

// JSON-RPC helper types
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// marshalHttpResponse marshals an HttpResponseWrap to JSON
func marshalHttpResponse(resp HttpResponseWrap) []byte {
	data, _ := json.Marshal(resp)
	return data
}

// sendJSONRPCResponse sends a successful JSON-RPC response
func sendJSONRPCResponse(id interface{}, result interface{}) HttpResponseWrap {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	jsonData, _ := json.Marshal(resp)
	return HttpResponseWrap{
		Error: "",
		Data: HttpResponse{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     map[string][]string{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

// sendJSONRPCError sends a JSON-RPC error response
func sendJSONRPCError(id interface{}, code int, message string) HttpResponseWrap {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}

	jsonData, _ := json.Marshal(resp)
	return HttpResponseWrap{
		Error: "",
		Data: HttpResponse{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     map[string][]string{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

// sendErrorResponse sends a plain HTTP error response
func sendErrorResponse(message string, statusCode int) HttpResponseWrap {
	return HttpResponseWrap{
		Error: "",
		Data: HttpResponse{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
			Header:     map[string][]string{"Content-Type": []string{"text/plain"}},
			Data:       []byte(message),
		},
	}
}
