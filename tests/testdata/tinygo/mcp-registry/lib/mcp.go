package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func InitGenesis(req InitGenesisRequest) []byte {
	// ONLY store the initialization data for later use in RoleChanged
	// Do NOT register contracts or call other contracts - we don't have roles yet!
	initDataBz, _ := json.Marshal(req)
	wasmx.StorageStore([]byte(STORAGE_INIT_DATA), initDataBz)

	// Store the OAuth client configuration for quick access
	paramsBz, _ := json.Marshal(req.Params)
	wasmx.StorageStore([]byte(STORAGE_PARAMS), paramsBz)

	LoggerInfo("MCP Registry init data stored", []string{
		"redirect_uris_count", fmt.Sprintf("%d", len(req.Params.RedirectURIs)),
		"initial_contracts_count", fmt.Sprintf("%d", len(req.InitialContracts)),
	})

	return []byte(`{"success": true}`)
}

// InitializeTables initializes the MCP registry tables
func InitializeTables() []byte {
	LoggerInfo("InitializeTables called", nil)
	// Load initialization data
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		LoggerError("Init data not found", nil)
		return []byte(`{"error": "init data not found"}`)
	}
	LoggerInfo("Init data loaded", []string{"size", fmt.Sprintf("%d", len(initDataBz))})

	var initData InitGenesisRequest
	if err := json.Unmarshal(initDataBz, &initData); err != nil {
		LoggerError("Failed to unmarshal init data", []string{"error", err.Error()})
		return []byte(`{"error": "failed to unmarshal init data"}`)
	}

	// Register initial contracts if provided
	if len(initData.InitialContracts) > 0 {
		for _, contract := range initData.InitialContracts {
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

			// Update HTTP routing (now we have roles to call other contracts)
			updateHttpRoute(contract.RoutePrefix, contract.ContractAddress)

			LoggerInfo("Initial MCP contract registered", []string{
				"contract_address", contract.ContractAddress,
				"route_prefix", contract.RoutePrefix,
				"tool_count", fmt.Sprintf("%d", len(tools)),
			})
		}
	}

	// Register HTTP routes with HTTP registry
	RegisterHttpRoutes(&RegisterHttpRoutesRequest{})

	LoggerInfo("MCP Registry initialized on RoleChanged", nil)
	return []byte(`{"success": true}`)
}

func GetParams(req GetParamsRequest) []byte {
	paramsBz := wasmx.StorageLoad([]byte(STORAGE_PARAMS))
	if len(paramsBz) == 0 {
		return []byte(`{"error": "params not initialized"}`)
	}
	return paramsBz
}

