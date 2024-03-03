package wasi

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	mem "mythos/v1/x/wasmx/vm/memory/common"
)

func WriteMemDefaultMalloc(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (int32, error) {
	datalen := int32(len(data))
	ptr, err := AllocateMemDefaultMalloc(vm, datalen)
	if err != nil {
		return 0, err
	}
	err = mem.WriteMem(callframe, data, ptr)
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

func WriteDynMemDefaultMalloc(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (uint64, error) {
	ptr, err := WriteMemDefaultMalloc(vm, callframe, data)
	if err != nil {
		return 0, err
	}
	return BuildPtr64(ptr, int32(len(data))), nil
}

func AllocateMemDefaultMalloc(vm *wasmedge.VM, size int32) (int32, error) {
	result, err := vm.Execute(types.MEMORY_EXPORT_MALLOC, size)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func BuildPtr64(ptr int32, datalen int32) uint64 {
	return (uint64(ptr) << uint64(32)) | uint64(datalen)
}
