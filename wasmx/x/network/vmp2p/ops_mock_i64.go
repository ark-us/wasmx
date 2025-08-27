package vmp2p

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateWasmxP2PJsonMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxP2P1Mock_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func BuildWasmxP2P1Mock_i64(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	logger := ctx_.GetContext().Logger().With("chain_id", ctx_.GetContext().ChainID())
	ctx := &Context{Context: ctx_, Logger: logger}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("StartNodeWithIdentity", StartNodeWithIdentityMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("GetNodeInfo", GetNodeInfoMock, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ConnectPeer", ConnectPeerMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("SendMessage", SendMessageMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("SendMessageToPeers", SendMessageToPeersMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ConnectChatRoom", ConnectChatRoomMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("SendMessageToChatRoom", SendMessageToChatRoomMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("CloseNode", CloseNodeMock, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("DisconnectChatRoom", DisconnectChatRoomMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("DisconnectPeer", DisconnectPeerMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		// TODO move to vmmc
		vm.BuildFn("StartStateSyncRequest", StartStateSyncRequestMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("StartStateSyncResponse", StartStateSyncResponseMock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
	}

	return vm.BuildModule(rnh, HOST_WASMX_ENV_P2P, ctx, fndefs)
}
