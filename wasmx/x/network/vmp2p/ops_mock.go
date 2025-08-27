package vmp2p

import (
	_ "embed"

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
