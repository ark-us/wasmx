package vm

import (
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/interpreters"
	memas "mythos/v1/x/wasmx/vm/memory/assemblyscript"
	memc "mythos/v1/x/wasmx/vm/memory/common"
	memtay "mythos/v1/x/wasmx/vm/memory/taylor"
	"mythos/v1/x/wasmx/vm/wasmutils"
)

var (
	REQUIRED_IBC_EXPORTS   = []string{}
	REQUIRED_EWASM_EXPORTS = []string{"codesize", "main", "instantiate"}
	// TODO enable and check these
	REQUIRED_CW8_EXPORTS = []string{"interface_version_8", "allocate", "deallocate", "instantiate"}
	ALLOWED_CW8_EXPORTS  = []string{"interface_version_8", "allocate", "deallocate", "instantiate", "execute", "query", "migrate", "reply", "sudo", "ibc_channel_open", "ibc_channel_connect", "ibc_channel_close", "ibc_packet_receive", "ibc_packet_ack", "ibc_packet_timeout"}
)

func InitiateSysEnv1(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	sys := BuildSysEnv(context)
	var cleanups []func()
	cleanups = append(cleanups, sys.Release)
	err := contractVm.RegisterModule(sys)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateWasmxEnv1(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	wasmx := BuildWasmxEnv1(context)
	env := BuildAssemblyScriptEnv(context)

	var cleanups []func()
	cleanups = append(cleanups, wasmx.Release)
	err := contractVm.RegisterModule(wasmx)
	if err != nil {
		return cleanups, err
	}
	cleanups = append(cleanups, env.Release)
	err = contractVm.RegisterModule(env)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateKeccak256(newvm func() memc.IVm) (memc.IVm, []func(), error) {
	var cleanups []func()
	var err error
	keccakVm := newvm()
	err = keccakVm.InstantiateWasm("", interpreters.Keccak256Util)
	if err != nil {
		return nil, cleanups, err
	}
	cleanups = append(cleanups, keccakVm.Release)
	return keccakVm, cleanups, nil
}

func InitiateWasmxEnv2(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	keccakVm, cleanups, err := InitiateKeccak256(contractVm.New)
	if err != nil {
		return cleanups, err
	}
	context.ContractRouter["keccak256"] = &ContractContext{Vm: keccakVm}

	wasmx := BuildWasmxEnv2(context)
	env := BuildAssemblyScriptEnv(context)
	cleanups = append(cleanups, wasmx.Release)

	err = contractVm.RegisterModule(wasmx)
	if err != nil {
		return cleanups, err
	}
	cleanups = append(cleanups, env.Release)
	err = contractVm.RegisterModule(env)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InstantiateWasmxConsensusJson(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()
	var err error
	wasmx := BuildWasmxConsensusJson1(context)
	err = contractVm.RegisterModule(wasmx)
	if err != nil {
		return cleanups, err
	}
	cleanups = append(cleanups, wasmx.Release)
	return cleanups, nil
}

func InitiateInterpreter(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()
	err := wasmutils.InstantiateWasm(contractVm, dep.FilePath, nil)

	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateWasi(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()

	// TODO better
	env1 := BuildWasiWasmxEnv(context)
	cleanups = append(cleanups, env1.Release)
	err := contractVm.RegisterModule(env1)
	if err != nil {
		return cleanups, err
	}

	keccakVm, cleanups2, err := InitiateKeccak256()
	cleanups = append(cleanups, cleanups2...)
	if err != nil {
		return cleanups, err
	}
	context.ContractRouter["keccak256"] = &ContractContext{Vm: keccakVm}

	return cleanups, nil
}

func InitiateEwasmTypeEnv(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	ewasmEnv := BuildEwasmEnv(context)
	var cleanups []func()
	cleanups = append(cleanups, ewasmEnv.Release)
	err := contractVm.RegisterModule(ewasmEnv)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateCosmWasmEnv8(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error) {
	wasmx := BuildCosmWasm_8(context)
	var cleanups []func()
	cleanups = append(cleanups, wasmx.Release)
	err := contractVm.RegisterModule(wasmx)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

var SystemDepHandler = map[string]func(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error){}

type ExecuteFunctionInterface func(context *Context, vm memc.IVm, funcName string, args []interface{}) ([]int32, error)

var ExecuteFunctionHandler = map[string]ExecuteFunctionInterface{}

var DependenciesMap = map[string]bool{}

var MemoryDepHandler = map[string]func(vm memc.IVm, mem memc.IMemory) memc.MemoryHandler{}

func init() {
	SystemDepHandler[types.SYS_ENV_1] = InitiateSysEnv1
	SystemDepHandler[types.WASMX_ENV_1] = InitiateWasmxEnv1
	SystemDepHandler[types.WASMX_ENV_2] = InitiateWasmxEnv2
	SystemDepHandler[types.WASI_SNAPSHOT_PREVIEW1] = InitiateWasi
	SystemDepHandler[types.WASI_UNSTABLE] = InitiateWasi
	SystemDepHandler[types.EWASM_ENV_1] = InitiateEwasmTypeEnv
	SystemDepHandler[types.CW_ENV_8] = InitiateCosmWasmEnv8
	SystemDepHandler[types.ROLE_INTERPRETER] = InitiateInterpreter
	SystemDepHandler[types.WASMX_CONSENSUS_JSON_1] = InstantiateWasmxConsensusJson

	ExecuteFunctionHandler[types.SYS_ENV_1] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.WASMX_ENV_1] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.WASMX_ENV_2] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.WASI_SNAPSHOT_PREVIEW1] = ExecuteWasiWrap
	ExecuteFunctionHandler[types.WASI_UNSTABLE] = ExecuteWasiWrap
	ExecuteFunctionHandler[types.EWASM_ENV_1] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.CW_ENV_8] = ExecuteCw8
	ExecuteFunctionHandler[types.ROLE_INTERPRETER] = ExecuteDefaultMain

	ExecuteFunctionHandler[types.INTERPRETER_EVM_SHANGHAI] = ExecuteDefaultMain
	ExecuteFunctionHandler[types.INTERPRETER_PYTHON] = ExecutePythonInterpreter
	ExecuteFunctionHandler[types.INTERPRETER_JS] = ExecuteJsInterpreter
	ExecuteFunctionHandler[types.INTERPRETER_FSM] = ExecuteFSM

	DependenciesMap[types.EWASM_VM_EXPORT] = true
	DependenciesMap[types.WASMX_VM_EXPORT] = true
	DependenciesMap[types.SYS_VM_EXPORT] = true
	DependenciesMap[types.WASMX_CONS_VM_EXPORT] = true

	MemoryDepHandler[types.WASMX_MEMORY_ASSEMBLYSCRIPT] = memas.NewMemoryHandlerAS
	MemoryDepHandler[types.WASMX_MEMORY_TAYLOR] = memtay.NewMemoryHandlerTay
}

func SetSystemDepHandler(
	key string,
	handler func(context *Context, contractVm memc.IVm, dep *types.SystemDep) ([]func(), error),
) {
	SystemDepHandler[key] = handler
}

func SetExecuteFunctionHandler(
	key string,
	handler ExecuteFunctionInterface,
) {
	ExecuteFunctionHandler[key] = handler
}

func GetExecuteFunctionHandler(systemDeps []types.SystemDep) ExecuteFunctionInterface {
	if len(systemDeps) > 0 {
		depName := systemDeps[0].Label
		executeFn, ok := ExecuteFunctionHandler[depName]
		if ok {
			return executeFn
		}
	}
	return ExecuteDefaultMain
}

func ExecuteDefault(context *Context, contractVm memc.IVm, funcName string) ([]int32, error) {
	return contractVm.Call(funcName, []interface{}{})
}

func ExecuteDefaultContract(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	if funcName == types.ENTRY_POINT_EXECUTE || funcName == types.ENTRY_POINT_QUERY {
		funcName = "main"
	}
	return contractVm.Call(funcName, []interface{}{})
}

func ExecuteDefaultMain(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	return contractVm.Call("main", []interface{}{})
}

func ExecuteFSM(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	if funcName == types.ENTRY_POINT_EXECUTE || funcName == types.ENTRY_POINT_QUERY || funcName == types.ENTRY_POINT_INSTANTIATE {
		funcName = "main"
	}
	return contractVm.Call(funcName, []interface{}{})
}
