package vm

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	vmtypes "mythos/v1/x/wasmx/vm"
)

const HOST_WASMX_ENV_P2P_VER1 = "wasmx_p2p_1"

var HOST_WASMX_ENV_EXPORT = "wasmx_p2p_"

var HOST_WASMX_ENV_P2P = "p2p"


type Peer struct {
	Id   string `json:"id"`
	Port string `json:"port"`
	Host string `json:"host"`
}

type Context struct {
	vmtypes.Context
	node    *host.Host
	streams map[string]network.Stream
}

type StartNodeWithIdentityRequest struct {
	Port       string `json:"port"`
	ProtocolId string `json:"protocolId"`
	PrivateKey []byte `json:"pk"`
}

type StartNodeWithIdentityResponse struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

type ConnectPeerRequest struct {
	ProtocolId string `json:"protocolId"`
	Peer       string `json:"peer"`
}

type ConnectPeerResponse struct{}

type SendMessageToPeersRequest struct {
	Msg        []byte   `json:"msg"`
	ProtocolId string   `json:"protocolId"`
	Peers      []string `json:"peers"`
}

type SendMessageToPeersResponse struct{}

type MsgStart struct {
	PrivateKey []byte `json:"pk"`
	ProtocolId string `json:"protocolId"`
	Node       Peer   `json:"node"`
	Peers      []Peer `json:"peers"`
}

type MsgStart2 struct {
	ProtocolIdd string `json:"protocolIdd"`
	ProtocolId  string `json:"protocolId"`
	PK          []byte `json:"pk"`
}

type CalldataStart struct {
	Start MsgStart `json:"start"`
}
