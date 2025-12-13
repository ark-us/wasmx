package lib

import (
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

// GetBlockTime returns the current block time
func GetBlockTime() int64 {
	block := wasmx.GetCurrentBlock()
	return int64(block.Height)
}
