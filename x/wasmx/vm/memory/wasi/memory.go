package wasi

import (
	"mythos/v1/x/wasmx/types"
	memc "mythos/v1/x/wasmx/vm/memory/common"
)

func WriteMemDefaultMalloc(vm memc.IVm, mem memc.IMemory, data []byte) (int32, error) {
	datalen := int32(len(data))
	ptr, err := AllocateMemDefaultMalloc(vm, datalen)
	if err != nil {
		return 0, err
	}
	err = mem.Write(ptr, data)
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

func WriteDynMemDefaultMalloc(vm memc.IVm, mem memc.IMemory, data []byte) (uint64, error) {
	ptr, err := WriteMemDefaultMalloc(vm, mem, data)
	if err != nil {
		return 0, err
	}
	return BuildPtr64(ptr, int32(len(data))), nil
}

func AllocateMemDefaultMalloc(vm memc.IVm, size int32) (int32, error) {
	result, err := vm.Call(types.MEMORY_EXPORT_MALLOC, []interface{}{size})
	if err != nil {
		return 0, err
	}
	return result[0], nil
}

func BuildPtr64(ptr int32, datalen int32) uint64 {
	return (uint64(ptr) << uint64(32)) | uint64(datalen)
}
