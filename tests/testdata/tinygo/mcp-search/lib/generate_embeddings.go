package lib

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
)

// GenerateEmbeddings generates embeddings for data and stores them in the database
func GenerateEmbeddings(req GenerateEmbeddingsRequest) []byte {
	fmt.Println("=== GenerateEmbeddings called ===")
	fmt.Println("Items provided:", len(req.Items))

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

	// Build stdin data: first line is config, second line is items (or empty for mock data)
	config := map[string]interface{}{
		"connection_string": initData.ConnectionString,
		"database_name":     initData.DatabaseName,
		"dimension":         initData.EmbeddingDimension,
	}
	configJSON, _ := json.Marshal(config)

	var stdinData string
	if len(req.Items) > 0 {
		// Convert items to JSON for stdin
		itemsJSON, _ := json.Marshal(req.Items)
		stdinData = string(configJSON) + "\n" + string(itemsJSON)
		fmt.Println("Using provided items, count:", len(req.Items))
	} else {
		// No items provided, script will use its mock data
		stdinData = string(configJSON) + "\n"
		fmt.Println("No items provided, script will use mock data")
	}

	// Execute the Python script
	fmt.Println("Executing bulk embedding script...")
	result, err := wasmxcore.ExecuteCliCommand("python3", []string{scriptPath}, stdinData)

	fmt.Println("Script execution completed")
	fmt.Println("Exit code:", result.ExitCode)
	fmt.Println("Stdout length:", len(result.Stdout))
	fmt.Println("Stderr length:", len(result.Stderr))

	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return createErrorResponse("Failed to execute embedding script", err.Error())
	}

	if result.ExitCode != 0 {
		fmt.Println("ERROR: Script failed")
		fmt.Println("Stderr:", result.Stderr)
		return createErrorResponse("Embedding script failed", result.Stderr)
	}

	// Parse the output - should be JSON with success and items_stored
	var scriptResult struct {
		Success      bool   `json:"success"`
		ItemsStored  int    `json:"items_stored"`
		Error        string `json:"error,omitempty"`
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
