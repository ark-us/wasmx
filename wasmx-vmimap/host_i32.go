package vmimap

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxImapVM_i32(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("Connect", Connect, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", Close, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Listen", Listen, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Count", Count, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("UIDSearch", UIDSearch, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ListMailboxes", ListMailboxes, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Fetch", Fetch, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("CreateFolder", CreateFolder, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ServerStart", ServerStart, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ServerClose", ServerClose, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),

		// TODO
		// Move
		// RenameMailbox
	}

	return vm.BuildModule(rnh, "imap", context, fndefs)
}
