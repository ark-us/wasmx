package vm

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	sdkerr "cosmossdk.io/errors"
	log "cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InitiateWasm(context *Context, rnh memc.RuntimeHandler, wasmFilePath string, aotFilePath string, wasmbuffer []byte, systemDeps []types.SystemDep, hasCoreRole bool) error {
	// set default
	if len(systemDeps) == 0 {
		label := types.DEFAULT_SYS_DEP
		systemDeps = append(systemDeps, types.SystemDep{Role: label, Label: label})
	}
	err := initiateWasmDeps(context, rnh, systemDeps, hasCoreRole)
	if err != nil {
		return err
	}
	contractVm := rnh.GetVm()

	if wasmFilePath != "" || aotFilePath != "" || len(wasmbuffer) > 0 {
		err = contractVm.InstantiateWasm(wasmFilePath, aotFilePath, wasmbuffer)
		if err != nil {
			return err
		}
	}
	return nil
}

// finds registered host APIs for the given role
func initiateWasmDeps(context *Context, rnh memc.RuntimeHandler, systemDeps []types.SystemDep, hasCoreRole bool) error {
	for _, systemDep := range systemDeps {
		// system deps of system deps
		err := initiateWasmDeps(context, rnh, systemDep.Deps, hasCoreRole)
		if err != nil {
			return err
		}

		_, isProtected := getFirstKeyFromMapping(types.PROTECTED_HOST_APIS, []string{systemDep.Role, systemDep.Label})

		if isProtected && !hasCoreRole {
			handler, found := getFirstKeyFromMapping(SystemDepHandlerMock, []string{systemDep.Role, systemDep.Label})
			if found {
				err := handler(context, rnh, &systemDep)
				if err != nil {
					return err
				}
			}
			// we do not allow a contract with no roles to access protected APIs, below
			continue
		}
		handler, found := getFirstKeyFromMapping(SystemDepHandler, []string{systemDep.Role, systemDep.Label})
		if found {
			err := handler(context, rnh, &systemDep)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getFirstKeyFromMapping[T any](vmap map[string]T, keys []string) (T, bool) {
	var zero T
	for _, key := range keys {
		if v, ok := vmap[key]; ok {
			return v, true
		}
	}
	return zero, false
}

func getRuntimeHandler(newIVmFn memc.NewIVmFn, ctx sdk.Context, systemDeps []types.SystemDep, pinned bool) memc.RuntimeHandler {
	if !pinned {
		// also check system deps
		for _, dep := range systemDeps {
			if dep.Pinned {
				pinned = true
				break
			}
		}
	}
	vm := newIVmFn(ctx, pinned)
	handler := getRuntimeHandlerFromDeps(vm, systemDeps)
	if handler != nil {
		return handler
	}
	// default is assemblyscript memory
	// if we change this, we should add the dependency in the system contracts
	return RuntimeDepHandler[types.WASMX_MEMORY_ASSEMBLYSCRIPT](vm, systemDeps)
}

func getRuntimeHandlerFromDeps(vm memc.IVm, systemDeps []types.SystemDep) memc.RuntimeHandler {
	for _, systemDep := range systemDeps {
		handler, found := RuntimeDepHandler[systemDep.Role]
		if !found {
			handler, found = RuntimeDepHandler[systemDep.Label]
		}
		if found {
			return handler(vm, systemDeps)
		}
	}
	// look in dep.Deps
	for _, systemDep := range systemDeps {
		handler := getRuntimeHandlerFromDeps(vm, systemDep.Deps)
		if handler != nil {
			return handler
		}
	}
	return nil
}

func AnalyzeWasm(ctx sdk.Context, vmMeta memc.IWasmVmMeta, wasmbuffer []byte) (memc.AnalysisReport, error) {
	report := memc.AnalysisReport{}
	meta, err := vmMeta.AnalyzeWasm(ctx, wasmbuffer)
	if err != nil {
		return report, err
	}

	uniqueDeps := make(map[string]bool)
	exports := meta.ListExports()
	imports := meta.ListImports()

	for _, mexport := range exports {
		fname := mexport.Name()
		var dep string
		for key, ok := range DependenciesMap {
			if !ok {
				continue
			}
			if strings.Contains(fname, key) {
				dep = parseDependencyOrHexAddr(fname, key)
			}
		}
		if fname == types.EWASM_ENV_0 {
			dep = types.EWASM_ENV_1
		} else if fname == types.CW_ENV_8 {
			dep = parseDependencyOrHexAddr(fname, types.CW_VM_EXPORT)
		}
		if dep != "" {
			err := VerifyEnv(dep, imports)
			if err != nil {
				return report, sdkerr.Wrapf(types.ErrCreateFailed, "wasm module requires imports not supported by the %s version: %s", fname, err.Error())
			}
			if _, found := uniqueDeps[dep]; !found {
				report.Dependencies = append(report.Dependencies, dep)
				uniqueDeps[dep] = true
			}
		}
	}

	for _, mimport := range imports {
		fname := mimport.ModuleName()
		var dep string

		if strings.Contains(fname, types.WASI_VM_EXPORT) {
			dep = parseDependencyOrHexAddr(fname, types.WASI_VM_EXPORT)
		}

		if dep != "" {
			err := VerifyEnv(dep, imports)
			if err != nil {
				return report, sdkerr.Wrapf(types.ErrCreateFailed, "wasm module requires imports not supported by the %s version: %s", fname, err.Error())
			}
			if _, found := uniqueDeps[dep]; !found {
				report.Dependencies = append(report.Dependencies, dep)
				uniqueDeps[dep] = true
			}
		}
	}

	// make deterministic: order alphabetically
	sort.Strings(report.Dependencies)

	return report, nil
}

func VerifyEnv(version string, imports []memc.WasmImport) error {
	// TODO check that all imports are supported by the given version

	// for _, mimport := range imports {
	// 	fmt.Println("Import:", mimport.GetModuleName(), mimport.GetExternalName())
	// }
	return nil
}

// Returns the hex address of the interpreter if exists or the version string
func parseDependencyOrHexAddr(contractVersion string, part string) string {
	dep := contractVersion
	if strings.Contains(contractVersion, part) {
		v := contractVersion[len(part):]
		if len(v) > 2 && v[0:2] == "0x" {
			dep = v
		}
	}
	return dep
}

func GetVmLogger(
	logger func(ctx sdk.Context) log.Logger,
	chainId string,
	contractAddress string,
) func(ctx sdk.Context) log.Logger {
	newlogger := func(ctx sdk.Context) log.Logger {
		moduleName := fmt.Sprintf("x/%s_%s_%s", types.ModuleName, chainId, contractAddress)
		return logger(ctx).With(log.ModuleKey, moduleName)
	}
	return newlogger
}

func GetVmLoggerExtended(
	logger func(ctx sdk.Context) log.Logger,
	chainId string,
	contractAddress string,
) func(ctx sdk.Context) log.Logger {
	newlogger := func(ctx sdk.Context) log.Logger {
		moduleName := fmt.Sprintf("x/%s_extended_%s_%s", types.ModuleName, chainId, contractAddress)
		return logger(ctx).With(log.ModuleKey, moduleName)
	}
	return newlogger
}

func LoggerExtended(ctx *Context) log.Logger {
	newmodule := GetVmLoggerExtended(ctx.Logger, ctx.Env.Chain.ChainIdFull, ctx.Env.Contract.Address.String())
	return newmodule(ctx.Ctx)
}

func ExecuteWasmInterpreted(
	goRoutineGroup *errgroup.Group,
	goContextParent context.Context,
	ctx sdk.Context,
	logger func(ctx sdk.Context) log.Logger,
	funcName string,
	env types.Env,
	msg []byte,
	kvstore prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	contractInfo types.ContractDependency,
	dependencies []types.ContractDependency,
	isdebug bool,
	inBackground bool,
	app types.Application,
	newIVmFn memc.NewIVmFn,
) (types.ContractResponse, error) {
	var err error
	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, sdkerr.Wrapf(err, "could not decode wasm execution message")
	}

	var contractRouter ContractRouter = make(map[string]*Context)
	rnh := getRuntimeHandler(newIVmFn, ctx, contractInfo.SystemDeps, contractInfo.Pinned)
	defer func() {
		rnh.GetVm().Cleanup()
	}()
	context := &Context{
		GoRoutineGroup:  goRoutineGroup,
		GoContextParent: goContextParent,
		Ctx:             ctx,
		Logger:          GetVmLogger(logger, env.Chain.ChainIdFull, env.Contract.Address.String()),
		GasMeter:        gasMeter,
		Env:             &env,
		ContractStore:   kvstore,
		CosmosHandler:   cosmosHandler,
		App:             app,
		ContractRouter:  contractRouter,
		NativeHandler:   NativeMap,
		dbIterators:     map[int32]types.Iterator{},
		RuntimeHandler:  rnh,
		ContractInfo:    &contractInfo,
		newIVmFn:        newIVmFn,
	}
	context.Env.CurrentCall.CallData = ethMsg.Data
	for i := range dependencies {
		dep := dependencies[i]
		context.ContractRouter[dep.Address.String()] = &Context{ContractInfo: &dep}
	}

	// add itself
	contractstr := env.Contract.Address.String()
	err = InitiateWasm(context, rnh, "", "", nil, contractInfo.SystemDeps, contractInfo.Role != "")
	if err != nil {
		return types.ContractResponse{}, err
	}
	context.RuntimeHandler = rnh

	setExecutionBytecode(context, rnh, funcName)
	context.ContractInfo.Bytecode = context.Env.Contract.Bytecode
	context.ContractInfo.CodeHash = context.Env.Contract.CodeHash
	context.ContractRouter[contractstr] = context
	executeHandler := GetExecuteFunctionHandler(contractInfo.SystemDeps)

	if inBackground {
		err = types.AddBackgroundProcesses(goContextParent, &types.BackgroundProcess{
			Label:          contractstr,
			RuntimeHandler: rnh,
			ExecuteHandler: func(funcName_ string) ([]byte, error) {
				_, err := executeHandler(context, rnh.GetVm(), funcName_, make([]interface{}, 0))
				if err != nil {
					return nil, err
				}
				return context.FinishData, nil
			},
		})
		if err != nil {
			return types.ContractResponse{}, err
		}
	}

	_, err = executeHandler(context, rnh.GetVm(), funcName, make([]interface{}, 0))
	// sp, err2 := contractVm.Execute("get_sp")
	if err != nil {
		wrapErr := sdkerr.Wrapf(
			err,
			"chain_id: %s; contract: %s; entry point: %s; revert: %s",
			context.Ctx.ChainID(),
			context.Env.Contract.Address.String(),
			funcName,
			string(context.FinishData),
		)

		resp := handleContractErrorResponse(rnh.GetVm(), context.FinishData, isdebug, wrapErr)
		if isdebug {
			// we don't fail for debug/tracing transactions
			return resp, nil
		}
		return resp, wrapErr
		// runCleanups(cleanups)
		// return types.ContractResponse{}, err
	}

	response := handleContractResponse(context, rnh.GetVm(), isdebug)
	return response, nil
}

func ExecuteWasm(
	goRoutineGroup *errgroup.Group,
	goContextParent context.Context,
	ctx sdk.Context,
	logger func(ctx sdk.Context) log.Logger,
	funcName string,
	env types.Env,
	msg []byte,
	kvstore prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	contractInfo types.ContractDependency,
	dependencies []types.ContractDependency,
	isdebug bool,
	inBackground bool,
	app types.Application,
	newIVmFn memc.NewIVmFn,
) (types.ContractResponse, error) {
	var err error
	var ethMsg types.WasmxExecutionMessage
	err = json.Unmarshal(msg, &ethMsg)
	if err != nil {
		return types.ContractResponse{}, sdkerr.Wrapf(err, "could not decode wasm execution message")
	}

	var contractRouter ContractRouter = make(map[string]*Context)
	rnh := getRuntimeHandler(newIVmFn, ctx, contractInfo.SystemDeps, contractInfo.Pinned)
	defer func() {
		rnh.GetVm().Cleanup()
	}()
	context := &Context{
		GoRoutineGroup:  goRoutineGroup,
		GoContextParent: goContextParent,
		Ctx:             ctx,
		Logger:          GetVmLogger(logger, env.Chain.ChainIdFull, env.Contract.Address.String()),
		GasMeter:        gasMeter,
		Env:             &env,
		ContractStore:   kvstore,
		CosmosHandler:   cosmosHandler,
		ContractRouter:  contractRouter,
		App:             app,
		NativeHandler:   NativeMap,
		dbIterators:     map[int32]types.Iterator{},
		RuntimeHandler:  rnh,
		ContractInfo:    &contractInfo,
		newIVmFn:        newIVmFn,
	}
	context.Env.CurrentCall.CallData = ethMsg.Data

	// native implementations
	found := context.NativeHandler.IsPrecompile(env.Contract.Address.Bytes())
	if found {
		data, err := context.NativeHandler.Execute(context, env.Contract.Address, ethMsg.Data)
		if err != nil {
			return types.ContractResponse{Data: data}, err
		}
		return types.ContractResponse{Data: data}, nil
	}

	for i := range dependencies {
		dep := dependencies[i]
		context.ContractRouter[dep.Address.String()] = &Context{ContractInfo: &dep}
	}
	// add itself
	contractstr := env.Contract.Address.String()
	err = InitiateWasm(context, rnh, contractInfo.CodeFilePath, contractInfo.AotFilePath, nil, contractInfo.SystemDeps, contractInfo.Role != "")
	if err != nil {
		return types.ContractResponse{}, err
	}
	context.RuntimeHandler = rnh

	setExecutionBytecode(context, rnh, funcName)
	context.ContractInfo.Bytecode = context.Env.Contract.Bytecode
	context.ContractInfo.CodeHash = context.Env.Contract.CodeHash
	context.ContractRouter[contractstr] = context
	executeHandler := GetExecuteFunctionHandler(contractInfo.SystemDeps)
	if inBackground {
		err = types.AddBackgroundProcesses(goContextParent, &types.BackgroundProcess{
			Label:          contractstr,
			RuntimeHandler: rnh,
			ExecuteHandler: func(funcName_ string) ([]byte, error) {
				_, err := executeHandler(context, rnh.GetVm(), funcName_, make([]interface{}, 0))
				if err != nil {
					return nil, err
				}
				return context.FinishData, nil
			},
		})
		if err != nil {
			return types.ContractResponse{}, err
		}
	}
	_, err = executeHandler(context, rnh.GetVm(), funcName, make([]interface{}, 0))
	if err != nil {
		wrapErr := sdkerr.Wrapf(
			err,
			"chain_id: %s; contract: %s; entry point: %s; revert: %s",
			context.Ctx.ChainID(),
			context.Env.Contract.Address.String(),
			funcName,
			hex.EncodeToString(context.FinishData),
		)
		resp := handleContractErrorResponse(rnh.GetVm(), context.FinishData, isdebug, wrapErr)
		if isdebug {
			return resp, nil
		}
		return resp, wrapErr
		// runCleanups(cleanups)
		// return types.ContractResponse{}, err
	}
	response := handleContractResponse(context, rnh.GetVm(), isdebug)
	return response, nil
}

// deploymentBytecode = constructorBytecode + runtimeBytecode
// codesize/codecopy at deployment = deploymentBytecode + args
// codesize/codecopy at runtime execution = runtimeBytecode
func setExecutionBytecode(context *Context, rnh memc.RuntimeHandler, funcName string) {
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
	vm := rnh.GetVm()
	fnList := vm.GetFunctionList()

	if slices.Contains(fnList, "evm_bytecode") {
		activeMemory, err := rnh.GetMemory()
		if err != nil {
			return
		}
		retvalues, err := vm.Call("evm_bytecode", []interface{}{}, context.GasMeter)
		if err != nil {
			return
		}

		memoffset := retvalues[0]
		constructorLength := retvalues[1]
		runtimeLength := retvalues[2]
		executionBytecode, err := activeMemory.Read(memoffset+constructorLength, runtimeLength)
		if err != nil {
			return
		}

		if funcName == types.ENTRY_POINT_INSTANTIATE {
			constructorBytecode, err := activeMemory.Read(memoffset, constructorLength)
			if err != nil {
				return
			}
			executionBytecode = append(constructorBytecode, executionBytecode...)
			executionBytecode = append(executionBytecode, context.Env.CurrentCall.CallData...)
		}

		context.Env.Contract.Bytecode = executionBytecode
	}
}

func handleContractResponse(context *Context, vm memc.IVm, isdebug bool) types.ContractResponse {
	data := context.FinishData
	logs := context.Logs
	messages := context.Messages
	var events []types.Event
	// module and contract address for the main transaction are added later
	// TODO fixme duplicate keys are shown once
	// we may need to add indexes or use arrays for values
	for i, log := range logs {
		contractAddressStr := log.ContractAddress.String()
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
			Value: contractAddressStr,
		})
		attributes = append(attributes, types.EventAttribute{
			Key:   types.AttributeKeyDependency,
			Value: log.SystemDependency,
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
	events = append(events, context.CosmosEvents...)

	var mem []byte
	if isdebug {
		mem = getMemory(vm)
	}

	return types.ContractResponse{
		Data:           data,
		Events:         events,
		MemorySnapshot: mem,
		Messages:       messages,
	}
}

func handleContractErrorResponse(vm memc.IVm, data []byte, isdebug bool, err error) types.ContractResponse {
	var mem []byte
	if isdebug {
		mem = getMemory(vm)
	}

	return types.ContractResponse{
		Data:           data,
		MemorySnapshot: mem,
		ErrorMessage:   err.Error(),
	}
}

func getMemory(vm memc.IVm) []byte {
	activeMemory, err := vm.GetMemory()
	if err != nil {
		return nil
	}
	membz, err := activeMemory.Read(int32(0), int32(activeMemory.Size()))
	if err != nil {
		return nil
	}
	dst := make([]byte, len(membz))
	copy(dst, membz)
	return dst
}
