package lib

import (
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// LoggerInfo logs an info message
func LoggerInfo(message string, keyvals []string) {
	wasmx.LoggerInfo(MODULE_NAME, message, keyvals)
}

// LoggerError logs an error message
func LoggerError(message string, keyvals []string) {
	wasmx.LoggerError(MODULE_NAME, message, keyvals)
}

// LoggerDebug logs an error message
func LoggerDebug(message string, keyvals []string) {
	wasmx.LoggerDebug(MODULE_NAME, message, keyvals)
}

func Revert(message string) {
	LoggerDebug("revert", []string{"err", message, "module", MODULE_NAME})
	wasmx.RevertWithModule(MODULE_NAME, message)
}

// GetBlockTime returns the current block time
func GetBlockTime() int64 {
	return time.Now().Unix()
}
