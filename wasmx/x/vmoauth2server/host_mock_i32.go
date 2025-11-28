package vmoauth2server

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxOAuth2ServerMock_i32(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("Initialize", InitializeMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RegisterClient", RegisterClientMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("UpdateClient", UpdateClientMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("DeleteClient", DeleteClientMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("GetClient", GetClientMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ValidateClient", ValidateClientMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("CreateAuthorizationCode", CreateAuthorizationCodeMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ValidateAuthorizationCode", ValidateAuthorizationCodeMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("IssueAccessToken", IssueAccessTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("IssueRefreshToken", IssueRefreshTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ValidateAccessToken", ValidateAccessTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RefreshAccessToken", RefreshAccessTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RevokeToken", RevokeTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("IntrospectToken", IntrospectTokenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("CleanupExpiredTokens", CleanupExpiredTokensMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "oauth2server", context, fndefs)
}
