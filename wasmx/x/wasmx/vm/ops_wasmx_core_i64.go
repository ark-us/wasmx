package vm

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxCoreEnvi64(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("migrateContractStateByStorageType", coreMigrateContractStateByStorageType, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("externalCall", coreExternalCall, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("grpcRequest", coreWasmxGrpcRequest, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("startTimeout", coreWasmxStartTimeout, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("cancelTimeout", coreWasmxCancelTimeout, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("startBackgroundProcess", coreWasmxStartBackgroundProcess, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("writeToBackgroundProcess", coreWasmxWriteToBackgroundProcess, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("readFromBackgroundProcess", coreWasmxReadFromBackgroundProcess, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		// TODO
		// env.AddFunction("endBackgroundProcess", NewFunction(functype_i32_, wasmxEndBackgroundProcess, context, 0))

		vm.BuildFn("storageLoadGlobal", coreWasmxStorageLoadGlobal, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("storageStoreGlobal", coreWasmxStorageStoreGlobal, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("storageDeleteGlobal", coreWasmxStorageDeleteGlobal, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("storageHasGlobal", coreWasmxStorageHasGlobal, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("storageResetGlobal", coreWasmxStorageResetGlobal, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
	}

	return vm.BuildModule(rnh, "wasmxcore", context, fndefs)
}
