package lib

import (
	"encoding/json"
	"fmt"
	"strings"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	postgresql "github.com/loredanacirstea/wasmx-env-postgresql/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
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

	// Parse limit - use from arguments or fall back to SearchConfig.DefaultLimit
	searchConfig := loadSearchConfig()
	if searchConfig == nil {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Search configuration not found"}},
			IsError: true,
		}
	}

	limit := searchConfig.DefaultLimit
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

	// Convert []float32 to []interface{} for vectorSearch
	queryEmbeddingInterface := make([]interface{}, len(queryEmbedding))
	for i, v := range queryEmbedding {
		queryEmbeddingInterface[i] = v
	}

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

	// Load script path from init data environment variables
	initData := loadInitData()
	scriptsPath := initData.EnvironmentVars["SCRIPTS_FOLDER"]
	if scriptsPath == "" {
		// Fallback to default path if not configured
		scriptsPath = "./tests/testdata/tinygo/wasmx-mcp-search/scripts"
	}
	scriptPath := scriptsPath + "/generate_embedding_local.py"

	LoggerInfo("About to execute embedding script", []string{"script", scriptPath, "query", query})

	result, err := wasmxcore.ExecuteCliCommand("python3", []string{scriptPath, query}, "")

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
	var embedding []float32
	if err := json.Unmarshal([]byte(output), &embedding); err != nil {
		LoggerError("Failed to parse embedding JSON", []string{"error", err.Error(), "output", output})
		return nil
	}

	LoggerInfo("Successfully generated embedding", []string{"dimensions", fmt.Sprintf("%d", len(embedding))})
	return embedding
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// loadSearchConfig loads the search configuration from storage
func loadSearchConfig() *SearchConfig {
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		return nil
	}

	var initData InitGenesisRequest
	if err := json.Unmarshal(initDataBz, &initData); err != nil {
		return nil
	}

	return &initData.SearchConfig
}

// ensureDatabaseConnection ensures a database connection exists, creating one if needed
func ensureDatabaseConnection() string {
	// Load search configuration
	searchConfig := loadSearchConfig()
	if searchConfig == nil {
		LoggerError("Search config not found, cannot establish connection", nil)
		return ""
	}

	// Set default metric if not specified
	metric := searchConfig.EmbeddingMetric
	if metric == "" {
		metric = "cosine"
	}

	// Establish PostgreSQL connection with vector embeddings support
	options := map[string]interface{}{
		"enable_embeddings":   true,
		"embedding_dimension": searchConfig.EmbeddingDimension,
		"embedding_metric":    metric,
		"maxconns":            50,
		"minconns":            5,
	}

	connReq := &postgresql.SqlConnectionRequest{
		Connection: searchConfig.Database.ConnectionString,
		DbName:     searchConfig.Database.DatabaseName,
		Id:         "mcp_search_main",
		Options:    options,
	}

	connResp := postgresql.Connect(connReq)

	if connResp.Error != "" {
		LoggerError("Failed to connect to PostgreSQL", []string{"error", connResp.Error})
		return ""
	}

	// Connection ID is always the same
	connID := "mcp_search_main"

	LoggerInfo("Database connection established", []string{"conn_id", connID})
	return connID
}

// vectorSearch performs similarity search using vector embeddings across all configured tables
func vectorSearch(userID string, arguments map[string]interface{}) ExecuteToolResponse {
	LoggerInfo("vectorSearch called", []string{"user_id", userID})

	// Parse query embedding
	queryEmbeddingRaw, ok := arguments["query_embedding"].([]interface{})
	if !ok {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'query_embedding' argument"}},
			IsError: true,
		}
	}

	// Convert to float32 slice
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
			return ExecuteToolResponse{
				Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Invalid embedding value at index %d", i)}},
				IsError: true,
			}
		}
	}

	// Load search configuration
	searchConfig := loadSearchConfig()
	if searchConfig == nil {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Search configuration not found"}},
			IsError: true,
		}
	}

	// Parse limit - use from arguments or fall back to SearchConfig.DefaultLimit
	limit := searchConfig.DefaultLimit
	if limitRaw, ok := arguments["limit"]; ok {
		switch v := limitRaw.(type) {
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
	}

	// Get or establish database connection
	connID := ensureDatabaseConnection()
	if connID == "" {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Failed to establish database connection"}},
			IsError: true,
		}
	}

	// Search across all configured tables
	allResults := make([]SearchResult, 0)

	for _, tableConfig := range searchConfig.Tables {
		// Build query for this table
		query := fmt.Sprintf(`
			SELECT
				%s as id,
				%s as cache_text,
				%s <=> $1::vector AS distance
			FROM %s
			ORDER BY %s <=> $1::vector
			LIMIT $2
		`,
			tableConfig.IDColumn,
			tableConfig.CacheTextColumn,
			tableConfig.EmbeddingColumn,
			tableConfig.TableName,
			tableConfig.EmbeddingColumn,
		)

		// Execute query
		embeddingJSON, _ := json.Marshal(queryEmbedding)
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

		queryResp := postgresql.Query(queryReq)
		if queryResp.Error != "" {
			LoggerError("Error querying table", []string{"table", tableConfig.TableName, "error", queryResp.Error})
			continue
		}

		// Parse results
		var tableResults []map[string]interface{}
		if len(queryResp.Data) > 0 {
			if err := json.Unmarshal(queryResp.Data, &tableResults); err != nil {
				LoggerError("Error parsing results", []string{"table", tableConfig.TableName, "error", err.Error()})
				continue
			}
		}

		// Add to all results with source table
		for _, result := range tableResults {
			allResults = append(allResults, SearchResult{
				Source:    tableConfig.TableName,
				ID:        fmt.Sprintf("%v", result["id"]),
				CacheText: fmt.Sprintf("%v", result["cache_text"]),
				Distance:  float32(result["distance"].(float64)),
			})
		}

		LoggerInfo("Search results", []string{"table", tableConfig.TableName, "count", fmt.Sprintf("%d", len(tableResults))})
	}

	// Sort all results by distance
	sortResultsByDistance(allResults)

	// Limit to top N overall
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	// Format response
	return formatSearchResults(allResults)
}

// SearchResult represents a single search result
type SearchResult struct {
	Source    string  `json:"source"`     // Table name
	ID        string  `json:"id"`         // Record ID
	CacheText string  `json:"cache_text"` // Human readable text
	Distance  float32 `json:"distance"`   // Similarity distance
}

// sortResultsByDistance sorts results by distance (ascending)
func sortResultsByDistance(results []SearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Distance < results[i].Distance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// formatSearchResults formats search results for display
func formatSearchResults(results []SearchResult) ExecuteToolResponse {
	if len(results) == 0 {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "No results found"}},
			IsError: false,
		}
	}

	// Build simple list of results with just the cache_text (JSON data)
	var responseText strings.Builder
	responseText.WriteString(fmt.Sprintf("Found %d results:\n\n", len(results)))

	for i, item := range results {
		responseText.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, item.CacheText))
	}

	return ExecuteToolResponse{
		Content: []ContentItem{{Type: "text", Text: responseText.String()}},
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
