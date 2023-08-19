package vm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

func wasiPythonStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
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

func wasiPythonStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
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

func wasiTinygoStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	valuebz, err := readMem(callframe, params[2], params[3])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "wasmx")
	ctx.ContractStore.Set(keybz, valuebz)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasiTinygoStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
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
	datalen := int32(len(data))
	ptr, err := tinygoAllocateMem(ctx, datalen)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	err = writeMem(callframe, data, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = (uint64(ptr) << uint64(32)) | uint64(datalen)
	return returns, wasmedge.Result_Success
}

func wasiQuickjsStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	valuebz, err := readMem(callframe, params[2], params[3])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "wasmx")
	ctx.ContractStore.Set(keybz, valuebz)
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

func wasiQuickjsStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
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
	datalen := int32(len(data))
	ptr, err := quickjsAllocateMem(ctx, datalen)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	err = writeMem(callframe, data, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// getCallData(ptr): ArrayBuffer
func getCallData2(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	ctx := context.(*Context)
	datalen := len(ctx.Env.CurrentCall.CallData)
	ptr, err := tinygoAllocateMem(ctx, int32(datalen))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	err = writeMem(callframe, ctx.Env.CurrentCall.CallData, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	newptr := (uint64(ptr) << uint64(32)) | uint64(datalen)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

func wasiSetReturnData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	ctx := context.(*Context)
	data, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.ReturnData = data
	return returns, wasmedge.Result_Success
}

func wasiCallClassic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := readMem(callframe, params[1], int32(45))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr, err := sdk.AccAddressFromBech32(string(addrbz))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	value := big.NewInt(params[2].(int64))
	calldata, err := readMem(callframe, params[3], params[4])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}

	var success int32
	var returnData []byte

	// Send funds
	if value.BitLen() > 0 {
		err = ctx.CosmosHandler.SendCoin(addr, value)
	}
	if err != nil {
		success = int32(2)
	} else {
		contractContext := GetContractContext(ctx, addr)
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			req := vmtypes.CallRequest{
				To:       addr,
				From:     ctx.Env.Contract.Address,
				Value:    value,
				GasLimit: big.NewInt(gasLimit),
				Calldata: calldata,
				Bytecode: contractContext.Bytecode,
				CodeHash: contractContext.CodeHash,
				FilePath: contractContext.FilePath,
				IsQuery:  false,
			}
			success, returnData = WasmxCall(ctx, req)
		}
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	datalen := int32(len(responsebz))
	ptr, err := wasiAllocateMem(ctx, datalen+4)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	// add length in 4 bytes
	responsebz = append(binary.BigEndian.AppendUint32([]byte{}, uint32(datalen)), responsebz...)
	err = writeMem(callframe, responsebz, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasiCallStatic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := readMem(callframe, params[1], int32(45))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr, err := sdk.AccAddressFromBech32(string(addrbz))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	calldata, err := readMem(callframe, params[2], params[3])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}

	var success int32
	var returnData []byte

	contractContext := GetContractContext(ctx, addr)
	if contractContext == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequest{
			To:       addr,
			From:     ctx.Env.Contract.Address,
			Value:    big.NewInt(0),
			GasLimit: big.NewInt(gasLimit),
			Calldata: calldata,
			Bytecode: contractContext.Bytecode,
			CodeHash: contractContext.CodeHash,
			FilePath: contractContext.FilePath,
			IsQuery:  true,
		}
		success, returnData = WasmxCall(ctx, req)
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	datalen := int32(len(responsebz))
	ptr, err := wasiAllocateMem(ctx, datalen+4)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	// add length in 4 bytes
	responsebz = append(binary.BigEndian.AppendUint32([]byte{}, uint32(datalen)), responsebz...)
	err = writeMem(callframe, responsebz, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasiTinygoLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO only in debug mode

	fmt.Println("LOG: ", string(keybz))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildWasiWasmxEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasmx")

	functype__i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i64i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i64i32i64i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32i32i32_, wasiPythonStorageStore, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32i32_i32, wasiPythonStorageLoad, context, 0))
	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i64, getCallData2, context, 0))
	env.AddFunction("setReturnData", wasmedge.NewFunction(functype_i32i32_, wasiSetReturnData, context, 0))

	env.AddFunction("storageStore2", wasmedge.NewFunction(functype_i32i32i32i32_, wasiTinygoStorageStore, context, 0))
	env.AddFunction("storageLoad2", wasmedge.NewFunction(functype_i32i32_i64, wasiTinygoStorageLoad, context, 0))

	env.AddFunction("storageStore3", wasmedge.NewFunction(functype_i32i32i32i32_, wasiQuickjsStorageStore, context, 0))
	env.AddFunction("storageLoad3", wasmedge.NewFunction(functype_i32i32_i32, wasiQuickjsStorageLoad, context, 0))

	env.AddFunction("callClassic", wasmedge.NewFunction(functype_i64i32i64i32i32_i32, wasiCallClassic, context, 0))
	env.AddFunction("callStatic", wasmedge.NewFunction(functype_i64i32i32i32_i32, wasiCallStatic, context, 0))

	env.AddFunction("log", wasmedge.NewFunction(functype_i32i32_, wasiTinygoLog, context, 0))
	return env
}

