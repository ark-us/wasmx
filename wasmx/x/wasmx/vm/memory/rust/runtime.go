package rust

import (
	"strings"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

type RuntimeHandler struct {
	vm           memc.IVm
	allocMemName string
	freeMemName  string
}

var _ memc.RuntimeHandler = (*RuntimeHandler)(nil)

func NewRuntimeHandler(vm memc.IVm, sysdeps []types.SystemDep) memc.RuntimeHandler {
	allocMemName := types.MEMORY_EXPORT_MALLOC
	freeMemName := types.MEMORY_EXPORT_FREE
	for _, dep := range sysdeps {
		if strings.Contains(dep.Role, types.MEMORY_ENTRYPOINT_ALLOC) {
			allocMemName = dep.Role[len(types.MEMORY_ENTRYPOINT_ALLOC):]
		}
		if strings.Contains(dep.Role, types.MEMORY_ENTRYPOINT_FREE) {
			freeMemName = dep.Role[len(types.MEMORY_ENTRYPOINT_FREE):]
		}
	}
	return RuntimeHandler{vm, allocMemName, freeMemName}
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
	panic("use ptrlen_i64")
}

func (h RuntimeHandler) ReadMemFromPtr(pointer []interface{}) ([]byte, error) {
	panic("use ptrlen_i64")
}
func (h RuntimeHandler) AllocateWriteMem(data []byte) ([]interface{}, error) {
	panic("use ptrlen_i64")
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

func (h RuntimeHandler) WriteMemDefaultMalloc(data []byte) (int32, error) {
	return AllocateAndWriteMem(h.vm, h.allocMemName, data)
}

func (h RuntimeHandler) WriteMemDefaultMallocI64(data []byte) (int64, error) {
	return AllocateAndWriteMemi64(h.vm, h.allocMemName, data)
}

func ReadMemFromPtr(mem memc.IMemory, pointer interface{}) ([]byte, error) {
	ptr, size := DecodePtrI64(pointer.(int64))
	data, err := mem.Read(ptr, size)
	if err != nil {
		return nil, err
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
	ptr, err := AllocateMemory(vm, allocMemName, datalen)
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

func AllocateMemory(vm memc.IVm, allocMemName string, size int32) (int32, error) {
	result, err := vm.Call(allocMemName, []interface{}{size}, nil)
	if err != nil {
		return 0, err
	}
	return result[0], nil
}

func FreeMemory(vm memc.IVm, freeMemName string, ptr int32) error {
	_, err := vm.Call(freeMemName, []interface{}{ptr}, nil)
	if err != nil {
		return err
	}
	return nil
}
