package vmp2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"cosmossdk.io/log"

	vmtypes "mythos/v1/x/wasmx/vm"
)

const HOST_WASMX_ENV_P2P_VER1 = "wasmx_p2p_1"

const HOST_WASMX_ENV_EXPORT = "wasmx_p2p_"

const HOST_WASMX_ENV_P2P = "p2p"

type ContextKey string

const P2PContextKey ContextKey = "p2p-context"

type Context struct {
	Context *vmtypes.Context
	Logger  log.Logger
}

// internal use
type ContractMessage struct {
	Msg             []byte `json:"msg"`
	ContractAddress string `json:"contract_address"`
	SenderAddress   string `json:"sender_address"`
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
	ProtocolID  string    `json:"protocolID"`
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
	mtx                  sync.Mutex
	Node                 *host.Host
	PubSub               *pubsub.PubSub
	Mdns                 MdnsService
	ProtocolContexts     map[string]*ProtocolContext // indexed by protocol ID
	CustomMessageHandler map[string]func(netmsg P2PMessage, contractAddress string, senderAddress string)
	ssctx                *StateSyncContext
}

type ProtocolContext struct {
	ChatRooms map[string]*ChatRoom      // indexed by topic
	Streams   map[string]network.Stream // indexed by peer address
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
	Contract   string   `json:"contract"`
	Sender     string   `json:"sender"`
	Msg        []byte   `json:"msg"`
	ProtocolId string   `json:"protocolId"`
	Peers      []string `json:"peers"`
}

type SendMessageToPeersResponse struct{}

type ConnectChatRoomRequest struct {
	ProtocolId string `json:"protocolId"`
	Topic      string `json:"topic"`
}

type ConnectChatRoomResponse struct {
	Error string `json:"error"`
}

type SendMessageToChatRoomRequest struct {
	Contract   string `json:"contract"`
	Sender     string `json:"sender"`
	Msg        []byte `json:"msg"`
	ProtocolId string `json:"protocolId"`
	Topic      string `json:"topic"`
}

type SendMessageToChatRoomResponse struct {
	Error string `json:"error"`
}

type SendMessageRequest struct {
	Contract   string `json:"contract"`
	Msg        []byte `json:"msg"`
	ProtocolId string `json:"protocolId"`
}

type SendMessageResponse struct{}

type DisconnectPeerRequest struct {
	ProtocolId string `json:"protocolId"`
	Peer       string `json:"peer"`
}

type DisconnectPeerResponse struct{}

type DisconnectChatRoomRequest struct {
	ProtocolId string `json:"protocolId"`
	Topic      string `json:"topic"`
}

type DisconnectChatRoomResponse struct{}

type StartStateSyncReqRequest struct {
	Height      int64  `json:"trust_height"`
	Hash        []byte `json:"trust_hash"`
	ProtocolId  string `json:"protocol_id"`
	PeerAddress string `json:"peer_address"`
}

type StartStateSyncReqResponse struct {
	Error string `json:"error"`
}

type StartStateSyncRespRequest struct {
	ProtocolId  string `json:"protocol_id"`
	PeerAddress string `json:"peer_address"`
}

type StartStateSyncRespResponse struct {
	Error string `json:"error"`
}

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

func (p *P2PContext) GetPeers(protocolID string) ([]string, error) {
	pctx, found := p.ProtocolContexts[protocolID]
	if !found {
		return nil, fmt.Errorf("protocol ID not registered: %s", protocolID)
	}
	peers := []string{}
	for peeraddr := range pctx.Streams {
		peers = append(peers, peeraddr)
	}
	return peers, nil
}

func (p *P2PContext) GetPeer(protocolID string, peer string) (network.Stream, bool) {
	pctx, found := p.ProtocolContexts[protocolID]
	if !found {
		return nil, false
	}
	stream, found := pctx.Streams[peer]
	return stream, found
}

func (p *P2PContext) AddPeer(protocolID string, peer string, stream network.Stream) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.ProtocolContexts[protocolID]
	if !found {
		p.ProtocolContexts[protocolID] = &ProtocolContext{ChatRooms: map[string]*ChatRoom{}, Streams: map[string]network.Stream{}}
	}
	p.ProtocolContexts[protocolID].Streams[peer] = stream
}

func (p *P2PContext) DeletePeer(protocolID string, peer string) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	pctx, found := p.ProtocolContexts[protocolID]
	if !found {
		return fmt.Errorf("protocol ID not registered: %s", protocolID)
	}
	delete(pctx.Streams, peer)
	return nil
}

func (p *P2PContext) GetChatRoom(protocolID string, topic string) (*ChatRoom, bool) {
	pctx, found := p.ProtocolContexts[protocolID]
	if !found {
		return nil, false
	}
	cr, found := pctx.ChatRooms[topic]
	return cr, found
}

func (p *P2PContext) AddChatRoom(protocolID string, topic string, cr *ChatRoom) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.ProtocolContexts[protocolID]
	if !found {
		p.ProtocolContexts[protocolID] = &ProtocolContext{ChatRooms: map[string]*ChatRoom{}, Streams: map[string]network.Stream{}}
	}
	p.ProtocolContexts[protocolID].ChatRooms[topic] = cr
}

func (p *P2PContext) DeleteChatRoom(protocolID string, topic string) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	pctx, found := p.ProtocolContexts[protocolID]
	if !found {
		return fmt.Errorf("protocol ID not registered: %s", protocolID)
	}
	delete(pctx.ChatRooms, topic)
	return nil
}

func (p *P2PContext) AddCustomHandler(name string, handler func(netmsg P2PMessage, contractAddress string, senderAddress string)) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.CustomMessageHandler == nil {
		p.CustomMessageHandler = map[string]func(netmsg P2PMessage, contractAddress string, senderAddress string){}
	}
	p.CustomMessageHandler[name] = handler
}

func (p *P2PContext) GetCustomHandler(name string) func(netmsg P2PMessage, contractAddress string, senderAddress string) {
	if p.CustomMessageHandler == nil {
		return nil
	}
	handler, ok := p.CustomMessageHandler[name]
	if !ok {
		return nil
	}
	return handler
}

func WithP2PEmptyContext(ctx context.Context) context.Context {
	p2pctx := &P2PContext{ProtocolContexts: map[string]*ProtocolContext{}}
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
