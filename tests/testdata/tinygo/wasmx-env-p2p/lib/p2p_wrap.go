package lib

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-env-utils"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func StartNodeWithIdentity(req StartNodeWithIdentityRequest) (WasmxResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return WasmxResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "StartNodeWithIdentity", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(StartNodeWithIdentity_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "StartNodeWithIdentity", []string{"response", string(out)})
	var resp WasmxResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return WasmxResponse{}, err
	}
	return resp, nil
}

func GetNodeInfo() (NetworkNode, error) {
	out := utils.PackedPtrToBytes(GetNodeInfo_())
	var node NetworkNode
	if err := json.Unmarshal(out, &node); err != nil {
		return NetworkNode{}, err
	}
	return node, nil
}

func ConnectPeer(req ConnectPeerRequest) (ConnectPeerResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ConnectPeerResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "ConnectPeer", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(ConnectPeer_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "ConnectPeer", []string{"response", string(out)})
	var resp ConnectPeerResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}

func SendMessage(req SendMessageRequest) (WasmxResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return WasmxResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "SendMessage", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(SendMessage_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "SendMessage", []string{"response", string(out)})
	var resp WasmxResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return WasmxResponse{}, err
	}
	return resp, nil
}

func SendMessageToPeers(req SendMessageToPeersRequest) (WasmxResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return WasmxResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "SendMessageToPeers", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(SendMessageToPeers_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "SendMessageToPeers", []string{"response", string(out)})
	var resp WasmxResponse
	if len(out) == 0 {
		return WasmxResponse{}, nil
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return WasmxResponse{}, err
	}
	return resp, nil
}

func ConnectChatRoom(req ConnectChatRoomRequest) (ConnectChatRoomResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ConnectChatRoomResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "ConnectChatRoom", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(ConnectChatRoom_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "ConnectChatRoom", []string{"response", string(out)})
	var resp ConnectChatRoomResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}

func SendMessageToChatRoom(req SendMessageToChatRoomRequest) (SendMessageToChatRoomResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return SendMessageToChatRoomResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "SendMessageToChatRoom", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(SendMessageToChatRoom_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "SendMessageToChatRoom", []string{"response", string(out)})
	var resp SendMessageToChatRoomResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}

func CloseNode() (WasmxResponse, error) {
	out := utils.PackedPtrToBytes(CloseNode_())
	var resp WasmxResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}

func DisconnectChatRoom(req DisconnectChatRoomRequest) (DisconnectChatRoomResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return DisconnectChatRoomResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "DisconnectChatRoom", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(DisconnectChatRoom_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "DisconnectChatRoom", []string{"response", string(out)})
	var resp DisconnectChatRoomResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}

func DisconnectPeer(req DisconnectPeerRequest) (DisconnectPeerResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return DisconnectPeerResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "DisconnectPeer", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(DisconnectPeer_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "DisconnectPeer", []string{"response", string(out)})
	var resp DisconnectPeerResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}

func StartStateSyncRequest(req StartStateSyncReqRequest) (StartStateSyncReqResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return StartStateSyncReqResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "StartStateSyncRequest", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(StartStateSyncRequest_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "StartStateSyncRequest", []string{"response", string(out)})
	var resp StartStateSyncReqResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}

func StartStateSyncResponse(req StartStateSyncResRequest) (StartStateSyncResResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return StartStateSyncResResponse{}, err
	}
	wasmx.LoggerDebugExtended(MODULE_NAME, "StartStateSyncResponse", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(StartStateSyncResponse_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(MODULE_NAME, "StartStateSyncResponse", []string{"response", string(out)})
	var resp StartStateSyncResResponse
	if len(out) > 0 {
		_ = json.Unmarshal(out, &resp)
	}
	return resp, nil
}
