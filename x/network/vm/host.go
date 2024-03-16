package vm

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"encoding/json"

	"bufio"
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	host "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	multiaddr "github.com/multiformats/go-multiaddr"

	"cosmossdk.io/log"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	"github.com/second-state/WasmEdge-go/wasmedge"

	ed25519 "github.com/cometbft/cometbft/crypto/ed25519"

	vmtypes "mythos/v1/x/wasmx/vm"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

func StartNodeWithIdentity(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req StartNodeWithIdentityRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	node, err := startNodeWithIdentityInternal(req.PrivateKey, req.Port)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("start p2p node with identity failed", "error", err.Error())
		return nil, wasmedge.Result_Fail
	}
	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil, wasmedge.Result_Fail
	}
	p2pctx.Node = &node

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.Context.Ctx.Logger().Info("started p2p node with identity", "ID", peerInfo.ID, "addresses", addrs)
	node.SetStreamHandler(protocol.ID(req.ProtocolId), ctx.handleStream)

	err = connectGossipSub(ctx, p2pctx, req.ProtocolId)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	response := StartNodeWithIdentityResponse{Error: "", Data: make([]byte, 0)}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func GetNodeInfo(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil, wasmedge.Result_Fail
	}
	node := *p2pctx.Node
	response := NodeInfo{Id: node.ID().String(), Ip: node.Addrs()[0].String()}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func ConnectPeer(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	response := ConnectPeerResponse{}
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req ConnectPeerRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	_, err = connectAndListenPeerInternal(ctx, req)
	if err != nil {
		response.Error = err.Error()
	}

	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// sends to all connected peers
func SendMessage(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	var req SendMessageRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Context.Ctx.Logger().Debug("could not unmarshal SendMessageRequest", "error", err.Error())
		return nil, wasmedge.Result_Fail
	}

	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil, wasmedge.Result_Fail
	}
	peers := []string{}
	for peeraddr := range p2pctx.Streams {
		peers = append(peers, peeraddr)
	}

	reqPeers := SendMessageToPeersRequest{
		Contract:   req.Contract,
		Msg:        req.Msg,
		ProtocolId: req.ProtocolId,
		Peers:      peers,
	}
	err = sendMessageToPeersCommon(ctx, reqPeers)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	response := SendMessageResponse{}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func SendMessageToPeers(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req SendMessageToPeersRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("send message to peers failed", "error", err.Error())
		return nil, wasmedge.Result_Fail
	}
	err = sendMessageToPeersCommon(ctx, req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	response := SendMessageToPeersResponse{}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func ConnectChatRoom(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req ConnectChatRoomRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("send message to chat room failed", "error", err.Error())
		return nil, wasmedge.Result_Fail
	}
	_, err = connectChatRoomAndListen(ctx, req.ProtocolId, req.Topic)
	// TODO send error to contract
	if err != nil {
		if err.Error() != ERROR_CTX_CANCELED {
			ctx.Context.Ctx.Logger().Error("Error chat room connection ", "error", err.Error(), "topic", req.Topic)
			return nil, wasmedge.Result_Fail
		}
		// remove chat room; it will be reconnected when needed
		p2pctx, err := GetP2PContext(ctx.Context)
		if err == nil {
			delete(p2pctx.ChatRooms, req.Topic)
		}
	}

	response := ConnectChatRoomResponse{}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func SendMessageToChatRoom(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req SendMessageToChatRoomRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("send message to chat room failed", "error", err.Error())
		return nil, wasmedge.Result_Fail
	}

	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil, wasmedge.Result_Fail
	}

	cr, found := p2pctx.ChatRooms[req.Topic]
	if !found {
		cr, err = connectChatRoomAndListen(ctx, req.ProtocolId, req.Topic)
		if err != nil {
			if err.Error() != ERROR_CTX_CANCELED {
				ctx.Context.Ctx.Logger().Error("Error chat room connection ", "error", err.Error(), "topic", req.Topic)
				return nil, wasmedge.Result_Fail
			}
			// remove chat room; it will be reconnected when needed
			p2pctx, err := GetP2PContext(ctx.Context)
			if err == nil {
				delete(p2pctx.ChatRooms, req.Topic)
			}
		}
	}
	// TODO send errors to contract
	if cr != nil {
		err = sendMessageToChatRoomInternal(ctx, cr, req)
		if err != nil {
			return nil, wasmedge.Result_Fail
		}
	}

	response := SendMessageToChatRoomResponse{}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func CloseNode(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil, wasmedge.Result_Fail
	}
	node := *p2pctx.Node
	err = node.Close()
	ctx.Context.Ctx.Logger().Info("closing p2p node")

	// TODO response
	if err != nil {
		ctx.Context.Ctx.Logger().Error("failed to close p2p node", "err", err.Error())
	}

	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildWasmxP2P1(ctx_ *vmtypes.Context) *wasmedge.Module {
	ctx := &Context{Context: ctx_}
	env := wasmedge.NewModule(HOST_WASMX_ENV_P2P)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	env.AddFunction("StartNodeWithIdentity", wasmedge.NewFunction(functype_i32_i32, StartNodeWithIdentity, ctx, 0))
	env.AddFunction("GetNodeInfo", wasmedge.NewFunction(functype__i32, GetNodeInfo, ctx, 0))
	env.AddFunction("CloseNode", wasmedge.NewFunction(functype__i32, CloseNode, ctx, 0))
	env.AddFunction("ConnectPeer", wasmedge.NewFunction(functype_i32_i32, ConnectPeer, ctx, 0))
	env.AddFunction("SendMessage", wasmedge.NewFunction(functype_i32_i32, SendMessage, ctx, 0))
	env.AddFunction("SendMessageToPeers", wasmedge.NewFunction(functype_i32_i32, SendMessageToPeers, ctx, 0))
	env.AddFunction("ConnectChatRoom", wasmedge.NewFunction(functype_i32_i32, ConnectChatRoom, ctx, 0))
	env.AddFunction("SendMessageToChatRoom", wasmedge.NewFunction(functype_i32_i32, SendMessageToChatRoom, ctx, 0))
	return env
}

