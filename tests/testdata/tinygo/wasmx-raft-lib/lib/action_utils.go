package lib

import (
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "strconv"

    wasmx "github.com/loredanacirstea/wasmx-env/lib"
    wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
)

// GetMajority: wrapper to storage's majority
func GetMajority(count int) int64 { return int64(count/2) + 1 }

// GetRandomInRange: deterministic-ish using block timestamp
func GetRandomInRange(min int64, max int64) (int64, error) {
    if max < min { return 0, errors.New("invalid range") }
    if min == max { return min, nil }
    ts := wasmx.GetTimestamp().UnixNano()
    span := max - min + 1
    if span <= 0 { return min, nil }
    return (ts%span + min), nil
}

// SignMessage signs string message with current state's ed25519 privkey (base64)
func SignMessage(msg string) (string, error) {
    st, err := GetCurrentState()
    if err != nil { return "", err }
    if st.ValidatorPrivkey == "" { return "", errors.New("empty validator privkey") }
    privbz, err := base64.StdEncoding.DecodeString(st.ValidatorPrivkey)
    if err != nil { return "", err }
    sig := wasmx.Ed25519Sign(privbz, []byte(msg))
    return base64.StdEncoding.EncodeToString(sig), nil
}

// PrepareAppendEntryMessage builds the JSON payload to send over gRPC, base64-encoded
func PrepareAppendEntryMessage(nodeId int32, nextIndex int64, lastIndex int64, lastIndexToSend int64, node NodeInfo, data AppendEntry) (string, error) {
    datastrBz, err := json.Marshal(&data)
    if err != nil { return "", err }
    datastr := string(datastrBz)
    signature, err := SignMessage(datastr)
    if err != nil { return "", err }
    dataBase64 := base64.StdEncoding.EncodeToString([]byte(datastr))
    msgstr := fmt.Sprintf(`{"run":{"event":{"type":"receiveHeartbeat","params":[{"key":"entry","value":"%s"},{"key":"signature","value":"%s"}]}}}`, dataBase64, signature)
    wasmx.LoggerDebug(MODULE_NAME, "diseminate append entry...", []string{"nodeId", Int32ToString(nodeId), "receiver", node.Address, "from", Int64ToString(nextIndex), "to", Int64ToString(lastIndexToSend), "last_index", Int64ToString(lastIndex)})
    return msgstr, nil
}

// SendGrpcJSONBase64 sends a base64-encoded JSON string via wasmxcore GrpcRequest
func SendGrpcJSONBase64(ip string, contract wasmx.Bech32String, jsonStr string) (*wasmxcore.GrpcResponse, error) {
    encoded := base64.StdEncoding.EncodeToString([]byte(jsonStr))
    return wasmxcore.GrpcRequest(ip, contract, encoded)
}

// Helpers to extract a param value by key
func GetParam(params []string, key string) (string, bool) {
    return "", false
}

// Parse string to int helpers
func ParseI64(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }
func ParseI32(s string) (int32, error) {
    v, err := strconv.ParseInt(s, 10, 32)
    return int32(v), err
}

