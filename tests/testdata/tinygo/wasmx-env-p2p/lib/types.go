package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "wasmx_p2p"

// Core node types
type NetworkNode struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port string `json:"port"`
	IP   string `json:"ip"`
}

type NodeInfo struct {
	Address   wasmx.Bech32String `json:"address"`
	Node      NetworkNode        `json:"node"`
	OutOfSync bool               `json:"outofsync"`
}

type NodeInfoResponse struct {
	NodeInfo *NodeInfo `json:"node_info,omitempty"`
	Error    string    `json:"error"`
}

// Requests/Responses
type WasmxResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

type StartNodeRequest struct {
	Port       string `json:"port"`
	ProtocolId string `json:"protocolId"`
}

type StartNodeWithIdentityRequest struct {
	Port       string `json:"port"`
	ProtocolId string `json:"protocolId"`
	PK         []byte `json:"pk"`
}

type SendMessageRequest struct {
	Contract   wasmx.Bech32String `json:"contract"`
	Msg        []byte             `json:"msg"`
	ProtocolId string             `json:"protocolId"`
}

type SendMessageToPeersRequest struct {
	Contract   wasmx.Bech32String `json:"contract"`
	Sender     wasmx.Bech32String `json:"sender"`
	Msg        []byte             `json:"msg"`
	ProtocolId string             `json:"protocolId"`
	Peers      []string           `json:"peers"`
}

type ConnectPeerRequest struct {
	ProtocolId string `json:"protocolId"`
	Peer       string `json:"peer"`
}

type ConnectPeerResponse struct{}

type DisconnectPeerRequest struct {
	ProtocolId string `json:"protocolId"`
	Peer       string `json:"peer"`
}

type DisconnectPeerResponse struct{}

type ConnectChatRoomRequest struct {
	ProtocolId string `json:"protocolId"`
	Topic      string `json:"topic"`
}

type ConnectChatRoomResponse struct {
	Error string `json:"error"`
}

type DisconnectChatRoomRequest struct {
	ProtocolId string `json:"protocolId"`
	Topic      string `json:"topic"`
}

type DisconnectChatRoomResponse struct{}

type SendMessageToChatRoomRequest struct {
	Contract   wasmx.Bech32String `json:"contract"`
	Sender     wasmx.Bech32String `json:"sender"`
	Msg        []byte             `json:"msg"`
	ProtocolId string             `json:"protocolId"`
	Topic      string             `json:"topic"`
}

type SendMessageToChatRoomResponse struct {
	Error string `json:"error"`
}

type P2PMessage struct {
	RoomID    string      `json:"roomId"`
	Message   []byte      `json:"message"`
	Timestamp string      `json:"timestamp"`
	Sender    NetworkNode `json:"sender"`
}

type StartStateSyncReqRequest struct {
	StartHeight                 int64              `json:"start_height"`
	TrustHeight                 int64              `json:"trust_height"`
	TrustHash                   []byte             `json:"trust_hash"`
	PeerAddress                 string             `json:"peer_address"`
	ProtocolId                  string             `json:"protocol_id"`
	Peers                       []string           `json:"peers"`
	CurrentNodeId               int32              `json:"current_node_id"`
	VerificationChainId         string             `json:"verification_chain_id"`
	VerificationContractAddress wasmx.Bech32String `json:"verification_contract_address"`
}

type StartStateSyncReqResponse struct {
	Error string `json:"error"`
}

type StartStateSyncResRequest struct {
	PeerAddress string `json:"peer_address"`
	ProtocolId  string `json:"protocol_id"`
}

type StartStateSyncResResponse struct {
	Error string `json:"error"`
}
