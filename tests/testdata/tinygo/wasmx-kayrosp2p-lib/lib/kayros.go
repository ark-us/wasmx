package lib

import (
	"encoding/hex"
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

// getDataType returns the formatted data_type for Kayros API
// Format: "wasmx_<chain_id>[_<data_type_id>]" converted to hex and right-padded to 32 bytes
func (kc *KayrosClient) getDataType() string {
	chainId := wasmx.GetChainId()
	dataType := fmt.Sprintf("wasmx_%s", chainId)

	// Append data_type_id if provided
	dataTypeId := sGet(CTX_DATA_TYPE_ID)
	if dataTypeId != "" {
		dataType = fmt.Sprintf("%s_%s", dataType, dataTypeId)
	}

	// Convert to bytes and right-pad to 32 bytes
	dataTypeBytes := make([]byte, 32)
	copy(dataTypeBytes, []byte(dataType))

	// Return as hex string
	return hex.EncodeToString(dataTypeBytes)
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
// GET /api/database/record?data_type=<data_type>&data_item=<tx_hash>
func (kc *KayrosClient) GetRecordByDataItem(txHash string) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/database/record?data_type=%s&data_item=%s",
		kc.getDataType(), txHash)

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

	return resp.Data.Data, nil
}

// RegisterTransaction registers a transaction with Kayros via POST /api/grpc/single-hash
func (kc *KayrosClient) RegisterTransaction(txHash string) (*KayrosRegistrationResponse, error) {
	endpoint := "/api/grpc/single-hash"

	reqBody := KayrosRegistrationRequest{
		DataType: kc.getDataType(),
		DataItem: txHash,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	respData, err := kc.makePostRequest(endpoint, bodyBytes)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRegistrationResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros registration response: %w", err)
	}

	return &kayrosResp, nil
}
