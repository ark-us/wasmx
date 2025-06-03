package ptrlen_i32

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// outdated, TO be replaced with (ptr: i32, len: i32) for pointer
// used by python interpreter

type RuntimeHandler struct {
	vm           memc.IVm
	allocMemName string
	freeMemName  string
}

var _ memc.RuntimeHandler = (*RuntimeHandler)(nil)

// TODO make the alloc/free function as a depedency requirement for contracts
// and give them as input to the handler
// right now we are using the Rust/TinyGo compatible alloc, free export names

// func NewRuntimeHandler(vm memc.IVm, allocMemName string, freeMemName string) memc.RuntimeHandler {
// 	return RuntimeHandler{vm, allocMemName, freeMemName}
// }

func NewRuntimeHandler(vm memc.IVm) memc.RuntimeHandler {
	return RuntimeHandler{vm, types.MEMORY_EXPORT_ALLOC, types.MEMORY_EXPORT_FREE}
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
	return 2
}

func (h RuntimeHandler) ReadMemFromPtr(pointer []interface{}) ([]byte, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return ReadMemFromPtr(mem, pointer[0].(int32), pointer[1].(int32))
}
func (h RuntimeHandler) AllocateWriteMem(data []byte) ([]interface{}, error) {
	ptr, len, err := AllocateAndWriteMem(h.vm, h.allocMemName, data)
	if err != nil {
		return []interface{}{}, err
	}
	return []interface{}{ptr, len}, nil
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

func ReadMemFromPtr(mem memc.IMemory, ptr, size int32) ([]byte, error) {
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

func AllocateAndWriteMem(vm memc.IVm, allocMemName string, data []byte) (int32, int32, error) {
	mem, err := vm.GetMemory()
	if err != nil {
		return 0, 0, err
	}
	datalen := int32(len(data))
	ptr, err := memc.AllocateMemory(vm, allocMemName, datalen)
	if err != nil {
		return 0, 0, err
	}
	err = mem.Write(ptr, data)
	if err != nil {
		return 0, 0, err
	}
	return ptr, datalen, nil
}
