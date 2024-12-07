package wazero

import (
	"fmt"

	"github.com/tetratelabs/wazero/api"

	memc "mythos/v1/x/wasmx/vm/memory/common"
)

var _ memc.IMemory = (*WazeroMemory)(nil)

type WazeroMemory struct {
	api.Memory
}

func (wm WazeroMemory) Read(ptr interface{}, size interface{}) ([]byte, error) {
	return ReadMem(wm.Memory, ptr.(int32), size.(int32))
}

func (wm WazeroMemory) Write(ptr interface{}, data []byte) error {
	return WriteMem(wm.Memory, ptr.(int32), data)
}

func ReadMem(mem api.Memory, ptr int32, length int32) ([]byte, error) {
	data, success := mem.Read(uint32(ptr), uint32(length))
	if !success {
		return nil, fmt.Errorf("memory failed to read from pointer %d, length %d", ptr, length)
	}
	result := make([]byte, length)
	copy(result, data)
	return result, nil
}

func WriteMem(mem api.Memory, ptr int32, data []byte) error {
	length := len(data)
	if length == 0 {
		return nil
	}
	success := mem.Write(uint32(ptr), data)
	if !success {
		return fmt.Errorf("memory failed to write to pointer %d, length %d", ptr, length)
	}
	return nil
}
