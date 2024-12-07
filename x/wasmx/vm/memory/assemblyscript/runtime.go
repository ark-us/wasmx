package assemblyscript

import (
	"encoding/binary"
	"fmt"

	"mythos/v1/x/wasmx/types"
	memc "mythos/v1/x/wasmx/vm/memory/common"
)

const AS_PTR_LENGHT_OFFSET = int32(4)
const AS_ARRAY_BUFFER_TYPE = int32(1)

// https://www.assemblyscript.org/runtime.html#memory-layout
// Name	   Offset	Type	Description
// mmInfo	-20	    usize	Memory manager info
// gcInfo	-16	    usize	Garbage collector info
// gcInfo2	-12	    usize	Garbage collector info
// rtId 	-8	    u32	    Unique id of the concrete class
// rtSize	-4	    u32	    Size of the data following the header
//           0		Payload starts here

type RuntimeHandlerAS struct {
	vm  memc.IVm
	mem memc.IMemory
}

var _ memc.RuntimeHandler = (*RuntimeHandlerAS)(nil)

func NewRuntimeHandlerAS(vm memc.IVm, mem memc.IMemory) memc.RuntimeHandler {
	return RuntimeHandlerAS{vm, mem}
}

func (h RuntimeHandlerAS) GetVm() memc.IVm {
	return h.vm
}

func (h RuntimeHandlerAS) GetMemory() memc.IMemory {
	return h.mem
}

func (h RuntimeHandlerAS) ReadMemFromPtr(pointer interface{}) ([]byte, error) {
	return ReadMemFromPtr(h.mem, pointer)
}

func (h RuntimeHandlerAS) AllocateWriteMem(data []byte) (int32, error) {
	return AllocateWriteMem(h.vm, h.mem, data)
}

func (RuntimeHandlerAS) ReadJsString(arr []byte) string {
	return ReadJsString(arr)
}

func (h RuntimeHandlerAS) ReadStringFromPtr(pointer interface{}) (string, error) {
	mm, err := ReadMemFromPtr(h.mem, pointer)
	if err != nil {
		return "", err
	}
	return ReadJsString(mm), nil
}

func ReadMemFromPtr(mem memc.IMemory, pointer interface{}) ([]byte, error) {
	lengthbz, err := mem.Read(pointer.(int32)-AS_PTR_LENGHT_OFFSET, int32(AS_PTR_LENGHT_OFFSET))
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lengthbz)
	data, err := mem.Read(pointer, int32(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AllocateMemVm(vm memc.IVm, mem memc.IMemory, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	args := []interface{}{size, AS_ARRAY_BUFFER_TYPE}
	result, err := vm.Call(types.MEMORY_EXPORT_AS, args)
	if err != nil {
		return 0, err
	}
	return result[0], nil
}

func AllocateWriteMem(vm memc.IVm, mem memc.IMemory, data []byte) (int32, error) {
	ptr, err := AllocateMemVm(vm, mem, int32(len(data)))
	if err != nil {
		return ptr, err
	}
	err = mem.Write(ptr, data)
	if err != nil {
		return ptr, err
	}
	return ptr, nil
}

func ReadJsString(arr []byte) string {
	msg := []byte{}
	for i, char := range arr {
		if i%2 == 0 {
			msg = append(msg, char)
		}
	}
	return string(msg)
}
