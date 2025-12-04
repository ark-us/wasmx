package lib

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	postgresql "github.com/loredanacirstea/wasmx-env-postgresql/lib"
)

// Populate handles bulk population of embeddings (for testing/initialization)
// This function does NOT require authentication and should only be used for initial setup
func Populate(req PopulateRequest) []byte {
	LoggerInfo("Populate called", []string{"items_count", fmt.Sprintf("%d", len(req.Items))})

	if len(req.Items) == 0 {
		response := PopulateResponse{
			Success:      false,
			ItemsStored:  0,
			ErrorMessage: "No items provided",
		}
		data, _ := json.Marshal(response)
		return data
	}

	// Get connection ID
	connIDBytes := wasmx.StorageLoad([]byte(STORAGE_DB_CONN))
	if len(connIDBytes) == 0 {
		response := PopulateResponse{
			Success:      false,
			ItemsStored:  0,
			ErrorMessage: "Database connection not initialized",
		}
		data, _ := json.Marshal(response)
		return data
	}
	connID := string(connIDBytes)

	itemsStored := 0
	var lastError string

	// Store each item
	for _, item := range req.Items {
		// Validate embedding
		if len(item.Embedding) == 0 {
			lastError = fmt.Sprintf("Empty embedding for key '%s'", item.Key)
			LoggerError("Skipping item with empty embedding", []string{"key", item.Key})
			continue
		}

		// Store the value in regular storage
		keyBytes := []byte(item.Key)
		valueBytes := []byte(item.Value)
		wasmx.StorageStore(keyBytes, valueBytes)

		// Store the embedding in PostgreSQL
		query := `
			INSERT INTO state_embeddings(key, embedding)
			VALUES($1, $2)
			ON CONFLICT(key) DO UPDATE SET embedding = EXCLUDED.embedding
		`

		// Encode embedding as JSON for the query parameter
		embeddingJSON, _ := json.Marshal(item.Embedding)
		embeddingParam, _ := json.Marshal(postgresql.SqlQueryParam{
			Type:  "jsonb",
			Value: string(embeddingJSON),
		})

		keyParam, _ := json.Marshal(postgresql.SqlQueryParam{
			Type:  "bytea",
			Value: item.Key,
		})

		execReq := &postgresql.SqlExecuteRequest{
			Id:     connID,
			Query:  query,
			Params: postgresql.Params{keyParam, embeddingParam},
		}

		execResp := postgresql.Execute(execReq)
		if execResp.Error != "" {
			lastError = fmt.Sprintf("Failed to store embedding for key '%s': %s", item.Key, execResp.Error)
			LoggerError("Failed to store embedding", []string{"key", item.Key, "error", execResp.Error})
			continue
		}

		itemsStored++
		LoggerInfo("Stored embedding", []string{"key", item.Key})
	}

	success := itemsStored > 0
	response := PopulateResponse{
		Success:      success,
		ItemsStored:  itemsStored,
		ErrorMessage: lastError,
	}

	if success {
		LoggerInfo("Populate completed", []string{
			"items_stored", fmt.Sprintf("%d", itemsStored),
			"total_items", fmt.Sprintf("%d", len(req.Items)),
		})
	}

	data, _ := json.Marshal(response)
	return data
}
