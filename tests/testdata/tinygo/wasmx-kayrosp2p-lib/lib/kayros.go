package lib

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	httpclient "github.com/loredanacirstea/wasmx-env-httpclient"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// KayrosClient provides methods to interact with Kayros API
type KayrosClient struct {
	config KayrosConfig
}

// NewKayrosClient creates a new Kayros API client with the given configuration
func NewKayrosClient(config KayrosConfig) *KayrosClient {
	return &KayrosClient{
		config: config,
	}
}

// getDataType returns the formatted data_type for Kayros API
// Format: "wasmx_<chain_id>[_<data_type_id>]" converted to hex and right-padded to 32 bytes
func (kc *KayrosClient) getDataType() wasmx.HexString {
	chainId := wasmx.GetChainId()
	dataType := fmt.Sprintf("wasmx_tx_%s", chainId)

	// Append data_type_id if provided
	dataTypeId := sGet(CTX_DATA_TYPE_ID)
	if dataTypeId != "" {
		dataType = fmt.Sprintf("%s_%s", dataType, dataTypeId)
	}

	// Convert to bytes and right-pad to 32 bytes
	dataTypeBytes := make([]byte, 32)
	copy(dataTypeBytes, []byte(dataType))

	// Return as hex string
	return wasmx.HexString(hex.EncodeToString(dataTypeBytes))
}

// getBlockDataType returns the data_type for block timestamp synchronization
// Format: "wasmx_<chain_id>[_<data_type_id>]_blocks" converted to hex and right-padded to 32 bytes
func (kc *KayrosClient) getBlockDataType() wasmx.HexString {
	chainId := wasmx.GetChainId()
	dataType := fmt.Sprintf("wasmx_b_%s", chainId)

	// Append data_type_id if provided
	dataTypeId := sGet(CTX_DATA_TYPE_ID)
	if dataTypeId != "" {
		dataType = fmt.Sprintf("%s_%s", dataType, dataTypeId)
	}

	// Convert to bytes and right-pad to 32 bytes
	dataTypeBytes := make([]byte, 32)
	copy(dataTypeBytes, []byte(dataType))

	// Return as hex string
	return wasmx.HexString(hex.EncodeToString(dataTypeBytes))
}

// makeRequest performs an HTTP GET request to the Kayros API
func (kc *KayrosClient) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", kc.config.ApiBaseUrl, endpoint)

	req := httpclient.HttpRequestWrap{
		Request: httpclient.HttpRequest{
			Method: "GET",
			Url:    url,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Data: []byte{},
		},
		ResponseHandler: httpclient.ResponseHandler{
			MaxSize:  1024 * 1024, // 1MB max response size
			FilePath: "",
		},
	}
	LoggerDebug("http request", []string{"url", url})
	resp := httpclient.Request(&req)

	if resp.Error != "" {
		return nil, fmt.Errorf("HTTP request failed: %s", resp.Error)
	}

	if resp.Data.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP request returned status %d: %s", resp.Data.StatusCode, resp.Data.Status)
	}

	return resp.Data.Data, nil
}

// GetRecordByDataItem retrieves a Kayros record by data_type and data_item (transaction hash)
func (kc *KayrosClient) GetRecordByDataItem(txHash wasmx.HexString) (*KayrosRecord, error) {
	return kc.GetRecord(kc.getDataType(), txHash)
}

// GetRecordByHash retrieves a Kayros record by hash_item (Kayros unique ID)
// GET /api/database/record-by-hash?hash_item=<hash_item>
func (kc *KayrosClient) GetRecordByHash(hashItem string) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/record-by-hash?hash_item=%s", hashItem)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRecordResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}

	if !kayrosResp.Success {
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	return &kayrosResp.Data, nil
}

