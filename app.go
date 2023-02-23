package main

import (
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
	// Calldata []byte
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
	// wrapper := context.(EwasmFunctionWrapper)
	returns := make([]interface{}, 1)
	returns[0] = 25
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
	fmt.Println("Go: finish", result)
	return returns, wasmedge.Result_Success
}

func buildEwasmEnv() *wasmedge.Module {
	var ewasmEnv = wasmedge.NewModule("env")

	getUseGas := wasmedge.NewFunction(wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64},
		[]wasmedge.ValType{},
	), ewasm_useGas, nil, 0)
	ewasmEnv.AddFunction("ethereum_useGas", getUseGas)

	getCallDataSize := wasmedge.NewFunction(wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	), ewasm_getCallDataSize, nil, 0)
	ewasmEnv.AddFunction("ethereum_getCallDataSize", getCallDataSize)

	getFinish := wasmedge.NewFunction(wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	), ewasm_finish, nil, 0)
	ewasmEnv.AddFunction("ethereum_finish", getFinish)

	return ewasmEnv
}

func main() {
	var err error
	wasmedge.SetLogErrorLevel()

	// globalVarType := wasmedge.NewGlobalType(wasmedge.ValType_I32, wasmedge.ValMut_Var)

	// global_sp := wasmedge.NewGlobal(globalVarType, int32(-32))
	// global_init := wasmedge.NewGlobal(globalVarType, int32(0))
	// global_cb_dest := wasmedge.NewGlobal(globalVarType, int32(0))

	var ewasmVm = wasmedge.NewVM()
	var ewasmEnv = buildEwasmEnv()
	// ewasmEnv.AddGlobal("sp", global_sp)
	// ewasmEnv.AddGlobal("init", global_init)
	// ewasmEnv.AddGlobal("cb_dest", global_cb_dest)
	ewasmVm.RegisterModule(ewasmEnv)
	fmt.Println("Go: eWasm module registered")
	ewasmVm.LoadWasmFile("./modules/ewasm.wasm")
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

	// contractEnv.AddGlobal("sp", global_sp)
	// contractEnv.AddGlobal("init", global_init)
	// contractEnv.AddGlobal("cb_dest", global_cb_dest)
	contractVm.RegisterModule(contractEnv)

	fmt.Println("Go: Contract module registered")

	// Instantiate wasm
	contractVm.LoadWasmFile("./modules/contract.wasm")
	fmt.Println("Go: Contract loaded")
	contractVm.Validate()
	contractVm.Instantiate()

	fmt.Println("Go: Contract instantiated")

	var res []interface{}
	res, err = contractVm.Execute("main")
	if err == nil {
		fmt.Println("Run main: ", res[0].(int32))
	} else {
		fmt.Println("Run main FAILED")
	}

	// global_sp.Release()
	// global_init.Release()
	// global_cb_dest.Release()

	contractVm.Release()
	contractEnv.Release()
	ewasmVm.Release()
	ewasmEnv.Release()
}
