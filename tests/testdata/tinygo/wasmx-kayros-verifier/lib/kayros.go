package lib

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	httpclient "github.com/loredanacirstea/wasmx-env-httpclient/lib"
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

// BuildRecordByHashPath returns the encoded Kayros API path for record lookup by hash.
func BuildRecordByHashPath(dataType string, hashItem string) string {
	params := url.Values{}
	params.Set("hash", hashItem)
	params.Set("data_type", dataType)
	return "/api/lightnet/database/record-by-hash?" + params.Encode()
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
	fmt.Println("--makeRequest.url--", url)
	resp := httpclient.Request(&req)
	fmt.Println("--makeRequest.resp--", resp.Error, resp.Data.StatusCode, string(resp.Data.Data))

	if resp.Error != "" {
		return nil, fmt.Errorf("HTTP request failed: %s", resp.Error)
	}
	if resp.Data.StatusCode != 200 {
		errMsg := ""
		if parsedErr, ok := parseKayrosError(resp.Data.Data); ok {
			errMsg = parsedErr
		}
		if errMsg != "" {
			return nil, fmt.Errorf(
				"HTTP request returned status %d: %s; kayros_error=%s; body=%s",
				resp.Data.StatusCode,
				resp.Data.Status,
				errMsg,
				string(resp.Data.Data),
			)
		}
		return nil, fmt.Errorf(
			"HTTP request returned status %d: %s; body=%s",
			resp.Data.StatusCode,
			resp.Data.Status,
			string(resp.Data.Data),
		)
	}
	return resp.Data.Data, nil
}

// GetRecordByDataItem retrieves a Kayros record by data_type and data_item (transaction hash)
func (kc *KayrosClient) GetRecordByDataItem(dataType string, txHash []byte) (*KayrosRecord, error) {
	return kc.GetRecord(dataType, txHash)
}