func startNodeWithIdentityInternal(_pk []byte, port string) (host.Host, error) {
	pk := ed25519.PrivKey(_pk)
	pkcrypto, err := crypto.UnmarshalEd25519PrivateKey(pk)
	if err != nil {
		return nil, err
	}
	identity := libp2p.Identity(pkcrypto)
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+port),
		libp2p.Ping(false),
		identity,
	)

	if err != nil {
		return nil, err
	}
	return node, nil
}

func connectPeerInternal(node host.Host, protocolID string, peeraddrstr string) (network.Stream, error) {
	ctx := context.Background()
	peeraddr, err := multiaddr.NewMultiaddr(peeraddrstr)
	if err != nil {
		return nil, err
	}
	peer, err := peerstore.AddrInfoFromP2pAddr(peeraddr)
	if err != nil {
		return nil, err
	}
	if err := node.Connect(ctx, *peer); err != nil {
		return nil, err
	}

	// open a stream, this stream will be handled by handleStream other end
	stream, err := node.NewStream(ctx, peer.ID, protocol.ID(protocolID))
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func connectAndListenPeerInternal(ctx *Context, req ConnectPeerRequest) (network.Stream, error) {
	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil, err
	}

	stream, err := connectPeerInternal(*p2pctx.Node, req.ProtocolId, req.Peer)
	if err != nil {
		ctx.Context.Ctx.Logger().Info("connect to peer failed", "peer", req.Peer, "protocol_id", req.ProtocolId, "error", err.Error())
		return nil, err
	}
	p2pctx.Streams[req.Peer] = stream

	ctx.Context.GoRoutineGroup.Go(func() error {
		intervalEnded := make(chan bool, 1)
		defer close(intervalEnded)
		go func(ctx *Context, p2pctx_ *P2PContext) {
			ctx.Context.Ctx.Logger().Info("goroutine peer connect started", "peer", req.Peer)
			defer ctx.Context.Ctx.Logger().Info("goroutine peer connect finished", "peer", req.Peer)

			stream, found := p2pctx_.Streams[req.Peer]
			if !found {
				ctx.Context.Ctx.Logger().Debug("stream not found: ", "peer", req.Peer)
				intervalEnded <- true
			}
			ctx.listenPeerStream(stream, req.Peer)
		}(ctx, p2pctx)

		select {
		case <-ctx.Context.GoContextParent.Done():
			return nil
		case <-intervalEnded:
			return nil
		}
	})
	return stream, nil
}

func sendMessageToPeersInternal(stream network.Stream, msg ContractMessage) error {
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	err = writeData(rw, msgbz)
	return err
}

func writeData(rw *bufio.ReadWriter, msg []byte) error {
	_, err := rw.WriteString(string(msg) + "\n")
	if err != nil {
		return err
	}
	err = rw.Flush()
	if err != nil {
		return err
	}
	return nil
}

func sendMessageToPeersCommon(ctx *Context, req SendMessageToPeersRequest) error {
	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil
	}

	msgReq := ContractMessage{
		Msg:             req.Msg,
		ContractAddress: req.Contract,
		SenderAddress:   ctx.Context.Env.Contract.Address,
	}

	// make sure peers are connected
	for _, peer := range req.Peers {
		stream, found := p2pctx.Streams[peer]
		if !found {
			stream, err = connectAndListenPeerInternal(ctx, ConnectPeerRequest{ProtocolId: req.ProtocolId, Peer: peer})
			if err != nil {
				ctx.Context.Ctx.Logger().Error("connect to peer failed", "peer", peer, "error", err.Error())
			}
		}
		if stream != nil {
			err := sendMessageToPeersInternal(stream, msgReq)
			if err != nil {
				if err.Error() == ERROR_STREAM_RESET {
					// we just remove the stream from the mapping
					// if later needed, it will try to reconnect
					delete(p2pctx.Streams, peer)
				}
			}
		}
	}
	return nil
}

// DiscoveryNotifee gets notified when we find a new peer via mDNS discovery
type DiscoveryNotifee struct {
	Node    host.Host
	Logger  log.Logger
	Context context.Context
}

