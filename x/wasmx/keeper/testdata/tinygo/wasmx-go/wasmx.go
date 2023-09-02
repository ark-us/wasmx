package wasmx

// #include <stdlib.h>
import "C"

import (
	"bytes"
	"encoding/json"
	"reflect"
	"unsafe"
)

//go:wasmimport wasmx storageStore
func StorageStore_(keyPtr, keyLen, valuePtr, valueLen uint32)

//go:wasmimport wasmx storageLoad
func StorageLoad_(keyPtr, keyLen uint32) uint64

//go:wasmimport wasmx getCallData
func GetCallData_() uint64

//go:wasmimport wasmx setReturnData
func SetReturnData_(dataPtr, dataLen uint32)

//go:wasmimport wasmx getEnv
func GetEnv_() uint64

//go:wasmimport wasmx callClassic
func CallClassic_(gasLimit uint64, addressPtr, valuePtr, calldPtr, calldLen uint32) uint64

//go:wasmimport wasmx callStatic
func CallStatic_(gasLimit uint64, addressPtr, calldPtr, calldLen uint32) uint64

//go:wasmimport wasmx getBlockHash
func GetBlockHash_(blockNumber uint64) uint64

//go:wasmimport wasmx getAccount
func GetAccount_(addrPtr uint32) uint64

//go:wasmimport wasmx getCodeHash
func GetCodeHash_(addrPtr uint32) uint64

//go:wasmimport wasmx getBalance
func GetBalance_(addrPtr uint32) uint64

//go:wasmimport wasmx keccak256
func Keccak256_(dataPtr, dataLen uint32) uint64

//go:wasmimport wasmx createAccount
func CreateAccount_(dataPtr, dataLen uint32) uint64

//go:wasmimport wasmx createAccount2
func CreateAccount2_(dataPtr, dataLen uint32) uint64

//go:wasmimport wasmx sendCosmosMsg
func SendCosmosMsg_(dataPtr, dataLen uint32) uint64

//go:wasmimport wasmx sendCosmosQuery
func SendCosmosQuery_(dataPtr, dataLen uint32) uint64

//go:wasmimport wasmx getGasLeft
func GetGasLeft_() uint64

//go:wasmimport wasmx bech32StringToBytes
func Bech32StringToBytes_(dataPtr, dataLen uint32) uint32

//go:wasmimport wasmx bech32BytesToString
func Bech32BytesToString_(dataPtr uint32) uint64

//go:wasmimport wasmx log
func Log_(ptr, size uint32)

type CallResult struct {
	Success int    `json:"success"`
	Data    []byte `json:"data"`
}

func StorageStore(key, value []byte) {
	keyPtr, keyLength := BytesToLeakedPtr(key)
	valuePtr, valueLength := BytesToLeakedPtr(value)
	StorageStore_(keyPtr, keyLength, valuePtr, valueLength)
}

func StorageLoad(key []byte) []byte {
	keyPtr, keyLength := BytesToLeakedPtr(key)
	ptr := StorageLoad_(keyPtr, keyLength)
	return bytesFromDynPtr(ptr)
}

func GetCallData() []byte {
	ptr := GetCallData_()
	return bytesFromDynPtr(ptr)
}

func SetReturnData(data []byte) {
	keyPtr, keyLength := BytesToLeakedPtr(data)
	SetReturnData_(keyPtr, keyLength)
}

func Bech32StringToBytes(addrBech32 string) []byte {
	addrStrPtr, addrStrLen := StringToLeakedPtr(addrBech32)
	ptr := Bech32StringToBytes_(addrStrPtr, addrStrLen)
	return PtrToBytes(ptr, 32)
}

func Bech32BytesToString(addr []byte) string {
	addrPtr, _ := BytesToLeakedPtr(PaddLeftTo32(addr))
	ptr := Bech32BytesToString_(addrPtr)
	data := bytesFromDynPtr(ptr)
	return string(data)
}

func Call(gasLimit uint64, addrBech32 string, value []byte, calldata []byte) (bool, []byte) {
	addrStrPtr, addrStrLen := StringToLeakedPtr(addrBech32)
	addrPtr := Bech32StringToBytes_(addrStrPtr, addrStrLen)

	valuePtr, _ := BytesToLeakedPtr(value)
	calldPtr, calldLength := BytesToLeakedPtr(calldata)

	ptr := CallClassic_(gasLimit, addrPtr, valuePtr, calldPtr, calldLength)
	res := bytesFromDynPtr(ptr)

	var calld CallResult
	err := json.Unmarshal(res, &calld)
	if err != nil {
		panic("Cannot decode json")
	}
	return calld.Success == 0, calld.Data
}

func CallStatic(gasLimit uint64, addrBech32 string, calldata []byte) (bool, []byte) {
	addrStrPtr, addrStrLen := StringToLeakedPtr(addrBech32)
	addrPtr := Bech32StringToBytes_(addrStrPtr, addrStrLen)

	calldPtr, calldLength := BytesToLeakedPtr(calldata)

	ptr := CallStatic_(gasLimit, addrPtr, calldPtr, calldLength)
	res := bytesFromDynPtr(ptr)

	var calld CallResult
	err := json.Unmarshal(res, &calld)
	if err != nil {
		panic("Cannot decode json")
	}
	return calld.Success == 0, calld.Data
}

type WasmxLog struct {
	Data   []byte
	Topics [][32]byte
}

// log a message to the console using _log.
func Log(data []byte, topics [][32]byte) {
	encoded, _ := json.Marshal(WasmxLog{Data: data, Topics: topics})
	ptr, size := BytesToLeakedPtr(encoded)
	Log_(ptr, size)
}

func splitPtr(ptr uint64) (uint32, uint32) {
	dataPtr := uint32(ptr >> 32)
	dataSize := uint32(ptr)
	return dataPtr, dataSize
}

func bytesFromDynPtr(ptr uint64) []byte {
	dataPtr, dataLen := splitPtr(ptr)
	return PtrToBytes(dataPtr, dataLen)
}

func PaddLeftTo32(data []byte) []byte {
	length := len(data)
	if length >= 32 {
		return data
	}
	data = append(bytes.Repeat([]byte{0}, 32-length), data...)
	return data
}

// PtrToString returns a string from WebAssembly compatible numeric types
// representing its pointer and length.
func PtrToString(ptr uint32, size uint32) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

// StringToPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
// The returned pointer aliases the string hence the string must be kept alive
// until ptr is no longer needed.
func StringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}

// StringToLeakedPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
// The pointer is not automatically managed by TinyGo hence it must be freed by the host (TODO)
func StringToLeakedPtr(s string) (uint32, uint32) {
	size := C.ulong(len(s))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), s)
	return uint32(uintptr(ptr)), uint32(size)
}

func BytesToLeakedPtr(data []byte) (uint32, uint32) {
	size := C.ulong(len(data))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), data)
	return uint32(uintptr(ptr)), uint32(size)
}

func PtrToBytes(ptr uint32, size uint32) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(ptr),
		Len:  uintptr(size),
		Cap:  uintptr(size),
	}))
}
