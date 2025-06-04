package wasmx

// #include <stdlib.h>
import "C"

import (
	"bytes"
	"encoding/json"
	"unsafe"
)

//go:wasmimport imap ConnectWithPassword
func ConnectWithPassword_(reqPtr int64) int64

//go:wasmimport imap ConnectOAuth2
func ConnectOAuth2_(reqPtr int64) int64

//go:wasmimport imap Close
func Close_(reqPtr int64) int64

//go:wasmimport imap Listen
func Listen_(reqPtr int64) int64

//go:wasmimport imap Count
func Count_(reqPtr int64) int64

//go:wasmimport imap UIDSearch
func UIDSearch_(reqPtr int64) int64

//go:wasmimport imap ListMailboxes
func ListMailboxes_(reqPtr int64) int64

//go:wasmimport imap Fetch
func Fetch_(reqPtr int64) int64

//go:wasmimport imap CreateFolder
func CreateFolder_(reqPtr int64) int64

func StorageStore(key, value []byte) {
	Log([]byte("storagestore"), [][32]byte{})
	keyPtr, keyLength := BytesToPtr(key)
	valuePtr, valueLength := BytesToPtr(value)
	StorageStore_(keyPtr, keyLength, valuePtr, valueLength)
}

func StorageLoad(key []byte) []byte {
	keyPtr, keyLength := BytesToPtr(key)
	valPtr, valLen := StorageLoad_(keyPtr, keyLength)
	return PtrToBytes(valPtr, valLen)
}

func GetCallData() []byte {
	ptr, len := GetCallData_()
	return PtrToBytes(ptr, len)
}

func SetFinishData(data []byte) {
	keyPtr, keyLength := BytesToPtr(data)
	SetFinishData_(keyPtr, keyLength)
}

func SetReturnData(data []byte) {
	keyPtr, keyLength := BytesToPtr(data)
	SetReturnData_(keyPtr, keyLength)
}

func SetExitCode(code int32, data []byte) {
	keyPtr, keyLength := BytesToPtr(data)
	SetExitCode_(code, keyPtr, keyLength)
}

func Bech32StringToBytes(addrBech32 string) []byte {
	addrStrPtr, addrStrLen := StringToPtr(addrBech32)
	ptr := Bech32StringToBytes_(addrStrPtr, addrStrLen)
	return PtrToBytes(ptr, 32)
}

func Bech32BytesToString(addr []byte) string {
	addrPtr, _ := BytesToPtr(PaddLeftTo32(addr))
	ptr, len := Bech32BytesToString_(addrPtr)
	data := PtrToBytes(ptr, len)
	return string(data)
}

func Call(gasLimit int64, addrBech32 string, value []byte, calldata []byte) (bool, []byte) {
	addrStrPtr, addrStrLen := StringToPtr(addrBech32)
	addrPtr := Bech32StringToBytes_(addrStrPtr, addrStrLen)

	valuePtr, _ := BytesToPtr(value)
	calldPtr, calldLength := BytesToPtr(calldata)

	ptr, len := CallClassic_(gasLimit, addrPtr, valuePtr, calldPtr, calldLength)
	res := PtrToBytes(ptr, len)

	var calld CallResult
	err := json.Unmarshal(res, &calld)
	if err != nil {
		panic("Cannot decode json")
	}
	return calld.Success == 0, calld.Data
}

func CallStatic(gasLimit int64, addrBech32 string, calldata []byte) (bool, []byte) {
	addrStrPtr, addrStrLen := StringToPtr(addrBech32)
	addrPtr := Bech32StringToBytes_(addrStrPtr, addrStrLen)

	calldPtr, calldLength := BytesToPtr(calldata)

	ptr, len := CallStatic_(gasLimit, addrPtr, calldPtr, calldLength)
	res := PtrToBytes(ptr, len)

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
	ptr, size := BytesToPtr(encoded)
	Log_(ptr, size)
}

func Println(data string) {
	ptr, size := BytesToPtr([]byte(data))
	Println_(ptr, size)
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
	return string(PtrToBytes(ptr, size))
}

func StringToPtr(s string) (*uint8, uint32) {
	buf := []byte(s)                 // convert string to []byte
	return &buf[0], uint32(len(buf)) // return pointer to buffer and length
}

func BytesToPtr(data []byte) (*uint8, uint32) {
	buf := make([]uint8, len(data))  // allocate buffer
	copy(buf, data)                  // copy data into it
	return &buf[0], uint32(len(buf)) // return pointer and size
}

func PtrToBytes(ptr *uint8, size uint32) []byte {
	return unsafe.Slice(ptr, size)
}

// *uint8, uint32
// return int32(uintptr(unsafe.Pointer(ptr))), int32(len(buf))
