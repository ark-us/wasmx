package vm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/native"
	"mythos/v1/x/wasmx/vm/wasmutils"
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

func InitiateWasm(context *Context, filePath string, wasmbuffer []byte, systemDeps []types.SystemDep) (*wasmedge.VM, []func(), error) {
	wasmedge.SetLogErrorLevel()
	// wasmedge.SetLogDebugLevel()
	// conf := wasmedge.NewConfigure()
	// conf.SetStatisticsInstructionCounting(true)
	// conf.SetStatisticsTimeMeasuring(true)
	// contractVm := wasmedge.NewVMWithConfig(conf)
	contractVm := wasmedge.NewVM()
	var cleanups []func()
	var err error

	// first in, last cleaned up
	cleanups = append(cleanups, contractVm.Release)

	// set default
	if len(systemDeps) == 0 {
		label := types.EWASM_VM_EXPORT + "1"
		systemDeps = append(systemDeps, types.SystemDep{Role: label, Label: label})
	}

	cleanups, err = initiateWasmDeps(context, contractVm, systemDeps, cleanups)
	if err != nil {
		return nil, cleanups, err
	}

	if filePath != "" || len(wasmbuffer) > 0 {
		err = wasmutils.InstantiateWasm(contractVm, filePath, wasmbuffer)
	}
	return contractVm, cleanups, err
}

func initiateWasmDeps(context *Context, contractVm *wasmedge.VM, systemDeps []types.SystemDep, cleanups []func()) ([]func(), error) {
	for _, systemDep := range systemDeps {
		// system deps of system deps
		cleanups, err := initiateWasmDeps(context, contractVm, systemDep.Deps, cleanups)
		if err != nil {
			return cleanups, err
		}

		handler, found := SystemDepHandler[systemDep.Role]
		if !found {
			handler, found = SystemDepHandler[systemDep.Label]
		}
		if found {
			releaseFn, err := handler(context, contractVm, &systemDep)
			cleanups = append(cleanups, releaseFn...)
			if err != nil {
				return cleanups, err
			}
		}
	}
	return cleanups, nil
}

// run in inverse order
func runCleanups(cleanups []func()) {
	for i := len(cleanups) - 1; i >= 0; i-- {
		cleanups[i]()
	}
}

