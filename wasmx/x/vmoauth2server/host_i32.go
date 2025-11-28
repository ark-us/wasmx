package vmoauth2server

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxOAuth2Server_i32(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("Initialize", Initialize, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RegisterClient", RegisterClient, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("UpdateClient", UpdateClient, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("DeleteClient", DeleteClient, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("GetClient", GetClient, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ValidateClient", ValidateClient, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("CreateAuthorizationCode", CreateAuthorizationCode, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ValidateAuthorizationCode", ValidateAuthorizationCode, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("IssueAccessToken", IssueAccessToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("IssueRefreshToken", IssueRefreshToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ValidateAccessToken", ValidateAccessToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RefreshAccessToken", RefreshAccessToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RevokeToken", RevokeToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("IntrospectToken", IntrospectToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("CleanupExpiredTokens", CleanupExpiredTokens, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "oauth2server", context, fndefs)
}
