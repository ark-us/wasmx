package vm

import (
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/interpreters"
	memas "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/assemblyscript"
	membase "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/base"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	memptrlen_i32 "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/ptrlen_i32"
	memptrlen_i64 "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/ptrlen_i64"
	memrust_i64 "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/rust"
	memtay "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/taylor"
)

var (
	REQUIRED_IBC_EXPORTS   = []string{}
	REQUIRED_EWASM_EXPORTS = []string{"codesize", "main", "instantiate"}
	// TODO enable and check these
	REQUIRED_CW8_EXPORTS = []string{"interface_version_8", "allocate", "deallocate", "instantiate"}
	ALLOWED_CW8_EXPORTS  = []string{"interface_version_8", "allocate", "deallocate", "instantiate", "execute", "query", "migrate", "reply", "sudo", "ibc_channel_open", "ibc_channel_connect", "ibc_channel_close", "ibc_packet_receive", "ibc_packet_ack", "ibc_packet_timeout"}
)

func InitiateSysEnv1(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	sys, err := BuildSysEnv(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(sys)
	if err != nil {
		return err
	}
	return nil
}

func InitiateWasmxEnv1(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxEnv1(context, rnh)
	if err != nil {
		return err
	}
	env, err := BuildAssemblyScriptEnv(context, rnh)
	if err != nil {
		return err
	}
	vm := rnh.GetVm()

	err = vm.RegisterModule(wasmx)
	if err != nil {
		return err
	}
	err = vm.RegisterModule(env)
	if err != nil {
		return err
	}
	return nil
}

func InitiateKeccak256(ctx sdk.Context, newvm memc.NewIVmFn) (memc.RuntimeHandler, error) {
	var err error
	keccakVm := newvm(ctx, true)
	// TODO cache aot keccak
	err = keccakVm.InstantiateWasm("", "", interpreters.Keccak256Util)
	if err != nil {
		return nil, err
	}
	keccakRnh := membase.NewRuntimeHandlerBase(keccakVm)
	return keccakRnh, nil
}

func InitiateWasmxEnv2(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	vm := rnh.GetVm()
	keccakRnh, err := InitiateKeccak256(context.Ctx, vm.New)
	if err != nil {
		return err
	}
	context.ContractRouter["keccak256"] = &Context{RuntimeHandler: keccakRnh}

	wasmx, err := BuildWasmxEnvi32(context, rnh)
	if err != nil {
		return err
	}
	env, err := BuildAssemblyScriptEnv(context, rnh)
	if err != nil {
		return err
	}
	err = vm.RegisterModule(wasmx)
	if err != nil {
		return err
	}
	err = vm.RegisterModule(env)
	if err != nil {
		return err
	}
	return nil
}

func InitiateAssemblyScript(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	vm := rnh.GetVm()
	env, err := BuildAssemblyScriptEnv(context, rnh)
	if err != nil {
		return err
	}
	err = vm.RegisterModule(env)
	if err != nil {
		return err
	}
	return nil
}

func InitiateWasmxEnvi64(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	return InitiateWasmxEnv(context, rnh, dep, BuildWasmxEnvi64)
}

func InitiateWasmxEnvi32(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	return InitiateWasmxEnv(context, rnh, dep, BuildWasmxEnvi32)
}

func InitiateWasmxEnv(
	context *Context,
	rnh memc.RuntimeHandler,
	dep *types.SystemDep,
	buildWasmxEnv func(context *Context, rnh memc.RuntimeHandler) (interface{}, error),
) error {
	vm := rnh.GetVm()
	keccakRnh, err := InitiateKeccak256(context.Ctx, vm.New)
	if err != nil {
		return err
	}
	context.ContractRouter["keccak256"] = &Context{RuntimeHandler: keccakRnh}

	wasmx, err := buildWasmxEnv(context, rnh)
	if err != nil {
		return err
	}
	err = vm.RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InitiateWasmxCoreEnvi64(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	return InitiateWasmxCoreEnv(context, rnh, dep, BuildWasmxCoreEnvi64)
}

func InitiateWasmxCoreEnvi32(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	return InitiateWasmxCoreEnv(context, rnh, dep, BuildWasmxCoreEnvi32)
}

func InitiateWasmxCoreEnv(
	context *Context,
	rnh memc.RuntimeHandler,
	dep *types.SystemDep,
	buildWasmxEnv func(context *Context, rnh memc.RuntimeHandler) (interface{}, error),
) error {
	vm := rnh.GetVm()
	core, err := buildWasmxEnv(context, rnh)
	if err != nil {
		return err
	}
	err = vm.RegisterModule(core)
	if err != nil {
		return err
	}
	return nil
}

func InitiateInterpreter(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	err := rnh.GetVm().InstantiateWasm(dep.CodeFilePath, dep.AotFilePath, nil)
	if err != nil {
		return err
	}
	return nil
}

func InitiateWasi(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	vm := rnh.GetVm()
	wasi, err := BuildWasiEnv(context, rnh)
	if err != nil {
		return sdkerr.Wrapf(err, "could not build wasi module")
	}
	err = vm.RegisterModule(wasi)
	if err != nil {
		return sdkerr.Wrapf(err, "could not register wasi module")
	}
	return nil
}

func InitiateWasmxEnvRusti64(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	vm := rnh.GetVm()
	// TODO javascript interpreter change!!
	env1, err := BuildWasmxEnvRusti64(context, rnh)
	if err != nil {
		return sdkerr.Wrapf(err, "could not build wasmx wasi module")
	}
	err = vm.RegisterModule(env1)
	if err != nil {
		return sdkerr.Wrapf(err, "could not register env module")
	}
	keccakRnh, err := InitiateKeccak256(context.Ctx, vm.New)
	if err != nil {
		return sdkerr.Wrapf(err, "initiate keccak256")
	}
	context.ContractRouter["keccak256"] = &Context{RuntimeHandler: keccakRnh}

	return nil
}

func InitiateEwasmTypeEnv(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	ewasmEnv, err := BuildEwasmEnv(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(ewasmEnv)
	if err != nil {
		return err
	}
	return nil
}

func InitiateCosmWasmEnv8(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildCosmWasm_8(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

var SystemDepHandler = map[string]func(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error{}

var SystemDepHandlerMock = map[string]func(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error{}

type ExecuteFunctionInterface func(context *Context, vm memc.IVm, funcName string, args []interface{}, interpreted bool) ([]int32, error)

type ExecuteFunctionHandlerInterface func(systemDeps []types.SystemDep) ExecuteFunctionInterface

var ExecuteFunctionHandler = map[string]ExecuteFunctionHandlerInterface{}

var DependenciesMap = map[string]bool{}

var RuntimeDepHandler = map[string]func(vm memc.IVm, sysdeps []types.SystemDep) memc.RuntimeHandler{}

func init() {
	SystemDepHandler[types.SYS_ENV_1] = InitiateSysEnv1
	SystemDepHandler[types.WASMX_ENV_1] = InitiateWasmxEnv1
	SystemDepHandler[types.WASMX_ENV_2] = InitiateWasmxEnv2
	SystemDepHandler[types.WASMX_ENVi32_2] = InitiateWasmxEnvi32
	SystemDepHandler[types.WASMX_ENVi64_2] = InitiateWasmxEnvi64
	SystemDepHandler[types.WASMX_ENV_RUSTi64_2] = InitiateWasmxEnvRusti64
	SystemDepHandler[types.WASMX_CORE_ENVi32_1] = InitiateWasmxCoreEnvi32
	SystemDepHandler[types.WASMX_CORE_ENVi64_1] = InitiateWasmxCoreEnvi64
	SystemDepHandler[types.WASI_SNAPSHOT_PREVIEW1] = InitiateWasi
	SystemDepHandler[types.WASI_UNSTABLE] = InitiateWasi
	SystemDepHandler[types.EWASM_ENV_1] = InitiateEwasmTypeEnv
	SystemDepHandler[types.CW_ENV_8] = InitiateCosmWasmEnv8
	SystemDepHandler[types.ROLE_INTERPRETER] = InitiateInterpreter
	SystemDepHandler[types.WASMX_CONSENSUS_JSON_1] = InstantiateWasmxConsensusJson_i32
	SystemDepHandler[types.WASMX_CONSENSUS_JSON_i32_1] = InstantiateWasmxConsensusJson_i32
	SystemDepHandler[types.WASMX_CONSENSUS_JSON_i64_1] = InstantiateWasmxConsensusJson_i64

	// language-specific imports
	SystemDepHandler[types.WASMX_MEMORY_ASSEMBLYSCRIPT] = InitiateAssemblyScript

	// TODO these interpreter exceptions should be removed and entrypoints should follow standard rules
	ExecuteFunctionHandler[types.INTERPRETER_EVM_SHANGHAI] = ExecuteFunctionWrapperNoop(ExecuteDefaultMain)
	ExecuteFunctionHandler[types.INTERPRETER_PYTHON] = ExecuteFunctionWrapperNoop(ExecutePythonInterpreter)
	ExecuteFunctionHandler[types.INTERPRETER_JS] = ExecuteFunctionWrapperNoop(ExecuteJsInterpreter)

	ExecuteFunctionHandler[types.ROLE_INTERPRETER] = ExecuteDefaultInterpreter

	// note: interpreter initialization will fall here, because the contract does not have a role at that time

	ExecuteFunctionHandler[types.SYS_ENV_1] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.WASMX_ENV_1] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.WASMX_ENV_2] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.WASMX_ENVi32_2] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.WASMX_ENVi64_2] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.WASMX_CORE_ENVi32_1] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.WASMX_CORE_ENVi64_1] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.WASI_SNAPSHOT_PREVIEW1] = ExecuteFunctionWrapperNoop(ExecuteWasiWrap)
	ExecuteFunctionHandler[types.WASI_UNSTABLE] = ExecuteFunctionWrapperNoop(ExecuteWasiWrap)
	ExecuteFunctionHandler[types.EWASM_ENV_1] = ExecuteFunctionWrapperNoop(ExecuteDefaultContract)
	ExecuteFunctionHandler[types.CW_ENV_8] = ExecuteFunctionWrapperNoop(ExecuteCw8)

	DependenciesMap[types.EWASM_VM_EXPORT] = true
	DependenciesMap[types.WASMX_VM_EXPORT] = true
	DependenciesMap[types.WASMX_VM_CORE_EXPORT] = true
	DependenciesMap[types.SYS_VM_EXPORT] = true
	DependenciesMap[types.WASMX_CONS_VM_EXPORT] = true
	DependenciesMap[types.MEMORY_EXPORT] = true

	RuntimeDepHandler[types.WASMX_MEMORY_ASSEMBLYSCRIPT] = memas.NewRuntimeHandlerAS
	RuntimeDepHandler[types.WASMX_MEMORY_TAYLOR] = memtay.NewRuntimeHandlerTay
	RuntimeDepHandler[types.WASMX_MEMORY_RUST_i64] = memrust_i64.NewRuntimeHandler
	RuntimeDepHandler[types.WASMX_MEMORY_PTRLEN_i64] = memptrlen_i64.NewRuntimeHandler
	RuntimeDepHandler[types.WASMX_MEMORY_PTRLEN_i32] = memptrlen_i32.NewRuntimeHandler
}

