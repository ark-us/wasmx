package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// HandleHttpRequest handles incoming HTTP requests for this tool contract
// Supports both:
// 1. Path-based routing (OpenAI style): POST /tools/userdata/get_favorite_color with body {"color": "blue"}
// 2. JSON-RPC style: POST /tools/userdata with body {"method": "tools/call", "params": {...}}
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

	if req.Method != "POST" {
		return marshalHttpResponse(sendErrorResponse("Method not allowed", 405))
	}

	// Get request body
	body := req.Data

	// Try to parse as JSON-RPC first
	var rpcReq JSONRPCRequest
	if err := json.Unmarshal(body, &rpcReq); err == nil && rpcReq.Method != "" {
		// JSON-RPC style request
		switch rpcReq.Method {
		case "tools/list":
			return marshalHttpResponse(handleToolsList(&rpcReq, userID))
		case "tools/call":
			return marshalHttpResponse(handleToolsCall(&rpcReq, userID))
		default:
			return marshalHttpResponse(sendJSONRPCError(rpcReq.ID, -32601, "Method not found: "+rpcReq.Method))
		}
	}

	// Path-based routing (OpenAI style)
	// Extract tool name from path: /tools/userdata/get_favorite_color -> get_favorite_color
	toolName := extractToolNameFromPath(path)
	if toolName == "" {
		return marshalHttpResponse(sendPlainJSONError("Invalid path: could not extract tool name", 400))
	}

	// Parse body as tool arguments
	var arguments map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &arguments); err != nil {
			return marshalHttpResponse(sendPlainJSONError("Invalid JSON body: "+err.Error(), 400))
		}
	}
	if arguments == nil {
		arguments = map[string]interface{}{}
	}

	// Execute the tool
	executeReq := ExecuteToolRequest{
		ToolName:  toolName,
		Arguments: arguments,
		UserID:    userID,
	}

	resultBz := ExecuteTool(executeReq)

	// Parse result
	var toolResult ExecuteToolResponse
	if err := json.Unmarshal(resultBz, &toolResult); err != nil {
		return marshalHttpResponse(sendPlainJSONError("Failed to parse tool result: "+err.Error(), 500))
	}

	// Return plain JSON response (not JSON-RPC) for path-based requests
	return marshalHttpResponse(sendPlainJSONResponse(toolResult))
}

// extractToolNameFromPath extracts the tool name from a path like /tools/userdata/get_favorite_color
func extractToolNameFromPath(path string) string {
	// Remove leading slash and split
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// Expected format: tools/userdata/<tool_name>
	if len(parts) >= 3 && parts[0] == "tools" && parts[1] == "userdata" {
		return parts[2]
	}
	return ""
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
			"token": token,
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
	// Parse params
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return sendJSONRPCError(req.ID, -32602, "Invalid params: expected object")
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return sendJSONRPCError(req.ID, -32602, "Missing or invalid 'name' parameter")
	}

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

	resultBz := ExecuteTool(executeReq)

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

// sendPlainJSONError sends a plain JSON error response (for path-based routing)
func sendPlainJSONError(message string, statusCode int) HttpResponseWrap {
	errorResp := map[string]interface{}{
		"error": message,
	}
	jsonData, _ := json.Marshal(errorResp)
	return HttpResponseWrap{
		Error: "",
		Data: HttpResponse{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
			Header:     map[string][]string{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

// sendPlainJSONResponse sends a plain JSON response (for path-based routing)
func sendPlainJSONResponse(result ExecuteToolResponse) HttpResponseWrap {
	statusCode := 200
	if result.IsError {
		statusCode = 400
	}

	// Build response based on content
	var responseData []byte
	if len(result.Content) == 1 && result.Content[0].Type == "text" {
		// Single text content - try to parse as JSON, otherwise wrap
		text := result.Content[0].Text
		var jsonCheck interface{}
		if err := json.Unmarshal([]byte(text), &jsonCheck); err == nil {
			// Already valid JSON
			responseData = []byte(text)
		} else {
			// Wrap in result object
			resp := map[string]interface{}{
				"result": text,
			}
			responseData, _ = json.Marshal(resp)
		}
	} else {
		// Multiple content items or non-text - wrap in content array
		resp := map[string]interface{}{
			"content": result.Content,
		}
		if result.IsError {
			resp["isError"] = true
		}
		responseData, _ = json.Marshal(resp)
	}

	return HttpResponseWrap{
		Error: "",
		Data: HttpResponse{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
			Header:     map[string][]string{"Content-Type": []string{"application/json"}},
			Data:       responseData,
		},
	}
}
