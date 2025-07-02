package vmsql

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxSqlVMMock_i64(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		// Connect(req) -> resp
		vm.BuildFn("Connect", ConnectMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("Close", CloseMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("Ping", PingMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("Execute", ExecuteMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("BatchAtomic", BatchAtomicMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("Query", QueryMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
	}

	return vm.BuildModule(rnh, "sql", context, fndefs)
}
