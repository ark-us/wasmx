package vmoauth2client

import (
	vmhttpclient "github.com/loredanacirstea/wasmx/x/vmhttpclient"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func GetRedirectUrlMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &GetRedirectUrlResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ExchangeCodeForTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ExchangeCodeForTokenResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func RefreshTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RefreshTokenResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func Oauth2ClientGetMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &Oauth2ClientGetResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func Oauth2ClientDoMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &vmhttpclient.HttpResponseWrap{Error: ""}
	return prepareResponse(rnh, response)
}

func Oauth2ClientPostMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &vmhttpclient.HttpResponseWrap{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildWasmxOAuth2ClientMock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("GetRedirectUrl", GetRedirectUrlMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ExchangeCodeForToken", ExchangeCodeForTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RefreshToken", RefreshTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Get", Oauth2ClientGetMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Do", Oauth2ClientDoMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Post", Oauth2ClientPostMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "oauth2client", context, fndefs)
}
