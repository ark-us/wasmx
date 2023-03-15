package ewasm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"wasmx/x/wasmx/types"
)

func ewasm_wrapper(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	wrapper := context.(EwasmFunctionWrapper)
	fmt.Println("Go: ewasm_wrapper entering", wrapper.Name, params)
	returns, err := wrapper.Vm.Execute(wrapper.Name, params...)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("Go: ewasm_wrapper: leaving", wrapper.Name, returns)
	return returns, wasmedge.Result_Success
}

func InstantiateWasm(contractVm *wasmedge.VM, filePath string, wasmbuffer []byte) error {
	var err error
	if wasmbuffer == nil {
		err = contractVm.LoadWasmFile(filePath)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed %s", filePath)
		}
	} else {
		err = contractVm.LoadWasmBuffer(wasmbuffer)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed from buffer")
		}
	}
	err = contractVm.Validate()
	if err != nil {
		return sdkerrors.Wrapf(err, "wasm validate failed")
	}
	err = contractVm.Instantiate()
	if err != nil {
		return sdkerrors.Wrapf(err, "wasm instantiate failed")
	}
	return nil
}

func InitiateWasm(context *Context, filePath string, wasmbuffer []byte, systemDeps []string) (*wasmedge.VM, []func(), error) {
	wasmedge.SetLogErrorLevel()
	// wasmedge.SetLogDebugLevel()
	// conf := wasmedge.NewConfigure()
	// conf.SetStatisticsInstructionCounting(true)
	// conf.SetStatisticsTimeMeasuring(true)
	// contractVm := wasmedge.NewVMWithConfig(conf)
	contractVm := wasmedge.NewVM()
	var cleanups []func()

	// set default
	if len(systemDeps) == 0 {
		systemDeps = append(systemDeps, EWASM_VM_EXPORT+"1")
	}

	for _, systemDep := range systemDeps {
		handler := SystemDepHandler[systemDep]
		releaseFn, err := handler(context, contractVm)
		cleanups = append(cleanups, releaseFn...)
		if err != nil {
			return nil, cleanups, err
		}
	}

	err := InstantiateWasm(contractVm, filePath, wasmbuffer)
	return contractVm, cleanups, err
}

func runCleanups(cleanups []func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}
}

func buildExecutionContextClassic(filePath string, env types.Env, storeKey []byte, conf *wasmedge.Configure, systemDeps []string) (*ContractContext, error) {
	contractCtx := &ContractContext{
		FilePath:         filePath,
		ContractStoreKey: storeKey,
		SystemDeps:       systemDeps,
	}
	return contractCtx, nil
}

func AnalyzeWasm(wasmbuffer []byte) (types.AnalysisReport, error) {
	report := types.AnalysisReport{}
	loader := wasmedge.NewLoader()
	ast, err := loader.LoadBuffer(wasmbuffer)
	if err != nil {
		return report, err
	}
	imports := ast.ListImports()
	exports := ast.ListExports()

	for _, mexport := range exports {
		fname := mexport.GetExternalName()
		var part string
		if strings.Contains(fname, EWASM_VM_EXPORT) {
			part = EWASM_VM_EXPORT
		} else if strings.Contains(fname, EWASM_INTERPRETER_EXPORT) {
			part = EWASM_INTERPRETER_EXPORT
		}
		if part != "" {
			dep := parseDependency(fname, EWASM_VM_EXPORT)
			err := VerifyEnv(dep, imports)
			if err != nil {
				return report, sdkerrors.Wrapf(types.ErrCreateFailed, "wasm module requires imports not supported by the %s version: %s", fname, err.Error())
			}
			report.Dependencies = append(report.Dependencies, dep)
		}
	}

	ast.Release()
	loader.Release()
	return report, nil
}

// Returns the hex address of the interpreter if exists or the version string
func parseDependency(contractVersion string, part string) string {
	dep := contractVersion
	if strings.Contains(contractVersion, part) {
		v := contractVersion[len(part):]
		if len(v) > 2 && v[0:2] == "0x" {
			dep = v
		}
	}
	return dep
}

