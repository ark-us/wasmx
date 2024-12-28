package rust

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

type RuntimeHandlerRust struct {
	vm memc.IVm
}

var _ memc.RuntimeHandler = (*RuntimeHandlerRust)(nil)

func NewRuntimeHandlerRust(vm memc.IVm) memc.RuntimeHandler {
	return RuntimeHandlerRust{vm}
}

func (h RuntimeHandlerRust) GetVm() memc.IVm {
	return h.vm
}

func (h RuntimeHandlerRust) GetMemory() (memc.IMemory, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return mem, nil
}

func (h RuntimeHandlerRust) ReadMemFromPtr(pointer interface{}) ([]byte, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return ReadMemFromPtr(mem, pointer)
}
func (h RuntimeHandlerRust) AllocateWriteMem(data []byte) (interface{}, error) {
	ptr, err := AllocateAndWriteMem(h.vm, data)
	if err != nil {
		return int64(0), err
	}
	ptr64 := BuildPtrI64(ptr, int32(len(data)))
	return ptr64, nil
}
func (RuntimeHandlerRust) ReadJsString(arr []byte) string {
	return ReadJsString(arr)
}
func (h RuntimeHandlerRust) ReadStringFromPtr(pointer interface{}) (string, error) {
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

func AllocateAndWriteMem(vm memc.IVm, data []byte) (int32, error) {
	mem, err := vm.GetMemory()
	if err != nil {
		return 0, err
	}
	datalen := int32(len(data))
	ptr, err := AllocateMemory(vm, datalen)
	if err != nil {
		return 0, err
	}
	err = mem.Write(ptr, data)
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

func AllocateAndWriteMemi64(vm memc.IVm, data []byte) (int64, error) {
	ptr, err := AllocateAndWriteMem(vm, data)
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

func AllocateMemory(vm memc.IVm, size int32) (int32, error) {
	result, err := vm.Call(types.MEMORY_EXPORT_ALLOC, []interface{}{size}, nil)
	if err != nil {
		return 0, err
	}
	return result[0], nil
}

func FreeMemory(vm memc.IVm, ptr int32) error {
	_, err := vm.Call(types.MEMORY_EXPORT_FREE, []interface{}{ptr}, nil)
	if err != nil {
		return err
	}
	return nil
}
