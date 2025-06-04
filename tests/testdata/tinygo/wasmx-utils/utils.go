package utils

// #include <stdlib.h>
import "C"

import (
	"bytes"
	"reflect"
	"unsafe"
)

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
