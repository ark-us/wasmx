package lib

import (
	"encoding/json"
	"fmt"

	postgresql "github.com/loredanacirstea/wasmx-env-postgresql/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// getToolDefinitions returns the tool definitions using stored descriptions
func getToolDefinitions() []map[string]interface{} {
	// Load init data to get tool descriptions
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		// Return empty if not initialized
		return []map[string]interface{}{}
	}

	var initData InitGenesisRequest
	if err := json.Unmarshal(initDataBz, &initData); err != nil {
		return []map[string]interface{}{}
	}

	tools := []map[string]interface{}{
		{
			"name":        "search_knowledge",
			"description": initData.ToolDescriptions.SearchKnowledge.Description,
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": initData.ToolDescriptions.SearchKnowledge.QueryDescription,
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": fmt.Sprintf("Maximum number of results to return (default: %d)", initData.SearchConfig.DefaultLimit),
						"default":     initData.SearchConfig.DefaultLimit,
					},
				},
				"required": []string{"query"},
			},
		},
	}

	// Add vector_search if description provided
	if initData.ToolDescriptions.VectorSearch.Description != "" {
		tools = append(tools, map[string]interface{}{
			"name":        "vector_search",
			"description": initData.ToolDescriptions.VectorSearch.Description,
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query_embedding": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "number"},
						"description": "The query vector embedding to search for similar items",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": fmt.Sprintf("Maximum number of results to return (default: %d)", initData.SearchConfig.DefaultLimit),
						"default":     initData.SearchConfig.DefaultLimit,
					},
				},
				"required": []string{"query_embedding"},
			},
		})
	}

	// Add store_embedding if description provided
	if initData.ToolDescriptions.StoreEmbedding.Description != "" {
		tools = append(tools, map[string]interface{}{
			"name":        "store_embedding",
			"description": initData.ToolDescriptions.StoreEmbedding.Description,
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "The unique key for this item",
					},
					"value": map[string]interface{}{
						"type":        "string",
						"description": "The value/content to store",
					},
					"embedding": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "number"},
						"description": "The vector embedding for this item",
					},
				},
				"required": []string{"key", "value", "embedding"},
			},
		})
	}

	return tools
}

// InitGenesis stores initialization data and establishes database connection
func InitGenesis(req InitGenesisRequest) []byte {
	// Set defaults in search config
	if req.SearchConfig.EmbeddingDimension == 0 {
		req.SearchConfig.EmbeddingDimension = 768 // sentence-transformers/all-mpnet-base-v2
	}
	if req.SearchConfig.EmbeddingMetric == "" {
		req.SearchConfig.EmbeddingMetric = "cosine"
	}
	if req.SearchConfig.Database.DatabaseName == "" {
		req.SearchConfig.Database.DatabaseName = "mcp_search"
	}
	if req.SearchConfig.Database.ConnectionString == "" {
		req.SearchConfig.Database.ConnectionString = "postgresql://localhost:5432/postgres"
	}
	if req.SearchConfig.DefaultLimit == 0 {
		req.SearchConfig.DefaultLimit = 10
	}

	// Store init data for later use in RoleChanged
	initDataBz, _ := json.Marshal(req)
	wasmx.StorageStore([]byte(STORAGE_INIT_DATA), initDataBz)

	// Establish PostgreSQL connection with vector embeddings support
	options := map[string]interface{}{
		"enable_embeddings":   true,
		"embedding_dimension": req.SearchConfig.EmbeddingDimension,
		"embedding_metric":    req.SearchConfig.EmbeddingMetric,
		"maxconns":            50,
		"minconns":            5,
	}

	connReq := &postgresql.SqlConnectionRequest{
		Connection: req.SearchConfig.Database.ConnectionString,
		DbName:     req.SearchConfig.Database.DatabaseName,
		Id:         "mcp_search_main",
		Options:    options,
	}

	connResp := postgresql.Connect(connReq)
	if connResp.Error != "" {
		LoggerError("Failed to connect to PostgreSQL", []string{"error", connResp.Error})
		return []byte(`{"error": "failed to connect to database"}`)
	}

	// Store connection ID
	wasmx.StorageStore([]byte(STORAGE_DB_CONN), []byte("mcp_search_main"))

	LoggerInfo("MCP Search initialized", []string{
		"route_prefix", req.RoutePrefix,
		"db_name", req.SearchConfig.Database.DatabaseName,
		"tables_count", string(rune(len(req.SearchConfig.Tables))),
		"embedding_dimension", string(rune(req.SearchConfig.EmbeddingDimension)),
		"embedding_metric", req.SearchConfig.EmbeddingMetric,
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
			"use_oauth2":       false,
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
