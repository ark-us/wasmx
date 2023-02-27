package ewasm

import (
	"encoding/json"
	"fmt"

	"wasmx/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

const coreOpcodesModule = "../ewasm/contracts/ewasm.wasm"

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

func ExecuteWasm(filePath string, funcName string, env types.Env, messageInfo types.MessageInfo, msg []byte) (types.ContractResponse, error) {
	var err error

	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, err
	}

	wasmedge.SetLogErrorLevel()
	// conf := wasmedge.NewConfigure()
	// conf.SetStatisticsInstructionCounting(true)
	// stat := wasmedge.NewStatistics()
	// loader := wasmedge.NewLoaderWithConfig(conf)
	// validator := wasmedge.NewValidatorWithConfig(conf)
	// executor := wasmedge.NewExecutorWithConfigAndStatistics(conf, stat)

	store := wasmedge.NewStore()
	ewasmVm := wasmedge.NewVMWithStore(store)
	callvalue, ok := sdk.NewIntFromString(messageInfo.Funds[0].Amount)
	if !ok {
		return types.ContractResponse{}, fmt.Errorf("invalid funds")
	}

	context := Context{
		Env:         env,
		CallContext: messageInfo,
		Calldata:    ethMsg.Data,
		Callvalue:   callvalue.BigInt(),
	}
	ewasmEnv := BuildEwasmEnv(&context)

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
	res, err = contractVm.Execute(funcName)
	fmt.Println("--vm res", res, err)
	fmt.Println("--vm context.ReturnData", context.ReturnData)

	contractVm.Release()
	contractEnv.Release()
	ewasmVm.Release()
	ewasmEnv.Release()

	response := types.ContractResponse{
		Data: context.ReturnData,
	}

	if err != nil {
		return response, err
	}

	return response, nil
}

func ExecuteWasmClassic(filePath string, funcName string, env types.Env, messageInfo types.MessageInfo, msg []byte) (types.ContractResponse, error) {
	var err error

	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, err
	}

	wasmedge.SetLogErrorLevel()
	// conf := wasmedge.NewConfigure()
	// conf.SetStatisticsInstructionCounting(true)
	// stat := wasmedge.NewStatistics()
	// loader := wasmedge.NewLoaderWithConfig(conf)
	// validator := wasmedge.NewValidatorWithConfig(conf)
	// executor := wasmedge.NewExecutorWithConfigAndStatistics(conf, stat)

	store := wasmedge.NewStore()
	contractVm := wasmedge.NewVMWithStore(store)

	callvalue := sdk.NewInt(0)
	if messageInfo.Funds != nil && len(messageInfo.Funds) > 0 {
		callvalue_, ok := sdk.NewIntFromString(messageInfo.Funds[0].Amount)
		if !ok {
			return types.ContractResponse{}, fmt.Errorf("invalid funds")
		}
		callvalue = callvalue_
	}

	context := Context{
		Env:         env,
		CallContext: messageInfo,
		Calldata:    ethMsg.Data,
		Callvalue:   callvalue.BigInt(),
	}
	ewasmEnv := BuildEwasmEnv(&context)

	contractVm.RegisterModule(ewasmEnv)

	fmt.Println("Go: Contract module registered")

	// Instantiate wasm
	contractVm.LoadWasmFile(filePath)
	fmt.Println("Go: Contract loaded")
	contractVm.Validate()
	contractVm.Instantiate()

	fmt.Println("Go: Contract instantiated")

	var res []interface{}
	res, err = contractVm.Execute(funcName)
	fmt.Println("--vm res", res, err)
	fmt.Println("--vm context.ReturnData", context.ReturnData)

	contractVm.Release()
	ewasmEnv.Release()

	response := types.ContractResponse{
		Data: context.ReturnData,
	}

	if err != nil {
		return response, err
	}

	return response, nil
}