func SetSystemDepHandler(
	key string,
	handler func(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error,
) {
	SystemDepHandler[key] = handler
}

func SetSystemDepHandlerMock(
	key string,
	handler func(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error,
) {
	SystemDepHandlerMock[key] = handler
}

func SetExecuteFunctionHandler(
	key string,
	handler ExecuteFunctionInterface,
) {
	ExecuteFunctionHandler[key] = ExecuteFunctionWrapperNoop(handler)
}

func GetExecuteFunctionHandler(systemDeps []types.SystemDep) ExecuteFunctionInterface {
	for _, dep := range systemDeps {
		executeFn, ok := ExecuteFunctionHandler[dep.Label]
		if ok {
			return executeFn(systemDeps)
		}
		executeFn, ok = ExecuteFunctionHandler[dep.Role]
		if ok {
			return executeFn(systemDeps)
		}
	}
	// look in dep.Deps
	for _, systemDep := range systemDeps {
		handler := GetExecuteFunctionHandler(systemDep.Deps)
		if handler != nil {
			return handler
		}
	}
	return ExecuteDefaultMain
}

func GetExecuteFunctionHandlerForLabels(systemDeps []types.SystemDep) ExecuteFunctionInterface {
	for _, dep := range systemDeps {
		executeFn, ok := ExecuteFunctionHandler[dep.Label]
		if ok {
			return executeFn(systemDeps)
		}
	}
	// look in dep.Deps
	for _, systemDep := range systemDeps {
		handler := GetExecuteFunctionHandlerForLabels(systemDep.Deps)
		if handler != nil {
			return handler
		}
	}
	return ExecuteDefaultMain
}

func ExecuteDefault(context *Context, contractVm memc.IVm, funcName string, interpreted bool) ([]int32, error) {
	return contractVm.Call(funcName, []interface{}{}, context.GasMeter)
}

func ExecuteDefaultContract(context *Context, contractVm memc.IVm, funcName string, args []interface{}, interpreted bool) ([]int32, error) {
	if funcName == types.ENTRY_POINT_EXECUTE || funcName == types.ENTRY_POINT_QUERY {
		funcName = "main"
	}
	if interpreted && funcName == types.ENTRY_POINT_INSTANTIATE {
		funcName = "main"
	}
	return contractVm.Call(funcName, []interface{}{}, context.GasMeter)
}

func ExecuteDefaultMain(context *Context, contractVm memc.IVm, funcName string, args []interface{}, interpreted bool) ([]int32, error) {
	return contractVm.Call("main", []interface{}{}, context.GasMeter)
}

func ExecuteFunctionWrapperNoop(fn ExecuteFunctionInterface) ExecuteFunctionHandlerInterface {
	return func(systemDeps []types.SystemDep) ExecuteFunctionInterface {
		return fn
	}
}

func ExecuteDefaultInterpreter(systemDeps []types.SystemDep) ExecuteFunctionInterface {
	handler := GetExecuteFunctionHandlerForLabels(systemDeps)
	return func(context *Context, contractVm memc.IVm, funcName string, args []interface{}, interpreted bool) ([]int32, error) {
		return handler(context, contractVm, funcName, args, true)
	}
}
