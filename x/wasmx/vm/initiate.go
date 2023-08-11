package vm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/interpreters"
	"mythos/v1/x/wasmx/vm/wasmutils"
)

var (
	REQUIRED_IBC_EXPORTS   = []string{}
	REQUIRED_EWASM_EXPORTS = []string{"codesize", "main", "instantiate"}
	// TODO enable and check these
	REQUIRED_CW8_EXPORTS = []string{"interface_version_8", "allocate", "deallocate", "instantiate"}
	ALLOWED_CW8_EXPORTS  = []string{"interface_version_8", "allocate", "deallocate", "instantiate", "execute", "query", "migrate", "reply", "sudo", "ibc_channel_open", "ibc_channel_connect", "ibc_channel_close", "ibc_packet_receive", "ibc_packet_ack", "ibc_packet_timeout"}
)

func InitiateSysEnv1(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	sys := BuildSysEnv(context)
	var cleanups []func()
	cleanups = append(cleanups, sys.Release)
	err := contractVm.RegisterModule(sys)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateWasmxEnv1(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
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

func InitiateWasmxEnv2(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()
	var err error
	keccakVm := wasmedge.NewVM()
	err = wasmutils.InstantiateWasm(keccakVm, "", interpreters.Keccak256Util)
	if err != nil {
		return cleanups, err
	}
	cleanups = append(cleanups, keccakVm.Release)
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

func InitiateInterpreter(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()
	err := wasmutils.InstantiateWasm(contractVm, dep.FilePath, nil)

	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateWasi(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()

	// TODO better
	env1 := BuildWasiWasmxEnv(context)
	cleanups = append(cleanups, env1.Release)
	err := contractVm.RegisterModule(env1)
	if err != nil {
		return cleanups, err
	}

	return cleanups, nil
}

func InitiateEwasmTypeEnv(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	ewasmEnv := BuildEwasmEnv(context)
	var cleanups []func()
	cleanups = append(cleanups, ewasmEnv.Release)
	err := contractVm.RegisterModule(ewasmEnv)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateCosmWasmEnv8(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	wasmx := BuildCosmWasm_8(context)
	var cleanups []func()
	cleanups = append(cleanups, wasmx.Release)
	err := contractVm.RegisterModule(wasmx)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

var SystemDepHandler = map[string]func(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error){}

type ExecuteFunctionInterface func(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error)

var ExecuteFunctionHandler = map[string]ExecuteFunctionInterface{}

func init() {
	SystemDepHandler[types.SYS_ENV_1] = InitiateSysEnv1
	SystemDepHandler[types.WASMX_ENV_1] = InitiateWasmxEnv1
	SystemDepHandler[types.WASMX_ENV_2] = InitiateWasmxEnv2
	SystemDepHandler[types.WASI_SNAPSHOT_PREVIEW1] = InitiateWasi
	SystemDepHandler[types.WASI_UNSTABLE] = InitiateWasi
	SystemDepHandler[types.EWASM_ENV_1] = InitiateEwasmTypeEnv
	SystemDepHandler[types.CW_ENV_8] = InitiateCosmWasmEnv8
	SystemDepHandler[types.ROLE_INTERPRETER] = InitiateInterpreter

	ExecuteFunctionHandler[types.SYS_ENV_1] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.WASMX_ENV_1] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.WASMX_ENV_2] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.WASI_SNAPSHOT_PREVIEW1] = ExecuteWasi
	ExecuteFunctionHandler[types.WASI_UNSTABLE] = ExecuteWasi
	ExecuteFunctionHandler[types.EWASM_ENV_1] = ExecuteDefaultContract
	ExecuteFunctionHandler[types.CW_ENV_8] = ExecuteCw8
	ExecuteFunctionHandler[types.ROLE_INTERPRETER] = ExecuteDefaultMain

	ExecuteFunctionHandler[types.INTERPRETER_EVM_SHANGHAI] = ExecuteDefaultMain
	ExecuteFunctionHandler[types.INTERPRETER_PYTHON] = ExecutePythonInterpreter
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

func ExecuteDefault(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	return contractVm.Execute(funcName)
}

func ExecuteDefaultContract(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	if funcName != types.ENTRY_POINT_INSTANTIATE {
		funcName = "main"
	}
	return contractVm.Execute(funcName)
}

func ExecuteDefaultMain(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	return contractVm.Execute("main")
}

func VerifyEnv(version string, imports []*wasmedge.ImportType) error {
	// TODO check that all imports are supported by the given version

	// for _, mimport := range imports {
	// 	fmt.Println("Import:", mimport.GetModuleName(), mimport.GetExternalName())
	// }
	return nil
}