func ExecuteWasi(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	var res []interface{}
	var err error
	dir := filepath.Dir(context.Env.Contract.FilePath)
	resultFile := path.Join(dir, types.WasiResultFile)

	// WASI standard does not have instantiate
	// this is only for wasmx contracts (e.g. compiled with tinygo, javy)
	// TODO consider extracting this in a dependency
	if funcName == types.ENTRY_POINT_INSTANTIATE {
		fnNames, _ := contractVm.GetFunctionList()
		found := false
		for _, name := range fnNames {
			// note that custom entries do not have access to WASI endpoints at this time
			if name == "main.instantiate" {
				found = true
				res, err = contractVm.Execute("main.instantiate")
			}
		}
		if !found {
			return nil, nil
		}
	} else {
		// WASI standard - no args, no return
		res, err = contractVm.Execute("_start") // tinygo main
	}
	if err != nil {
		return nil, err
	}

	// if returndata is set, we use this
	if len(context.ReturnData) == 0 {
		result, err := os.ReadFile(resultFile)
		if err != nil {
			return nil, err
		}
		context.ReturnData = result
	}

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

res = ""
entrypoint = %s
if len(sys.argv) > 1 and sys.argv[1] != "":
	inputObject = sys.argv[1]
	input = json.loads(inputObject)
	res = entrypoint(input)
else:
	res = entrypoint()

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

func ExecuteJsInterpreter(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	if funcName == "execute" || funcName == "query" {
		funcName = "main"
	}

	wasimodule := contractVm.GetImportModule(wasmedge.WASI)
	dir := filepath.Dir(context.Env.Contract.FilePath)
	fileName := filepath.Base(context.Env.Contract.FilePath)
	inputFileName := "main.js"
	inputFile := path.Join(dir, inputFileName)
	strcontent := fmt.Sprintf(`
import * as std from "std";
import * as os from "os";
import * as contract from "./%s";

const inputData = std.parseExtJSON(scriptArgs[1]);
const res = contract.%s(inputData);

const filename = "./%s"
const fd = os.open(filename, "rw");

const f = std.open(filename, "w");
f.puts(res);
f.close();

	`, fileName, funcName, types.WasiResultFile)

	err := os.WriteFile(inputFile, []byte(strcontent), 0644)
	if err != nil {
		return nil, err
	}

	wasimodule.InitWasi(
		[]string{
			``,
			inputFileName,
			string(context.Env.CurrentCall.CallData),
		},
		[]string{},
		[]string{
			fmt.Sprintf(`.:%s`, dir),
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

func tinygoAllocateMem(ctx *Context, size int32) (int32, error) {
	addr := ctx.Env.Contract.Address
	contractCtx, ok := ctx.ContractRouter[addr.String()]
	if !ok {
		return int32(0), sdkerr.Wrapf(sdkerr.Error{}, "contract context not found for address %s", addr.String())
	}
	return tinygoAllocateMemVm(contractCtx.Vm, size)
}

func tinygoAllocateMemVm(vm *wasmedge.VM, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	result, err := vm.Execute("malloc", size)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func quickjsAllocateMem(ctx *Context, size int32) (int32, error) {
	addr := ctx.Env.Contract.Address
	contractCtx, ok := ctx.ContractRouter[addr.String()]
	if !ok {
		return int32(0), sdkerr.Wrapf(sdkerr.Error{}, "contract context not found for address %s", addr.String())
	}
	return quickjsAllocateMemVm(contractCtx.Vm, size)
}

func quickjsAllocateMemVm(vm *wasmedge.VM, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	result, err := vm.Execute("malloc", size)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func allocateInputNear(vm *wasmedge.VM, input []byte) (int32, error) {
	inputLen := len(input)

	// Allocate memory for the input, and get a pointer to it.
	// Include a byte for the NULL terminator we add below.
	allocateResult, err := vm.Execute("malloc", int32(inputLen+1))
	if err != nil {
		return 0, err
	}
	inputPointer := allocateResult[0].(int32)

	// Write the subject into the memory.
	mod := vm.GetActiveModule()
	mem := mod.FindMemory("memory")
	// C-string terminates by NULL.
	input = append(input, []byte{0}...)
	err = mem.SetData(input, uint(inputPointer), uint(inputLen+1))
	if err != nil {
		return 0, err
	}
	return inputPointer, nil
}

func allocateInputJs(vm *wasmedge.VM, input []byte) (int32, error) {
	inputLen := len(input)

	// Allocate memory for the input, and get a pointer to it.
	// Include a byte for the NULL terminator we add below.
	allocateResult, err := vm.Execute("stackAlloc", int32(inputLen+1))
	if err != nil {
		return 0, err
	}
	inputPointer := allocateResult[0].(int32)

	// Write the subject into the memory.
	mod := vm.GetActiveModule()
	mem := mod.FindMemory("memory")
	memData, err := mem.GetData(uint(inputPointer), uint(inputLen+1))
	if err != nil {
		return 0, err
	}
	copy(memData, input)

	// C-string terminates by NULL.
	memData[inputLen] = 0

	return inputPointer, nil
}
