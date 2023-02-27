package ewasm

import (
	"encoding/hex"
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

func useGas(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	fmt.Println("Go: useGas")
	return returns, wasmedge.Result_Success
}

func getGasLeft(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getGasLeft")
	returns := make([]interface{}, 1)
	returns[0] = int64(1000000)
	return returns, wasmedge.Result_Success
}

func storageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: storageLoad")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func storageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: storageLoad")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getBalance(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBalance")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getExternalBalance(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getExternalBalance")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getAddress(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getAddress")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getCaller(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getCaller")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getCallValue(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getCallValue")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getCallDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getCallDataSize")
	ctx := context.(Context)
	returns := make([]interface{}, 1)
	returns[0] = len(ctx.Calldata)
	return returns, wasmedge.Result_Success
}

func callDataCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: callDataCopy")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getReturnDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getReturnDataSize")
	returns := make([]interface{}, 1)
	returns[0] = 20
	return returns, wasmedge.Result_Success
}

func returnDataCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: returnDataCopy")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getCodeSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getCodeSize")
	returns := make([]interface{}, 1)
	returns[0] = int32(100000)
	return returns, wasmedge.Result_Success
}

func getExternalCodeSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getExternalCodeSize")
	returns := make([]interface{}, 1)
	returns[0] = int32(100000)
	return returns, wasmedge.Result_Success
}

func codeCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: codeCopy")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func externalCodeCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: externalCodeCopy")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getExternalCodeHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getExternalCodeHash")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getTxGasPrice(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getTxGasPrice")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getTxOrigin(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getTxOrigin")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getBlockNumber(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBlockNumber")
	returns := make([]interface{}, 1)
	returns[0] = int64(1)
	return returns, wasmedge.Result_Success
}

func getBlockCoinbase(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBlockCoinbase")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getBlockHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBlockCoinbase")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getBlockGasLimit(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBlockGasLimit")
	returns := make([]interface{}, 1)
	returns[0] = int64(10000000)
	return returns, wasmedge.Result_Success
}

func getBlockTimestamp(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBlockTimestamp")
	returns := make([]interface{}, 1)
	returns[0] = int64(1000000000)
	return returns, wasmedge.Result_Success
}

