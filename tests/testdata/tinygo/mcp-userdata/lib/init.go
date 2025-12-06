package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// getToolDefinitions returns the hardcoded tool definitions for this contract
func getToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "set_favorite_color",
			"description": "Set the user's favorite color",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"color": map[string]interface{}{
						"type":        "string",
						"description": "The favorite color to set",
					},
				},
				"required": []string{"color"},
			},
		},
		{
			"name":        "get_favorite_color",
			"description": "Get the user's favorite color",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "list_items",
			"description": "List all items for the user",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// InitGenesis stores initialization data during instantiation
func InitGenesis(req InitGenesisRequest) []byte {
	// Store init data for later use in RoleChanged
	initDataBz, _ := json.Marshal(req)
	wasmx.StorageStore([]byte(STORAGE_INIT_DATA), initDataBz)

	LoggerInfo("MCP Userdata init data stored", []string{
		"route_prefix", req.RoutePrefix,
	})

	return []byte(`{"success": true}`)
}

// OnRoleChanged registers this contract with the MCP registry when role is assigned
func OnRoleChanged() []byte {
	// Load stored init data
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		LoggerError("Init data not found", nil)
		return []byte(`{"error": "init data not found"}`)
	}

	var initData InitGenesisRequest
	if err := json.Unmarshal(initDataBz, &initData); err != nil {
		LoggerError("Failed to unmarshal init data", []string{"error", err.Error()})
		return []byte(`{"error": "failed to unmarshal init data"}`)
	}

	// Get MCP registry address
	mcpRegistryAddr := wasmx.GetAddressByRole(wasmx.ROLE_MCP_REGISTRY)
	if mcpRegistryAddr == "" {
		LoggerError("MCP registry not found", nil)
		return []byte(`{"error": "mcp registry not found"}`)
	}

	// Get hardcoded tool definitions
	tools := getToolDefinitions()
	toolsJSON, _ := json.Marshal(tools)

	// Register with MCP registry
	registerMsg := map[string]interface{}{
		"register_mcp_contract": map[string]interface{}{
			"contract_address": string(wasmx.GetAddress()),
			"route_prefix":     initData.RoutePrefix,
			"tools_json":       string(toolsJSON),
			"use_oauth2":       true,
		},
	}
	msgBz, _ := json.Marshal(registerMsg)
	ok, data := wasmx.CallSimple(mcpRegistryAddr, msgBz, false, MODULE_NAME)
	if !ok {
		LoggerError("Failed to register with MCP registry", []string{"error", string(data)})
		return []byte(`{"error": "failed to register"}`)
	}

	LoggerInfo("Registered with MCP registry", []string{
		"route_prefix", initData.RoutePrefix,
	})

	return []byte(`{"success": true}`)
}
