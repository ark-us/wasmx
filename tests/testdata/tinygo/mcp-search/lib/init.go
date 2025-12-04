package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	postgresql "github.com/loredanacirstea/wasmx-env-postgresql/lib"
)

// getToolDefinitions returns the hardcoded tool definitions for this contract
func getToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "search_knowledge",
			"description": "Search for information in the knowledge base using natural language queries. Returns relevant information about blockchain, smart contracts, consensus mechanisms, cryptography, and DeFi.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query in natural language (e.g., 'blockchain technology', 'smart contracts')",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 5)",
						"default":     5,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "vector_search",
			"description": "Advanced search using pre-computed vector embeddings. For direct embedding-based similarity search.",
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
						"description": "Maximum number of results to return (default: 10)",
						"default":     10,
					},
				},
				"required": []string{"query_embedding"},
			},
		},
		{
			"name":        "store_embedding",
			"description": "Store a key-value pair with its vector embedding for future similarity search",
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
		},
	}
}

// InitGenesis stores initialization data and establishes database connection
func InitGenesis(req InitGenesisRequest) []byte {
	// Set defaults
	if req.EmbeddingDimension == 0 {
		req.EmbeddingDimension = 1536 // OpenAI default
	}
	if req.EmbeddingMetric == "" {
		req.EmbeddingMetric = "cosine"
	}
	if req.DatabaseName == "" {
		req.DatabaseName = "mcp_search"
	}
	if req.ConnectionString == "" {
		req.ConnectionString = "postgresql://localhost:5432/postgres"
	}

	// Store init data for later use in RoleChanged
	initDataBz, _ := json.Marshal(req)
	wasmx.StorageStore([]byte(STORAGE_INIT_DATA), initDataBz)

	// Establish PostgreSQL connection with vector embeddings support
	options := map[string]interface{}{
		"enable_embeddings":   true,
		"embedding_dimension": req.EmbeddingDimension,
		"embedding_metric":    req.EmbeddingMetric,
		"maxconns":            50,
		"minconns":            5,
	}

	connReq := &postgresql.SqlConnectionRequest{
		Connection: req.ConnectionString,
		DbName:     req.DatabaseName,
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
		"db_name", req.DatabaseName,
		"embedding_dimension", string(rune(req.EmbeddingDimension)),
		"embedding_metric", req.EmbeddingMetric,
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
