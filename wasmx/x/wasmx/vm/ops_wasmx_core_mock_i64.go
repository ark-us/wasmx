package vm

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxCoreEnvMocki64(context *Context, rnh memc.RuntimeHandler, modname string) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("migrateContractStateByStorageType", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("externalCall", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("grpcRequest", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("startTimeout", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("cancelTimeout", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("startBackgroundProcess", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("writeToBackgroundProcess", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("readFromBackgroundProcess", MockWithPanic, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		// TODO
		// env.AddFunction("endBackgroundProcess", NewFunction(functype_i32_, wasmxEndBackgroundProcess, context, 0))
	}
	if modname == "" {
		modname = "wasmxcore"
	}
	return vm.BuildModule(rnh, modname, context, fndefs)
}
