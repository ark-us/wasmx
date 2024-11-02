package vm

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
)

var (
	SSTORE_GAS_WASMX = 20_000
)

// cosmos-sdk key-value max sizes
// MaxKeyLength: 128K - 1
// MaxValueLength: 2G - 1

type WasmxJsonLog struct {
	Type   string
	Data   []byte
	Topics [][32]byte
}

// getCallData(): ArrayBuffer
func getCallData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, ctx.Env.CurrentCall.CallData)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetCaller(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addr := types.PaddLeftTo32(ctx.Env.CurrentCall.Sender.Bytes())
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, addr)
	if err != nil {
		return nil, wasmedge.Result_Fail

	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetAddress(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addr := types.PaddLeftTo32(ctx.Env.Contract.Address.Bytes())
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, addr)
	if err != nil {
		return nil, wasmedge.Result_Fail

	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// storageStore(key: ArrayBuffer, value: ArrayBuffer)
func wasmxStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	key, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_WASMX), "wasmx")
	ctx.ContractStore.Set(key, data)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// storageLoad(key: ArrayBuffer): ArrayBuffer
func wasmxStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	keybz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	// if len(data) == 0 {
	// 	data = make([]byte, 32)
	// }
	newptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

func wasmxStorageLoadRange(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req StorageRange
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	values := make([][]byte, 0)
	startKey := req.StartKey
	endKey := req.EndKey
	if len(startKey) == 0 {
		startKey = nil
	}
	if len(endKey) == 0 {
		endKey = nil
	}

	var iter types.Iterator
	if req.Reverse {
		iter = ctx.ContractStore.ReverseIterator(startKey, endKey)
	} else {
		iter = ctx.ContractStore.Iterator(startKey, endKey)
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		values = append(values, iter.Value())
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	newptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

func wasmxStorageLoadRangePairs(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req StorageRange
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	pairs := make([]StoragePair, 0)
	startKey := req.StartKey
	endKey := req.EndKey
	if len(startKey) == 0 {
		startKey = nil
	}
	if len(endKey) == 0 {
		endKey = nil
	}

	var iter types.Iterator
	if req.Reverse {
		iter = ctx.ContractStore.ReverseIterator(startKey, endKey)
	} else {
		iter = ctx.ContractStore.Iterator(startKey, endKey)
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		pair := StoragePair{Key: iter.Key(), Value: iter.Value()}
		pairs = append(pairs, pair)
	}
	response := StoragePairs{Values: pairs}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	newptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

func wasmxLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var wlog WasmxJsonLog
	err = json.Unmarshal(data, &wlog)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	logtype := wlog.Type
	// cosmos log attributes cannot be empty
	if logtype == "" {
		logtype = LOG_TYPE_WASMX
	}
	dependency := types.DEFAULT_SYS_DEP
	if len(ctx.Env.Contract.SystemDeps) > 0 {
		dependency = ctx.Env.Contract.SystemDeps[0]
	}
	log := WasmxLog{
		Type:             logtype,
		ContractAddress:  ctx.Env.Contract.Address,
		SystemDependency: dependency,
		Data:             wlog.Data,
		Topics:           wlog.Topics,
	}
	ctx.Logs = append(ctx.Logs, log)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// SSTORE key_ptr: i32, value_ptr: i32,
func storageStoreGlobal(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// SLOAD key_ptr: i32, result_ptr: i32
func storageLoadGlobal(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxGetReturnData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	newptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, ctx.ReturnData)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

func wasmxGetFinishData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	newptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, ctx.FinishData)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

func wasmxSetFinishData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.FinishData = data
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxFinish(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = data
	ctx.ReturnData = data
	// TODO fixme wasmedge fails if we terminate
	// terminate the WASM execution
	// return returns, wasmedge.Result_Terminate
	return returns, wasmedge.Result_Success
}

func wasmxRevert(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = data
	ctx.ReturnData = data
	return returns, wasmedge.Result_Fail
}

// message: usize, fileName: usize, line: u32, column: u32
func asAbort(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, _ := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	fileName, _ := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[1])
	ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx_env_1: ABORT: %s, %s. line: %d, column: %d", ctx.MemoryHandler.ReadJsString(message), ctx.MemoryHandler.ReadJsString(fileName), params[2], params[3]))
	return wasmxRevert(context, callframe, params)
}

func asConsoleLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, err := ctx.MemoryHandler.ReadStringFromPtr(callframe, params[0])
	if err == nil {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.log: %s", message))
	} else {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.log error: %s", err.Error()))
	}
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asConsoleInfo(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, err := ctx.MemoryHandler.ReadStringFromPtr(callframe, params[0])
	if err == nil {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.info: %s", message))
	} else {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.info error: %s", err.Error()))
	}
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asConsoleError(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, err := ctx.MemoryHandler.ReadStringFromPtr(callframe, params[0])
	if err == nil {
		ctx.Logger(ctx.Ctx).Error(fmt.Sprintf("wasmx: console.error: %s", message))
	} else {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.error error: %s", err.Error()))
	}
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asConsoleDebug(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, _ := ctx.MemoryHandler.ReadStringFromPtr(callframe, params[0])
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("wasmx: console.debug: %s", message))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asDateNow(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = float64(time.Now().UTC().UnixMilli())
	return returns, wasmedge.Result_Success
}

// TODO - move this only for non-deterministic contracts
func asSeed(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = float64(binary.LittleEndian.Uint64(b[:]))
	return returns, wasmedge.Result_Success
}

// function env.trace?(message: usize, n: i32, a0..4?: f64): void
// function env.seed?(): f64

func BuildWasmxEnv1(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasmx")
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)

	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i32, getCallData, context, 0))
	env.AddFunction("getCaller", wasmedge.NewFunction(functype__i32, wasmxGetCaller, context, 0))
	env.AddFunction("getAddress", wasmedge.NewFunction(functype__i32, wasmxGetAddress, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoad, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32_, wasmxStorageStore, context, 0))
	env.AddFunction("storageLoadRange", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoadRange, context, 0))
	env.AddFunction("storageLoadRangePairs", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoadRangePairs, context, 0))
	env.AddFunction("log", wasmedge.NewFunction(functype_i32_, wasmxLog, context, 0))
	env.AddFunction("getReturnData", wasmedge.NewFunction(functype__i32, wasmxGetReturnData, context, 0))
	env.AddFunction("getFinishData", wasmedge.NewFunction(functype__i32, wasmxGetFinishData, context, 0))
	env.AddFunction("setFinishData", wasmedge.NewFunction(functype_i32_, wasmxSetFinishData, context, 0))
	// TODO some precompiles use setReturnData instead of setFinishData
	env.AddFunction("setReturnData", wasmedge.NewFunction(functype_i32_, wasmxSetFinishData, context, 0))
	env.AddFunction("finish", wasmedge.NewFunction(functype_i32_, wasmxFinish, context, 0))
	env.AddFunction("revert", wasmedge.NewFunction(functype_i32_, wasmxRevert, context, 0))
	env.AddFunction("storageLoad_global", wasmedge.NewFunction(functype_i32i32_, storageLoadGlobal, context, 0))
	env.AddFunction("storageStore_global", wasmedge.NewFunction(functype_i32i32_, storageStoreGlobal, context, 0))

	return env
}

func BuildAssemblyScriptEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("env")
	functype_i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype__f64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_F64},
	)

	env.AddFunction("abort", wasmedge.NewFunction(functype_i32i32i32i32_, asAbort, context, 0))
	env.AddFunction("console.log", wasmedge.NewFunction(functype_i32_, asConsoleLog, context, 0))
	env.AddFunction("console.info", wasmedge.NewFunction(functype_i32_, asConsoleInfo, context, 0))
	env.AddFunction("console.error", wasmedge.NewFunction(functype_i32_, asConsoleError, context, 0))
	env.AddFunction("console.debug", wasmedge.NewFunction(functype_i32_, asConsoleDebug, context, 0))
	env.AddFunction("Date.now", wasmedge.NewFunction(functype__f64, asDateNow, context, 0))
	env.AddFunction("seed", wasmedge.NewFunction(functype__f64, asSeed, context, 0))
	return env
}
