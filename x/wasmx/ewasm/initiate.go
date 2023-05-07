package ewasm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/ewasm/wasmutils"
)

const coreOpcodesModule = "../ewasm/contracts/ewasm.wasm"

var (
	EWASM_VM_EXPORT          = "ewasm_env_"
	EWASM_INTERPRETER_EXPORT = "ewasm_ewasm_"

	REQUIRED_IBC_EXPORTS   = []string{}
	REQUIRED_EWASM_EXPORTS = []string{"codesize", "main", "instantiate"}
	// codesize_constructor
)

func InitiateWasmTypeEnv(context *Context, contractVm *wasmedge.VM) ([]func(), error) {
	ewasmEnv := BuildEwasmEnv(context)
	var cleanups []func()
	cleanups = append(cleanups, ewasmEnv.Release)
	err := contractVm.RegisterModule(ewasmEnv)
	if err != nil {
		return cleanups, err
	}
	return cleanups, nil
}

func InitiateWasmTypeInterpreter(context *Context, contractVm *wasmedge.VM) ([]func(), error) {
	var cleanups []func()
	contractEnv := wasmedge.NewModule("ewasm")
	ewasmVm := wasmedge.NewVM()
	ewasmEnv := BuildEwasmEnv(context)
	cleanups = append(cleanups, contractEnv.Release, ewasmVm.Release, ewasmEnv.Release)

	err := ewasmVm.RegisterModule(ewasmEnv)
	if err != nil {
		return cleanups, err
	}
	err = wasmutils.InstantiateWasm(ewasmVm, coreOpcodesModule, nil)
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
	SystemDepHandler["ewasm_env_1"] = InitiateWasmTypeEnv
	SystemDepHandler["ewasm_ewasm_1"] = InitiateWasmTypeInterpreter
}

func VerifyEnv(version string, imports []*wasmedge.ImportType) error {
	// TODO check that all imports are supported by the given version

	// for _, mimport := range imports {
	// 	fmt.Println("Import:", mimport.GetModuleName(), mimport.GetExternalName())
	// }
	return nil
}
