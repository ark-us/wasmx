package lib

import (
	"encoding/json"
	"fmt"
	"strings"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	postgresql "github.com/loredanacirstea/wasmx-env-postgresql/lib"
)

// ExecuteTool handles tool execution requests
func ExecuteTool(req ExecuteToolRequest) []byte {
	var response ExecuteToolResponse

	switch req.ToolName {
	case "search_knowledge":
		response = searchKnowledge(req.UserID, req.Arguments)
	case "vector_search":
		response = vectorSearch(req.UserID, req.Arguments)
	case "store_embedding":
		response = storeEmbedding(req.UserID, req.Arguments)
	default:
		response = ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Unknown tool: " + req.ToolName}},
			IsError: true,
		}
	}

	data, _ := json.Marshal(response)
	return data
}

// searchKnowledge performs text-based search by generating embeddings from the query
func searchKnowledge(userID string, arguments map[string]interface{}) ExecuteToolResponse {
	LoggerInfo("searchKnowledge called", []string{"user_id", userID})

	// Parse query text
	query, ok := arguments["query"].(string)
	if !ok || query == "" {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'query' argument"}},
			IsError: true,
		}
	}

	// Parse limit
	limit := 5
	if limitRaw, ok := arguments["limit"]; ok {
		switch v := limitRaw.(type) {
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
	}

	LoggerInfo("searchKnowledge query", []string{"query", query, "limit", fmt.Sprintf("%d", limit)})

	// Generate embedding from the query text using local model via Python script
	queryEmbedding := generateQueryEmbedding(query)
	if queryEmbedding == nil {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Failed to generate embedding for query. Please check that Python dependencies are installed (pip install -r requirements.txt)."}},
			IsError: true,
		}
	}

	fmt.Println("=== Converting embedding to interface{} slice ===")
	// Convert []float32 to []interface{} for vectorSearch
	queryEmbeddingInterface := make([]interface{}, len(queryEmbedding))
	for i, v := range queryEmbedding {
		queryEmbeddingInterface[i] = v
	}
	fmt.Println("Conversion complete, length:", len(queryEmbeddingInterface))

	// Now perform vector search with the generated embedding
	vectorArgs := map[string]interface{}{
		"query_embedding": queryEmbeddingInterface,
		"limit":           limit,
	}

	return vectorSearch(userID, vectorArgs)
}

// generateQueryEmbedding generates embeddings using OpenAI API via Python script
func generateQueryEmbedding(query string) []float32 {
	LoggerInfo("Generating embedding for query", []string{"query", query})

	// Get the absolute path to the embedding script
	// In production, this should be configurable via InitGenesis
	// Using local model (no API key required)
	scriptPath := "/Users/user/dev/blockchain/wasmx/tests/testdata/tinygo/mcp-search/scripts/generate_embedding_local.py"

	// Just run the Python script directly - assume dependencies are already installed
	// User should run: pip install -r requirements.txt before starting the test
	cmdStr := fmt.Sprintf("python3 %s %s", scriptPath, query)
	fmt.Println("=== About to execute embedding command ===")
	fmt.Println("Command:", cmdStr)
	LoggerInfo("About to execute embedding script", []string{"script", scriptPath, "query", query, "full_command", cmdStr})

	result, err := wasmxcore.ExecuteCliCommand("python3", []string{scriptPath, query}, "")

	fmt.Println("=== Command execution completed ===")
	fmt.Println("Has error:", err != nil)
	if result != nil {
		fmt.Println("Exit code:", result.ExitCode)
		fmt.Println("Stdout length:", len(result.Stdout))
		fmt.Println("Stderr length:", len(result.Stderr))
	}
	LoggerInfo("Script execution completed", []string{"has_error", fmt.Sprintf("%v", err != nil)})

	if err != nil {
		LoggerError("Failed to execute embedding script", []string{"error", err.Error()})
		return nil
	}

	if result.ExitCode != 0 {
		LoggerError("Embedding script failed", []string{
			"exit_code", fmt.Sprintf("%d", result.ExitCode),
			"stderr", result.Stderr,
			"stdout", result.Stdout,
		})
		return nil
	}

	// Parse the JSON output
	output := strings.TrimSpace(result.Stdout)
	fmt.Println("=== Parsing embedding output ===")
	fmt.Println("Output length:", len(output))
	fmt.Println("First 100 chars:", output[:min(100, len(output))])
	LoggerInfo("Embedding script output", []string{"output_length", fmt.Sprintf("%d", len(output))})

	// Check if output is an error object
	if strings.HasPrefix(output, "{") && strings.Contains(output, "\"error\"") {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal([]byte(output), &errorResp); err == nil {
			LoggerError("Embedding generation error", []string{"error", errorResp.Error})
			return nil
		}
	}

	// Parse embedding array
	fmt.Println("=== About to unmarshal JSON ===")
	var embedding []float32
	if err := json.Unmarshal([]byte(output), &embedding); err != nil {
		fmt.Println("JSON unmarshal error:", err.Error())
		LoggerError("Failed to parse embedding JSON", []string{"error", err.Error(), "output", output})
		return nil
	}

	fmt.Println("=== Embedding parsed successfully ===")
	fmt.Println("Dimensions:", len(embedding))
	LoggerInfo("Successfully generated embedding", []string{"dimensions", fmt.Sprintf("%d", len(embedding))})
	return embedding
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ensureDatabaseConnection ensures a database connection exists, creating one if needed
func ensureDatabaseConnection() string {
	fmt.Println("=== ensureDatabaseConnection called ===")

	// Always try to establish a fresh connection
	// The stored connection ID may not be valid anymore
	fmt.Println("Establishing new database connection...")

	// Load init data to get connection parameters
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		LoggerError("Init data not found, cannot establish connection", nil)
		fmt.Println("ERROR: Init data not found")
		return ""
	}

	var initData InitGenesisRequest
	if err := json.Unmarshal(initDataBz, &initData); err != nil {
		LoggerError("Failed to unmarshal init data", []string{"error", err.Error()})
		fmt.Println("ERROR: Failed to unmarshal init data:", err.Error())
		return ""
	}

	fmt.Println("Init data loaded:", initData.DatabaseName, initData.ConnectionString)

	// Establish PostgreSQL connection with vector embeddings support
	options := map[string]interface{}{
		"enable_embeddings":   true,
		"embedding_dimension": initData.EmbeddingDimension,
		"embedding_metric":    initData.EmbeddingMetric,
		"maxconns":            50,
		"minconns":            5,
	}

	connReq := &postgresql.SqlConnectionRequest{
		Connection: initData.ConnectionString,
		DbName:     initData.DatabaseName,
		Id:         "mcp_search_main",
		Options:    options,
	}

	fmt.Println("Calling postgresql.Connect...")
	connResp := postgresql.Connect(connReq)
	fmt.Println("Connect returned, error:", connResp.Error)

	if connResp.Error != "" {
		LoggerError("Failed to connect to PostgreSQL", []string{"error", connResp.Error})
		fmt.Println("ERROR: Failed to connect:", connResp.Error)
		return ""
	}

	// Connection ID is always the same
	connID := "mcp_search_main"
	fmt.Println("Connection established:", connID)

	LoggerInfo("Database connection established", []string{"conn_id", connID})
	return connID
}

