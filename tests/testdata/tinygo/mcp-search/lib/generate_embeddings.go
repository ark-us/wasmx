package lib

import (
	"encoding/json"
	"fmt"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// GenerateEmbeddings generates embeddings for data and stores them in the database
func GenerateEmbeddings(req GenerateEmbeddingsRequest) []byte {
	LoggerInfo("GenerateEmbeddings called", []string{"items_count", fmt.Sprintf("%d", len(req.Items))})

	// Path to the bulk embedding generation script
	scriptPath := "/Users/user/dev/blockchain/wasmx/tests/testdata/tinygo/mcp-search/scripts/generate_bulk_embeddings.py"

	// Load init data to get database connection parameters
	initDataBz := wasmx.StorageLoad([]byte(STORAGE_INIT_DATA))
	if len(initDataBz) == 0 {
		return createErrorResponse("Init data not found", "cannot generate embeddings without database config")
	}

	var initData InitGenesisRequest
	if err := json.Unmarshal(initDataBz, &initData); err != nil {
		return createErrorResponse("Failed to unmarshal init data", err.Error())
	}

	// Build environment variables map
	envVars := make(map[string]string)
	envVars["MCP_SEARCH_DB_CONNECTION"] = initData.SearchConfig.Database.ConnectionString
	envVars["MCP_SEARCH_DB_NAME"] = initData.SearchConfig.Database.DatabaseName
	envVars["MCP_SEARCH_EMBEDDING_DIMENSION"] = fmt.Sprintf("%d", initData.SearchConfig.EmbeddingDimension)

	// Add any custom environment variables from init data
	if initData.EnvironmentVars != nil {
		for key, value := range initData.EnvironmentVars {
			envVars[key] = value
		}
	}

	LoggerInfo("Executing embedding script", []string{"script", scriptPath})

	// Execute python3 directly with environment variables
	result, err := wasmxcore.ExecuteCliCommandWithEnv("python3", []string{scriptPath}, "", envVars)

	if err != nil {
		LoggerError("Failed to execute embedding script", []string{"error", err.Error()})
		return createErrorResponse("Failed to execute embedding script", err.Error())
	}

	if result.ExitCode != 0 {
		errorDetails := result.Stderr
		if errorDetails == "" {
			errorDetails = "No error output captured"
		}
		LoggerError("Embedding script failed", []string{"exit_code", fmt.Sprintf("%d", result.ExitCode), "stderr", errorDetails})
		return createErrorResponse("Embedding script failed", errorDetails)
	}

	// Parse the output - should be JSON with success and items_stored
	var scriptResult struct {
		Success     bool   `json:"success"`
		ItemsStored int    `json:"items_stored"`
		Error       string `json:"error,omitempty"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &scriptResult); err != nil {
		LoggerError("Failed to parse script output", []string{"error", err.Error()})
		return createErrorResponse("Failed to parse embedding script output", err.Error())
	}

	LoggerInfo("Embedding script completed", []string{"success", fmt.Sprintf("%v", scriptResult.Success), "items_stored", fmt.Sprintf("%d", scriptResult.ItemsStored)})

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
