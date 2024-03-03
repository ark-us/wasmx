package assemblyscript

import (
	"encoding/binary"
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	mem "mythos/v1/x/wasmx/vm/memory/common"
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

func ReadMemFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error) {
	lengthbz, err := mem.ReadMem(callframe, pointer.(int32)-AS_PTR_LENGHT_OFFSET, int32(AS_PTR_LENGHT_OFFSET))
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lengthbz)
	data, err := mem.ReadMem(callframe, pointer, int32(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AllocateMemVm(vm *wasmedge.VM, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	result, err := vm.Execute(types.MEMORY_EXPORT_AS, size, AS_ARRAY_BUFFER_TYPE)
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
	err = mem.WriteMem(callframe, data, ptr)
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
