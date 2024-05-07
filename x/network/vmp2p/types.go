package vmp2p

import (
	"context"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	mcodec "mythos/v1/codec"
	vmtypes "mythos/v1/x/wasmx/vm"
)

const HOST_WASMX_ENV_P2P_VER1 = "wasmx_p2p_1"

var HOST_WASMX_ENV_EXPORT = "wasmx_p2p_"

var HOST_WASMX_ENV_P2P = "p2p"

type ContextKey string

const P2PContextKey ContextKey = "p2p-context"

type Context struct {
	Context *vmtypes.Context
}

// internal use
type ContractMessage struct {
	Msg             []byte                    `json:"msg"`
	ContractAddress mcodec.AccAddressPrefixed `json:"contract_address"`
	SenderAddress   mcodec.AccAddressPrefixed `json:"sender_address"`
}

// sent to contracts
type P2PMessage struct {
	RoomId    string    `json:"roomId"`
	Message   []byte    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Sender    NodeInfo  `json:"sender"`
}

// internal use
type ChatRoomMessage struct {
	ContractMsg []byte    `json:"msg"`
	RoomId      string    `json:"roomId"`
	Timestamp   time.Time `json:"timestamp"`
	Sender      NodeInfo  `json:"sender"`
}

type NodeInfo struct {
	Id   string `json:"id"`
	Host string `json:"host"`
	Port string `json:"port"`
	Ip   string `json:"ip"` // can be empty if host is set
}

type MdnsService interface {
	Start() error
	Close() error
}

type P2PContext struct {
	Node      *host.Host
	PubSub    *pubsub.PubSub
	Mdns      MdnsService
	ChatRooms map[string]*ChatRoom
	Streams   map[string]network.Stream
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

type ConnectPeerResponse struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

type SendMessageToPeersRequest struct {
	Contract   mcodec.AccAddressPrefixed `json:"contract"`
	Msg        []byte                    `json:"msg"`
	ProtocolId string                    `json:"protocolId"`
	Peers      []string                  `json:"peers"`
}

type SendMessageToPeersResponse struct{}

type ConnectChatRoomRequest struct {
	ProtocolId string `json:"protocolId"`
	Topic      string `json:"topic"`
}

type ConnectChatRoomResponse struct{}

type SendMessageToChatRoomRequest struct {
	Contract   mcodec.AccAddressPrefixed `json:"contract"`
	Msg        []byte                    `json:"msg"`
	ProtocolId string                    `json:"protocolId"`
	Topic      string                    `json:"topic"`
}

type SendMessageToChatRoomResponse struct{}

type SendMessageRequest struct {
	Contract   mcodec.AccAddressPrefixed `json:"contract"`
	Msg        []byte                    `json:"msg"`
	ProtocolId string                    `json:"protocolId"`
}

type SendMessageResponse struct{}

type MsgStart struct {
	PrivateKey []byte     `json:"pk"`
	ProtocolId string     `json:"protocolId"`
	Node       NodeInfo   `json:"node"`
	Peers      []NodeInfo `json:"peers"`
}

type MsgStart2 struct {
	ProtocolIdd string `json:"protocolIdd"`
	ProtocolId  string `json:"protocolId"`
	PK          []byte `json:"pk"`
}

type CalldataStart struct {
	Start MsgStart `json:"start"`
}

func WithP2PEmptyContext(ctx context.Context) context.Context {
	p2pctx := &P2PContext{Streams: map[string]network.Stream{}, ChatRooms: map[string]*ChatRoom{}}
	return context.WithValue(ctx, P2PContextKey, p2pctx)
}

func WithP2PContext(ctx context.Context, p2pctx *P2PContext) context.Context {
	return context.WithValue(ctx, P2PContextKey, p2pctx)
}

func GetP2PContext(ctx *vmtypes.Context) (*P2PContext, error) {
	p2pctx_ := ctx.GoContextParent.Value(P2PContextKey)
	p2pctx := (p2pctx_).(*P2PContext)
	if p2pctx == nil {
		return nil, fmt.Errorf("p2p context not set")
	}
	return p2pctx, nil
}
