package vm

import (
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/interpreters"
	memas "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/assemblyscript"
	membase "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/base"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
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

	wasmx, err := BuildWasmxEnv2(context, rnh)
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

func InstantiateWasmxConsensusJson(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	var err error
	wasmx, err := BuildWasmxConsensusJson1(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
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
	env1, err := BuildWasiWasmxEnv(context, rnh)
	if err != nil {
		return sdkerr.Wrapf(err, "could not build wasmx wasi module")
	}
	vm := rnh.GetVm()
	err = vm.RegisterModule(env1)
	if err != nil {
		return sdkerr.Wrapf(err, "could not register env module")
	}
	err = vm.InitWasi([]string{``}, []string{}, []string{})
	if err != nil {
		return sdkerr.Wrapf(err, "could not register WASI module")
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

type ExecuteFunctionInterface func(context *Context, vm memc.IVm, funcName string, args []interface{}) ([]int32, error)

var ExecuteFunctionHandler = map[string]ExecuteFunctionInterface{}

var DependenciesMap = map[string]bool{}

var RuntimeDepHandler = map[string]func(vm memc.IVm) memc.RuntimeHandler{}

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

	RuntimeDepHandler[types.WASMX_MEMORY_ASSEMBLYSCRIPT] = memas.NewRuntimeHandlerAS
	RuntimeDepHandler[types.WASMX_MEMORY_TAYLOR] = memtay.NewRuntimeHandlerTay
}

func SetSystemDepHandler(
	key string,
	handler func(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error,
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
	for _, dep := range systemDeps {
		executeFn, ok := ExecuteFunctionHandler[dep.Label]
		if ok {
			return executeFn
		}
	}
	return ExecuteDefaultMain
}

func ExecuteDefault(context *Context, contractVm memc.IVm, funcName string) ([]int32, error) {
	return contractVm.Call(funcName, []interface{}{}, context.GasMeter)
}

func ExecuteDefaultContract(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	if funcName == types.ENTRY_POINT_EXECUTE || funcName == types.ENTRY_POINT_QUERY {
		funcName = "main"
	}
	return contractVm.Call(funcName, []interface{}{}, context.GasMeter)
}

func ExecuteDefaultMain(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	return contractVm.Call("main", []interface{}{}, context.GasMeter)
}

func ExecuteFSM(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	if funcName == types.ENTRY_POINT_EXECUTE || funcName == types.ENTRY_POINT_QUERY || funcName == types.ENTRY_POINT_INSTANTIATE {
		funcName = "main"
	}
	return contractVm.Call(funcName, []interface{}{}, context.GasMeter)
}
