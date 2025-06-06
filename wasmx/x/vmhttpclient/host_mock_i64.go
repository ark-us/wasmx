package vmhttpclient

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxHttpClientMock_i64(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("Request", RequestMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
	}

	return vm.BuildModule(rnh, "httpclient", context, fndefs)
}
