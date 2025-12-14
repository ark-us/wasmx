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
	block := wasmx.GetCurrentBlock()
	return int64(block.Timestamp)
}

// AddressInList checks if an address is in a list of addresses
func AddressInList(address string, list []string) bool {
	for _, addr := range list {
		if addr == address {
			return true
		}
	}
	return false
}

// RemoveAddressFromList removes an address from a list
func RemoveAddressFromList(address string, list []string) []string {
	result := make([]string, 0, len(list))
	for _, addr := range list {
		if addr != address {
			result = append(result, addr)
		}
	}
	return result
}
