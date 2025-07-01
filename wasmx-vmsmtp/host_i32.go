package vmsmtp

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxSmtpVM_i32(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("ConnectWithPassword", ConnectWithPassword, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ConnectOAuth2", ConnectOAuth2, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", Close, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Quit", Quit, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Extension", Extension, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Noop", Noop, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SendMail", SendMail, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Verify", Verify, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SupportsAuth", SupportsAuth, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("MaxMessageSize", MaxMessageSize, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("BuildMail", BuildMail, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("StartServer", StartServer, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "smtp", context, fndefs)
}
