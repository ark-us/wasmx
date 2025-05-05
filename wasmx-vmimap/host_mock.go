package vmimap

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ConnectWithPasswordMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* ConnectWithPasswordMock: %s", ctx.ContractInfo.Address.String())
	response := &ImapConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ConnectOAuth2Mock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* ConnectOAuth2Mock: %s", ctx.ContractInfo.Address.String())
	response := &ImapConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* CloseMock: %s", ctx.ContractInfo.Address.String())
	response := &ImapCloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ListenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* ListenMock: %s", ctx.ContractInfo.Address.String())
	response := &ImapListenResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func FetchMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* FetchMock: %s", ctx.ContractInfo.Address.String())
	response := &ImapFetchResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CreateFolderMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* CreateFolderMock: %s", ctx.ContractInfo.Address.String())
	response := &ImapCreateFolderResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildWasmxImapVMMock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("ConnectWithPassword", ConnectWithPasswordMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ConnectOAuth2", ConnectOAuth2Mock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", CloseMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Listen", ListenMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Fetch", FetchMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("CreateFolder", CreateFolderMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "imap", context, fndefs)
}