var TRY_RECONNECT = 10
var RECONNECT_TIMEOUT = time.Second * 2

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *DiscoveryNotifee) HandlePeerFound(pi peerstore.AddrInfo) {
	n.Logger.Info("discovered new peer", "ID", pi.ID)
	// err := n.Node.Connect(n.Context, pi)
	// context.Background()
	err := tryPeerConnect(n.Context, n, pi, 0)
	if err != nil {
		fmt.Println("error connecting to peer", pi.ID, err)
		n.Logger.Info("error connecting to peer", "ID", pi.ID, "error", err)
	}
}

func tryPeerConnect(ctx context.Context, n *DiscoveryNotifee, pi peerstore.AddrInfo, retryCount int32) error {
	err := n.Node.Connect(ctx, pi)
	if err != nil {
		if retryCount < int32(TRY_RECONNECT) {
			time.Sleep(RECONNECT_TIMEOUT)
			return tryPeerConnect(ctx, n, pi, retryCount+1)
		} else {
			return err
		}
	}
	return nil
}

func connectChatRoomAndListen(ctx *Context, protocolId string, topic string) (*ChatRoom, error) {
	p2pctx, err := GetP2PContext(ctx.Context)
	if err != nil {
		ctx.Context.Ctx.Logger().Error("p2pcontext not found")
		return nil, err
	}

	err = connectGossipSub(ctx, p2pctx, protocolId)
	if err != nil {
		return nil, err
	}

	cr, err := connectChatRoomInternal(ctx, p2pctx, *p2pctx.Node, protocolId, topic)
	if err != nil {
		return nil, err
	}

	ctx.Context.GoRoutineGroup.Go(func() error {
		intervalEnded := make(chan bool, 1)
		defer close(intervalEnded)
		go func(ctx_ *Context, p2pctx_ *P2PContext) {
			fmt.Println("goroutine room connect started", topic)
			defer fmt.Println("goroutine room connect finished", topic)

			cr, found := p2pctx_.ChatRooms[topic]
			if !found {
				ctx.Context.Ctx.Logger().Debug("chat room not found: ", "topic", topic)
				intervalEnded <- true
			}
			err := listenChatRoomStream(cr)
			if err != nil {
				fmt.Println("connect room found, err", found, err)
				intervalEnded <- true
			}

		}(ctx, p2pctx)

		select {
		case <-ctx.Context.GoContextParent.Done():
			return nil
		case <-intervalEnded:
			return nil
		}
	})
	return cr, nil
}

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

func connectGossipSub(ctx *Context, p2pctx *P2PContext, protocolId string) error {
	goctx := ctx.Context.GoContextParent
	// ctx := context.Background()
	if p2pctx.PubSub == nil {
		// create a new PubSub service using the GossipSub router
		// we can use pubsub.NewJSONTracer to trace messages
		ps, err := pubsub.NewGossipSub(goctx, *p2pctx.Node)
		if err != nil {
			return err
		}
		p2pctx.PubSub = ps
	}

	// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
	// This lets us automatically discover peers on the same LAN and connect to them.
	if p2pctx.Mdns == nil {
		// setup local mDNS discovery
		logger := ctx.Context.Ctx.Logger()
		s := mdns.NewMdnsService(*p2pctx.Node, DiscoveryServiceTag, &DiscoveryNotifee{Node: *p2pctx.Node, Logger: logger, Context: goctx})
		err := s.Start()
		if err != nil {
			return err
		}
		p2pctx.Mdns = s
	}
	return nil
}

func connectChatRoomInternal(ctx *Context, p2pctx *P2PContext, node host.Host, protocolId string, topic string) (*ChatRoom, error) {
	cr, err := JoinChatRoom(ctx, p2pctx.PubSub, node.ID(), topic)
	if err != nil {
		return nil, err
	}
	p2pctx.ChatRooms[topic] = cr
	return cr, nil
}

func listenChatRoomStream(cr *ChatRoom) error {
	go readDataChatRoomStd(cr)
	return nil
}

func readDataChatRoomStd(cr *ChatRoom) {
	fmt.Println("reading stream data from room: ", cr.roomName)
	for {
		select {
		case m := <-cr.Messages:
			cr.ctx.handleChatRoomMessage(m)
		case <-cr.ctx.Context.GoContextParent.Done():
			return
		}
	}
}

func sendMessageToChatRoomInternal(ctx *Context, cr *ChatRoom, req SendMessageToChatRoomRequest) error {
	msgReq := ContractMessage{
		Msg:             req.Msg,
		ContractAddress: req.Contract,
		SenderAddress:   ctx.Context.Env.Contract.Address,
	}
	msgbz, err := json.Marshal(msgReq)
	if err != nil {
		return err
	}
	err = cr.Publish(msgbz)
	if err != nil {
		return err
	}
	return nil
}

// defaultNick generates a nickname based on the last 8 chars of a peer ID.
func defaultNick(p peerstore.ID) string {
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), shortID(p))
}

// shortID returns the last 8 chars of a base58-encoded peer id.
func shortID(p peerstore.ID) string {
	pretty := p.String()
	return pretty[len(pretty)-8:]
}
