package vmp2p

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateWasmxP2PJson_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxP2P1_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func BuildWasmxP2P1_i64(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	logger := ctx_.GetContext().Logger().With("chain_id", ctx_.GetContext().ChainID())
	ctx := &Context{Context: ctx_, Logger: logger}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("StartNodeWithIdentity", StartNodeWithIdentity, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("GetNodeInfo", GetNodeInfo, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ConnectPeer", ConnectPeer, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("SendMessage", SendMessage, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("SendMessageToPeers", SendMessageToPeers, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ConnectChatRoom", ConnectChatRoom, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("SendMessageToChatRoom", SendMessageToChatRoom, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("CloseNode", CloseNode, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("DisconnectChatRoom", DisconnectChatRoom, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("DisconnectPeer", DisconnectPeer, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		// TODO move to vmmc
		vm.BuildFn("StartStateSyncRequest", StartStateSyncRequest, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("StartStateSyncResponse", StartStateSyncResponse, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
	}

	return vm.BuildModule(rnh, HOST_WASMX_ENV_P2P, ctx, fndefs)
}
