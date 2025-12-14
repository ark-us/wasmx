package lib

import (
	"encoding/hex"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func LoggerInfo(msg string, args []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, args)
}

func LoggerError(msg string, args []string) {
	wasmx.LoggerError(MODULE_NAME, msg, args)
}

func LoggerDebug(msg string, args []string) {
	wasmx.LoggerDebug(MODULE_NAME, msg, args)
}

func Revert(message string) {
	LoggerDebug("revert", []string{"err", message, "module", MODULE_NAME})
	wasmx.RevertWithModule(MODULE_NAME, message)
}

// hexToBytes converts a hex string to bytes
func hexToBytes(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return []byte{}
	}
	return bytes
}
