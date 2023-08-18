package wasmx

// #include <stdlib.h>
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

//go:wasmimport wasmx storageStore2
func StorageStore_(keyPtr, keyLen, valuePtr, valueLen uint32)

//go:wasmimport wasmx storageLoad2
func StorageLoad_(keyPtr, keyLen uint32) uint64

//go:wasmimport wasmx getCallData
func GetCallData_() uint64

//go:wasmimport wasmx setReturnData
func SetReturnData_(dataPtr, dataLen uint32)

//go:wasmimport wasmx log
func Log_(ptr, size uint32)

func StorageStore(key, value string) {
	Log(fmt.Sprintf("tinygo!! storageStore %s, %s", key, value))
	keyPtr, keyLength := StringToLeakedPtr(key)
	valuePtr, valueLength := StringToLeakedPtr(value)
	StorageStore_(keyPtr, keyLength, valuePtr, valueLength)
}

func StorageLoad(key string) string {
	keyPtr, keyLength := StringToLeakedPtr(key)
	ptr := StorageLoad_(keyPtr, keyLength)
	dataPtr := uint32(ptr >> 32)
	dataSize := uint32(ptr)
	return PtrToString(dataPtr, dataSize)
}

func GetCallData() string {
	ptr := GetCallData_()
	dataPtr := uint32(ptr >> 32)
	dataSize := uint32(ptr)
	data := PtrToString(dataPtr, dataSize)
	return data
}

func SetReturnData(data string) {
	keyPtr, keyLength := StringToLeakedPtr(data)
	SetReturnData_(keyPtr, keyLength)
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
// The pointer is not automatically managed by TinyGo hence it must be freed by the host.
func StringToLeakedPtr(s string) (uint32, uint32) {
	size := C.ulong(len(s))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), s)
	return uint32(uintptr(ptr)), uint32(size)
}