func RegisterHttpRoutes(req *RegisterHttpRoutesRequest) []byte {
	self := string(wasmx.GetAddress())
	LoggerInfo("RegisterHttpRoutes called", []string{"self", self})

	// Get HTTP registry address by role
	httpRegistryAddr := wasmx.GetAddressByRole(wasmx.ROLE_HTTP_SERVER)
	LoggerInfo("HTTP registry address", []string{"addr", string(httpRegistryAddr)})

	routes := []string{
		"/sse",
	}

	for _, rt := range routes {
		LoggerInfo("Registering route", []string{"route", rt, "contract", self})
		msg := map[string]interface{}{
			"set_route": map[string]interface{}{
				"route":            rt,
				"contract_address": self,
			},
		}
		bz, _ := json.Marshal(msg)
		ok, data := wasmx.CallSimple(httpRegistryAddr, bz, false, MODULE_NAME)
		if !ok {
			LoggerError("failed to set route", []string{"route", rt, "error", string(data)})
		} else {
			LoggerInfo("Successfully registered route", []string{"route", rt})
		}
	}

	LoggerInfo("MCP Registry HTTP routes registered", []string{"registry", string(httpRegistryAddr)})
	return []byte(`{"success": true}`)
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

func extractToken(req *HttpRequestIncoming) string {
	// Try Authorization header first (OAuth Bearer token)
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

func validateToken(token string) (string, bool) {
	if token == "" {
		return "", false
	}

	// Validate token via OAuth2 server contract
	userID, err := callOAuth2ValidateAccessToken(token)
	if err != nil {
		return "", false
	}

	return userID, true
}

func HandleHttpRequest(req HttpRequestIncoming) []byte {
	// Extract path from Url or RequestURI (remove query string)
	path := req.Url
	if path == "" {
		path = req.RequestURI
	}
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Route the request based on the URL path
	var resp HttpResponseWrap
	switch path {
	case "/":
		resp = handleRoot(&req)
	case "/sse":
		resp = handleSSE(&req)
	default:
		resp = HttpResponseWrap{
			Error: "",
			Data: HttpResponse{
				StatusCode: 404,
				Status:     "404 Not Found",
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Data:       []byte("Not Found"),
			},
		}
	}

	// Marshal and return
	respBz, _ := json.Marshal(resp)
	return respBz
}

func handleRoot(req *HttpRequestIncoming) HttpResponseWrap {
	response := map[string]interface{}{
		"message": "MCP Registry Server",
		"version": "1.0.0",
	}

	return sendJSONResponse(response, 200)
}

func handleSSE(req *HttpRequestIncoming) HttpResponseWrap {
	// Extract token from Authorization header
	token := extractToken(req)
	if token == "" {
		return HttpResponseWrap{
			Error: "",
			Data: HttpResponse{
				StatusCode: 401,
				Status:     "401 Unauthorized",
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Data:       []byte("Unauthorized: missing token"),
			},
		}
	}

	// Validate token and get user_id
	userID, valid := validateToken(token)
	if !valid {
		return HttpResponseWrap{
			Error: "",
			Data: HttpResponse{
				StatusCode: 401,
				Status:     "401 Unauthorized",
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Data:       []byte("Unauthorized: invalid token"),
			},
		}
	}

	// Get request body
	body := req.Data

	// Parse JSON-RPC request
	var rpcReq JSONRPCRequest
	if err := json.Unmarshal(body, &rpcReq); err != nil {
		return sendJSONRPCError(nil, -32700, "Parse error: "+err.Error())
	}

	// Route based on JSON-RPC method
	switch rpcReq.Method {
	case "initialize":
		return handleInitialize(&rpcReq, userID)
	case "tools/list":
		return handleToolsList(&rpcReq, userID)
	case "tools/call":
		return handleToolsCall(&rpcReq, userID)
	default:
		return sendJSONRPCError(rpcReq.ID, -32601, "Method not found: "+rpcReq.Method)
	}
}

func sendJSONResponse(data interface{}, statusCode int) HttpResponseWrap {
	jsonData, _ := json.Marshal(data)
	return HttpResponseWrap{
		Error: "",
		Data: HttpResponse{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d OK", statusCode),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

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
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

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
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       jsonData,
		},
	}
}

func handleInitialize(req *JSONRPCRequest, userID string) HttpResponseWrap {
	// MCP initialize response
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "MCP Registry Server",
			"version": "1.0.0",
		},
	}

	return sendJSONRPCResponse(req.ID, result)
}

func handleToolsList(req *JSONRPCRequest, userID string) HttpResponseWrap {
	// Get all registered contracts
	addresses := getRegisteredList()
	var allTools []ToolsListEntry

	for _, addr := range addresses {
		registration := getContractRegistration(addr)
		if registration == nil || !registration.Active {
			continue
		}

		// Add tools from this contract
		for _, tool := range registration.Tools {
			allTools = append(allTools, ToolsListEntry{
				Name:            tool.Name,
				Description:     tool.Description,
				InputSchema:     tool.InputSchema,
				ContractAddress: registration.Address,
				RoutePrefix:     registration.RoutePrefix,
			})
		}
	}

	result := map[string]interface{}{
		"tools": allTools,
	}

	return sendJSONRPCResponse(req.ID, result)
}

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

	// Find the contract that provides this tool
	addresses := getRegisteredList()
	var targetContract *ContractRegistration

	for _, addr := range addresses {
		registration := getContractRegistration(addr)
		if registration == nil || !registration.Active {
			continue
		}

		for _, tool := range registration.Tools {
			if tool.Name == toolName {
				targetContract = registration
				break
			}
		}
		if targetContract != nil {
			break
		}
	}

	if targetContract == nil {
		return sendJSONRPCError(req.ID, -32602, "Tool not found: "+toolName)
	}

	// Call the contract that provides this tool
	executeMsg := map[string]interface{}{
		"execute_tool": map[string]interface{}{
			"tool_name":  toolName,
			"arguments":  arguments,
			"user_id":    userID,
		},
	}
	msgBz, _ := json.Marshal(executeMsg)
	ok, data := wasmx.CallSimple(wasmx.Bech32String(targetContract.Address), msgBz, false, MODULE_NAME)
	if !ok {
		return sendJSONRPCError(req.ID, -32603, "Tool execution failed: "+string(data))
	}

	// Parse the tool result
	var toolResult struct {
		Content []map[string]interface{} `json:"content"`
		IsError bool                     `json:"isError,omitempty"`
	}
	if err := json.Unmarshal(data, &toolResult); err != nil {
		return sendJSONRPCError(req.ID, -32603, "Failed to parse tool result: "+err.Error())
	}

	// Return the result
	result := map[string]interface{}{
		"content": toolResult.Content,
	}

	if toolResult.IsError {
		result["isError"] = true
	}

	return sendJSONRPCResponse(req.ID, result)
}
