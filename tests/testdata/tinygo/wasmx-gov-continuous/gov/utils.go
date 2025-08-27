package gov

import (
	"math/big"

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

// Big integer utilities
func NewBigZero() *big.Int {
	return new(big.Int)
}

func NewBigFromString(s string) *big.Int {
	z := new(big.Int)
	z.SetString(s, 10)
	return z
}

func NewBigFromUint64(i uint64) *big.Int {
	return new(big.Int).SetUint64(i)
}

func NewBigPow10(exp int) *big.Int {
	result := new(big.Int)
	base := big.NewInt(10)
	exponent := big.NewInt(int64(exp))
	return result.Exp(base, exponent, nil)
}

// Utility constants
const MaxMetadataLen = 10000
