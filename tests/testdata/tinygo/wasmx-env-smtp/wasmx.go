package wasmx

// #include <stdlib.h>
import "C"

import (
	"bytes"
	"encoding/json"
	"unsafe"
)

//go:wasmimport wasmx storageStore
func StorageStore_(keyPtr *uint8, keyLen uint32, valuePtr *uint8, valueLen uint32)

//go:wasmimport wasmx storageLoad
func StorageLoad_(keyPtr *uint8, keyLen uint32) (*uint8, uint32)

//go:wasmimport wasmx getCallData
func GetCallData_() (*uint8, uint32)

//go:wasmimport wasmx setFinishData
func SetFinishData_(dataPtr *uint8, dataLen uint32)

//go:wasmimport wasmx setReturnData
func SetReturnData_(dataPtr *uint8, dataLen uint32)

//go:wasmimport wasmx setExitCode
func SetExitCode_(code int32, dataPtr *uint8, dataLen uint32)

//go:wasmimport wasmx getEnv
func GetEnv_() (*uint8, uint32)

//go:wasmimport wasmx callClassic
func CallClassic_(gasLimit int64, addressPtr *uint8, valuePtr *uint8, calldPtr *uint8, calldLen uint32) (*uint8, uint32)

//go:wasmimport wasmx callStatic
func CallStatic_(gasLimit int64, addressPtr *uint8, calldPtr *uint8, calldLen uint32) (*uint8, uint32)

//go:wasmimport wasmx getBlockHash
func GetBlockHash_(blockNumber int64) (*uint8, uint32)

//go:wasmimport wasmx getAccount
func GetAccount_(addrPtr *uint8) (*uint8, uint32)

//go:wasmimport wasmx getCodeHash
func GetCodeHash_(addrPtr *uint8) (*uint8, uint32)

//go:wasmimport wasmx getBalance
func GetBalance_(addrPtr *uint8) (*uint8, uint32)

//go:wasmimport wasmx keccak256
func Keccak256_(dataPtr *uint8, dataLen uint32) (*uint8, uint32)

//go:wasmimport wasmx createAccount
func CreateAccount_(dataPtr *uint8, dataLen uint32) (*uint8, uint32)

//go:wasmimport wasmx createAccount2
func CreateAccount2_(dataPtr *uint8, dataLen uint32) (*uint8, uint32)

//go:wasmimport wasmx sendCosmosMsg
func SendCosmosMsg_(dataPtr *uint8, dataLen uint32) (*uint8, uint32)

//go:wasmimport wasmx sendCosmosQuery
func SendCosmosQuery_(dataPtr *uint8, dataLen uint32) (*uint8, uint32)

//go:wasmimport wasmx getGasLeft
func GetGasLeft_() int64

//go:wasmimport wasmx bech32StringToBytes
func Bech32StringToBytes_(dataPtr *uint8, dataLen uint32) *uint8

//go:wasmimport wasmx bech32BytesToString
func Bech32BytesToString_(dataPtr *uint8) (*uint8, uint32)

//go:wasmimport wasmx log
func Log_(ptr *uint8, size uint32)

//go:wasmimport wasmx println
func Println_(ptr *uint8, size uint32)

type CallResult struct {
	Success int    `json:"success"`
	Data    []byte `json:"data"`
}

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