// GetRecordByHash retrieves a Kayros record by hash_item and data_type
func (kc *KayrosClient) GetRecordByHash(dataType string, hashItem string) (*KayrosRecord, error) {
	endpoint := BuildRecordByHashPath(dataType, hashItem)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var hashResp kayrosHashRecordResponse
	if err := json.Unmarshal(respData, &hashResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if errMsg, ok := parseKayrosError(respData); ok {
		return nil, fmt.Errorf("Kayros API error: %s", errMsg)
	}
	record, err := hashResp.toKayrosRecord()
	if err != nil {
		return nil, err
	}
	return record, nil
}

// GetRecordWithPrev retrieves a Kayros record by previous hash
func (kc *KayrosClient) GetRecordWithPrev(dataType string, prevHash string) (*KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/lightnet/database/record-with-prev?data_type=%s&prev_hash=%s",
		dataType, prevHash)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var record KayrosRecord
	if err := json.Unmarshal(respData, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if errMsg, ok := parseKayrosError(respData); ok {
		return nil, fmt.Errorf("Kayros API error: %s", errMsg)
	}
	return &record, nil
}

// GetRecordsFromPrev retrieves multiple Kayros records from a previous UUID with a limit
func (kc *KayrosClient) GetRecordsFromPrev(dataType string, uuid string, limit int) ([]KayrosRecord, error) {
	endpoint := fmt.Sprintf("/api/lightnet/database/records-since-prev?data_type=%s&uuid=%s&limit=%d",
		dataType, uuid, limit)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var kayrosResp kayrosRecordsResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if kayrosResp.Error != "" {
		return nil, fmt.Errorf("Kayros API error: %s", kayrosResp.Error)
	}
	records := make([]KayrosRecord, 0, len(kayrosResp.Records))
	for _, rec := range kayrosResp.Records {
		mapped, err := rec.toKayrosRecord()
		if err != nil {
			continue
		}
		records = append(records, *mapped)
	}
	return records, nil
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
	LoggerDebug("http request", []string{"url", url, "method", "POST", "data", string(body)})
	resp := httpclient.Request(&req)

	if resp.Error != "" {
		return nil, fmt.Errorf("HTTP POST request failed: %s; request=%s", resp.Error, string(body))
	}
	if resp.Data.StatusCode != 200 && resp.Data.StatusCode != 201 {
		errMsg := ""
		if parsedErr, ok := parseKayrosError(resp.Data.Data); ok {
			errMsg = parsedErr
		}
		if errMsg != "" {
			return nil, fmt.Errorf(
				"HTTP POST request returned status %d: %s; kayros_error=%s; body=%s; request=%s",
				resp.Data.StatusCode,
				resp.Data.Status,
				errMsg,
				string(resp.Data.Data),
				string(body),
			)
		}
		return nil, fmt.Errorf(
			"HTTP POST request returned status %d: %s; body=%s; request=%s",
			resp.Data.StatusCode,
			resp.Data.Status,
			string(resp.Data.Data),
			string(body),
		)
	}
	LoggerDebug("http request", []string{"url", url, "method", "POST", "response", string(resp.Data.Data)})

	return resp.Data.Data, nil
}

// GetBlockTimestamp retrieves or registers a block hash with Kayros to get a deterministic timestamp
func (kc *KayrosClient) GetBlockTimestamp(blockDataType string, blockHashHex []byte) (string, error) {
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

	var lastErr error
	for i := 0; i < 10; i++ {
		time.Sleep(300 * time.Millisecond)
		record, err = kc.GetBlockRecord(blockDataType, blockHashHex)
		if err == nil && record != nil {
			break
		}
		if err != nil && !isNotFoundError(err) {
			return "", fmt.Errorf("failed to retrieve block record after registration: %w", err)
		}
		lastErr = err
	}
	if record == nil {
		if lastErr == nil {
			lastErr = fmt.Errorf("record not found after registration")
		}
		return "", fmt.Errorf("failed to retrieve block record after registration: %w", lastErr)
	}

	timestamp, err := TimeuuidHexToTimestamp(record.UuidHex)
	if err != nil {
		return "", fmt.Errorf("failed to extract timestamp from record UUID: %w", err)
	}

	LoggerInfo("block hash registered and queried from Kayros", []string{
		"uuid", record.UuidHex,
		"timestamp", timestamp,
		"data_type", blockDataType,
	})
	return timestamp, nil
}

// GetBlockRecord retrieves a block record from Kayros by block hash
func (kc *KayrosClient) GetBlockRecord(blockDataType string, blockHashHex []byte) (*KayrosRecord, error) {
	return kc.GetRecord(blockDataType, blockHashHex)
}

// GetRecord retrieves a record from Kayros by data_type and data_item
func (kc *KayrosClient) GetRecord(dataType string, dataItem []byte) (*KayrosRecord, error) {
	dataItemHex := hex.EncodeToString(dataItem)
	endpoint := fmt.Sprintf("/api/lightnet/database/record?data_type=%s&data_item=%s", dataType, dataItemHex)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var recordsResp kayrosRecordsResponse
	if err := json.Unmarshal(respData, &recordsResp); err != nil {
		var record KayrosRecord
		if err2 := json.Unmarshal(respData, &record); err2 != nil {
			return nil, fmt.Errorf("failed to unmarshal Kayros records response: %w; record decode error: %v", err, err2)
		}
		if len(record.HashItem) == 0 {
			return nil, fmt.Errorf("unexpected Kayros record response without hash_item")
		}
		return &record, nil
	}
	if recordsResp.Error != "" {
		return nil, fmt.Errorf("Kayros API error: %s", recordsResp.Error)
	}
	if len(recordsResp.Records) == 0 {
		return nil, fmt.Errorf("Kayros API record not found")
	}
	record, err := recordsResp.Records[0].toKayrosRecord()
	if err != nil {
		return nil, err
	}
	return record, nil
}

// GetLevelHash retrieves a rollup hash by data_type, level, and position.
func (kc *KayrosClient) GetLevelHash(dataType string, level int, position int) (*LevelHashEntry, error) {
	if strings.TrimSpace(dataType) == "" {
		return nil, fmt.Errorf("data_type is required for level hash")
	}
	reqBody := browseLevelsRequest{
		Table:    "levels",
		DataType: dataType,
		Level:    &level,
		Position: int64Ptr(int64(position)),
		Limit:    1,
	}
	reqbz, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respData, err := kc.makePostRequest("/api/lightnet/database/browse", reqbz)
	if err != nil {
		return nil, err
	}

	var browseResp browseDatabaseResponse
	if err := json.Unmarshal(respData, &browseResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if browseResp.Error != "" {
		return nil, fmt.Errorf("Kayros API error: %s", browseResp.Error)
	}
	if len(browseResp.Rows) == 0 {
		return nil, fmt.Errorf("Kayros API record not found")
	}

	var row levelRow
	if err := json.Unmarshal(browseResp.Rows[0], &row); err != nil {
		return nil, fmt.Errorf("failed to parse level row: %w", err)
	}
	hashHex, err := base64ToHex(row.Hash)
	if err != nil {
		return nil, fmt.Errorf("invalid level hash encoding: %w", err)
	}
	return &LevelHashEntry{
		Position: int(row.Position),
		HashHex:  hashHex,
		UuidHex:  "",
	}, nil
}

// GetLevelRange retrieves a set of rollup hashes for a data_type and level.
func (kc *KayrosClient) GetLevelRange(dataType string, level int, limit int) ([]LevelHashEntry, error) {
	if strings.TrimSpace(dataType) == "" {
		return nil, fmt.Errorf("data_type is required for level range")
	}
	if limit <= 0 {
		limit = 50
	}
	reqBody := browseLevelsRequest{
		Table:    "levels",
		DataType: dataType,
		Limit:    limit,
	}
	reqbz, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respData, err := kc.makePostRequest("/api/lightnet/database/browse", reqbz)
	if err != nil {
		return nil, err
	}

	var browseResp browseDatabaseResponse
	if err := json.Unmarshal(respData, &browseResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if browseResp.Error != "" {
		return nil, fmt.Errorf("Kayros API error: %s", browseResp.Error)
	}
	if len(browseResp.Rows) == 0 {
		return []LevelHashEntry{}, nil
	}

	entries := make([]LevelHashEntry, 0, len(browseResp.Rows))
	for _, raw := range browseResp.Rows {
		var row levelRow
		if err := json.Unmarshal(raw, &row); err != nil {
			continue
		}
		if row.Level != level {
			continue
		}
		hashHex, err := base64ToHex(row.Hash)
		if err != nil {
			continue
		}
		entries = append(entries, LevelHashEntry{
			Position: int(row.Position),
			HashHex:  hashHex,
			UuidHex:  "",
		})
		if len(entries) >= limit {
			break
		}
	}
	return entries, nil
}

// VerifyHashBatch checks hash existence at a level and start position.
func (kc *KayrosClient) VerifyHashBatch(dataType string, level int, start int64, hashes []string) (bool, string) {
	if strings.TrimSpace(dataType) == "" {
		return false, "missing data_type"
	}
	if len(hashes) == 0 {
		return true, ""
	}
	reqBody := verifyHashBatchRequest{
		DataType: dataType,
		Level:    level,
		Start:    start,
		Hashes:   hashes,
	}
	reqbz, err := json.Marshal(&reqBody)
	if err != nil {
		return false, "failed to marshal verify-hash-batch request"
	}
	respData, err := kc.makePostRequest("/api/lightnet/merkle/verify-hash-batch", reqbz)
	if err != nil {
		return false, err.Error()
	}
	var resp verifyHashBatchResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return false, "failed to parse verify-hash-batch response"
	}
	if resp.Error != "" {
		return false, resp.Error
	}
	if resp.Mismatches > 0 {
		return false, fmt.Sprintf("mismatches=%d", resp.Mismatches)
	}
	for _, r := range resp.Results {
		if r != 1 {
			return false, "missing hash in batch"
		}
	}
	return true, ""
}

// GetProofPath retrieves the merkle proof for a hash_item.
func (kc *KayrosClient) GetProofPath(dataType string, hashItem string) (*ProofPathData, error) {
	endpoint := fmt.Sprintf("/api/lightnet/merkle-proof?hash=%s&data_type=%s", hashItem, dataType)

	respData, err := kc.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var rawResp struct {
		Success     bool     `json:"success"`
		Error       string   `json:"error,omitempty"`
		Message     string   `json:"message,omitempty"`
		DataType    string   `json:"data_type"`
		HashItem    string   `json:"hash_item"`
		Proof       []string `json:"proof"`
		Root        string   `json:"root"`
		Position    int64    `json:"position"`
		Levels      int32    `json:"levels"`
		LevelCounts []int32  `json:"level_counts"`
		LevelStarts []int64  `json:"level_starts"`
	}
	if err := json.Unmarshal(respData, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}
	if !rawResp.Success {
		return nil, fmt.Errorf("Kayros API error: %s", rawResp.Message)
	}

	hashItemBytes, err := decodeHex(rawResp.HashItem)
	if err != nil {
		return nil, fmt.Errorf("invalid hash_item hex: %w", err)
	}

	return &ProofPathData{
		DataType:    rawResp.DataType,
		HashItem:    hashItemBytes,
		Proof:       rawResp.Proof,
		Root:        rawResp.Root,
		Position:    rawResp.Position,
		Levels:      rawResp.Levels,
		LevelCounts: rawResp.LevelCounts,
		LevelStarts: rawResp.LevelStarts,
	}, nil
}

// GetBlockRecordByHashItem retrieves a block record from Kayros by hash_item
func (kc *KayrosClient) GetBlockRecordByHashItem(dataType string, hashItem string) (*KayrosRecord, error) {
	return kc.GetRecordByHash(dataType, hashItem)
}

// RegisterBlockHash registers a block hash with Kayros for timestamp synchronization
func (kc *KayrosClient) RegisterBlockHash(blockDataType string, blockHashHex []byte) (*KayrosRegistrationResponse, error) {
	reqBody := KayrosRegistrationRequest{
		DataType: blockDataType,
		DataItem: blockHashHex,
	}
	return kc.RegisterData(reqBody)
}

// RegisterTransaction registers a transaction hash with Kayros for ordering
func (kc *KayrosClient) RegisterTransaction(dataType string, txHash []byte) (*KayrosRegistrationResponse, error) {
	reqBody := KayrosRegistrationRequest{
		DataType: dataType,
		DataItem: txHash,
	}
	return kc.RegisterData(reqBody)
}

// RegisterData registers a data item with Kayros
func (kc *KayrosClient) RegisterData(reqBody KayrosRegistrationRequest) (*KayrosRegistrationResponse, error) {
	endpoint := "/api/lightnet/grpc/single-hash"
	wireReq := kayrosRegistrationWireRequest{
		DataType: reqBody.DataType,
		DataItem: hex.EncodeToString(reqBody.DataItem),
	}
	reqbz, err := json.Marshal(&wireReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respData, err := kc.makePostRequest(endpoint, reqbz)
	if err != nil {
		return nil, err
	}

	var kayrosResp KayrosRegistrationResponse
	if err := json.Unmarshal(respData, &kayrosResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kayros response: %w", err)
	}

	LoggerDebugExtended(`Kayros registration`, []string{"endpoint", endpoint, "success", fmt.Sprintf("%v", kayrosResp.Success), "kayros_hash", kayrosResp.Hash, "timeuuid", kayrosResp.TimeUUID})

	if !kayrosResp.Success {
		errMsg := kayrosResp.Error
		if errMsg == "" {
			errMsg = kayrosResp.Message
		}
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("Failed Kayros registration: %s", errMsg)
	}
	return &kayrosResp, nil
}

type browseLevelsRequest struct {
	Table    string `json:"table"`
	Limit    int    `json:"limit,omitempty"`
	DataType string `json:"data_type,omitempty"`
	Level    *int   `json:"level,omitempty"`
	Position *int64 `json:"position,omitempty"`
}

type browseDatabaseResponse struct {
	Rows    []json.RawMessage `json:"rows"`
	Error   string            `json:"error,omitempty"`
	Message string            `json:"message,omitempty"`
}

type levelRow struct {
	DataType string `json:"data_type"`
	Level    int    `json:"level"`
	Position int64  `json:"position"`
	Hash     string `json:"hash"`
	TS       string `json:"ts"`
}

type kayrosHashRecordResponse struct {
	DataType string `json:"data_type"`
	DataItem string `json:"data_item"`
	HashType string `json:"hash_type"`
	HashItem string `json:"hash_item"`
	Position int64  `json:"position"`
	TS       string `json:"ts"`
	PrevHash string `json:"prev_hash"`
}

func (resp kayrosHashRecordResponse) toKayrosRecord() (*KayrosRecord, error) {
	if resp.DataType == "" || resp.HashItem == "" {
		return nil, fmt.Errorf("Kayros API record not found")
	}
	dataItem, err := base64.StdEncoding.DecodeString(resp.DataItem)
	if err != nil {
		return nil, fmt.Errorf("invalid data_item encoding: %w", err)
	}
	hashItem, err := base64.StdEncoding.DecodeString(resp.HashItem)
	if err != nil {
		return nil, fmt.Errorf("invalid hash_item encoding: %w", err)
	}
	prevHash, err := decodeBase64Optional(resp.PrevHash)
	if err != nil {
		return nil, fmt.Errorf("invalid prev_hash encoding: %w", err)
	}
	uuidHex, err := uuidStringToHex(resp.TS)
	if err != nil {
		return nil, fmt.Errorf("invalid ts uuid %q: %w", resp.TS, err)
	}
	timestamp, err := TimeuuidHexToTimestamp(uuidHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp from uuid %q: %w", resp.TS, err)
	}

	return &KayrosRecord{
		DataType:    resp.DataType,
		DataTypeHex: hex.EncodeToString([]byte(resp.DataType)),
		DataItem:    dataItem,
		UuidHex:     uuidHex,
		HashItem:    hashItem,
		PrevHash:    prevHash,
		HashType:    resp.HashType,
		Position:    resp.Position,
		Timestamp:   timestamp,
	}, nil
}

type kayrosRecordApi struct {
	DataType string `json:"data_type"`
	DataItem []byte `json:"data_item"`
	HashItem []byte `json:"hash_item"`
	HashType string `json:"hash_type"`
	Position int64  `json:"position"`
	PrevHash []byte `json:"prev_hash"`
	TS       string `json:"ts"`
}

type kayrosRecordsResponse struct {
	Count   int               `json:"count"`
	Records []kayrosRecordApi `json:"records"`
	Error   string            `json:"error,omitempty"`
	Message string            `json:"message,omitempty"`
}

type kayrosRegistrationWireRequest struct {
	DataType string `json:"data_type"`
	DataItem string `json:"data_item"`
}

type verifyHashBatchRequest struct {
	DataType string   `json:"data_type"`
	Level    int      `json:"level"`
	Start    int64    `json:"start"`
	Hashes   []string `json:"hashes"`
}

type verifyHashBatchResponse struct {
	DataType   string `json:"data_type"`
	Level      int    `json:"level"`
	Start      int64  `json:"start"`
	Count      int    `json:"count"`
	Results    []int  `json:"results"`
	Matches    int    `json:"matches"`
	Mismatches int    `json:"mismatches"`
	Error      string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (resp kayrosRecordApi) toKayrosRecord() (*KayrosRecord, error) {
	if resp.DataType == "" || len(resp.HashItem) == 0 {
		return nil, fmt.Errorf("Kayros API record not found")
	}
	uuidHex, err := uuidStringToHex(resp.TS)
	if err != nil {
		return nil, fmt.Errorf("invalid ts uuid %q: %w", resp.TS, err)
	}
	timestamp, err := TimeuuidHexToTimestamp(uuidHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp from uuid %q: %w", resp.TS, err)
	}

	return &KayrosRecord{
		DataType:    resp.DataType,
		DataTypeHex: hex.EncodeToString([]byte(resp.DataType)),
		DataItem:    resp.DataItem,
		UuidHex:     uuidHex,
		HashItem:    resp.HashItem,
		PrevHash:    resp.PrevHash,
		HashType:    resp.HashType,
		Position:    resp.Position,
		Timestamp:   timestamp,
	}, nil
}

func parseKayrosError(respData []byte) (string, bool) {
	var errResp KayrosErrorResponse
	if err := json.Unmarshal(respData, &errResp); err != nil {
		return "", false
	}
	if errResp.Error != "" {
		return errResp.Error, true
	}
	if errResp.Message != "" {
		return errResp.Message, true
	}
	return "", false
}

func base64ToHex(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty value")
	}
	bz, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bz), nil
}

func decodeBase64Optional(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []byte{}, nil
	}
	return base64.StdEncoding.DecodeString(value)
}

func uuidStringToHex(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty uuid")
	}
	value = strings.ReplaceAll(value, "-", "")
	if len(value) != 32 {
		return "", fmt.Errorf("invalid uuid length")
	}
	_, err := hex.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("invalid uuid hex %q: %w", value, err)
	}
	return strings.ToLower(value), nil
}

func int64Ptr(value int64) *int64 {
	return &value
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "status 404") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "record not found")
}
