package lib

import wasmx "github.com/loredanacirstea/wasmx-env/lib"

func Revert(msg string) {
	wasmx.Revert([]byte(msg))
}

func LoggerInfo(msg string, args []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, args)
}

func LoggerError(msg string, args []string) {
	wasmx.LoggerError(MODULE_NAME, msg, args)
}
