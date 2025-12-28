package lib

import (
	"encoding/base64"
	"encoding/hex"

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

// HexHashToBase64 converts a hex-encoded hash to base64 format
// This is needed because mempool stores hashes in base64, but Kayros uses hex
func HexHashToBase64(hexHash string) (string, error) {
	hashBytes, err := hex.DecodeString(hexHash)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hashBytes), nil
}
