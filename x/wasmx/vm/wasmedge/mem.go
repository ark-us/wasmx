package wasmedge

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	memc "mythos/v1/x/wasmx/vm/memory/common"
)

var _ memc.IMemory = (*WasmEdgeMemory)(nil)

type WasmEdgeMemory struct {
	*wasmedge.Memory
}

func (wm WasmEdgeMemory) Read(ptr interface{}, size interface{}) ([]byte, error) {
	return ReadMem(wm.Memory, ptr.(int32), size.(int32))
}

func (wm WasmEdgeMemory) Write(ptr interface{}, data []byte) error {
	return WriteMem(wm.Memory, ptr.(int32), data)
}

func ReadMem(mem *wasmedge.Memory, ptr int32, length int32) ([]byte, error) {
	data, err := mem.GetData(uint(ptr), uint(length))
	if err != nil {
		return nil, err
	}
	result := make([]byte, length)
	copy(result, data)
	return result, nil
}

func WriteMem(mem *wasmedge.Memory, ptr int32, data []byte) error {
	length := len(data)
	if length == 0 {
		return nil
	}
	err := mem.SetData(data, uint(ptr), uint(length))
	return err
}
