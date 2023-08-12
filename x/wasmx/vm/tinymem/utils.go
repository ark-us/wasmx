// from https://github.com/tetratelabs/tinymem/blob/main/tinymem.go

package tinymem

import (
	"unsafe"
)

// PtrToString returns a string from WebAssembly compatible numeric types
// representing its pointer and length.
func PtrToString(ptr uintptr, size uint32) string {
	// Get a slice view of the underlying bytes in the stream.
	s := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
	return *(*string)(unsafe.Pointer(&s))
}

// StringToPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
func StringToPtr(s string) (uintptr, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return unsafePtr, uint32(len(buf))
}
