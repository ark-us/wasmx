package lib

import (
	"encoding/json"
	"fmt"
	"net/http"

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

// makeRequest performs an HTTP GET request to the Kayros API
func (kc *KayrosClient) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", kc.config.ApiBaseUrl, endpoint)

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}
	if kc.config.ApiUserKey != "" {
		headers["X-User-Key"] = []string{kc.config.ApiUserKey}
	}

	req := httpclient.HttpRequestWrap{
		Request: httpclient.HttpRequest{
			Method: "GET",
			Url:    url,
			Header: headers,
			Data:   []byte{},
		},
		ResponseHandler: httpclient.ResponseHandler{
			MaxSize:  1024 * 1024,
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
func (kc *KayrosClient) GetRecordByDataItem(dataType wasmx.HexString, txHash wasmx.HexString) (*KayrosRecord, error) {
	return kc.GetRecord(dataType, txHash)
}

// GetRecordByHash retrieves a Kayros record by hash_item
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
func (kc *KayrosClient) GetRecordWithPrev(dataType wasmx.HexString, uuid string) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/record-with-prev?data_type=%s&uuid=%s",
		dataType, uuid)

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
func (kc *KayrosClient) GetRecordsFromPrev(dataType wasmx.HexString, uuid string, limit int) ([]KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/records-since-prev?data_type=%s&uuid=%s&limit=%d",
		dataType, uuid, limit)

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

	userKey := kc.config.ApiUserKey

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
			MaxSize:  1024 * 1024,
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
func (kc *KayrosClient) GetBlockTimestamp(blockDataType wasmx.HexString, blockHashHex wasmx.HexString) (string, error) {
	record, err := kc.GetBlockRecord(blockDataType, blockHashHex)
	if err == nil && record != nil {
		timestamp, err := TimeuuidHexToTimestamp(record.UuidHex)
		if err != nil {
			return "", fmt.Errorf("failed to extract timestamp from existing record UUID: %w: %s", err, record.UuidHex)
		}
		return timestamp, nil
	}

	LoggerDebug("registering block hash with Kayros", []string{"blockHash", string(blockHashHex)})

	_, err = kc.RegisterBlockHash(blockDataType, blockHashHex)
	if err != nil {
		return "", fmt.Errorf("failed to register block hash: %w", err)
	}

	record, err = kc.GetBlockRecord(blockDataType, blockHashHex)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve block record after registration: %w", err)
	}

	timestamp, err := TimeuuidHexToTimestamp(record.UuidHex)
	if err != nil {
		return "", fmt.Errorf("failed to extract timestamp from record UUID: %w", err)
	}

	LoggerInfo("block hash registered and queried from Kayros", []string{
		"uuid", record.UuidHex,
		"timestamp", timestamp,
	})
	return timestamp, nil
}

// GetBlockRecord retrieves a block record from Kayros by block hash
func (kc *KayrosClient) GetBlockRecord(blockDataType wasmx.HexString, blockHashHex wasmx.HexString) (*KayrosRecord, error) {
	return kc.GetRecord(blockDataType, blockHashHex)
}

// GetRecord retrieves a record from Kayros by data_type and data_item
func (kc *KayrosClient) GetRecord(dataType wasmx.HexString, dataItem wasmx.HexString) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/record?data_type=%s&data_item=%s", dataType, dataItem)

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
	if len(kayrosResp.Data.Records) < 1 {
		return nil, fmt.Errorf("Kayros API record not found: %s", kayrosResp.Message)
	}
	return &kayrosResp.Data.Records[0], nil
}

// GetLevelHash retrieves a rollup hash by level and position.
func (kc *KayrosClient) GetLevelHash(level int, position int) (*LevelHashEntry, error) {
	endpoint := fmt.Sprintf("/api/database/level-hash?level=%d&position=%d", level, position)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp LevelHashResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if !kayrosResp.Success {
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	return &LevelHashEntry{
		Position: kayrosResp.Data.Position,
		HashHex:  kayrosResp.Data.HashHex,
		UuidHex:  "",
	}, nil
}

// GetLevelRange retrieves a set of rollup hashes for a level.
func (kc *KayrosClient) GetLevelRange(level int, limit int) ([]LevelHashEntry, error) {
	endpoint := fmt.Sprintf("/api/database/level-range?level=%d&limit=%d", level, limit)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp LevelRangeResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if !kayrosResp.Success {
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	return kayrosResp.Data.Entries, nil
}

// GetProofPath retrieves the proof path for a hash_item.
func (kc *KayrosClient) GetProofPath(hashItem string) (*ProofPathData, error) {
	endpoint := fmt.Sprintf("/api/proof?hash_item=%s", hashItem)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp ProofPathResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if !kayrosResp.Success {
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}

	return &kayrosResp.Data, nil
}

// GetBlockRecordByHashItem retrieves a block record from Kayros by hash_item
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
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Message)
	}
	return &kayrosResp.Data, nil
}

// RegisterBlockHash registers a block hash with Kayros for timestamp synchronization
func (kc *KayrosClient) RegisterBlockHash(blockDataType wasmx.HexString, blockHashHex wasmx.HexString) (*KayrosRegistrationResponse, error) {
	reqBody := KayrosRegistrationRequest{
		DataType: blockDataType,
		DataItem: blockHashHex,
	}
	return kc.RegisterData(reqBody)
}

// RegisterTransaction registers a transaction hash with Kayros for ordering
func (kc *KayrosClient) RegisterTransaction(dataType wasmx.HexString, txHash wasmx.HexString) (*KayrosRegistrationResponse, error) {
	reqBody := KayrosRegistrationRequest{
		DataType: dataType,
		DataItem: txHash,
	}
	return kc.RegisterData(reqBody)
}

// RegisterData registers a data item with Kayros
func (kc *KayrosClient) RegisterData(reqBody KayrosRegistrationRequest) (*KayrosRegistrationResponse, error) {
	endpoint := "/api/grpc/single-hash"
	reqbz, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respData, err := kc.makePostRequest(endpoint, reqbz)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRegistrationResponseWrap
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}

	LoggerDebugExtended(`Kayros registration`, []string{"endpoint", endpoint, "success", fmt.Sprintf("%v", kayrosResp.Success), "kayros_hash", string(kayrosResp.Data.ComputedHash), "timeuuid", string(kayrosResp.Data.TimeUUID)})

	if !kayrosResp.Success {
		return nil, fmt.Errorf("Failed Kayros registration request: %s", kayrosResp.Message)
	}
	if !kayrosResp.Data.Success {
		return nil, fmt.Errorf("Failed Kayros registration: %s; %s", kayrosResp.Message, kayrosResp.Data.Message)
	}
	return &kayrosResp.Data, nil
}
