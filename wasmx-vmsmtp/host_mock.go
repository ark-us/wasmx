package vmsmtp

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ConnectWithPasswordMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* ConnectWithPasswordMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ConnectOAuth2Mock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* ConnectOAuth2Mock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* CloseMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpCloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func QuitMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* QuitMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpQuitResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ExtensionMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* ExtensionMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpExtensionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func NoopMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* NoopMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpNoopResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func SendMailMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* SendMailMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpSendMailResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func VerifyMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* VerifyMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpVerifyResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func SupportsAuthMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* SupportsAuthMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpSupportsAuthResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func MaxMessageSizeMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* MaxMessageSizeMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpMaxMessageSizeResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildMailMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Ctx.Logger().Info("* BuildMailMock: %s", ctx.ContractInfo.Address.String())
	response := &SmtpBuildMailResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildWasmxSmtpVMMock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("ConnectWithPassword", ConnectWithPasswordMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ConnectOAuth2", ConnectOAuth2Mock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", CloseMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Quit", QuitMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Extension", ExtensionMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Noop", NoopMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SendMail", SendMailMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Verify", VerifyMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SupportsAuth", SupportsAuthMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("MaxMessageSize", MaxMessageSizeMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("BuildMail", BuildMailMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "smtp", context, fndefs)
}
