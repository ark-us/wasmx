package common

import (
	"bytes"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IFnVal = func(context interface{}, mod RuntimeHandler, params []interface{}) ([]interface{}, error)

type IFn interface {
	Name() string
	InputTypes() []interface{}
	OutputTypes() []interface{}
	Fn(context interface{}, mod RuntimeHandler, params []interface{}) ([]interface{}, error)
	Cost() int32
}

type IMemory interface {
	Read(ptr interface{}, size interface{}) ([]byte, error)
	Write(ptr interface{}, data []byte) error
}

type IVm interface {
	Call(name string, args []interface{}) ([]int32, error)
	GetMemory() (IMemory, error)
	New(ctx sdk.Context) IVm
	Cleanup()
	InstantiateWasm(filePath string, wasmbuffer []byte) error
	RegisterModule(mod interface{}) error
	BuildModule(rnh RuntimeHandler, modname string, context interface{}, fndefs []IFn) (interface{}, error)
	BuildFn(fnname string, fnval IFnVal, inputTypes []interface{}, outputTypes []interface{}, cost int32) IFn
	ValType_I32() interface{}
	ValType_I64() interface{}
	ValType_F64() interface{}
}

type RuntimeHandler interface {
	GetVm() IVm
	GetMemory() IMemory
	ReadMemFromPtr(pointer interface{}) ([]byte, error)
	AllocateWriteMem(data []byte) (int32, error)
	ReadStringFromPtr(pointer interface{}) (string, error)
	ReadJsString(arr []byte) string
}

func ReadMemUntilNull(mem IMemory, pointer interface{}) ([]byte, error) {
	result := []byte{}
	ptr := pointer.(int32)
	bz, err := mem.Read(ptr, 1)
	if err != nil {
		return nil, err
	}
	for bz[0] != 0 {
		result = append(result, bz[0])
		ptr = ptr + 1
		bz, err = mem.Read(ptr, 1)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func WriteMemBoundBySize(mem IMemory, data []byte, ptr interface{}, size interface{}) error {
	if len(data) < size.(int) {
		size = int32(len(data))
	}
	return mem.Write(ptr, data[0:size.(int)])
}

func WriteBigInt(mem IMemory, value *big.Int, pointer interface{}) error {
	data := value.FillBytes(make([]byte, 32))
	return mem.Write(pointer, data)
}

func ReadBigInt(mem IMemory, pointer interface{}, size interface{}) (*big.Int, error) {
	data, err := mem.Read(pointer, size)
	if err != nil {
		return nil, err
	}
	x := new(big.Int)
	x.SetBytes(data)
	return x, nil
}

func ReadI64(mem IMemory, pointer interface{}, size interface{}) (int64, error) {
	x, err := ReadBigInt(mem, pointer, size)
	if err != nil {
		return 0, err
	}
	if !x.IsInt64() {
		return 0, fmt.Errorf("ReadI32 overflow")
	}
	return x.Int64(), nil
}

func ReadI32(mem IMemory, pointer interface{}, size interface{}) (int32, error) {
	xi64, err := ReadI64(mem, pointer, size)
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

func ToInt32Slice(input []interface{}) ([]int32, error) {
	result := make([]int32, len(input))
	for i, v := range input {
		val, ok := v.(int32)
		if !ok {
			return nil, fmt.Errorf("value at index %d is not an int32", i)
		}
		result[i] = val
	}
	return result, nil
}

func FromInt32Slice(input []int32) []interface{} {
	result := make([]interface{}, len(input))
	for i, v := range input {
		result[i] = v
	}
	return result
}
