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

//go:wasmimport wasmx getGasLeft
func GetGasLeft_() int64

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

// missing host APIs mirrored from AssemblyScript sdk

//go:wasmimport wasmx getEnv
func GetEnv_() int64

//go:wasmimport wasmx getCaller
func GetCaller_() int64

//go:wasmimport wasmx getAddress
func GetAddress_() int64

//go:wasmimport wasmx getCurrentBlock
func GetCurrentBlock_() int64

//go:wasmimport wasmx storageDelete
func StorageDelete_(keyPtr int64)

//go:wasmimport wasmx storageDeleteRange
func StorageDeleteRange_(reqPtr int64)

//go:wasmimport wasmx storageLoadRange
func StorageLoadRange_(reqPtr int64) int64

//go:wasmimport wasmx storageLoadRangePairs
func StorageLoadRangePairs_(reqPtr int64) int64

//go:wasmimport wasmx getFinishData
func GetFinishData_() int64

//go:wasmimport wasmx emitCosmosEvents
func EmitCosmosEvents_(ptr int64)

//go:wasmimport wasmx sha256
func Sha256_(ptr int64) int64

//go:wasmimport wasmx MerkleHash
func MerkleHash_(ptr int64) int64

//go:wasmimport wasmx ed25519Sign
func Ed25519Sign_(privPtr int64, msgPtr int64) int64

//go:wasmimport wasmx ed25519Verify
func Ed25519Verify_(pubPtr int64, sigPtr int64, msgPtr int64) int64

//go:wasmimport wasmx ed25519PubToHex
func Ed25519PubToHex_(pubPtr int64) int64

//go:wasmimport wasmx validate_bech32_address
func ValidateBech32Address_(ptr int64) int64

//go:wasmimport wasmx addr_canonicalize
func Bech32StringToBytes_(dataPtr int64) int64

//go:wasmimport wasmx addr_humanize
func Bech32BytesToString_(dataPtr int64) int64

//go:wasmimport wasmx addr_equivalent
func AddrEquivalent_(addr1 int64, addr2 int64) int64

//go:wasmimport wasmx addr_humanize_mc
func AddrHumanizeMC_(addrPtr int64, prefixPtr int64) int64

//go:wasmimport wasmx addr_canonicalize_mc
func AddrCanonicalizeMC_(strPtr int64) int64

//go:wasmimport wasmx getAddressByRole
func GetAddressByRole_(rolePtr int64) int64

//go:wasmimport wasmx getRoleByAddress
func GetRoleByAddress_(addrPtr int64) int64

//go:wasmimport wasmx executeCosmosMsg
func ExecuteCosmosMsg_(ptr int64) int64

//go:wasmimport wasmx decodeCosmosTxToJson
func DecodeCosmosTxToJson_(ptr int64) int64

//go:wasmimport wasmx verifyCosmosTx
func VerifyCosmosTx_(ptr int64) int64
