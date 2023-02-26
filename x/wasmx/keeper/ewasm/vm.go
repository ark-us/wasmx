package ewasm

import (
	"encoding/json"
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

const coreOpcodesModule = "./ewasm/contracts/ewasm.wasm"

type WasmEthMessage struct {
	Readonly bool
	Data     []byte
}

type Context struct {
	Calldata   []byte
	ReturnData []byte
}

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
}

func ewasm_wrapper(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	wrapper := context.(EwasmFunctionWrapper)
	fmt.Println("Go: ewasm_wrapper entering", wrapper.Name)
	returns, err := wrapper.Vm.Execute(wrapper.Name, params...)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("Go: ewasm_wrapper: leaving")
	return returns, wasmedge.Result_Success
}

func ewasm_getCallDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: getCallDataSize")
	ctx := context.(Context)
	returns := make([]interface{}, 1)
	returns[0] = len(ctx.Calldata)
	return returns, wasmedge.Result_Success
}

func ewasm_useGas(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// Set the returns
	returns := make([]interface{}, 1)
	fmt.Println("Go: useGas")
	return returns, wasmedge.Result_Success
}

func ewasm_finish(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
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

	// Set the returns
	returns := make([]interface{}, 1)
	returns[0] = result
	ctx.ReturnData = result
	fmt.Println("Go: finish", result)
	return returns, wasmedge.Result_Success
}

func buildEwasmEnv(context Context) *wasmedge.Module {
	var ewasmEnv = wasmedge.NewModule("env")

	getUseGas := wasmedge.NewFunction(wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64},
		[]wasmedge.ValType{},
	), ewasm_useGas, context, 0)
	ewasmEnv.AddFunction("ethereum_useGas", getUseGas)

	getCallDataSize := wasmedge.NewFunction(wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	), ewasm_getCallDataSize, context, 0)
	ewasmEnv.AddFunction("ethereum_getCallDataSize", getCallDataSize)

	getFinish := wasmedge.NewFunction(wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	), ewasm_finish, context, 0)
	ewasmEnv.AddFunction("ethereum_finish", getFinish)

	return ewasmEnv
}

func ExecuteWasm(filePath string, funcName string, msg []byte) ([]byte, error) {
	var err error

	var ethMsg WasmEthMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return nil, err
	}

	wasmedge.SetLogErrorLevel()

	store := wasmedge.NewStore()
	ewasmVm := wasmedge.NewVMWithStore(store)
	context := Context{Calldata: ethMsg.Data}
	ewasmEnv := buildEwasmEnv(context)

	ewasmVm.RegisterModule(ewasmEnv)
	fmt.Println("Go: eWasm module registered")
	ewasmVm.LoadWasmFile(coreOpcodesModule)
	ewasmVm.Validate()
	ewasmVm.Instantiate()
	fmt.Println("Go: eWasm module instantiate")

	ewasmFnList, ewasmFnTypes := ewasmVm.GetFunctionList()
	fmt.Println("Go: ewasmFnList", ewasmFnList, ewasmFnTypes)

	var contractVm = wasmedge.NewVM()
	var contractEnv = wasmedge.NewModule("ewasm")

	for i, name := range ewasmFnList {
		data := EwasmFunctionWrapper{Name: name, Vm: ewasmVm}
		fntype := ewasmFnTypes[i]

		var wrappedFn = wasmedge.NewFunction(fntype, ewasm_wrapper, data, 0)
		contractEnv.AddFunction(name, wrappedFn)
	}

	contractVm.RegisterModule(contractEnv)

	fmt.Println("Go: Contract module registered")

	// Instantiate wasm
	contractVm.LoadWasmFile(filePath)
	fmt.Println("Go: Contract loaded")
	contractVm.Validate()
	contractVm.Instantiate()

	fmt.Println("Go: Contract instantiated")

	var res []interface{}
	_, err = contractVm.Execute(funcName)

	contractVm.Release()
	contractEnv.Release()
	ewasmVm.Release()
	ewasmEnv.Release()

	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return make([]byte, 0), nil
	}

	return context.ReturnData, nil
}
