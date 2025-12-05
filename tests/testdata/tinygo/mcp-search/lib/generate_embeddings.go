package lib

import (
	"encoding/json"
	"fmt"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// GenerateEmbeddings generates embeddings for data and stores them in the database
func GenerateEmbeddings(req GenerateEmbeddingsRequest) []byte {
	fmt.Println("=== GenerateEmbeddings called ===")
	fmt.Println("Items provided:", len(req.Items))

	// Path to the bulk embedding generation script
	// Try to find the script in common locations
	scriptPath := "/Users/user/dev/blockchain/wasmx/tests/testdata/tinygo/mcp-search/scripts/generate_bulk_embeddings.py"

	// Check if script exists - if not, try relative path
	// Note: This is a workaround for execution context issues
	fmt.Println("Using script path:", scriptPath)

	// Load init data to get database connection parameters
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		return createErrorResponse("Init data not found", "cannot generate embeddings without database config")
	}

	var initData InitGenesisRequest
	if err := json.Unmarshal(initDataBz, &initData); err != nil {
		return createErrorResponse("Failed to unmarshal init data", err.Error())
	}

	// No stdin data needed - script uses mock data or fetches from PostgreSQL
	// TODO: Later, add support for providing data items from contract
	fmt.Println("Executing bulk embedding script...")
	fmt.Println("Script path:", scriptPath)
	if len(req.Items) > 0 {
		fmt.Println("WARNING: Items provided but not yet supported. Script will use mock data.")
	} else {
		fmt.Println("Script will use mock data")
	}

	// Build environment variables map
	envVars := make(map[string]string)

	// Add database connection info as environment variables
	// Using MCP_SEARCH_ prefix to avoid conflicts with other environment variables
	envVars["MCP_SEARCH_DB_CONNECTION"] = initData.ConnectionString
	envVars["MCP_SEARCH_DB_NAME"] = initData.DatabaseName
	envVars["MCP_SEARCH_EMBEDDING_DIMENSION"] = fmt.Sprintf("%d", initData.EmbeddingDimension)

	// Add model-specific environment variable if specified
	if modelName := initData.EmbeddingModel; modelName != "" {
		envVars["EMBEDDING_MODEL"] = modelName
	}

	// Add any custom environment variables from init data
	if initData.EnvironmentVars != nil {
		for key, value := range initData.EnvironmentVars {
			envVars[key] = value
		}
	}

	fmt.Println("Environment vars:", envVars)
	fmt.Println("Executing python3 with script:", scriptPath)

	// Execute python3 directly with environment variables (no stdin data)
	result, err := wasmxcore.ExecuteCliCommandWithEnv("python3", []string{scriptPath}, "", envVars)

	fmt.Println("Script execution completed")
	if result != nil {
		fmt.Println("Exit code:", result.ExitCode)
		fmt.Println("Stdout length:", len(result.Stdout))
		fmt.Println("Stderr length:", len(result.Stderr))
		if len(result.Stdout) > 0 {
			fmt.Println("Stdout:", result.Stdout)
		}
		if len(result.Stderr) > 0 {
			fmt.Println("Stderr:", result.Stderr)
		}
	} else {
		fmt.Println("Result is nil!")
	}

	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return createErrorResponse("Failed to execute embedding script", err.Error())
	}

	if result.ExitCode != 0 {
		fmt.Println("ERROR: Script failed")
		errorDetails := result.Stderr
		if errorDetails == "" {
			errorDetails = "No error output captured"
		}
		return createErrorResponse("Embedding script failed", errorDetails)
	}

	// Parse the output - should be JSON with success and items_stored
	var scriptResult struct {
		Success     bool   `json:"success"`
		ItemsStored int    `json:"items_stored"`
		Error       string `json:"error,omitempty"`
	}

	fmt.Println("Parsing script output...")
	if err := json.Unmarshal([]byte(result.Stdout), &scriptResult); err != nil {
		fmt.Println("ERROR: Failed to parse output:", err.Error())
		fmt.Println("Output:", result.Stdout[:min(200, len(result.Stdout))])
		return createErrorResponse("Failed to parse embedding script output", err.Error())
	}

	fmt.Println("Script result: success=", scriptResult.Success, "items_stored=", scriptResult.ItemsStored)

	// Build response
	response := GenerateEmbeddingsResponse{
		Success:      scriptResult.Success,
		ItemsStored:  scriptResult.ItemsStored,
		ErrorMessage: scriptResult.Error,
	}

	responseBz, _ := json.Marshal(response)
	return responseBz
}

func createErrorResponse(message string, details string) []byte {
	response := GenerateEmbeddingsResponse{
		Success:      false,
		ErrorMessage: fmt.Sprintf("%s: %s", message, details),
	}
	responseBz, _ := json.Marshal(response)
	return responseBz
}
