package wasmx

// #include <stdlib.h>
import "C"

import (
	"bytes"
	"encoding/json"
	"math/big"
	"reflect"
	"unsafe"

	sdkmath "cosmossdk.io/math"
)

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

func StorageStore(key, value []byte) {
	Log([]byte("storagestore"), [][32]byte{})
	StorageStore_(BytesToPackedPtr(key), BytesToPackedPtr(value))
}

func StorageLoad(key []byte) []byte {
	packed := StorageLoad_(BytesToPackedPtr(key))
	return PackedPtrToBytes(packed)
}

func GetCallData() []byte {
	packed := GetCallData_()
	return PackedPtrToBytes(packed)
}

func SetFinishData(data []byte) {
	SetFinishData_(BytesToPackedPtr(data))
}

func SetReturnData(data []byte) {
	SetReturnData_(BytesToPackedPtr(data))
}

func Finish(data []byte) {
	Finish_(BytesToPackedPtr(data))
}

func Revert(data []byte) {
	Revert_(BytesToPackedPtr(data))
}

func Bech32StringToBytes(addrBech32 string) []byte {
	ptr := Bech32StringToBytes_(StringToPackedPtr(addrBech32))
	return PackedPtrToBytes(ptr)
}

func Bech32BytesToString(addr []byte) string {
	packed := Bech32BytesToString_(BytesToPackedPtr(PaddLeftTo32(addr)))
	data := PackedPtrToBytes(packed)
	return string(data)
}

func CallInternal(addrBech32 string, value *sdkmath.Int, calldata []byte, gasLimit *big.Int, isQuery bool) (bool, []byte) {
	req := &SimpleCallRequestRaw{
		To:       addrBech32,
		Value:    value,
		GasLimit: gasLimit,
		Calldata: calldata,
		IsQuery:  isQuery,
	}

	reqbz, err := json.Marshal(req)
	if err != nil {
		Revert([]byte(err.Error()))
	}
	ptr := BytesToPackedPtr(reqbz)
	packed := Call_(ptr)
	res := PackedPtrToBytes(packed)

	var calld CallResult
	err = json.Unmarshal(res, &calld)
	if err != nil {
		Revert([]byte("Cannot decode json: " + err.Error()))
	}
	return calld.Success == 0, calld.Data
}

func Call(addrBech32 string, value *sdkmath.Int, calldata []byte, gasLimit *big.Int) (bool, []byte) {
	return CallInternal(addrBech32, value, calldata, gasLimit, false)
}

func CallStatic(addrBech32 string, calldata []byte, gasLimit *big.Int) (bool, []byte) {
	return CallInternal(addrBech32, nil, calldata, gasLimit, false)
}

type WasmxLog struct {
	Data   []byte
	Topics [][32]byte
}

// log a message to the console using _log.
func Log(data []byte, topics [][32]byte) {
	encoded, _ := json.Marshal(WasmxLog{Data: data, Topics: topics})
	Log_(BytesToPackedPtr(encoded))
}

func LoggerInfo(msg string, parts []string) {
	LoggerInfo_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerError(msg string, parts []string) {
	LoggerError_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDebug(msg string, parts []string) {
	LoggerDebug_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDebugExtended(msg string, parts []string) {
	LoggerDebugExtended_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDataToPackedPtr(msg string, parts []string) int64 {
	databz, err := json.Marshal(&LoggerLog{Msg: msg, Parts: parts})
	if err != nil {
		Revert([]byte("cannot marshal LoggerLog" + err.Error()))
	}
	return BytesToPackedPtr(databz)
}

func PaddLeftTo32(data []byte) []byte {
	length := len(data)
	if length >= 32 {
		return data
	}
	data = append(bytes.Repeat([]byte{0}, 32-length), data...)
	return data
}

func PtrToString(ptr *uint8, size uint32) string {
	// return string(PtrToBytes(ptr, size))
	return unsafe.String(ptr, size)
}

func StringToPtr(s string) (*uint8, uint32) {
	size := C.ulong(len(s))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), s)
	return (*uint8)(ptr), uint32(size)
}

func StringToPackedPtr(s string) int64 {
	ptr, len := StringToPtr(s)
	return PackPtr(ptr, len)
}

func BytesToPackedPtr(data []byte) int64 {
	ptr, len := BytesToPtr(data)
	return PackPtr(ptr, len)
}

func BytesToPtr(data []byte) (*uint8, uint32) {
	size := C.ulong(len(data))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), data)
	return (*uint8)(ptr), uint32(size)
}

func PtrToBytes(ptr *uint8, size uint32) []byte {
	bz := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(ptr)),
		Len:  int(size),
		Cap:  int(size),
	}))
	out := make([]byte, size)
	copy(out, bz)
	return out
}

func SplitPtr(packed int64) (*uint8, uint32) {
	ptr := uint32(packed >> 32)
	size := uint32(packed & 0xffffffff)
	return (*uint8)(unsafe.Pointer(uintptr(ptr))), size
}

func PackPtr(ptr *uint8, size uint32) int64 {
	offset := uint32(uintptr(unsafe.Pointer(ptr))) // convert pointer to memory offset
	return (int64(offset) << 32) | int64(size)
}

func PackedPtrToBytes(ptr int64) []byte {
	dataPtr, dataLen := SplitPtr(ptr)
	return PtrToBytes(dataPtr, dataLen)
}
