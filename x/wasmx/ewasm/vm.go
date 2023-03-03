package ewasm

import (
	"encoding/json"
	"fmt"
	"strings"

	"wasmx/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/second-state/WasmEdge-go/wasmedge"
)

const coreOpcodesModule = "../ewasm/contracts/ewasm.wasm"

var (
	INTERPRETER_EXPORT     = "ewasm_interface_version_"
	REQUIRED_IBC_EXPORTS   = []string{}
	REQUIRED_EWASM_EXPORTS = []string{"codesize", "main", "instantiate"}
	// codesize_constructor
)

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

func checkRequiredIbcExports(exports []string) bool {
	// TODO
	return true
}

func InstantiateWasm(contractVm *wasmedge.VM, filePath string, wasmbuffer []byte) error {
	var err error
	if wasmbuffer == nil {
		err = contractVm.LoadWasmFile(filePath)
		if err != nil {
			return err
		}
	} else {
		err = contractVm.LoadWasmBuffer(wasmbuffer)
		if err != nil {
			return err
		}
	}
	err = contractVm.Validate()
	if err != nil {
		return err
	}
	err = contractVm.Instantiate()
	if err != nil {
		return err
	}
	return nil
}

func InitiateWasm(context *Context, filePath string, wasmbuffer []byte) (*wasmedge.VM, *wasmedge.Module, error) {
	wasmedge.SetLogErrorLevel()
	// wasmedge.SetLogDebugLevel()
	// conf := wasmedge.NewConfigure()
	// conf.SetStatisticsInstructionCounting(true)
	// conf.SetStatisticsTimeMeasuring(true)
	// contractVm := wasmedge.NewVMWithConfig(conf)
	ewasmEnv := BuildEwasmEnv(context)
	contractVm := wasmedge.NewVM()
	err := contractVm.RegisterModule(ewasmEnv)
	if err != nil {
		return contractVm, ewasmEnv, err
	}
	// We also register the interpreter
	err = contractVm.RegisterWasmFile("ewasm", coreOpcodesModule)
	if err != nil {
		return contractVm, ewasmEnv, err
	}
	err = InstantiateWasm(contractVm, filePath, wasmbuffer)
	return contractVm, ewasmEnv, err
}

