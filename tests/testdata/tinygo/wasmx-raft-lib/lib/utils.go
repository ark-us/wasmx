package lib

import (
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// Simple helpers for logging and conversions
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
	// AS: LoggerDebug("revert", ["err", message, "module", MODULE_NAME])
	LoggerDebug("revert", []string{"err", message, "module", MODULE_NAME})
	wasmx.RevertWithModule(MODULE_NAME, message)
}

func Int32ToString(v int32) string { return fmt.Sprintf("%d", v) }
func Int64ToString(v int64) string { return fmt.Sprintf("%d", v) }