func buildExecutionContextClassic(filePath string, bytecode []byte, codeHash []byte, storeKey []byte, systemDeps []types.SystemDep, conf *wasmedge.Configure) *ContractContext {
	contractCtx := &ContractContext{
		FilePath:         filePath,
		ContractStoreKey: storeKey,
		SystemDeps:       systemDeps,
		Bytecode:         bytecode,
		CodeHash:         codeHash,
	}
	return contractCtx
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
		var dep string
		if strings.Contains(fname, types.EWASM_VM_EXPORT) {
			dep = parseDependency(fname, types.EWASM_VM_EXPORT)
		} else if strings.Contains(fname, types.WASMX_VM_EXPORT) {
			dep = parseDependency(fname, types.WASMX_VM_EXPORT)
		} else if strings.Contains(fname, types.SYS_VM_EXPORT) {
			dep = parseDependency(fname, types.SYS_VM_EXPORT)
		} else if fname == types.EWASM_ENV_0 {
			dep = types.EWASM_ENV_1
		} else if fname == types.CW_ENV_8 {
			dep = parseDependency(fname, types.CW_VM_EXPORT)
		}
		if dep != "" {
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

func AotCompile(inPath string, outPath string) error {
	// Create Configure
	// conf := wasmedge.NewConfigure(wasmedge.THREADS, wasmedge.EXTENDED_CONST, wasmedge.TAIL_CALL, wasmedge.MULTI_MEMORIES)

	// Create Compiler
	// compiler := wasmedge.NewCompilerWithConfig(conf)
	compiler := wasmedge.NewCompiler()

	// Compile WASM AOT
	err := compiler.Compile(inPath, outPath)
	if err != nil {
		fmt.Println("Go: Compile WASM to AOT mode Failed!!")
		compiler.Release()
		return err
	}

	// conf.Release()
	compiler.Release()
	return nil
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

func ExecuteWasmInterpreted(
	ctx sdk.Context,
	funcName string,
	env types.Env,
	msg []byte,
	storeKey []byte, kvstore types.KVStore,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
	dependencies []types.ContractDependency,
	isdebug bool,
) (types.ContractResponse, error) {
	var err error
	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not decode wasm execution message")
	}

	var cleanups []func()
	conf := wasmedge.NewConfigure()
	cleanups = append(cleanups, conf.Release)

	var contractRouter ContractRouter = make(map[string]*ContractContext)
	context := &Context{
		Ctx:            ctx,
		GasMeter:       gasMeter,
		Env:            &env,
		ContractStore:  kvstore,
		CosmosHandler:  cosmosHandler,
		ContractRouter: contractRouter,
	}
	context.Env.CurrentCall.CallData = ethMsg.Data
	for _, dep := range dependencies {
		contractContext := buildExecutionContextClassic(dep.FilePath, dep.Bytecode, dep.CodeHash, dep.StoreKey, dep.SystemDeps, conf)
		if err != nil {
			return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not build dependenci execution context for %s", dep.Address)
		}
		context.ContractRouter[dep.Address.String()] = contractContext
	}
	// add itself
	selfContext := buildExecutionContextClassic("", []byte{}, []byte{}, storeKey, systemDeps, conf)
	if err != nil {
		return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not build dependenci execution context for self %s", env.Contract.Address.String())
	}
	context.ContractRouter[env.Contract.Address.String()] = selfContext
	contractVm, _cleanups, err := InitiateWasm(context, "", nil, systemDeps)
	cleanups = append(cleanups, _cleanups...)
	if err != nil {
		runCleanups(cleanups)
		return types.ContractResponse{}, err
	}
	selfContext.Vm = contractVm
	setExecutionBytecode(context, contractVm, funcName)
	selfContext.Bytecode = context.Env.Contract.Bytecode
	selfContext.CodeHash = context.Env.Contract.CodeHash

	executeHandler := GetExecuteFunctionHandler(systemDeps)
	_, err = executeHandler(context, contractVm, types.ENTRY_POINT_EXECUTE)
	// sp, err2 := contractVm.Execute("get_sp")
	if err != nil {
		if isdebug {
			resp := handleContractErrorResponse(contractVm, context.ReturnData, isdebug, err)
			runCleanups(cleanups)
			return resp, nil
		}
		runCleanups(cleanups)
		return types.ContractResponse{}, err
	}

	response := handleContractResponse(contractVm, context.ReturnData, context.Logs, isdebug)

	runCleanups(cleanups)
	return response, nil
}

func ExecuteWasm(
	ctx sdk.Context,
	filePath string,
	funcName string,
	env types.Env,
	msg []byte,
	storeKey []byte, kvstore types.KVStore,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
	dependencies []types.ContractDependency,
	isdebug bool,
) (types.ContractResponse, error) {
	var err error
	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not decode wasm execution message")
	}

	// native implementations
	hexaddr := types.EvmAddressFromAcc(env.Contract.Address).Hex()
	nativePrecompile, found := native.NativeMap[hexaddr]
	if found {
		data := nativePrecompile(ethMsg.Data)
		return types.ContractResponse{Data: data}, nil
	}

	var cleanups []func()
	conf := wasmedge.NewConfigure()
	cleanups = append(cleanups, conf.Release)

	var contractRouter ContractRouter = make(map[string]*ContractContext)
	context := &Context{
		Ctx:            ctx,
		GasMeter:       gasMeter,
		Env:            &env,
		ContractStore:  kvstore,
		CosmosHandler:  cosmosHandler,
		ContractRouter: contractRouter,
	}
	context.Env.CurrentCall.CallData = ethMsg.Data
	for _, dep := range dependencies {
		contractContext := buildExecutionContextClassic(dep.FilePath, dep.Bytecode, dep.CodeHash, dep.StoreKey, dep.SystemDeps, conf)
		if err != nil {
			return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not build dependency execution context for %s", dep.Address)
		}
		context.ContractRouter[dep.Address.String()] = contractContext
	}
	// add itself
	selfContext := buildExecutionContextClassic(filePath, []byte{}, []byte{}, storeKey, systemDeps, conf)
	if err != nil {
		return types.ContractResponse{}, sdkerrors.Wrapf(err, "could not build dependenci execution context for self %s", env.Contract.Address.String())
	}
	context.ContractRouter[env.Contract.Address.String()] = selfContext

	contractVm, _cleanups, err := InitiateWasm(context, filePath, nil, systemDeps)
	cleanups = append(cleanups, _cleanups...)
	if err != nil {
		runCleanups(cleanups)
		return types.ContractResponse{}, err
	}
	selfContext.Vm = contractVm

	setExecutionBytecode(context, contractVm, funcName)
	selfContext.Bytecode = context.Env.Contract.Bytecode
	selfContext.CodeHash = context.Env.Contract.CodeHash

	executeHandler := GetExecuteFunctionHandler(systemDeps)
	_, err = executeHandler(context, contractVm, funcName)
	if err != nil {
		wrapErr := sdkerrors.Wrapf(err, "revert: %s", hex.EncodeToString(context.ReturnData))
		if isdebug {
			resp := handleContractErrorResponse(contractVm, context.ReturnData, isdebug, wrapErr)
			runCleanups(cleanups)
			return resp, nil
		}
		runCleanups(cleanups)
		return types.ContractResponse{}, err
	}

	response := handleContractResponse(contractVm, context.ReturnData, context.Logs, isdebug)

	runCleanups(cleanups)
	return response, nil
}

// deploymentBytecode = constructorBytecode + runtimeBytecode
// codesize/codecopy at deployment = deploymentBytecode + args
// codesize/codecopy at runtime execution = runtimeBytecode
func setExecutionBytecode(context *Context, contractVm *wasmedge.VM, funcName string) {
	// for interpreted code
	// TODO improve detection of interpreted code
	if len(context.Env.Contract.Bytecode) > 0 {
		if funcName == types.ENTRY_POINT_INSTANTIATE {
			context.Env.Contract.Bytecode = append(context.Env.Contract.Bytecode, context.Env.CurrentCall.CallData...)
		}

		runtimeLen, err := hex.DecodeString(fmt.Sprintf("%064x", len(context.Env.Contract.Bytecode)))
		if err != nil {
			return
		}
		calldLen, err := hex.DecodeString(fmt.Sprintf("%064x", len(context.Env.CurrentCall.CallData)))
		if err != nil {
			return
		}
		context.Env.CurrentCall.CallData = append(append(append(
			runtimeLen,
			context.Env.Contract.Bytecode...),
			calldLen...),
			context.Env.CurrentCall.CallData...,
		)
		return
	}

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

		if funcName == types.ENTRY_POINT_INSTANTIATE {
			constructorBytecode, err := activeMemory.GetData(uint(memoffset), uint(constructorLength))
			if err != nil {
				return
			}
			executionBytecode = append(constructorBytecode, executionBytecode...)
			executionBytecode = append(executionBytecode, context.Env.CurrentCall.CallData...)
		}

		context.Env.Contract.Bytecode = executionBytecode
	}
}

