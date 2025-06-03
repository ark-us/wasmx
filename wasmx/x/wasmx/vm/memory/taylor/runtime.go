package taylor

import (
	"encoding/binary"
	"fmt"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

const AS_PTR_LENGHT_OFFSET = int32(4)
const AS_ARRAY_BUFFER_TYPE = int32(1)
const MEMORY_EXPORT_ALLOCATE = "alloc_buffer"
const BUFFER_VALUE_OFFSET = 12

// BUFFER: ptr - 4 bytes ref|type - 4 bytes length - 4 bytes value ptr - value

type RuntimeHandlerTay struct {
	vm memc.IVm
}

var _ memc.RuntimeHandler = (*RuntimeHandlerTay)(nil)

func NewRuntimeHandlerTay(vm memc.IVm) memc.RuntimeHandler {
	return RuntimeHandlerTay{vm}
}

func (h RuntimeHandlerTay) GetVm() memc.IVm {
	return h.vm
}

func (h RuntimeHandlerTay) GetMemory() (memc.IMemory, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return mem, nil
}

func (h RuntimeHandlerTay) PtrParamsLength() int {
	return 1
}

func (h RuntimeHandlerTay) ReadMemFromPtr(pointer []interface{}) ([]byte, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return ReadMemFromPtr(mem, pointer[0])
}
func (h RuntimeHandlerTay) AllocateWriteMem(data []byte) ([]interface{}, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return []interface{}{}, err
	}
	ptr, err := AllocateWriteMem(h.vm, mem, data)
	if err != nil {
		return []interface{}{}, err
	}
	return []interface{}{ptr}, nil
}
func (RuntimeHandlerTay) ReadJsString(arr []byte) string {
	return ReadJsString(arr)
}
func (h RuntimeHandlerTay) ReadStringFromPtr(pointer interface{}) (string, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return "", err
	}
	bz, err := memc.ReadMemUntilNull(mem, pointer)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func ReadMemFromPtr(mem memc.IMemory, pointer interface{}) ([]byte, error) {
	lengthbz, err := mem.Read(pointer.(int32)+AS_PTR_LENGHT_OFFSET, int32(AS_PTR_LENGHT_OFFSET))
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lengthbz)
	data, err := mem.Read(pointer.(int32)+BUFFER_VALUE_OFFSET, int32(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AllocateMemVm(vm memc.IVm, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no VM instance found")
	}
	result, err := vm.Call(MEMORY_EXPORT_ALLOCATE, []interface{}{size}, nil)
	if err != nil {
		return 0, err
	}
	return result[0], nil
}

func AllocateWriteMem(vm memc.IVm, mem memc.IMemory, data []byte) (int32, error) {
	ptr, err := AllocateMemVm(vm, int32(len(data)))
	if err != nil {
		return ptr, err
	}
	valptr := ptr + BUFFER_VALUE_OFFSET
	err = mem.Write(valptr, data)
	if err != nil {
		return ptr, err
	}
	return ptr, nil
}

func ReadJsString(arr []byte) string {
	return string(arr)
}
