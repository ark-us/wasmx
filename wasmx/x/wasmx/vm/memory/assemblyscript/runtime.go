package assemblyscript

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
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
	vm memc.IVm
}

var _ memc.RuntimeHandler = (*RuntimeHandlerAS)(nil)

func NewRuntimeHandlerAS(vm memc.IVm, _ []types.SystemDep) memc.RuntimeHandler {
	return RuntimeHandlerAS{vm}
}

func (h RuntimeHandlerAS) GetVm() memc.IVm {
	return h.vm
}

func (h RuntimeHandlerAS) GetMemory() (memc.IMemory, error) {
	return h.vm.GetMemory()
}

func (h RuntimeHandlerAS) PtrParamsLength() int {
	return 1
}

func (h RuntimeHandlerAS) ReadMemFromPtr(pointer []interface{}) ([]byte, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return ReadMemFromPtr(mem, pointer[0])
}

func (h RuntimeHandlerAS) AllocateWriteMem(data []byte) ([]interface{}, error) {
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

func (RuntimeHandlerAS) ReadJsString(arr []byte) string {
	return ReadJsString(arr)
}

func (h RuntimeHandlerAS) ReadStringFromPtr(pointer interface{}) (string, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return "", err
	}
	mm, err := ReadMemFromPtr(mem, pointer)
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
	data, err := mem.ReadRaw(pointer, int32(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AllocateMemVm(vm memc.IVm, mem memc.IMemory, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no VM instance found")
	}
	args := []interface{}{size, AS_ARRAY_BUFFER_TYPE}
	result, err := vm.Call(types.MEMORY_EXPORT_AS, args, nil)
	if err != nil {
		return 0, err
	}
	return result[0], nil
}

func AllocateWriteMem(vm memc.IVm, mem memc.IMemory, data []byte) (interface{}, error) {
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

// ReadJsString converts a byte slice to a UTF-16 encoded string.
func ReadJsString(data []byte) string {
	// Ensure the data has an even number of bytes (required for UTF-16 encoding)
	if len(data)%2 != 0 {
		data = append([]byte{0}, data...)
	}

	// Convert bytes to uint16 values
	u16 := make([]uint16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u16[i/2] = uint16(data[i]) | uint16(data[i+1])<<8
	}

	// Decode UTF-16 to a Go string
	runes := utf16.Decode(u16)
	return string(runes)
}