// GetRecordWithPrev retrieves a Kayros record with its previous hash by UUID
// GET /api/database/record-with-prev?data_type=<data_type>&uuid=<uuid>
func (kc *KayrosClient) GetRecordWithPrev(uuid string) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/record-with-prev?data_type=%s&uuid=%s",
		kc.getDataType(), uuid)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRecordResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}

	if !kayrosResp.Success {
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	return &kayrosResp.Data, nil
}

// GetRecordsFromPrev retrieves multiple Kayros records from a previous UUID with a limit
// GET /api/database/records-since-prev?data_type=<data_type>&uuid=<uuid>&limit=<limit>
func (kc *KayrosClient) GetRecordsFromPrev(uuid string, limit int) ([]KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/records-since-prev?data_type=%s&uuid=%s&limit=%d",
		kc.getDataType(), uuid, limit)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRecordsResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}

	if !kayrosResp.Success {
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	return kayrosResp.Data.Records, nil
}

// makePostRequest performs an HTTP POST request to the Kayros API
func (kc *KayrosClient) makePostRequest(endpoint string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s%s", kc.config.ApiBaseUrl, endpoint)

	// Get user key from context
	userKey := sGet("kayros_user_key")

	req := httpclient.HttpRequestWrap{
		Request: httpclient.HttpRequest{
			Method: "POST",
			Url:    url,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-User-Key":   []string{userKey},
			},
			Data: body,
		},
		ResponseHandler: httpclient.ResponseHandler{
			MaxSize:  1024 * 1024, // 1MB max response size
			FilePath: "",
		},
	}
	reqbz, _ := json.Marshal(&req)
	LoggerDebug("http request", []string{"url", url, "method", "POST", "data", string(body), "req", string(reqbz)})
	resp := httpclient.Request(&req)

	if resp.Error != "" {
		return nil, fmt.Errorf("HTTP POST request failed: %s", resp.Error)
	}

	if resp.Data.StatusCode != 200 && resp.Data.StatusCode != 201 {
		return nil, fmt.Errorf("HTTP POST request returned status %d: %s", resp.Data.StatusCode, resp.Data.Status)
	}
	LoggerDebug("http request", []string{"url", url, "method", "POST", "response", string(resp.Data.Data)})

	return resp.Data.Data, nil
}

// GetBlockTimestamp retrieves or registers a block hash with Kayros to get a deterministic timestamp
// This ensures all validators use the same timestamp for the same block
// Returns the Kayros timestamp for the block in RFC3339Nano format
func (kc *KayrosClient) GetBlockTimestamp(blockHashHex wasmx.HexString) (string, error) {
	// First, try to get the existing record
	record, err := kc.GetBlockRecord(blockHashHex)
	if err == nil && record != nil {
		// Record exists, extract timestamp from TimeUUID for RFC3339Nano format
		timestamp, err := TimeuuidHexToTimestamp(record.UuidHex)
		if err != nil {
			return "", fmt.Errorf("failed to extract timestamp from existing record UUID: %w: %s", err, record.UuidHex)
		}
		LoggerDebug("block hash already registered with Kayros", []string{
			"blockHash", string(blockHashHex),
			"timestamp", timestamp,
		})
		return timestamp, nil
	}

	// Record doesn't exist, register it
	LoggerDebug("registering block hash with Kayros", []string{"blockHash", string(blockHashHex)})
	_, err = kc.RegisterBlockHash(blockHashHex)
	if err != nil {
		return "", fmt.Errorf("failed to register block hash: %w", err)
	}

	// IMPORTANT: Query separately to get the earliest record (not registration response)
	// This handles race conditions when multiple nodes register the same hash simultaneously
	// The Kayros API returns the earliest registration (limit 1) when querying by data_type + data_item
	record, err = kc.GetBlockRecord(blockHashHex)
	if err != nil {
		return "", fmt.Errorf("failed to query block record after registration: %w", err)
	}
	if record == nil {
		return "", fmt.Errorf("block record not found after registration")
	}

	// Extract timestamp from the earliest registration's TimeUUID
	timestamp, err := TimeuuidHexToTimestamp(record.UuidHex)
	if err != nil {
		return "", fmt.Errorf("failed to extract timestamp from UUID: %w: %s", err, record.UuidHex)
	}

	LoggerInfo("block hash registered and queried from Kayros", []string{
		"blockHash", string(blockHashHex),
		"timeUUID", record.UuidHex,
		"timestamp", timestamp,
		"hashItem", record.HashItemHex,
	})

	return timestamp, nil
}

