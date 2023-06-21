package vm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/interpreters"
	"mythos/v1/x/wasmx/vm/wasmutils"
)

var (
	EWASM_VM_EXPORT = "ewasm_env_"
	WASMX_VM_EXPORT = "wasmx_env_"

	REQUIRED_IBC_EXPORTS   = []string{}
	REQUIRED_EWASM_EXPORTS = []string{"codesize", "main", "instantiate"}
	// codesize_constructor
)

// interface_version
// interpreter_name / address

// ewasm_env_1 // current ewasm interface
// wasmx_env_1 // simplest wasmx version 1 interface
// wasmx_env_2 // wasmx version 2 with env information

// interpreter_evm_shanghai
// interpreter_ewasm_shanghai

func InitiateWasmxWasmx1(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
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

func InitiateWasmxWasmx2(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
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

var SystemDepHandler = map[string]func(context *Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error){}

func init() {
	SystemDepHandler[types.WASMX_ENV_1] = InitiateWasmxWasmx1
	SystemDepHandler[types.WASMX_ENV_2] = InitiateWasmxWasmx2
	SystemDepHandler[types.EWASM_ENV_1] = InitiateEwasmTypeEnv
	SystemDepHandler[types.ROLE_INTERPRETER] = InitiateInterpreter
}

func VerifyEnv(version string, imports []*wasmedge.ImportType) error {
	// TODO check that all imports are supported by the given version

	// for _, mimport := range imports {
	// 	fmt.Println("Import:", mimport.GetModuleName(), mimport.GetExternalName())
	// }
	return nil
}
