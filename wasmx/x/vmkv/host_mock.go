package vmkv

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ConnectMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvCloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func GetMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvGetResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func HasMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvHasResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func SetMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvSetResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func DeleteMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvDeleteResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func IteratorMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvIteratorResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func NewBatchMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvNewBatchResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func PrintMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, nil
}

func StatsMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := map[string]string{}
	return prepareResponse(rnh, response)
}

func BuildWasmxKvVMMock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	// follow cosmos-db interface
	fndefs := []memc.IFn{
		vm.BuildFn("Connect", ConnectMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", CloseMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Get", GetMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Has", HasMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Set", SetMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Delete", DeleteMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Iterator", IteratorMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("NewBatch", NewBatchMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("Print", PrintMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("Stats", StatsMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "kvdb", context, fndefs)
}
