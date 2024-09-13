package taylor

import (
	"encoding/binary"
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"

	mem "mythos/v1/x/wasmx/vm/memory/common"
)

const AS_PTR_LENGHT_OFFSET = int32(4)
const AS_ARRAY_BUFFER_TYPE = int32(1)
const MEMORY_EXPORT_ALLOCATE = "alloc_buffer"
const BUFFER_VALUE_OFFSET = 12

// BUFFER: ptr - 4 bytes ref|type - 4 bytes length - 4 bytes value ptr - value

type MemoryHandlerTay struct{}

func (MemoryHandlerTay) ReadMemFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error) {
	return ReadMemFromPtr(callframe, pointer)
}
func (MemoryHandlerTay) AllocateWriteMem(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (int32, error) {
	return AllocateWriteMem(vm, callframe, data)
}
func (MemoryHandlerTay) ReadJsString(arr []byte) string {
	return ReadJsString(arr)
}
func (MemoryHandlerTay) ReadStringFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) (string, error) {
	bz, err := mem.ReadMemUntilNull(callframe, pointer)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func ReadMemFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error) {
	lengthbz, err := mem.ReadMem(callframe, pointer.(int32)+AS_PTR_LENGHT_OFFSET, int32(AS_PTR_LENGHT_OFFSET))
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lengthbz)
	data, err := mem.ReadMem(callframe, pointer.(int32)+BUFFER_VALUE_OFFSET, int32(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AllocateMemVm(vm *wasmedge.VM, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	result, err := vm.Execute(MEMORY_EXPORT_ALLOCATE, size)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func AllocateWriteMem(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (int32, error) {
	ptr, err := AllocateMemVm(vm, int32(len(data)))
	if err != nil {
		return ptr, err
	}
	valptr := ptr + BUFFER_VALUE_OFFSET
	err = mem.WriteMem(callframe, data, valptr)
	if err != nil {
		return ptr, err
	}
	return ptr, nil
}

func ReadJsString(arr []byte) string {
	return string(arr)
}
