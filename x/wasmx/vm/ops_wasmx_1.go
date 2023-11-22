package vm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"mythos/v1/x/wasmx/types"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

var (
	SSTORE_GAS_WASMX = 20_000
)

type WasmxJsonLog struct {
	Type   string
	Data   []byte
	Topics [][32]byte
}

// getCallData(): ArrayBuffer
func getCallData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	fmt.Println("---getCallData--", ctx.Env.CurrentCall.CallData)
	ptr, err := allocateWriteMem(ctx, callframe, ctx.Env.CurrentCall.CallData)
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
	ptr, err := allocateWriteMem(ctx, callframe, addr)
	if err != nil {
		return nil, wasmedge.Result_Fail

	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// storageStore(key: ArrayBuffer, value: ArrayBuffer)
func wasmxStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	key, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data, err := readMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	ctx := context.(*Context)
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_WASMX), "wasmx")

	ctx.ContractStore.Set(key, data)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// storageLoad(key: ArrayBuffer): ArrayBuffer
func wasmxStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	keybz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = make([]byte, 32)
	}
	newptr, err := allocateWriteMem(ctx, callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

func wasmxLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
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

func wasmxFinish(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.ReturnData = data
	// fmt.Println("---getCalwasmxFinishlData--", data)
	returns := make([]interface{}, 0)
	// terminate the WASM execution
	// return returns, wasmedge.Result_Terminate
	return returns, wasmedge.Result_Success
}

func wasmxRevert(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = data
	ctx.ReturnData = data
	return returns, wasmedge.Result_Fail
}

// message: usize, fileName: usize, line: u32, column: u32
func asAbort(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	message, _ := readMemFromPtr(callframe, params[0])
	fileName, _ := readMemFromPtr(callframe, params[1])
	ctx := context.(*Context)
	ctx.GetContext().Logger().Debug(fmt.Sprintf("wasmx_env_1: ABORT: %s, %s. line: %d, column: %d", readJsString(message), readJsString(fileName), params[2], params[3]))
	return wasmxRevert(context, callframe, params)
}

func asConsoleLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	message, _ := readMemFromPtr(callframe, params[0])
	ctx := context.(*Context)
	fmt.Println("-asConsoleLog", readJsString(message))
	ctx.GetContext().Logger().Debug(fmt.Sprintf("wasmx_env_1: console.log: %s", readJsString(message)))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func readJsString(arr []byte) string {
	msg := []byte{}
	for i, char := range arr {
		if i%2 == 0 {
			msg = append(msg, char)
		}
	}
	return string(msg)
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
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoad, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32_, wasmxStorageStore, context, 0))
	env.AddFunction("log", wasmedge.NewFunction(functype_i32_, wasmxLog, context, 0))
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

	env.AddFunction("abort", wasmedge.NewFunction(functype_i32i32i32i32_, asAbort, context, 0))
	env.AddFunction("console.log", wasmedge.NewFunction(functype_i32_, asConsoleLog, context, 0))
	return env
}

const AS_PTR_LENGHT_OFFSET = int32(4)
const AS_ARRAY_BUFFER_TYPE = int32(1)

// https://www.assemblyscript.org/runtime.html#memory-layout
// Name	   Offset	Type	Description
// mmInfo	-20	    usize	Memory manager info
// gcInfo	-16	    usize	Garbage collector info
// gcInfo2	-12	    usize	Garbage collector info
// rtId 	-8	    u32	    Unique id of the concrete class
// rtSize	-4	    u32	    Size of the data following the header
//           0		Payload starts here

func readMemFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error) {
	lengthbz, err := readMem(callframe, pointer.(int32)-AS_PTR_LENGHT_OFFSET, int32(AS_PTR_LENGHT_OFFSET))
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lengthbz)
	data, err := readMem(callframe, pointer, int32(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func allocateMemVm(vm *wasmedge.VM, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	fmt.Println("--allocateMemVm--", types.MEMORY_EXPORT_AS, size, AS_ARRAY_BUFFER_TYPE)
	result, err := vm.Execute(types.MEMORY_EXPORT_AS, size, AS_ARRAY_BUFFER_TYPE)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func allocateWriteMem(ctx *Context, callframe *wasmedge.CallingFrame, data []byte) (int32, error) {
	ptr, err := allocateMemVm(ctx.MustGetVmFromContext(), int32(len(data)))
	if err != nil {
		return ptr, err
	}
	err = writeMem(callframe, data, ptr)
	if err != nil {
		return ptr, err
	}
	return ptr, nil
}
