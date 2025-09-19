package vmp2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const HOST_WASMX_ENV_P2P_VER1 = "wasmx_p2p_1"
const HOST_WASMX_ENV_P2P_VER1_i32 = "wasmx_p2p_json_i32_1"
const HOST_WASMX_ENV_P2P_VER1_i64 = "wasmx_p2p_json_i64_1"

const HOST_WASMX_ENV_P2P_EXPORT = "wasmx_p2p_"

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
	Node                    *host.Host
	PubSub                  *pubsub.PubSub
	Mdns                    MdnsService
	mtxProtocolContexts     sync.Mutex
	ProtocolContexts        map[string]*ProtocolContext // indexed by protocol ID
	mtxCustomMessageHandler sync.Mutex
	CustomMessageHandler    map[string]func(netmsg P2PMessage, contractAddress string, senderAddress string)
	ssctx                   *StateSyncContext
}

type ProtocolContext struct {
	mtxChatRooms sync.Mutex
	ChatRooms    map[string]*ChatRoom // indexed by topic
	mtxStreams   sync.Mutex
	Streams      map[string]network.Stream // indexed by peer address
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
	Height                      int64                      `json:"trust_height"`
	Hash                        []byte                     `json:"trust_hash"`
	ProtocolId                  string                     `json:"protocol_id"`
	PeerAddress                 string                     `json:"peer_address"`
	Peers                       []string                   `json:"peers"`
	CurrentNodeId               int32                      `json:"current_node_id"`
	VerificationChainId         string                     `json:"verification_chain_id"`
	VerificationContractAddress *mcodec.AccAddressPrefixed `json:"verification_contract_address"`
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

func (p *ProtocolContext) getChatRoom(topic string) *ChatRoom {
	p.mtxChatRooms.Lock()
	defer p.mtxChatRooms.Unlock()
	room, found := p.ChatRooms[topic]
	if !found {
		return nil
	}
	return room
}

func (p *ProtocolContext) setChatRoom(topic string, room *ChatRoom) {
	p.mtxChatRooms.Lock()
	defer p.mtxChatRooms.Unlock()
	p.ChatRooms[topic] = room
}

func (p *ProtocolContext) deleteChatRoom(topic string) {
	p.mtxChatRooms.Lock()
	defer p.mtxChatRooms.Unlock()
	delete(p.ChatRooms, topic)
}

func (p *ProtocolContext) getStream(peer string) network.Stream {
	p.mtxStreams.Lock()
	defer p.mtxStreams.Unlock()
	stream, found := p.Streams[peer]
	if !found {
		return nil
	}
	return stream
}

func (p *ProtocolContext) setStream(peer string, stream network.Stream) {
	p.mtxStreams.Lock()
	defer p.mtxStreams.Unlock()
	p.Streams[peer] = stream
}

func (p *ProtocolContext) deleteStream(peer string) {
	p.mtxStreams.Lock()
	defer p.mtxStreams.Unlock()
	delete(p.Streams, peer)
}

func (p *P2PContext) getProtocolContext(protocolID string) *ProtocolContext {
	p.mtxProtocolContexts.Lock()
	defer p.mtxProtocolContexts.Unlock()
	pctx, found := p.ProtocolContexts[protocolID]
	if !found {
		return nil
	}
	return pctx
}

func (p *P2PContext) setProtocolContext(protocolID string, pctx *ProtocolContext) {
	p.mtxProtocolContexts.Lock()
	defer p.mtxProtocolContexts.Unlock()
	p.ProtocolContexts[protocolID] = pctx
}

func (p *P2PContext) getCustomMessageHandler(protocolID string) (func(netmsg P2PMessage, contractAddress string, senderAddress string), bool) {
	p.mtxCustomMessageHandler.Lock()
	defer p.mtxCustomMessageHandler.Unlock()
	handler, found := p.CustomMessageHandler[protocolID]
	return handler, found
}

func (p *P2PContext) setCustomMessageHandler(protocolID string, handler func(netmsg P2PMessage, contractAddress string, senderAddress string)) {
	p.mtxCustomMessageHandler.Lock()
	defer p.mtxCustomMessageHandler.Unlock()
	p.CustomMessageHandler[protocolID] = handler
}

func (p *P2PContext) GetPeers(protocolID string) ([]string, error) {
	pctx := p.getProtocolContext(protocolID)
	if pctx == nil {
		return nil, fmt.Errorf("protocol ID not registered: %s", protocolID)
	}
	peers := []string{}
	for peeraddr := range pctx.Streams {
		peers = append(peers, peeraddr)
	}
	return peers, nil
}

func (p *P2PContext) GetPeer(protocolID string, peer string) (network.Stream, bool) {
	pctx := p.getProtocolContext(protocolID)
	if pctx == nil {
		return nil, false
	}

	stream := pctx.getStream(peer)
	if stream == nil {
		return nil, false
	}
	return stream, true
}

func (p *P2PContext) AddPeer(protocolID string, peer string, stream network.Stream) {
	pctx := p.getProtocolContext(protocolID)
	if pctx == nil {
		p.setProtocolContext(protocolID, &ProtocolContext{ChatRooms: map[string]*ChatRoom{}, Streams: map[string]network.Stream{}})
	}
	pctx = p.getProtocolContext(protocolID)
	pctx.setStream(peer, stream)
}

func (p *P2PContext) DeletePeer(protocolID string, peer string) error {
	pctx := p.getProtocolContext(protocolID)
	if pctx == nil {
		return fmt.Errorf("protocol ID not registered: %s", protocolID)
	}
	pctx.deleteStream(peer)
	return nil
}

func (p *P2PContext) GetChatRoom(protocolID string, topic string) (*ChatRoom, bool) {
	pctx := p.getProtocolContext(protocolID)
	if pctx == nil {
		return nil, false
	}
	cr := pctx.getChatRoom(topic)
	if cr == nil {
		return nil, false
	}
	return cr, true
}

func (p *P2PContext) AddChatRoom(protocolID string, topic string, cr *ChatRoom) {
	pctx := p.getProtocolContext(protocolID)
	if pctx == nil {
		p.setProtocolContext(protocolID, &ProtocolContext{ChatRooms: map[string]*ChatRoom{}, Streams: map[string]network.Stream{}})
	}
	pctx = p.getProtocolContext(protocolID)
	pctx.setChatRoom(topic, cr)
}

func (p *P2PContext) DeleteChatRoom(protocolID string, topic string) error {
	pctx := p.getProtocolContext(protocolID)
	if pctx == nil {
		return fmt.Errorf("protocol ID not registered: %s", protocolID)
	}
	pctx.deleteChatRoom(topic)
	return nil
}

func (p *P2PContext) AddCustomHandler(name string, handler func(netmsg P2PMessage, contractAddress string, senderAddress string)) {
	if p.CustomMessageHandler == nil {
		p.CustomMessageHandler = map[string]func(netmsg P2PMessage, contractAddress string, senderAddress string){}
	}
	p.setCustomMessageHandler(name, handler)
}

func (p *P2PContext) GetCustomHandler(name string) func(netmsg P2PMessage, contractAddress string, senderAddress string) {
	if p.CustomMessageHandler == nil {
		return nil
	}
	handler, ok := p.getCustomMessageHandler(name)
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

func GetP2PContext(goContextParent context.Context) (*P2PContext, error) {
	p2pctx_ := goContextParent.Value(P2PContextKey)
	p2pctx := (p2pctx_).(*P2PContext)
	if p2pctx == nil {
		return nil, fmt.Errorf("p2p context not set")
	}
	return p2pctx, nil
}

type VerifyCommitLightRequest struct {
	ChainId string                `json:"chain_id"`
	BlockID cmttypes.BlockID      `json:"block_id"`
	Height  int64                 `json:"height"`
	Commit  cmttypes.Commit       `json:"commit"`
	ValSet  cmttypes.ValidatorSet `json:"valset"`
}

type VerifyCommitLightResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}
