package ptrlen_i64

import (
	"fmt"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// outdated, TO be replaced with ptrlen_i32 (ptr: i32, len: i32) for pointer
// currently used by python interpreter (Rust)

type RuntimeHandler struct {
	vm           memc.IVm
	allocMemName string
	freeMemName  string
}

var _ memc.RuntimeHandler = (*RuntimeHandler)(nil)

func NewRuntimeHandler(vm memc.IVm) memc.RuntimeHandler {
	return RuntimeHandler{vm, types.MEMORY_EXPORT_MALLOC, types.MEMORY_EXPORT_FREE}
}

func (h RuntimeHandler) GetVm() memc.IVm {
	return h.vm
}

func (h RuntimeHandler) GetMemory() (memc.IMemory, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return mem, nil
}

func (h RuntimeHandler) PtrParamsLength() int {
	return 1
}

func (h RuntimeHandler) ReadMemFromPtr(pointer []interface{}) ([]byte, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return ReadMemFromPtr(mem, h.vm, h.freeMemName, pointer[0])
}
func (h RuntimeHandler) AllocateWriteMem(data []byte) ([]interface{}, error) {
	ptr, err := AllocateAndWriteMem(h.vm, h.allocMemName, data)
	if err != nil {
		return []interface{}{}, err
	}
	ptr64 := BuildPtrI64(ptr, int32(len(data)))
	return []interface{}{ptr64}, nil
}
func (RuntimeHandler) ReadJsString(arr []byte) string {
	return ReadJsString(arr)
}
func (h RuntimeHandler) ReadStringFromPtr(pointer interface{}) (string, error) {
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

func ReadMemFromPtr(mem memc.IMemory, vm memc.IVm, freeMemName string, pointer interface{}) ([]byte, error) {
	ptr, size := DecodePtrI64(pointer.(int64))
	data, err := mem.Read(ptr, size)
	if err != nil {
		return nil, err
	}
	err = memc.FreeMemory(vm, freeMemName, ptr)
	if err != nil {
		return nil, fmt.Errorf("cannot free memory: %s", err)
	}
	return data, nil
}

// ReadJsString converts a byte slice to a UTF-8 encoded string.
func ReadJsString(data []byte) string {
	return string(data)
}

func AllocateAndWriteMem(vm memc.IVm, allocMemName string, data []byte) (int32, error) {
	mem, err := vm.GetMemory()
	if err != nil {
		return 0, err
	}
	datalen := int32(len(data))
	ptr, err := memc.AllocateMemory(vm, allocMemName, datalen)
	if err != nil {
		return 0, err
	}
	err = mem.Write(ptr, data)
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

func AllocateAndWriteMemi64(vm memc.IVm, allocMemName string, data []byte) (int64, error) {
	ptr, err := AllocateAndWriteMem(vm, allocMemName, data)
	if err != nil {
		return 0, err
	}
	return BuildPtrI64(ptr, int32(len(data))), nil
}

func BuildPtrI64(ptr int32, datalen int32) int64 {
	return (int64(ptr) << int64(32)) | int64(datalen)
}

func DecodePtrI64(value int64) (int32, int32) {
	ptr := int32(value >> 32)
	len := int32(value & 0xFFFFFFFF)
	return ptr, len
}
