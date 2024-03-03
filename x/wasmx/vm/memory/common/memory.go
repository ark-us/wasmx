package common

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

func ReadMem(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) ([]byte, error) {
	ptr := pointer.(int32)
	length := size.(int32)
	mem := callframe.GetMemoryByIndex(0)
	if mem == nil {
		return nil, fmt.Errorf("could not find memory")
	}

	data, err := mem.GetData(uint(ptr), uint(length))
	if err != nil {
		return nil, err
	}
	result := make([]byte, length)
	copy(result, data)
	return result, nil
}

func WriteMem(callframe *wasmedge.CallingFrame, data []byte, pointer interface{}) error {
	ptr := pointer.(int32)
	length := len(data)
	if length == 0 {
		return nil
	}
	mem := callframe.GetMemoryByIndex(0)
	if mem == nil {
		return fmt.Errorf("no memory found")
	}
	err := mem.SetData(data, uint(ptr), uint(length))
	return err
}

func WriteMemBoundBySize(callframe *wasmedge.CallingFrame, data []byte, pointer interface{}, size interface{}) error {
	length := size.(int32)
	if len(data) < int(length) {
		length = int32(len(data))
	}
	return WriteMem(callframe, data[0:length], pointer)
}

func WriteBigInt(callframe *wasmedge.CallingFrame, value *big.Int, pointer interface{}) error {
	data := value.FillBytes(make([]byte, 32))
	return WriteMem(callframe, data, pointer)
}

func ReadBigInt(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) (*big.Int, error) {
	data, err := ReadMem(callframe, pointer, size)
	if err != nil {
		return nil, err
	}
	x := new(big.Int)
	x.SetBytes(data)
	return x, nil
}

func ReadI64(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) (int64, error) {
	x, err := ReadBigInt(callframe, pointer, size)
	if err != nil {
		return 0, err
	}
	if !x.IsInt64() {
		return 0, fmt.Errorf("ReadI32 overflow")
	}
	return x.Int64(), nil
}

func ReadI32(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) (int32, error) {
	xi64, err := ReadI64(callframe, pointer, size)
	if err != nil {
		return 0, err
	}
	xi32 := int32(xi64)
	if xi64 > int64(xi32) {
		return 0, fmt.Errorf("ReadI32 overflow")
	}
	return xi32, nil
}

func ReadAndFillWithZero(data []byte, start int32, length int32) []byte {
	dataLen := int32(len(data))
	end := start + length
	var value []byte
	if end >= dataLen {
		if len(data) > 0 {
			value = data[start:]
		}
		value = PadWithZeros(value, int(length))
	} else {
		value = data[start:end]
	}
	return value
}

func PaddRightToMultiple32(data []byte) []byte {
	length := len(data)
	c := length % 32
	if c > 0 {
		data = append(data, bytes.Repeat([]byte{0}, 32-c)...)
	}
	return data
}

func PaddLeftTo32(data []byte) []byte {
	length := len(data)
	if length >= 32 {
		return data
	}
	data = append(bytes.Repeat([]byte{0}, 32-length), data...)
	return data
}

func PadWithZeros(data []byte, targetLen int) []byte {
	dataLen := len(data)
	if targetLen <= dataLen {
		return data
	}
	data = append(data, bytes.Repeat([]byte{0}, targetLen-dataLen)...)
	return data
}
