package wasmx

// #include <stdlib.h>
import "C"

// tinygo does not support multi value return
// so we need to pack (*uint8, uint32) into a int64

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func wasmx_env_i64_2() {}

//go:wasmimport wasmx storageStore
func StorageStore_(keyPtr int64, valuePtr int64)

//go:wasmimport wasmx storageLoad
func StorageLoad_(keyPtr int64) int64

//go:wasmimport wasmx getCallData
func GetCallData_() int64

//go:wasmimport wasmx setFinishData
func SetFinishData_(dataPtr int64)

//go:wasmimport wasmx setReturnData
func SetReturnData_(dataPtr int64)

//go:wasmimport wasmx finish
func Finish_(dataPtr int64)

//go:wasmimport wasmx revert
func Revert_(dataPtr int64)

//go:wasmimport wasmx call
func Call_(reqPtr int64) int64

//go:wasmimport wasmx getChainId
func GetChainId_() int64

//go:wasmimport wasmx getBlockHash
func GetBlockHash_(blockNumber int64) int64

//go:wasmimport wasmx getAccount
func GetAccount_(addrPtr int64) int64

//go:wasmimport wasmx getCodeHash
func GetCodeHash_(addrPtr int64) int64

//go:wasmimport wasmx getBalance
func GetBalance_(addrPtr int64) int64

//go:wasmimport wasmx keccak256
func Keccak256_(dataPtr int64) int64

//go:wasmimport wasmx createAccount
func CreateAccount_(dataPtr int64) int64

//go:wasmimport wasmx createAccount2
func CreateAccount2_(dataPtr int64) int64

//go:wasmimport wasmx sendCosmosMsg
func SendCosmosMsg_(dataPtr int64) int64

//go:wasmimport wasmx sendCosmosQuery
func SendCosmosQuery_(dataPtr int64) int64

//go:wasmimport wasmx getGasLeft
func GetGasLeft_() int64

//go:wasmimport wasmx addr_canonicalize
func Bech32StringToBytes_(dataPtr int64) int64

//go:wasmimport wasmx addr_humanize
func Bech32BytesToString_(dataPtr int64) int64

//go:wasmimport wasmx log
func Log_(ptr int64)

//go:wasmimport wasmx LoggerInfo
func LoggerInfo_(ptr int64)

//go:wasmimport wasmx LoggerError
func LoggerError_(ptr int64)

//go:wasmimport wasmx LoggerDebug
func LoggerDebug_(ptr int64)

//go:wasmimport wasmx LoggerDebugExtended
func LoggerDebugExtended_(ptr int64)
