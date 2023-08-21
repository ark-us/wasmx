package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"
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
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "wasmx")
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

	datalen := int32(len(data))
	ptr, err := allocateMemDefaultMalloc(ctx.MustGetVmFromContext(), datalen)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	err = writeMem(callframe, data, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, datalen)
	return returns, wasmedge.Result_Success
}

// getCallData(ptr): ArrayBuffer
func wasiGetCallData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	ctx := context.(*Context)
	datalen := len(ctx.Env.CurrentCall.CallData)
	ptr, err := allocateMemDefaultMalloc(ctx.MustGetVmFromContext(), int32(datalen))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	err = writeMem(callframe, ctx.Env.CurrentCall.CallData, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(datalen))
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
	addrbz, err := readMem(callframe, params[1], params[2])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr, err := sdk.AccAddressFromBech32(string(addrbz))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	value, err := readBigInt(callframe, params[3], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	calldata, err := readMem(callframe, params[4], params[5])
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
	ptr, err := allocateMemDefaultMalloc(ctx.MustGetVmFromContext(), datalen)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	err = writeMem(callframe, responsebz, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, datalen)
	return returns, wasmedge.Result_Success
}

func wasiCallStatic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := readMem(callframe, params[1], params[2])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr, err := sdk.AccAddressFromBech32(string(addrbz))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	calldata, err := readMem(callframe, params[3], params[4])
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
	ptr, err := allocateMemDefaultMalloc(ctx.MustGetVmFromContext(), datalen)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	err = writeMem(callframe, responsebz, ptr)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, datalen)
	return returns, wasmedge.Result_Success
}

func wasiLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
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
	functype_i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i64i32i32i32i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i64i32i32i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)

	// TODO support any 32 bytes address, not only bech32 strings

	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32i32i32_, wasiStorageStore, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32i32_i64, wasiStorageLoad, context, 0))

	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i64, wasiGetCallData, context, 0))
	env.AddFunction("setReturnData", wasmedge.NewFunction(functype_i32i32_, wasiSetReturnData, context, 0))

	env.AddFunction("callClassic", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i64, wasiCallClassic, context, 0))
	env.AddFunction("callStatic", wasmedge.NewFunction(functype_i64i32i32i32i32_i64, wasiCallStatic, context, 0))

	env.AddFunction("log", wasmedge.NewFunction(functype_i32i32_, wasiLog, context, 0))
	return env
}

func ExecuteWasi(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	var res []interface{}
	var err error

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
from wasmx import set_returndata

%s

res = b''
entrypoint = %s
if len(sys.argv) > 1 and sys.argv[1] != "":
	inputObject = sys.argv[1]
	input = json.loads(inputObject)
	res = entrypoint(input)
else:
	res = entrypoint()

set_returndata(res or b'')
`, string(content), funcName)

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

let inputData;
if (scriptArgs.length > 1 && scriptArgs[1] != "") {
	inputData = std.parseExtJSON(scriptArgs[1]);
}
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

func allocateMemDefaultMalloc(vm *wasmedge.VM, size int32) (int32, error) {
	result, err := vm.Execute("malloc", size)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func buildPtr64(ptr int32, datalen int32) uint64 {
	return (uint64(ptr) << uint64(32)) | uint64(datalen)
}
