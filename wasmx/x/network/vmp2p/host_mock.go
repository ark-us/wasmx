package vmp2p

import (
	_ "embed"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func StartNodeWithIdentityMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartNodeWithIdentityResponse{Error: "", Data: make([]byte, 0)}
	return prepareResponse(rnh, response)
}

func GetNodeInfoMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &NodeInfo{Id: "", Ip: ""}
	return prepareResponse(rnh, response)
}

func ConnectPeerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ConnectPeerResponse{}
	return prepareResponse(rnh, response)
}

func SendMessageMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SendMessageResponse{}
	return prepareResponse(rnh, response)
}

func SendMessageToPeersMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SendMessageToPeersResponse{}
	return prepareResponse(rnh, response)
}

func ConnectChatRoomMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ConnectChatRoomResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func SendMessageToChatRoomMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SendMessageToChatRoomResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseNodeMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, nil
}

func DisconnectChatRoomMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &DisconnectChatRoomResponse{}
	return prepareResponse(rnh, response)
}

func DisconnectPeerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &DisconnectPeerResponse{}
	return prepareResponse(rnh, response)
}

func StartStateSyncRequestMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartStateSyncReqResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func StartStateSyncResponseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartStateSyncRespResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildWasmxP2P1Mock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	logger := ctx_.GetContext().Logger().With("chain_id", ctx_.GetContext().ChainID())
	ctx := &Context{Context: ctx_, Logger: logger}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("StartNodeWithIdentity", StartNodeWithIdentityMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("GetNodeInfo", GetNodeInfoMock, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ConnectPeer", ConnectPeerMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SendMessage", SendMessageMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SendMessageToPeers", SendMessageToPeersMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ConnectChatRoom", ConnectChatRoomMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SendMessageToChatRoom", SendMessageToChatRoomMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("CloseNode", CloseNodeMock, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("DisconnectChatRoom", DisconnectChatRoomMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("DisconnectPeer", DisconnectPeerMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),

		// TODO move to vmmc
		vm.BuildFn("StartStateSyncRequest", StartStateSyncRequestMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("StartStateSyncResponse", StartStateSyncResponseMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, HOST_WASMX_ENV_P2P, ctx, fndefs)
}
