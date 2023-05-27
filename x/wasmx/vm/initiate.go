package vm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/interpreters"
	"mythos/v1/x/wasmx/vm/wasmutils"
)

var (
	EWASM_VM_EXPORT          = "ewasm_env_"
	EWASM_INTERPRETER_EXPORT = "ewasm_ewasm_"
	WASMX_VM_EXPORT          = "wasmx_wasmx_"

	REQUIRED_IBC_EXPORTS   = []string{}
	REQUIRED_EWASM_EXPORTS = []string{"codesize", "main", "instantiate"}
	// codesize_constructor
)

// interface_version
// interpreter_name / address

// wasmx_wasmx_1 // simplest wasmx version 1 interface
// ewasm_env_1 // current ewasm interface
// wasmx_wasmx_2 // wasmx version 2 with env information

// interpreter_evm_shanghai
// interpreter_ewasm_shanghai

func InitiateWasmxWasmx1(context *Context, contractVm *wasmedge.VM) ([]func(), error) {
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

func InitiateWasmxWasmx2(context *Context, contractVm *wasmedge.VM) ([]func(), error) {
	var cleanups []func()
	keccakVm := wasmedge.NewVM()
	// err := wasmutils.InstantiateWasm(keccakVm, "", interpreters.Keccak256Util)
	err := wasmutils.InstantiateWasm(keccakVm, "../vm/interpreters/keccak256.so", nil)
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

func InitiateEvmInterpreter_1(context *Context, contractVm *wasmedge.VM) ([]func(), error) {
	var cleanups []func()
	// err := wasmutils.InstantiateWasm(contractVm, "", interpreters.EvmInterpreter_1)
	err := wasmutils.InstantiateWasm(contractVm, "../vm/interpreters/evm_shanghai.so", nil)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateEwasmTypeEnv(context *Context, contractVm *wasmedge.VM) ([]func(), error) {
	ewasmEnv := BuildEwasmEnv(context)
	var cleanups []func()
	cleanups = append(cleanups, ewasmEnv.Release)
	err := contractVm.RegisterModule(ewasmEnv)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateEwasmTypeInterpreter(context *Context, contractVm *wasmedge.VM) ([]func(), error) {
	var cleanups []func()
	contractEnv := wasmedge.NewModule("ewasm")
	ewasmVm := wasmedge.NewVM()
	ewasmEnv := BuildEwasmEnv(context)
	cleanups = append(cleanups, contractEnv.Release, ewasmVm.Release, ewasmEnv.Release)

	err := ewasmVm.RegisterModule(ewasmEnv)
	if err != nil {
		return cleanups, err
	}
	err = wasmutils.InstantiateWasm(ewasmVm, "", interpreters.EwasmInterpreter_1)
	if err != nil {
		return cleanups, err
	}

	ewasmFnList, ewasmFnTypes := ewasmVm.GetFunctionList()
	for i, name := range ewasmFnList {
		data := EwasmFunctionWrapper{Name: name, Vm: ewasmVm}
		fntype := ewasmFnTypes[i]
		wrappedFn := wasmedge.NewFunction(fntype, ewasm_wrapper, data, 0)
		contractEnv.AddFunction(name, wrappedFn)
	}

	// err = contractVm.RegisterModule(ewasmVm.GetActiveModule())
	err = contractVm.RegisterModule(contractEnv)
	if err != nil {
		return cleanups, err
	}

	return cleanups, nil
}

var SystemDepHandler = map[string]func(context *Context, contractVm *wasmedge.VM) ([]func(), error){}

func init() {
	SystemDepHandler[types.WASMX_WASMX_1] = InitiateWasmxWasmx1
	SystemDepHandler[types.WASMX_WASMX_2] = InitiateWasmxWasmx2
	SystemDepHandler[types.EWASM_ENV_1] = InitiateEwasmTypeEnv
	SystemDepHandler[types.INTERPRETER_EWASM_1] = InitiateEwasmTypeInterpreter
	SystemDepHandler[types.INTERPRETER_EVM_SHANGHAI] = InitiateEvmInterpreter_1
}

func VerifyEnv(version string, imports []*wasmedge.ImportType) error {
	// TODO check that all imports are supported by the given version

	// for _, mimport := range imports {
	// 	fmt.Println("Import:", mimport.GetModuleName(), mimport.GetExternalName())
	// }
	return nil
}