func InitiateWasmWithWrap(context *Context, filePath string, wasmbuffer []byte) (*wasmedge.VM, *wasmedge.Module, *wasmedge.VM, *wasmedge.Module, error) {
	var err error
	wasmedge.SetLogErrorLevel()
	contractVm := wasmedge.NewVM()
	contractEnv := wasmedge.NewModule("ewasm")
	ewasmVm := wasmedge.NewVM()
	ewasmEnv := BuildEwasmEnv(context)

	err = ewasmVm.RegisterModule(ewasmEnv)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	err = InstantiateWasm(ewasmVm, coreOpcodesModule, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	ewasmFnList, ewasmFnTypes := ewasmVm.GetFunctionList()
	for i, name := range ewasmFnList {
		data := EwasmFunctionWrapper{Name: name, Vm: ewasmVm}
		fntype := ewasmFnTypes[i]
		wrappedFn := wasmedge.NewFunction(fntype, ewasm_wrapper, data, 0)
		contractEnv.AddFunction(name, wrappedFn)
	}

	err = contractVm.RegisterModule(contractEnv)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	err = InstantiateWasm(contractVm, filePath, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return contractVm, contractEnv, ewasmVm, ewasmEnv, nil
}

func buildExecutionContextClassic(filePath string, env types.Env, storeKey []byte, conf *wasmedge.Configure) (*ContractContext, error) {
	contractCtx := &ContractContext{
		FilePath:         filePath,
		ContractStoreKey: storeKey,
	}
	return contractCtx, nil
}

func AnalyzeWasm(wasmbuffer []byte) (types.AnalysisReport, error) {
	report := types.AnalysisReport{}
	vm, ewasmEnv, err := InitiateWasm(&Context{}, "", wasmbuffer)
	if err != nil {
		return report, err
	}
	fnames, _ := vm.GetFunctionList()

	// TODO REQUIRED_EWASM_EXPORTS
	// TODO checkRequiredIbcExports

	for _, fname := range fnames {
		if strings.Contains(fname, INTERPRETER_EXPORT) {
			v := fname[24:]
			dep := v
			if len(v) > 2 && v[0:2] != "0x" {
				dep = "ewasm_" + v
			}
			report.Dependencies = append(report.Dependencies, dep)
		}
	}

	ewasmEnv.Release()
	vm.Release()

	return report, nil
}

// TODO remove
// func buildExecutionContextClassic_0(filePath string, env types.Env, kvstore types.KVStore, conf *wasmedge.Configure) (*ContractContext, error) {
// 	var err error
// 	loader := wasmedge.NewLoader()
// 	ast, err := loader.LoadFile(filePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	contractVm := wasmedge.NewVM()
// 	context := Context{
// 		Env:           &env,
// 		ContractStore: kvstore,
// 	}
// 	ewasmEnv := BuildEwasmEnv(&context)
// 	err = contractVm.RegisterModule(ewasmEnv)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = contractVm.GetValidator().Validate(ast)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = contractVm.LoadWasmAST(ast)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// err = contractVm.Validate()
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	contractCtx := &ContractContext{
// 		FilePath:      filePath,
// 		Vm:            contractVm,
// 		VmAst:         ast,
// 		ContractStore: kvstore,
// 		Context:       &context,
// 	}
// 	return contractCtx, nil
// }

// func buildExecutionContextClassic_(filePath string, env types.Env, kvstore types.KVStore, conf *wasmedge.Configure) (*ContractContext, error) {
// 	var err error
// 	stat := wasmedge.NewStatistics()
// 	loader := wasmedge.NewLoaderWithConfig(conf)
// 	validator := wasmedge.NewValidatorWithConfig(conf)
// 	executor := wasmedge.NewExecutorWithConfigAndStatistics(conf, stat)

// 	ast, err := loader.LoadFile(filePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = validator.Validate(ast)
// 	if err != nil {
// 		return nil, err
// 	}

// 	context := Context{
// 		Env:           &env,
// 		ContractStore: kvstore,
// 	}
// 	ewasmEnv := BuildEwasmEnv(&context)
// 	ewasmStore := wasmedge.NewStore()
// 	err = executor.RegisterImport(ewasmStore, ewasmEnv)
// 	if err != nil {
// 		return nil, err
// 	}

// 	contractCtx := &ContractContext{
// 		VmAst:         ast,
// 		VmExecutor:    executor,
// 		ContractStore: kvstore,
// 		Context:       &context,
// 	}
// 	return contractCtx, nil
// }

func ExecuteWasm(filePath string, funcName string, env types.Env, messageInfo types.MessageInfo, msg []byte, kvstore types.KVStore) (types.ContractResponse, error) {
	var err error

	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, err
	}
	context := &Context{
		Env:           &env,
		ContractStore: kvstore,
		CallContext:   messageInfo,
		Calldata:      ethMsg.Data,
		Callvalue:     messageInfo.Funds,
	}
	contractVm, ewasmEnv, err := InitiateWasm(context, filePath, nil)
	if err != nil {
		return types.ContractResponse{}, err
	}
	_, err = contractVm.Execute(funcName)
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

func ExecuteWasmWithDeps(ctx sdk.Context, filePath string, funcName string, env types.Env, messageInfo types.MessageInfo, msg []byte, storeKey []byte, kvstore types.KVStore, dependencies []types.ContractDependency, cosmosHandler types.WasmxCosmosHandler) (types.ContractResponse, error) {
	var err error

	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, err
	}
	conf := wasmedge.NewConfigure()

	store := wasmedge.NewStore()
	contractVm := wasmedge.NewVMWithStore(store)
	var contractRouter ContractRouter = make(map[string]ContractContext)
	context := &Context{
		Ctx:            ctx,
		Env:            &env,
		ContractStore:  kvstore,
		CallContext:    messageInfo,
		CosmosHandler:  cosmosHandler,
		Calldata:       ethMsg.Data,
		Callvalue:      messageInfo.Funds,
		ContractRouter: contractRouter,
	}
	for _, dep := range dependencies {
		contractContext, err := buildExecutionContextClassic(dep.FilePath, env, dep.StoreKey, conf)
		if err != nil {
			return types.ContractResponse{}, err
		}
		context.ContractRouter[dep.Address.String()] = *contractContext
	}
	// add itself
	selfContext, err := buildExecutionContextClassic(filePath, env, storeKey, conf)
	if err != nil {
		return types.ContractResponse{}, err
	}
	context.ContractRouter[env.Contract.Address.String()] = *selfContext

	ewasmEnv := BuildEwasmEnv(context)
	err = contractVm.RegisterModule(ewasmEnv)
	if err != nil {
		return types.ContractResponse{}, err
	}
	// stat := wasmedge.NewStatistics()

	// Instantiate wasm
	err = contractVm.LoadWasmFile(filePath)
	if err != nil {
		return types.ContractResponse{}, err
	}
	err = contractVm.Validate()
	if err != nil {
		return types.ContractResponse{}, err
	}
	err = contractVm.Instantiate()
	if err != nil {
		return types.ContractResponse{}, err
	}
	_, err = contractVm.Execute(funcName)

	contractVm.Release()
	ewasmEnv.Release()
	store.Release() // release after vm is released

	response := types.ContractResponse{
		Data: context.ReturnData,
	}

	if err != nil {
		return response, err
	}

	return response, nil
}