func handleContractResponse(contractVm *wasmedge.VM, data []byte, logs []WasmxLog, isdebug bool) types.ContractResponse {
	var events []types.Event
	// module and contract address for the main transaction are added later
	for i, log := range logs {
		var attributes []types.EventAttribute
		attributes = append(attributes, types.EventAttribute{
			Key:   types.AttributeKeyIndex,
			Value: fmt.Sprint(i),
		})
		attributes = append(attributes, types.EventAttribute{
			Key:   types.AttributeKeyData,
			Value: "0x" + hex.EncodeToString(log.Data),
		})
		attributes = append(attributes, types.EventAttribute{
			Key:   types.AttributeKeyEventType,
			Value: log.Type,
		})
		// logs can be from nested calls to other contracts
		attributes = append(attributes, types.EventAttribute{
			Key:   types.AttributeKeyCallContractAddress,
			Value: log.ContractAddress.String(),
		})
		for _, topic := range log.Topics {
			attributes = append(attributes, types.EventAttribute{
				// the topic is the indexed key
				Key:   types.AttributeKeyTopic,
				Value: "0x" + hex.EncodeToString(topic[:]),
			})
		}

		events = append(events, types.Event{
			Type:       types.EventTypeWasmxLog,
			Attributes: attributes,
		})
	}
	var mem []byte
	if isdebug {
		mem = getMemory(contractVm)
	}

	return types.ContractResponse{
		Data:           data,
		Events:         events,
		MemorySnapshot: mem,
	}
}

func handleContractErrorResponse(contractVm *wasmedge.VM, data []byte, isdebug bool, err error) types.ContractResponse {
	var mem []byte
	if isdebug {
		mem = getMemory(contractVm)
	}

	return types.ContractResponse{
		Data:           data,
		MemorySnapshot: mem,
		ErrorMessage:   err.Error(),
	}
}

const MEM_PAGE_SIZE = 64 * 1024 // 64KiB
func getMemory(contractVm *wasmedge.VM) []byte {
	activeMemory := contractVm.GetActiveModule().FindMemory("memory")
	if activeMemory == nil {
		return nil
	}
	pageSize := activeMemory.GetPageSize()
	membz, err := activeMemory.GetData(uint(0), uint(pageSize*MEM_PAGE_SIZE))
	if err != nil {
		return nil
	}
	dst := make([]byte, len(membz))
	copy(dst, membz)
	return dst
}
