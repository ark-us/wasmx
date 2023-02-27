package ewasm

import (
	"encoding/json"
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

const coreOpcodesModule = "./ewasm/contracts/ewasm.wasm"

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
	ewasmEnv := BuildEwasmEnv(context)

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
		return context.ReturnData, err
	}
	if len(res) == 0 {
		return make([]byte, 0), nil
	}

	return context.ReturnData, nil
}
