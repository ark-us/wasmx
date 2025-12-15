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

	LoggerInfo("MCP Registry init data stored", []string{
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
				Address:        contract.ContractAddress,
				RoutePrefix:    contract.RoutePrefix,
				Tools:          tools,
				RegisteredAt:   int64(currentBlock.Height),
				LastUpdatedAt:  int64(currentBlock.Height),
				Active:         true,
				UseOAuth2:      contract.UseOAuth2,
				UseTransaction: contract.UseTransaction,
				SubPaths:       contract.SubPaths,
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

func RegisterHttpRoutes(req *RegisterHttpRoutesRequest) []byte {
	self := string(wasmx.GetAddress())
	LoggerInfo("RegisterHttpRoutes called", []string{"self", self})

	// Get HTTP registry address by role
	httpRegistryAddr := wasmx.GetAddressByRole(wasmx.ROLE_HTTP_SERVER)
	LoggerInfo("HTTP registry address", []string{"addr", string(httpRegistryAddr)})

	routes := []string{
		"/sse",
		"/.well-known/ai-plugin.json",
		"/openapi.json",
		"/tools",
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
	fmt.Println("* mcpregistry", req.Url)
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
	switch {
	case path == "/":
		resp = handleRoot(&req)
	case path == "/sse":
		resp = handleSSE(&req)
	case path == "/.well-known/ai-plugin.json":
		resp = handleAIPluginManifest(&req)
	case path == "/openapi.json":
		resp = handleOpenAPISpec(&req)
	case strings.HasPrefix(path, "/tools/"):
		resp = handleToolExecution(&req, path)
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
	// Extract token from Authorization header (optional - will be checked per-tool)
	token := extractToken(req)
	var userID string
	if token != "" {
		// Validate token if provided
		var valid bool
		userID, valid = validateToken(token)
		if !valid {
			// Invalid token provided
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
	}
	// If no token provided, userID will be empty string

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
			"name":    "WasmX Provable MCP Registry Server",
			"version": "1.0.0",
		},
	}

	return sendJSONRPCResponse(req.ID, result)
}

func handleToolsList(req *JSONRPCRequest, userID string) HttpResponseWrap {
	// Get all registered contracts
	addresses := getRegisteredList()

	LoggerInfo("handleToolsList called", []string{
		"user_id", userID,
		"registered_count", fmt.Sprintf("%d", len(addresses)),
	})

	var allTools []ToolsListEntry

	for i, addr := range addresses {
		fmt.Printf("Processing address %d: %s\n", i, addr)
		registration := getContractRegistration(addr)
		if registration == nil {
			LoggerInfo("Registration not found", []string{"address", addr})
			continue
		}
		fmt.Printf("  -> Active: %v, Tools count: %d\n", registration.Active, len(registration.Tools))

		if !registration.Active {
			LoggerInfo("Registration inactive", []string{"address", addr})
			continue
		}

		LoggerInfo("Found active registration", []string{
			"address", addr,
			"route_prefix", registration.RoutePrefix,
			"tool_count", fmt.Sprintf("%d", len(registration.Tools)),
		})

		// Add tools from this contract
		for j, tool := range registration.Tools {
			fmt.Printf("  -> Tool %d: %s\n", j, tool.Name)
			allTools = append(allTools, ToolsListEntry{
				Name:            tool.Name,
				Description:     tool.Description,
				InputSchema:     tool.InputSchema,
				ContractAddress: registration.Address,
				RoutePrefix:     registration.RoutePrefix,
			})
		}
	}

	LoggerInfo("Returning tools list", []string{
		"total_tools", fmt.Sprintf("%d", len(allTools)),
	})

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

	// OAuth2 authentication is handled by HTTP server based on route's UseOAuth2 field
	// No need to check here - just pass user_id to the tool contract

	// Call the contract that provides this tool
	executeMsg := map[string]interface{}{
		"execute_tool": map[string]interface{}{
			"tool_name": toolName,
			"arguments": arguments,
			"user_id":   userID,
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

func handleAIPluginManifest(req *HttpRequestIncoming) HttpResponseWrap {
	// Get the base URL from headers
	scheme := "http"
	if len(req.Header["X-Forwarded-Proto"]) > 0 && strings.Contains(req.Header["X-Forwarded-Proto"][0], "https") {
		scheme = "https"
	}
	host := ""
	if len(req.Header["X-Forwarded-Host"]) > 0 {
		host = req.Header["X-Forwarded-Host"][0]
		// Handle comma-separated hosts (take first one)
		if idx := strings.Index(host, ","); idx != -1 {
			host = strings.TrimSpace(host[:idx])
		}
	}
	if host == "" && len(req.Header["Host"]) > 0 {
		host = req.Header["Host"][0]
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	manifest := map[string]interface{}{
		"schema_version":        "v1",
		"name_for_human":        "MCP Tools Registry",
		"name_for_model":        "mcp_tools",
		"description_for_human": "Access to blockchain-based MCP tools for executing code and managing data",
		"description_for_model": "Provides access to various MCP (Model Context Protocol) tools running on blockchain. Includes code execution, data management, and other utilities.",
		"auth": map[string]interface{}{
			"type":                       "oauth",
			"client_url":                 baseURL + "/oauth/authorize",
			"scope":                      "",
			"authorization_url":          baseURL + "/oauth/token",
			"authorization_content_type": "application/x-www-form-urlencoded",
		},
		"api": map[string]interface{}{
			"type": "openapi",
			"url":  baseURL + "/openapi.json",
		},
		"logo_url":       baseURL + "/logo.png",
		"contact_email":  "contact@mail.provable.dev",
		"legal_info_url": baseURL + "/legal",
	}

	return sendJSONResponse(manifest, 200)
}

func handleOpenAPISpec(req *HttpRequestIncoming) HttpResponseWrap {
	// Get the base URL from headers
	scheme := "http"
	if len(req.Header["X-Forwarded-Proto"]) > 0 && strings.Contains(req.Header["X-Forwarded-Proto"][0], "https") {
		scheme = "https"
	}
	host := ""
	if len(req.Header["X-Forwarded-Host"]) > 0 {
		host = req.Header["X-Forwarded-Host"][0]
		// Handle comma-separated hosts (take first one)
		if idx := strings.Index(host, ","); idx != -1 {
			host = strings.TrimSpace(host[:idx])
		}
	}
	if host == "" && len(req.Header["Host"]) > 0 {
		host = req.Header["Host"][0]
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	// Get all registered tools
	addresses := getRegisteredList()
	paths := make(map[string]interface{})

	for _, addr := range addresses {
		registration := getContractRegistration(addr)
		if registration == nil || !registration.Active {
			continue
		}

		// Create an endpoint for each tool
		for _, tool := range registration.Tools {
			pathKey := fmt.Sprintf("%s/%s", registration.RoutePrefix, tool.Name)

			// Convert JSON Schema properties to OpenAPI parameters
			parameters := []map[string]interface{}{}
			if props, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
				for propName, propSchema := range props {
					param := map[string]interface{}{
						"name":     propName,
						"in":       "query",
						"required": false,
						"schema":   propSchema,
					}
					// Check if this property is required
					if required, ok := tool.InputSchema["required"].([]interface{}); ok {
						for _, req := range required {
							if req.(string) == propName {
								param["required"] = true
								break
							}
						}
					}
					parameters = append(parameters, param)
				}
			}

			paths[pathKey] = map[string]interface{}{
				"post": map[string]interface{}{
					"operationId": tool.Name,
					"summary":     tool.Description,
					"description": tool.Description,
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": tool.InputSchema,
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"content": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"type": "object",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
		}
	}

	openapi := map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":       "Provable MCP Registry",
			"description": "API for accessing MCP tools, built on WASM-based blockchain tech wasmx.provable.dev, by provable.dev",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{
			{
				"url": baseURL,
			},
		},
		"paths": paths,
	}

	return sendJSONResponse(openapi, 200)
}

func handleToolExecution(req *HttpRequestIncoming, path string) HttpResponseWrap {
	// Extract tool name from path: /tools/{toolName}
	toolName := strings.TrimPrefix(path, "/tools/")
	if toolName == "" {
		return HttpResponseWrap{
			Error: "",
			Data: HttpResponse{
				StatusCode: 400,
				Status:     "400 Bad Request",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Data:       []byte(`{"error":"missing tool name"}`),
			},
		}
	}

	// Parse request body as tool arguments
	var arguments map[string]interface{}
	if len(req.Data) > 0 {
		if err := json.Unmarshal(req.Data, &arguments); err != nil {
			return HttpResponseWrap{
				Error: "",
				Data: HttpResponse{
					StatusCode: 400,
					Status:     "400 Bad Request",
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Data:       []byte(`{"error":"invalid JSON body"}`),
				},
			}
		}
	} else {
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
		return HttpResponseWrap{
			Error: "",
			Data: HttpResponse{
				StatusCode: 404,
				Status:     "404 Not Found",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Data:       []byte(fmt.Sprintf(`{"error":"tool not found: %s"}`, toolName)),
			},
		}
	}

	// OAuth2 authentication is handled by HTTP server based on route's UseOAuth2 field
	// Extract user_id from token if provided (HTTP server already validated it)
	var userID string
	token := extractToken(req)
	if token != "" {
		// Token was already validated by HTTP server, just extract user_id
		userID, _ = validateToken(token)
	}

	// Call the contract that provides this tool
	executeMsg := map[string]interface{}{
		"execute_tool": map[string]interface{}{
			"tool_name": toolName,
			"arguments": arguments,
			"user_id":   userID,
		},
	}
	msgBz, _ := json.Marshal(executeMsg)
	ok, data := wasmx.CallSimple(wasmx.Bech32String(targetContract.Address), msgBz, false, MODULE_NAME)
	if !ok {
		return HttpResponseWrap{
			Error: "",
			Data: HttpResponse{
				StatusCode: 500,
				Status:     "500 Internal Server Error",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Data:       []byte(fmt.Sprintf(`{"error":"tool execution failed: %s"}`, string(data))),
			},
		}
	}

	// Return the tool result directly
	return HttpResponseWrap{
		Error: "",
		Data: HttpResponse{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       data,
		},
	}
}
