package vmhttpserver

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxHttpServerMock_i64(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("StartWebServer", StartWebServerMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("SetRouteHandler", SetRouteHandlerMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("RemoveRouteHandler", RemoveRouteHandlerMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("Close", CloseMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		// temporary, these should be provided by a smart contract
		vm.BuildFn("GenerateJWT", GenerateJWTMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("VerifyJWT", VerifyJWTMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
	}

	return vm.BuildModule(rnh, "httpserver", context, fndefs)
}
