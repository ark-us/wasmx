package common

import (
	"bytes"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const VM_TERMINATE_ERROR = "terminate"

// GasMeter is a read-only version of the sdk gas meter
type Gas = uint64
type GasMeter interface {
	GasConsumed() Gas
	GasLimit() Gas
	GasRemaining() Gas
	ConsumeGas(gas uint64, descriptor string)
}

type IFnVal = func(context interface{}, mod RuntimeHandler, params []interface{}) ([]interface{}, error)

type IWasmVmMeta interface {
	LibVersion() string
	NewWasmVm(ctx sdk.Context, aot bool) IVm
	AnalyzeWasm(ctx sdk.Context, wasmbuffer []byte) (WasmMeta, error)
	AotCompile(ctx sdk.Context, inPath string, outPath string) error
}

type IFn interface {
	Name() string
	InputTypes() []interface{}
	OutputTypes() []interface{}
	Fn(context interface{}, mod RuntimeHandler, params []interface{}) ([]interface{}, error)
	Cost() int32
}

type WasmExport interface {
	Name() string
	// InputTypes() []interface{}
	// OutputTypes() []interface{}
	Fn() interface{}
}

type WasmImport interface {
	Name() string
	// InputTypes() []interface{}
	// OutputTypes() []interface{}
	Fn() interface{}
	ModuleName() string
}

type WasmMeta interface {
	ListImports() []WasmImport
	ListExports() []WasmExport
}

type IMemory interface {
	Size() uint32
	ReadRaw(ptr interface{}, size interface{}) ([]byte, error)
	WriteRaw(ptr interface{}, data []byte) error
	Read(ptr int32, size int32) ([]byte, error)
	Write(ptr int32, data []byte) error
}

type NewIVmFn = func(ctx sdk.Context, aot bool) IVm

type IVm interface {
	Call(name string, args []interface{}, gasMeter GasMeter) ([]int32, error)
	GetMemory() (IMemory, error)
	New(ctx sdk.Context, aot bool) IVm
	Cleanup()
	InstantiateWasm(filePath string, wasmbuffer []byte) error
	RegisterModule(mod interface{}) error
	BuildModule(rnh RuntimeHandler, modname string, context interface{}, fndefs []IFn) (interface{}, error)
	BuildFn(fnname string, fnval IFnVal, inputTypes []interface{}, outputTypes []interface{}, cost int32) IFn
	ValType_I32() interface{}
	ValType_I64() interface{}
	ValType_F64() interface{}
	GetFunctionList() []string
	FindGlobal(name string) interface{}
	ListRegisteredModule() []string
	InitWasi(args []string, envs []string, preopens []string) error
}

type RuntimeHandler interface {
	GetVm() IVm
	GetMemory() (IMemory, error)
	ReadMemFromPtr(pointer interface{}) ([]byte, error)
	AllocateWriteMem(data []byte) (int32, error)
	ReadStringFromPtr(pointer interface{}) (string, error)
	ReadJsString(arr []byte) string
}

// Contains static analysis info of the contract (the Wasm code to be precise).
// This type is returned by VM.AnalyzeCode().
type AnalysisReport struct {
	Dependencies []string
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
	_size := size.(int32)
	if len(data) < int(_size) {
		_size = int32(len(data))
	}
	return mem.WriteRaw(ptr, data[0:int(_size)])
}

func WriteBigInt(mem IMemory, value *big.Int, pointer interface{}) error {
	data := value.FillBytes(make([]byte, 32))
	return mem.WriteRaw(pointer, data)
}

func ReadBigInt(mem IMemory, pointer interface{}, size interface{}) (*big.Int, error) {
	data, err := mem.ReadRaw(pointer, size)
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

var _ IWasmVmMeta = (*WasmRuntimeMockVmMeta)(nil)

type WasmRuntimeMockVmMeta struct{}

// When cgo is disabled at build time, this returns an error at runtime.
func (WasmRuntimeMockVmMeta) LibVersion() string {
	return "0"
}

func (WasmRuntimeMockVmMeta) NewWasmVm(ctx sdk.Context, aot bool) IVm {
	return nil
}

func (WasmRuntimeMockVmMeta) AnalyzeWasm(_ sdk.Context, wasmbuffer []byte) (WasmMeta, error) {
	return nil, fmt.Errorf("runtime mock: AnalyzeWasm not implemented")
}

func (WasmRuntimeMockVmMeta) AotCompile(_ sdk.Context, inPath string, outPath string) error {
	return fmt.Errorf("runtime mock: AotCompile not implemented")
}