// vectorSearch performs similarity search using vector embeddings
func vectorSearch(userID string, arguments map[string]interface{}) ExecuteToolResponse {
	fmt.Println("=== vectorSearch called ===")
	LoggerInfo("vectorSearch called", []string{"user_id", userID})

	// Parse query embedding
	fmt.Println("=== Parsing query_embedding ===")
	queryEmbeddingRaw, ok := arguments["query_embedding"].([]interface{})
	if !ok {
		fmt.Println("ERROR: query_embedding is not []interface{}")
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'query_embedding' argument"}},
			IsError: true,
		}
	}
	fmt.Println("query_embedding length:", len(queryEmbeddingRaw))

	// Convert to float32 slice
	fmt.Println("=== Converting to float32 slice ===")
	queryEmbedding := make([]float32, len(queryEmbeddingRaw))
	for i, v := range queryEmbeddingRaw {
		switch val := v.(type) {
		case float64:
			queryEmbedding[i] = float32(val)
		case float32:
			queryEmbedding[i] = val
		case int:
			queryEmbedding[i] = float32(val)
		default:
			fmt.Println("ERROR: Invalid embedding value type at index", i)
			return ExecuteToolResponse{
				Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Invalid embedding value at index %d", i)}},
				IsError: true,
			}
		}
	}
	fmt.Println("Conversion complete, float32 slice length:", len(queryEmbedding))

	// Parse limit
	limit := 10
	if limitRaw, ok := arguments["limit"]; ok {
		switch v := limitRaw.(type) {
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
	}
	fmt.Println("Limit:", limit)

	// Get or establish database connection
	fmt.Println("=== Getting database connection ===")
	connID := ensureDatabaseConnection()
	if connID == "" {
		fmt.Println("ERROR: Failed to establish database connection")
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Failed to establish database connection"}},
			IsError: true,
		}
	}
	fmt.Println("Connection ID:", connID)

	// Build similarity search query
	// Using PostgreSQL pgvector similarity operators
	// The query will be constructed by the host-side PostgreSQL adapter
	fmt.Println("=== Building PostgreSQL query ===")
	query := `
		SELECT key, value, embedding <=> $1::vector AS distance
		FROM state_embeddings
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`

	// Encode embedding as JSON for the query parameter
	fmt.Println("=== Encoding embedding as JSON ===")
	embeddingJSON, _ := json.Marshal(queryEmbedding)
	fmt.Println("Embedding JSON length:", len(embeddingJSON))

	embeddingParam, _ := json.Marshal(postgresql.SqlQueryParam{
		Type:  "jsonb",
		Value: string(embeddingJSON),
	})

	limitParam, _ := json.Marshal(postgresql.SqlQueryParam{
		Type:  "",
		Value: limit,
	})

	queryReq := &postgresql.SqlQueryRequest{
		Id:     connID,
		Query:  query,
		Params: postgresql.Params{embeddingParam, limitParam},
	}

	fmt.Println("=== Executing PostgreSQL query ===")
	queryResp := postgresql.Query(queryReq)
	fmt.Println("=== Query completed ===")
	fmt.Println("Error:", queryResp.Error)
	fmt.Println("Data length:", len(queryResp.Data))

	if queryResp.Error != "" {
		fmt.Println("ERROR: Query failed:", queryResp.Error)
		LoggerError("Vector search query failed", []string{"error", queryResp.Error})
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Search failed: " + queryResp.Error}},
			IsError: true,
		}
	}

	// Parse results
	fmt.Println("=== Parsing query results ===")
	var results []map[string]interface{}
	if len(queryResp.Data) > 0 {
		if err := json.Unmarshal(queryResp.Data, &results); err != nil {
			fmt.Println("ERROR: Failed to unmarshal results:", err.Error())
			return ExecuteToolResponse{
				Content: []ContentItem{{Type: "text", Text: "Failed to parse search results"}},
				IsError: true,
			}
		}
	}
	fmt.Println("Results count:", len(results))

	// Format results
	if len(results) == 0 {
		fmt.Println("No results found")
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "No results found"}},
			IsError: false,
		}
	}

	// Build response text
	fmt.Println("=== Building response text ===")
	responseText := fmt.Sprintf("Found %d similar items:\n\n", len(results))
	for i, result := range results {
		key := result["key"]
		value := result["value"]
		distance := result["distance"]

		// Value is now fetched directly from PostgreSQL
		valueStr := ""
		if value != nil {
			valueStr = fmt.Sprintf("%v", value)
		}
		if valueStr == "" {
			valueStr = "(no value stored)"
		}

		responseText += fmt.Sprintf("%d. Key: %v, Distance: %v\n   Value: %s\n", i+1, key, distance, valueStr)
	}

	fmt.Println("=== Returning response ===")
	fmt.Println("Response length:", len(responseText))
	return ExecuteToolResponse{
		Content: []ContentItem{{Type: "text", Text: responseText}},
		IsError: false,
	}
}

