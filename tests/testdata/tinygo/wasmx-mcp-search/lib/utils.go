package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// Revert reverts the transaction with an error message
func Revert(msg string) {
	wasmx.Revert([]byte(msg))
}

// LoggerInfo logs info message
func LoggerInfo(msg string, args []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, args)
}

// LoggerError logs error message
func LoggerError(msg string, args []string) {
	wasmx.LoggerError(MODULE_NAME, msg, args)
}

// LoggerDebug logs debug message
func LoggerDebug(msg string, args []string) {
	wasmx.LoggerDebug(MODULE_NAME, msg, args)
}
