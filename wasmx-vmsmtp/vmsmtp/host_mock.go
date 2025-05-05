package vmsmtp

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ConnectMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlCloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func PingMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlPingResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ExecuteMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlExecuteResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BatchAtomicMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlExecuteResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func QueryMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlQueryResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildWasmxSqlVMMock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		// Connect(req) -> resp
		vm.BuildFn("Connect", ConnectMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", CloseMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Ping", PingMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Execute", ExecuteMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("BatchAtomic", BatchAtomicMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Query", QueryMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "sql", context, fndefs)
}