// storeEmbedding stores a key-value pair with its embedding
func storeEmbedding(userID string, arguments map[string]interface{}) ExecuteToolResponse {
	LoggerInfo("storeEmbedding called", []string{"user_id", userID})

	// Require authentication for storing data
	if userID == "" {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Authentication required: no user_id provided"}},
			IsError: true,
		}
	}

	// Parse key
	key, ok := arguments["key"].(string)
	if !ok {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'key' argument"}},
			IsError: true,
		}
	}

	// Parse value
	value, ok := arguments["value"].(string)
	if !ok {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'value' argument"}},
			IsError: true,
		}
	}

	// Parse embedding
	embeddingRaw, ok := arguments["embedding"].([]interface{})
	if !ok {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'embedding' argument"}},
			IsError: true,
		}
	}

	// Convert to float32 slice
	embedding := make([]float32, len(embeddingRaw))
	for i, v := range embeddingRaw {
		switch val := v.(type) {
		case float64:
			embedding[i] = float32(val)
		case float32:
			embedding[i] = val
		case int:
			embedding[i] = float32(val)
		default:
			return ExecuteToolResponse{
				Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Invalid embedding value at index %d", i)}},
				IsError: true,
			}
		}
	}

	// Store the value in regular storage
	keyBytes := []byte(key)
	valueBytes := []byte(value)
	wasmx.StorageStore(keyBytes, valueBytes)

	// Get or establish database connection
	connID := ensureDatabaseConnection()
	if connID == "" {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Failed to establish database connection"}},
			IsError: true,
		}
	}

	// Store the embedding in PostgreSQL
	// Using INSERT ... ON CONFLICT DO UPDATE to upsert
	query := `
		INSERT INTO state_embeddings(key, embedding)
		VALUES($1, $2)
		ON CONFLICT(key) DO UPDATE SET embedding = EXCLUDED.embedding
	`

	// Encode embedding as JSON for the query parameter
	embeddingJSON, _ := json.Marshal(embedding)
	embeddingParam, _ := json.Marshal(postgresql.SqlQueryParam{
		Type:  "jsonb",
		Value: string(embeddingJSON),
	})

	keyParam, _ := json.Marshal(postgresql.SqlQueryParam{
		Type:  "bytea",
		Value: key,
	})

	execReq := &postgresql.SqlExecuteRequest{
		Id:     connID,
		Query:  query,
		Params: postgresql.Params{keyParam, embeddingParam},
	}

	execResp := postgresql.Execute(execReq)
	if execResp.Error != "" {
		LoggerError("Failed to store embedding", []string{"error", execResp.Error})
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Failed to store embedding: " + execResp.Error}},
			IsError: true,
		}
	}

	LoggerInfo("Embedding stored successfully", []string{"key", key, "user_id", userID})

	return ExecuteToolResponse{
		Content: []ContentItem{{
			Type: "text",
			Text: fmt.Sprintf("Successfully stored embedding for key '%s'", key),
		}},
		IsError: false,
	}
}