func ExecuteWasm(
	ctx sdk.Context,
	filePath string,
	funcName string,
	env types.Env,
	messageInfo types.MessageInfo,
	msg []byte,
	storeKey []byte, kvstore types.KVStore,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []string,
	dependencies []types.ContractDependency,
) (types.ContractResponse, error) {
	var err error
	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not decode wasm execution message")
	}

	conf := wasmedge.NewConfigure()
	var contractRouter ContractRouter = make(map[string]ContractContext)
	context := &Context{
		Ctx:               ctx,
		GasMeter:          gasMeter,
		Env:               &env,
		ContractStore:     kvstore,
		CallContext:       messageInfo,
		CosmosHandler:     cosmosHandler,
		Calldata:          ethMsg.Data,
		Callvalue:         messageInfo.Funds,
		ContractRouter:    contractRouter,
		ExecutionBytecode: []byte{},
	}
	for _, dep := range dependencies {
		contractContext, err := buildExecutionContextClassic(dep.FilePath, env, dep.StoreKey, conf, dep.SystemDeps)
		if err != nil {
			return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not build dependenci execution context for %s", dep.Address)
		}
		context.ContractRouter[dep.Address.String()] = *contractContext
	}
	// add itself
	selfContext, err := buildExecutionContextClassic(filePath, env, storeKey, conf, systemDeps)
	if err != nil {
		return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not build dependenci execution context for self %s", env.Contract.Address.String())
	}
	context.ContractRouter[env.Contract.Address.String()] = *selfContext

	contractVm, cleanups, err := InitiateWasm(context, filePath, nil, systemDeps)
	if err != nil {
		runCleanups(cleanups)
		return types.ContractResponse{}, err
	}

	setExecutionBytecode(context, contractVm, funcName)

	_, err = contractVm.Execute(funcName)
	runCleanups(cleanups)
	conf.Release()
	contractVm.Release()

	response := handleContractResponse(funcName, context.ReturnData, context.Logs)
	if err != nil {
		return response, err
	}

	return response, nil
}

// deploymentBytecode = constructorBytecode + runtimeBytecode
// codesize/codecopy at deployment = deploymentBytecode + args
// codesize/codecopy at runtime execution = runtimeBytecode
func setExecutionBytecode(context *Context, contractVm *wasmedge.VM, funcName string) {
	fnList, _ := contractVm.GetFunctionList()

	if slices.Contains(fnList, "evm_bytecode") {
		retvalues, err := contractVm.Execute("evm_bytecode")
		if err != nil {
			return
		}

		memoffset := retvalues[0].(int32)
		constructorLength := retvalues[1].(int32)
		runtimeLength := retvalues[2].(int32)
		activeMemory := contractVm.GetActiveModule().FindMemory("memory")
		if activeMemory == nil {
			return
		}
		executionBytecode, err := activeMemory.GetData(uint(memoffset+constructorLength), uint(runtimeLength))
		if err != nil {
			return
		}

		if funcName == "instantiate" {
			constructorBytecode, err := activeMemory.GetData(uint(memoffset), uint(constructorLength))
			if err != nil {
				return
			}
			executionBytecode = append(constructorBytecode, executionBytecode...)
			executionBytecode = append(executionBytecode, context.Calldata...)
		}

		context.ExecutionBytecode = executionBytecode
	}
}

func handleContractResponse(funcName string, data []byte, logs []EwasmLog) types.ContractResponse {
	var events []types.Event
	for i, log := range logs {
		var attributes []types.EventAttribute
		attributes = append(attributes, types.EventAttribute{
			Key:   AttributeKeyIndex,
			Value: fmt.Sprint(i),
		})
		attributes = append(attributes, types.EventAttribute{
			Key:   AttributeKeyData,
			Value: "0x" + hex.EncodeToString(log.Data),
		})
		for j, topic := range log.Topics {
			attributes = append(attributes, types.EventAttribute{
				Key:   AttributeKeyTopic + fmt.Sprint(j),
				Value: "0x" + hex.EncodeToString(topic),
			})
		}
		events = append(events, types.Event{
			Type:       EventTypeEwasmLog + fmt.Sprint(i),
			Attributes: attributes,
		})
	}

	return types.ContractResponse{
		Data:   data,
		Events: events,
	}
}
