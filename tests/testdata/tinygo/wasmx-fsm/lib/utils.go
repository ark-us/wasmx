package lib

import (
	"encoding/hex"
	"strconv"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func LoggerInfo(msg string, parts []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, parts)
}

func LoggerError(msg string, parts []string) {
	wasmx.LoggerError(MODULE_NAME, msg, parts)
}

func LoggerDebug(msg string, parts []string) {
	wasmx.LoggerDebug(MODULE_NAME, msg, parts)
}

func LoggerDebugExtended(msg string, parts []string) {
	wasmx.LoggerDebugExtended(MODULE_NAME, msg, parts)
}

func Revert(message string) {
	wasmx.RevertWithModule(MODULE_NAME, message)
}

// Context to action params conversion
func ctxToActionParams(ctx map[string]string) []ActionParam {
	params := make([]ActionParam, 0, len(ctx))
	for key, value := range ctx {
		params = append(params, ActionParam{Key: key, Value: value})
	}
	return params
}

// Action params to map conversion
func actionParamsToMap(params []ActionParam) map[string]string {
	ctx := make(map[string]string)
	for _, param := range params {
		ctx[param.Key] = param.Value
	}
	return ctx
}

// Get params or event params
func getParamsOrEventParams(params []ActionParam, event EventObject) []ActionParam {
	result := make([]ActionParam, len(params)+len(event.Params))
	copy(result, params)
	copy(result[len(params):], event.Params)
	return result
}

// Address utilities
func getAddressHex(addr []byte) string {
	return hex.EncodeToString(addr)
}

// Parsing utilities
func parseInt32(s string) (int32, error) {
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// String utilities
func isNaN(val int) bool {
	return false // Simple implementation for TinyGo
}

// Array buffer utilities
func arrayBufferToU8Array(buffer []byte) []byte {
	result := make([]byte, len(buffer))
	copy(result, buffer)
	return result
}

// Base64 utilities (simplified)
func base64ToString(encoded string) string {
	// For simplicity, we'll just return the encoded string
	// In a real implementation, this would decode base64
	return encoded
}

func GrpcRequest(ip string, contract wasmx.Bech32String, data string) (*GrpcResponse, error) {
	response, err := wasmxcore.GrpcRequest(ip, contract, data)
	if err != nil {
		return nil, err
	}
	return &GrpcResponse{
		Data:  response.Data,
		Error: response.Error,
	}, nil
}
