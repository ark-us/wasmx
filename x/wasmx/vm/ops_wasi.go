package vm

import (
	"encoding/binary"
	"fmt"
	"mythos/v1/x/wasmx/types"
	"os"
	"path"
	"path/filepath"

	sdkerr "cosmossdk.io/errors"
	"github.com/second-state/WasmEdge-go/wasmedge"
)

func wasiStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	valuebz, err := readMem(callframe, params[2], params[3])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "ewasm")
	ctx.ContractStore.Set(keybz, valuebz)

	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasiStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = types.EMPTY_BYTES32
	}

	lenData := int32(len(data))
	ptr, err := wasiAllocateMem(ctx, lenData+4) // add 4 bytes for length
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	// add length in 4 bytes
	data = append(binary.BigEndian.AppendUint32([]byte{}, uint32(lenData)), data...)

	err = writeMem(callframe, data, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
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
	dir := filepath.Dir(context.Env.Contract.FilePath)
	resultFile := path.Join(dir, types.WasiResultFile)

	// WASI does not have instantiate
	if funcName == types.ENTRY_POINT_INSTANTIATE {
		return nil, nil
	}

	// WASI standard - no args, no return
	res, err := contractVm.Execute("_start") // tinygo main
	if err != nil {
		return nil, err
	}

	result, err := os.ReadFile(resultFile)
	if err != nil {
		return nil, err
	}
	context.ReturnData = result

	return res, err
}

func ExecutePythonInterpreter(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	if funcName == "execute" || funcName == "query" {
		funcName = "main"
	}

	wasimodule := contractVm.GetImportModule(wasmedge.WASI)
	dir := filepath.Dir(context.Env.Contract.FilePath)
	inputFile := path.Join(dir, "main.py")
	content, err := os.ReadFile(context.Env.Contract.FilePath)
	if err != nil {
		return nil, err
	}
	// TODO import the module or make sure json & sys are not imported one more time
	strcontent := fmt.Sprintf(`
import sys
import json

%s

inputObject = sys.argv[1]
input = json.loads(inputObject)

res = %s(input)

resfilepath = "%s"

file1 = open(resfilepath, "w")
file1.write(res or "")
file1.close()

	`, string(content), funcName, types.WasiResultFile)

	err = os.WriteFile(inputFile, []byte(strcontent), 0644)
	if err != nil {
		return nil, err
	}

	wasimodule.InitWasi(
		[]string{
			``,
			"main.py",
			string(context.Env.CurrentCall.CallData),
		},
		// os.Environ(), // The envs
		[]string{},
		// The mapping preopens
		[]string{
			// ".:.",
			// fmt.Sprintf(`%s:.`, dir),
			// fmt.Sprintf(`%s:%s`, dir, dir),
			fmt.Sprintf(`.:%s`, dir),
			// fmt.Sprintf(`/:%s`, dir),
		},
	)
	return ExecuteWasi(context, contractVm, types.ENTRY_POINT_EXECUTE)
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
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}
