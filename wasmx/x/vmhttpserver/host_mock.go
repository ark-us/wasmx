package vmhttpserver

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func StartWebServerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartWebServerResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func SetRouteHandlerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SetRouteHandlerResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func RemoveRouteHandlerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RemoveRouteHandlerResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &CloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func GenerateJWTMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &GenerateJWTResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func VerifyJWTMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &VerifyJWTResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildWasmxHttpServerMock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("StartWebServer", StartWebServerMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SetRouteHandler", SetRouteHandlerMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RemoveRouteHandler", RemoveRouteHandlerMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", CloseMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),

		// temporary, these should be provided by a smart contract
		vm.BuildFn("GenerateJWT", GenerateJWTMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("VerifyJWT", VerifyJWTMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "httpserver", context, fndefs)
}