func getBlockDifficulty(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBlockDifficulty")
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

func getChainId(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getChainId")
	returns := make([]interface{}, 1)
	returns[0] = int32(1000)
	return returns, wasmedge.Result_Success
}

func getBaseFee(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getBaseFee")
	returns := make([]interface{}, 1)
	returns[0] = int32(1)
	return returns, wasmedge.Result_Success
}

func call(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: call")
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

func callCode(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: callCode")
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

func callDelegate(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: callDelegate")
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

func callStatic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: callStatic")
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

func create(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: create")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func create2(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: create2")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func selfDestruct(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: selfDestruct")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func log(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: log")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func finish(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: finish")
	ctx := context.(Context)
	pointer := params[0].(int32)
	size := params[1].(int32)
	mem := callframe.GetMemoryByIndex(0)
	data, err := mem.GetData(uint(pointer), uint(size))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	result := make([]byte, size)
	copy(result, data)

	returns := make([]interface{}, 1)
	returns[0] = result
	ctx.ReturnData = result
	fmt.Println("Go: finish", result)
	return returns, wasmedge.Result_Success
}

func stop(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: stop")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func revert(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: revert")
	ctx := context.(Context)
	pointer := params[0].(int32)
	size := params[1].(int32)
	mem := callframe.GetMemoryByIndex(0)
	data, err := mem.GetData(uint(pointer), uint(size))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	result := make([]byte, size)
	copy(result, data)

	returns := make([]interface{}, 1)
	returns[0] = result
	ctx.ReturnData = result
	fmt.Println("Go: revert", result)
	return returns, wasmedge.Result_Fail
}

func sendCosmosMsg(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: sendCosmosMsg")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func sendCosmosQuery(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: sendCosmosQuery")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func debugPrinti32(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: debugPrinti32", params[0].(int32))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func debugPrinti64(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: debugPrinti64", params[0].(int64))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func debugPrintMemHex(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	pointer := params[0].(int32)
	size := params[1].(int32)
	mem := callframe.GetMemoryByIndex(0)
	data, _ := mem.GetData(uint(pointer), uint(size))
	fmt.Println("Go: debugPrintMemHex", hex.EncodeToString(data))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildEwasmEnv(context Context) *wasmedge.Module {
	ewasmEnv := wasmedge.NewModule("env")

	functype__ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{},
	)
	functype_i64_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64},
		[]wasmedge.ValType{},
	)
	functype__i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i64i32i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i64i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)

	ewasmEnv.AddFunction("ethereum_useGas", wasmedge.NewFunction(functype_i64_, useGas, context, 0))
	ewasmEnv.AddFunction("ethereum_getGasLeft", wasmedge.NewFunction(functype__i64, getGasLeft, context, 0))
	ewasmEnv.AddFunction("ethereum_storageLoad", wasmedge.NewFunction(functype_i32i32_, storageLoad, context, 0))
	ewasmEnv.AddFunction("ethereum_storageStore", wasmedge.NewFunction(functype_i32i32_, storageStore, context, 0))
	ewasmEnv.AddFunction("ethereum_getBalance", wasmedge.NewFunction(functype_i32_, getBalance, context, 0))
	ewasmEnv.AddFunction("ethereum_getExternalBalance", wasmedge.NewFunction(functype_i32i32_, getExternalBalance, context, 0))
	ewasmEnv.AddFunction("ethereum_getAddress", wasmedge.NewFunction(functype_i32_, getAddress, context, 0))
	ewasmEnv.AddFunction("ethereum_getCaller", wasmedge.NewFunction(functype_i32_, getCaller, context, 0))
	ewasmEnv.AddFunction("ethereum_getCallValue", wasmedge.NewFunction(functype_i32_, getCallValue, context, 0))
	ewasmEnv.AddFunction("ethereum_getCallDataSize", wasmedge.NewFunction(functype__i32, getCallDataSize, context, 0))
	ewasmEnv.AddFunction("ethereum_callDataCopy", wasmedge.NewFunction(functype_i32i32i32_, callDataCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_getReturnDataSize", wasmedge.NewFunction(functype__i32, getReturnDataSize, context, 0))
	ewasmEnv.AddFunction("ethereum_returnDataCopy", wasmedge.NewFunction(functype_i32i32i32_, returnDataCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_getCodeSize", wasmedge.NewFunction(functype__i32, getCodeSize, context, 0))
	ewasmEnv.AddFunction("ethereum_getExternalCodeSize", wasmedge.NewFunction(functype_i32_i32, getExternalCodeSize, context, 0))
	ewasmEnv.AddFunction("ethereum_codeCopy", wasmedge.NewFunction(functype_i32i32i32_, codeCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_externalCodeCopy", wasmedge.NewFunction(functype_i32i32i32i32_, externalCodeCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_getExternalCodeHash", wasmedge.NewFunction(functype_i32i32_, getExternalCodeHash, context, 0))
	ewasmEnv.AddFunction("ethereum_getTxGasPrice", wasmedge.NewFunction(functype_i32_, getTxGasPrice, context, 0))
	ewasmEnv.AddFunction("ethereum_getTxOrigin", wasmedge.NewFunction(functype_i32_, getTxOrigin, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockNumber", wasmedge.NewFunction(functype__i64, getBlockNumber, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockCoinbase", wasmedge.NewFunction(functype_i32_, getBlockCoinbase, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockHash", wasmedge.NewFunction(functype_i32i32_, getBlockHash, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockGasLimit", wasmedge.NewFunction(functype__i64, getBlockGasLimit, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockTimestamp", wasmedge.NewFunction(functype__i64, getBlockTimestamp, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockDifficulty", wasmedge.NewFunction(functype_i32_, getBlockDifficulty, context, 0))
	ewasmEnv.AddFunction("ethereum_getChainId", wasmedge.NewFunction(functype_i32_, getChainId, context, 0))
	ewasmEnv.AddFunction("ethereum_getBaseFee", wasmedge.NewFunction(functype_i32_, getBaseFee, context, 0))
	ewasmEnv.AddFunction("ethereum_call", wasmedge.NewFunction(functype_i64i32i32i32i32i32i32_i32, call, context, 0))
	ewasmEnv.AddFunction("ethereum_callCode", wasmedge.NewFunction(functype_i64i32i32i32i32i32i32_i32, callCode, context, 0))
	ewasmEnv.AddFunction("ethereum_callDelegate", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i32, callDelegate, context, 0))
	ewasmEnv.AddFunction("ethereum_callStatic", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i32, callStatic, context, 0))
	ewasmEnv.AddFunction("ethereum_create", wasmedge.NewFunction(functype_i32i32i32i32_, create, context, 0))
	ewasmEnv.AddFunction("ethereum_create2", wasmedge.NewFunction(functype_i32i32i32i32i32_, create2, context, 0))
	ewasmEnv.AddFunction("ethereum_selfDestruct", wasmedge.NewFunction(functype_i32_, selfDestruct, context, 0))
	ewasmEnv.AddFunction("ethereum_log", wasmedge.NewFunction(functype_i32i32i32i32i32i32i32_, log, context, 0))
	ewasmEnv.AddFunction("ethereum_finish", wasmedge.NewFunction(functype_i32i32_, finish, context, 0))
	ewasmEnv.AddFunction("ethereum_stop", wasmedge.NewFunction(functype__, stop, context, 0))
	ewasmEnv.AddFunction("ethereum_revert", wasmedge.NewFunction(functype_i32i32_, revert, context, 0))
	ewasmEnv.AddFunction("ethereum_sendCosmosMsg", wasmedge.NewFunction(functype_i32i32_i32, sendCosmosMsg, context, 0))
	ewasmEnv.AddFunction("ethereum_sendCosmosQuery", wasmedge.NewFunction(functype_i32i32_i32, sendCosmosQuery, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrinti32", wasmedge.NewFunction(functype_i32_, debugPrinti32, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrinti64", wasmedge.NewFunction(functype_i64_, debugPrinti64, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrintMemHex", wasmedge.NewFunction(functype_i32i32_, debugPrintMemHex, context, 0))

	return ewasmEnv
}
