package wasmx

// #include <stdlib.h>
import "C"

import (
	"encoding/json"
	"reflect"
	"runtime"
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

//go:wasmimport wasmx callClassic
func CallClassic_(gasLimit uint64, addressPtr, addressLen, valuePtr, calldPtr, calldLen uint32) uint64

//go:wasmimport wasmx callStatic
func CallStatic_(gasLimit uint64, addressPtr, addressLen, calldPtr, calldLen uint32) uint64

//go:wasmimport wasmx log
func Log_(ptr, size uint32)

type CallResult struct {
	Success int    `json:"success"`
	Data    []byte `json:"data"`
}

func StorageStore(key, value []byte) {
	// Log(fmt.Sprintf("tinygo!! storageStore %s, %s", key, value))
	keyPtr, keyLength := BytesToLeakedPtr(key)
	valuePtr, valueLength := BytesToLeakedPtr(value)
	StorageStore_(keyPtr, keyLength, valuePtr, valueLength)
}

func StorageLoad(key []byte) []byte {
	keyPtr, keyLength := BytesToLeakedPtr(key)
	ptr := StorageLoad_(keyPtr, keyLength)
	dataPtr := uint32(ptr >> 32)
	dataSize := uint32(ptr)
	return PtrToBytes(dataPtr, dataSize)
}

func GetCallData() []byte {
	ptr := GetCallData_()
	dataPtr := uint32(ptr >> 32)
	dataSize := uint32(ptr)
	data := PtrToBytes(dataPtr, dataSize)
	return data
}

func SetReturnData(data []byte) {
	keyPtr, keyLength := BytesToLeakedPtr(data)
	SetReturnData_(keyPtr, keyLength)
}

func Call(gasLimit uint64, addrBech32 string, value []byte, calldata []byte) (bool, []byte) {
	addrPtr, addrLen := StringToLeakedPtr(addrBech32)
	valuePtr, _ := BytesToLeakedPtr(value)
	calldPtr, calldLength := BytesToLeakedPtr(calldata)

	ptr := CallClassic_(gasLimit, addrPtr, addrLen, valuePtr, calldPtr, calldLength)
	dataPtr := uint32(ptr >> 32)
	dataSize := uint32(ptr)
	res := PtrToBytes(dataPtr, dataSize)

	var calld CallResult
	err := json.Unmarshal(res, &calld)
	if err != nil {
		panic("Cannot decode json")
	}
	return calld.Success == 0, calld.Data
}

func CallStatic(gasLimit uint64, addrBech32 string, calldata []byte) (bool, []byte) {
	addrPtr, addrLen := StringToLeakedPtr(addrBech32)
	calldPtr, calldLength := BytesToLeakedPtr(calldata)

	ptr := CallStatic_(gasLimit, addrPtr, addrLen, calldPtr, calldLength)
	dataPtr := uint32(ptr >> 32)
	dataSize := uint32(ptr)
	res := PtrToBytes(dataPtr, dataSize)

	var calld CallResult
	err := json.Unmarshal(res, &calld)
	if err != nil {
		panic("Cannot decode json")
	}
	return calld.Success == 0, calld.Data
}

// log a message to the console using _log.
func Log(message string) {
	ptr, size := StringToPtr(message)
	Log_(ptr, size)
	runtime.KeepAlive(message) // keep message alive until ptr is no longer needed.
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
