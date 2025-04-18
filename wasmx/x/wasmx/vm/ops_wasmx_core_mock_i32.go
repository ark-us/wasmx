package vm

import (
	"fmt"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func MockWithPanic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, fmt.Errorf("wasmx core not allowed")
}

func BuildWasmxCoreEnvMocki32(context *Context, rnh memc.RuntimeHandler, modname string) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("migrateContractStateByStorageType", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("externalCall", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("grpcRequest", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("startTimeout", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("cancelTimeout", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("startBackgroundProcess", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("writeToBackgroundProcess", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("readFromBackgroundProcess", MockWithPanic, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}
	if modname == "" {
		modname = "wasmxcore"
	}
	return vm.BuildModule(rnh, modname, context, fndefs)
}