// GetBlockRecord retrieves a block record from Kayros by block hash
func (kc *KayrosClient) GetBlockRecord(blockHashHex wasmx.HexString) (*KayrosRecord, error) {
	return kc.GetRecord(kc.getBlockDataType(), blockHashHex)
}

// GetBlockRecord retrieves a block record from Kayros by block hash
// GET /api/database/record?data_type=<blocks_data_type>&data_item=<block_hash>
func (kc *KayrosClient) GetRecord(dataType wasmx.HexString, dataItem wasmx.HexString) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/record?data_type=%s&data_item=%s&limit=1",
		dataType, dataItem)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRecordsResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}

	if !kayrosResp.Success {
		// Record not found or other error
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	if len(kayrosResp.Data.Records) < 1 {
		// Record not found or other error
		return nil, fmt.Errorf("Kayros API record not found: %s", kayrosResp.Message)
	}

	return &kayrosResp.Data.Records[0], nil
}

// GetBlockRecordByHashItem retrieves a block record from Kayros by hash_item (Kayros computed hash)
// GET /api/database/record-by-hash?hash_item=<hash_item>
func (kc *KayrosClient) GetBlockRecordByHashItem(hashItem string) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/record-by-hash?hash_item=%s", hashItem)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRecordResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}

	if !kayrosResp.Success {
		// Record not found or other error
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	return &kayrosResp.Data, nil
}

// RegisterBlockHash registers a block hash with Kayros for timestamp synchronization
// POST /api/grpc/single-hash with blocks data type
func (kc *KayrosClient) RegisterBlockHash(blockHashHex wasmx.HexString) (*KayrosRegistrationResponse, error) {
	reqBody := KayrosRegistrationRequest{
		DataType: kc.getBlockDataType(),
		DataItem: blockHashHex,
	}
	return kc.RegisterData(reqBody)
}

// RegisterTransaction registers a transaction with Kayros via POST /api/grpc/single-hash
func (kc *KayrosClient) RegisterTransaction(txHash string) (*KayrosRegistrationResponse, error) {
	reqBody := KayrosRegistrationRequest{
		DataType: kc.getDataType(),
		DataItem: wasmx.HexString(txHash),
	}
	return kc.RegisterData(reqBody)
}

// RegisterData registers data with Kayros via POST /api/grpc/single-hash
func (kc *KayrosClient) RegisterData(reqBody KayrosRegistrationRequest) (*KayrosRegistrationResponse, error) {
	endpoint := "/api/grpc/single-hash"
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	LoggerDebugExtended(`Kayros registration`, []string{"endpoint", endpoint})

	respData, err := kc.makePostRequest(endpoint, bodyBytes)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRegistrationResponseWrap
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros registration response: %w", err)
	}

	LoggerDebugExtended(`Kayros registration`, []string{"endpoint", endpoint, "success", strconv.FormatBool(kayrosResp.Success), "kayros_hash", string(kayrosResp.Data.ComputedHash), "timeuuid", string(kayrosResp.Data.TimeUUID)})

	if !kayrosResp.Success {
		return nil, fmt.Errorf("Failed Kayros registration request: %s", kayrosResp.Message)
	}
	if !kayrosResp.Data.Success {
		return nil, fmt.Errorf("Failed Kayros registration: %s; %s", kayrosResp.Message, kayrosResp.Data.Message)
	}
	return &kayrosResp.Data, nil
}
