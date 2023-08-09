package vm

import (
	"encoding/binary"
	"fmt"
	"mythos/v1/x/wasmx/types"

	sdkerr "cosmossdk.io/errors"
	"github.com/second-state/WasmEdge-go/wasmedge"
)

func wasiStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--wasiStorageStore--", params)

	ctx := context.(*Context)
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--keybz--", keybz)
	valuebz, err := readMem(callframe, params[2], params[3])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--valuebz--", valuebz)
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "ewasm")
	ctx.ContractStore.Set(keybz, valuebz)

	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasiStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--wasiStorageLoad--", params)

	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	keybz, err := readMem(callframe, params[0], params[1])
	fmt.Println("--keybz--", keybz, string(keybz), err)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	fmt.Println("--data--", data, string(data), err)
	if len(data) == 0 {
		data = types.EMPTY_BYTES32
	}

	lenData := int32(len(data))
	ptr, err := wasiAllocateMem(ctx, lenData+4) // add 4 bytes for length
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	initdata, err := readMem(callframe, ptr, lenData)
	fmt.Println("--emptydata0--", initdata, string(initdata), err)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	initdata, err = readMem(callframe, ptr-int32(4), int32(32))
	fmt.Println("--emptydata2--", initdata, string(initdata), err)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	// add length in 4 bytes
	data = append(binary.BigEndian.AppendUint32([]byte{}, uint32(lenData)), data...)
	fmt.Println("--data2--", data)

	err = writeMem(callframe, data, ptr)
	fmt.Println("--writeMem--", err)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	initdata, err = readMem(callframe, ptr, lenData+4)
	fmt.Println("--writtendata--", initdata, string(initdata), err)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func BuildWasiWasmxEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasmx")

	functype_i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32i32i32_, wasiStorageStore, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32i32_i32, wasiStorageLoad, context, 0))

	return env
}

func ExecuteWasi(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	fmt.Println("--ExecuteWasi--", funcName)

	wasimodule := contractVm.GetImportModule(wasmedge.WASI)
	wasimodule.InitWasi(
		[]string{
			``,
			`./testdata/main.py`,
			"222444",
		},
		// os.Environ(), // The envs
		[]string{},
		// The mapping preopens
		[]string{
			".:.",
			// '/': __dirname
		},
	)

	if funcName == types.ENTRY_POINT_INSTANTIATE {
		funcName = "_initialize"
		res, err := contractVm.Execute(funcName)
		fmt.Println("--ExecuteWasi-_instantiate-", res, err)
		// return res, err
		return nil, nil
	}

	// WASI standard - no args, no return
	funcName = "_start" // tinygo main
	res, err := contractVm.Execute(funcName)

	activeMemory := contractVm.GetActiveModule().FindMemory("memory")
	bz, err := readMemSimple(activeMemory, int32(0), int32(32))
	fmt.Println("--result mem--", bz, string(bz), err)

	fmt.Println("--ExecuteWasi--", res, err)

	// TODO bindings dependency? encoding of the JSON string

	return res, err
}

func readMemSimple(mem *wasmedge.Memory, pointer interface{}, size interface{}) ([]byte, error) {
	ptr := pointer.(int32)
	length := size.(int32)

	data, err := mem.GetData(uint(ptr), uint(length))
	if err != nil {
		return nil, err
	}
	result := make([]byte, length)
	copy(result, data)
	return result, nil
}

func wasiAllocateMem(ctx *Context, size int32) (int32, error) {
	addr := ctx.Env.Contract.Address
	contractCtx, ok := ctx.ContractRouter[addr.String()]
	if !ok {
		return int32(0), sdkerr.Wrapf(sdkerr.Error{}, "contract context not found for address %s", addr.String())
	}
	return wasiAllocateMemVm(contractCtx.Vm, size)
}

func wasiAllocateMemVm(vm *wasmedge.VM, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	result, err := vm.Execute("alloc", size)
	fmt.Println("==alloc==", result, err)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}
