package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// getToolDefinitions returns the hardcoded tool definitions for this contract
func getToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "execute_py",
			"description": "Execute the Python hello.py script that prints 'Hello world: ' + random number",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "execute_cli",
			"description": "Execute a CLI command with optional arguments and stdin",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The command to execute",
					},
					"args": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Command arguments",
					},
					"stdin": map[string]interface{}{
						"type":        "string",
						"description": "Standard input for the command",
					},
				},
				"required": []string{"command"},
			},
		},
	}
}

// InitGenesis stores initialization data during instantiation
func InitGenesis(req InitGenesisRequest) []byte {
	// Store init data for later use in RoleChanged
	initDataBz, _ := json.Marshal(req)
	wasmx.StorageStore([]byte(STORAGE_INIT_DATA), initDataBz)

	LoggerInfo("MCP Execute init data stored", []string{
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

	// Define subpaths with per-path configuration
	subpaths := []map[string]interface{}{
		{
			"path":            "/execute_py",
			"use_oauth2":      true,
			"use_transaction": false, // POST but read-only call, no transaction needed
		},
		{
			"path":            "/execute_cli",
			"use_oauth2":      true,
			"use_transaction": false, // POST but read-only call, no transaction needed
		},
	}

	// Register with MCP registry
	registerMsg := map[string]interface{}{
		"register_mcp_contract": map[string]interface{}{
			"contract_address": string(wasmx.GetAddress()),
			"route_prefix":     initData.RoutePrefix,
			"tools_json":       string(toolsJSON),
			"subpaths":         subpaths,
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

// loadInitData loads the stored initialization data
func loadInitData() InitGenesisRequest {
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		return InitGenesisRequest{}
	}

	var initData InitGenesisRequest
	json.Unmarshal(initDataBz, &initData)
	return initData
}
