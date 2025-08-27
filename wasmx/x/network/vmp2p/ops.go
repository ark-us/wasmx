package vmp2p

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"encoding/hex"
	"encoding/json"

	"bufio"
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	host "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	multiaddr "github.com/multiformats/go-multiaddr"

	errorsmod "cosmossdk.io/errors"

	"cosmossdk.io/log"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	ed25519 "github.com/cometbft/cometbft/crypto/ed25519"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"

	mcfg "github.com/loredanacirstea/wasmx/config"
)

func StartNodeWithIdentity(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req StartNodeWithIdentityRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}

	_, err = startNodeWithIdentityAndGossip(ctx.Context.GoContextParent, p2pctx, ctx.Logger, req.PrivateKey, req.Port, req.ProtocolId, ctx.handleStream)
	if err != nil {
		return nil, err
	}

	response := &StartNodeWithIdentityResponse{Error: "", Data: make([]byte, 0)}
	return prepareResponse(rnh, response)
}

func GetNodeInfo(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, errorsmod.Wrapf(err, "p2pcontext not found")
	}
	if p2pctx.Node == nil {
		ctx.Logger.Error(ERROR_NODE_NOT_INSTANTIATED)
		return nil, fmt.Errorf(ERROR_NODE_NOT_INSTANTIATED)
	}
	node := *p2pctx.Node
	response := &NodeInfo{Id: node.ID().String(), Ip: node.Addrs()[0].String()}
	return prepareResponse(rnh, response)
}

func ConnectPeer(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ConnectPeerResponse{}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ConnectPeerRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	_, err = connectAndListenPeerInternal(ctx.Context.GoContextParent, ctx.Context.GoRoutineGroup, ctx.Logger, req.ProtocolId, req.Peer, ctx.listenPeerStream)
	if err != nil {
		response.Error = err.Error()
	}

	return prepareResponse(rnh, response)
}

// sends to all connected peers
func SendMessage(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}

	var req SendMessageRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Logger.Debug("could not unmarshal SendMessageRequest", "error", err.Error())
		return nil, err
	}

	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}
	peers, err := p2pctx.GetPeers(req.ProtocolId)
	if err != nil {
		ctx.Logger.Error(err.Error())
		return nil, err
	}

	reqPeers := SendMessageToPeersRequest{
		Contract:   req.Contract,
		Msg:        req.Msg,
		ProtocolId: req.ProtocolId,
		Peers:      peers,
	}
	err = sendMessageToPeersCommon(ctx, reqPeers)
	if err != nil {
		return nil, err
	}

	response := &SendMessageResponse{}
	return prepareResponse(rnh, response)
}

func SendMessageToPeers(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SendMessageToPeersRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Logger.Error("send message to peers failed", "error", err.Error())
		return nil, err
	}
	err = sendMessageToPeersCommon(ctx, req)
	if err != nil {
		return nil, err
	}

	response := &SendMessageToPeersResponse{}
	return prepareResponse(rnh, response)
}

func ConnectChatRoom(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ConnectChatRoomResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ConnectChatRoomRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		response.Error = fmt.Sprintf("send message to chat room failed: %s", err.Error())
		ctx.Logger.Error(response.Error)
		return prepareResponse(rnh, response)
	}
	_, err = connectChatRoomAndListen(ctx, req.ProtocolId, req.Topic)
	// TODO send error to contract
	if err != nil {
		if err.Error() != ERROR_CTX_CANCELED {
			response.Error = fmt.Sprintf("error chat room connection: topic %s: %s", req.Topic, err.Error())
			ctx.Logger.Error("Error chat room connection ", "error", err.Error(), "topic", req.Topic)
			return prepareResponse(rnh, response)
		}
		// remove chat room; it will be reconnected when needed
		p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
		if err == nil {
			p2pctx.DeleteChatRoom(req.ProtocolId, req.Topic)
		}
	}
	return prepareResponse(rnh, response)
}

func SendMessageToChatRoom(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SendMessageToChatRoomResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SendMessageToChatRoomRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Logger.Error("send message to chat room failed", "error", err.Error())
		return nil, err
	}

	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}

	cr, found := p2pctx.GetChatRoom(req.ProtocolId, req.Topic)
	if !found {
		cr, err = connectChatRoomAndListen(ctx, req.ProtocolId, req.Topic)
		if err != nil {
			if err.Error() != ERROR_CTX_CANCELED {
				ctx.Logger.Error("Error chat room connection ", "error", err.Error(), "topic", req.Topic)
				return nil, err
			}
			// remove chat room; it will be reconnected when needed
			p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
			if err == nil {
				p2pctx.DeleteChatRoom(req.ProtocolId, req.Topic)
			}
		}
	}
	if cr != nil {
		err = sendMessageToChatRoomInternal(ctx, cr, req)
		if err != nil {
			response.Error = err.Error()
			return prepareResponse(rnh, response)
		}
	}
	return prepareResponse(rnh, response)
}

func CloseNode(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}
	if p2pctx.Node == nil {
		ctx.Logger.Error(ERROR_NODE_NOT_INSTANTIATED)
		return nil, fmt.Errorf(ERROR_NODE_NOT_INSTANTIATED)
	}
	node := *p2pctx.Node
	err = node.Close()
	ctx.Logger.Info("closing p2p node")

	// TODO response
	if err != nil {
		ctx.Logger.Error("failed to close p2p node", "err", err.Error())
	}

	returns := make([]interface{}, 0)
	return returns, nil
}

func DisconnectChatRoom(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req DisconnectChatRoomRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Logger.Error("disconnect chat room failed", "error", err.Error())
		return nil, err
	}
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}
	cr, found := p2pctx.GetChatRoom(req.ProtocolId, req.Topic)
	if found {
		cr.Unsubscribe()
		p2pctx.DeleteChatRoom(req.ProtocolId, req.Topic)
	}

	response := &DisconnectChatRoomResponse{}
	return prepareResponse(rnh, response)
}

func DisconnectPeer(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req DisconnectPeerRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Logger.Error("send message to chat room failed", "error", err.Error())
		return nil, err
	}
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}
	stream, found := p2pctx.GetPeer(req.ProtocolId, req.Peer)
	if found {
		stream.Close()
		p2pctx.DeletePeer(req.ProtocolId, req.Peer)
		ctx.Logger.Debug("p2p disconnect from peer", "protocolID", req.ProtocolId, "peer", req.Peer)
	}

	response := &DisconnectPeerResponse{}
	return prepareResponse(rnh, response)
}

// TODO this is temporary, to be replaced with consensus api methods like ApplySnapshotChunk, LoadSnapshotChunk, OfferSnapshot, ListSnapshots
// we currently dont use this method, because I could not figure out how to properly statesync after resetting the stores
func StartStateSyncRequest(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartStateSyncReqResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req StartStateSyncReqRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Logger.Error("state sync failed", "error", err.Error())
		return nil, err
	}
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}

	app := ctx.Context.App.(mcfg.MythosApp)

	connectToPeerFn := func() (network.Stream, error) {
		return connectAndListenPeerInternal(ctx.Context.GoContextParent, ctx.Context.GoRoutineGroup, ctx.Logger, req.ProtocolId, req.PeerAddress, ctx.listenPeerStream)
	}

	stream, found := p2pctx.GetPeer(req.ProtocolId, req.PeerAddress)
	if !found {
		stream, err = connectToPeerFn()
		if err != nil {
			ctx.Logger.Error("connect to peer failed", "peer", req.PeerAddress, "error", err.Error())
		}
	}

	sdklogger := ctx.Logger
	goContextParent := ctx.Context.GoContextParent
	interfaceRegistry := ctx.Context.CosmosHandler.Codec().InterfaceRegistry()
	jsonCdc := ctx.Context.CosmosHandler.JSONCodec()

	if stream != nil {
		tndcfg := app.GetTendermintConfig()
		tndcfg.StateSync.TrustHash = strings.ToUpper(hex.EncodeToString(req.Hash))
		tndcfg.StateSync.TrustHeight = req.Height
		ctx.Context.GoRoutineGroup.Go(func() error {
			time.Sleep(time.Second * 5)
			return startStateSyncRequest(goContextParent, sdklogger, interfaceRegistry, jsonCdc, tndcfg, ctx.Context.Ctx.ChainID(), *app.GetChainCfg(), app, app.GetRpcClient(), p2pctx, req.ProtocolId, req.PeerAddress, req.Peers, req.CurrentNodeId, stream, connectToPeerFn)

			// err = startStateSyncRequest(goContextParent, sdklogger, interfaceRegistry, jsonCdc, tndcfg, ctx.Context.Ctx.ChainID(), app, app.GetRpcClient(), p2pctx, req.ProtocolId, req.PeerAddress, stream, connectToPeerFn)
			// if err != nil {
			// 	response.Error = err.Error()
			// }
		})
	} else {
		response.Error = fmt.Sprintf("state sync failed: connect to peer failed: %s: %s", req.PeerAddress, err.Error())
	}

	return prepareResponse(rnh, response)
}

func StartStateSyncResponse(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartStateSyncRespResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req StartStateSyncRespRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Logger.Error("state sync failed", "error", err.Error())
		return nil, err
	}
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}

	app := ctx.Context.App.(mcfg.MythosApp)

	connectToPeerFn := func() (network.Stream, error) {
		return connectAndListenPeerInternal(ctx.Context.GoContextParent, ctx.Context.GoRoutineGroup, ctx.Logger, req.ProtocolId, req.PeerAddress, ctx.listenPeerStream)
	}

	stream, found := p2pctx.GetPeer(req.ProtocolId, req.PeerAddress)
	if !found {
		stream, err = connectToPeerFn()
		if err != nil {
			ctx.Logger.Error("connect to peer failed", "peer", req.PeerAddress, "error", err.Error())
		}
	}

	sdklogger := ctx.Logger
	goContextParent := ctx.Context.GoContextParent
	interfaceRegistry := ctx.Context.CosmosHandler.Codec().InterfaceRegistry()
	jsonCdc := ctx.Context.CosmosHandler.JSONCodec()

	if stream != nil {
		err = startStateSyncResponse(goContextParent, sdklogger, interfaceRegistry, jsonCdc, app.GetTendermintConfig(), ctx.Context.Ctx.ChainID(), *app.GetChainCfg(), app, app.GetRpcClient(), p2pctx, req.ProtocolId, req.PeerAddress, stream, connectToPeerFn)
		if err != nil {
			response.Error = err.Error()
		}
	} else {
		response.Error = fmt.Sprintf("state sync failed: connect to peer failed: %s: %s", req.PeerAddress, err.Error())
	}

	return prepareResponse(rnh, response)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

func startNodeWithIdentityAndGossip(
	goContextParent context.Context,
	p2pctx *P2PContext,
	logger log.Logger,
	privateKey []byte,
	port string,
	protocolId string,
	handleStream func(stream network.Stream),
) (host.Host, error) {
	var node host.Host
	var err error
	if p2pctx.Node != nil {
		node = *p2pctx.Node
		addladdr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/" + port)
		if err != nil {
			return nil, err
		}
		err = node.Network().Listen(addladdr)
		if err != nil {
			return nil, err
		}

		// print the node's PeerInfo in multiaddr format
		peerInfo := peerstore.AddrInfo{
			ID:    node.ID(),
			Addrs: node.Addrs(),
		}
		addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
		if err != nil {
			return nil, err
		}
		logger.Info("p2p node already started", "ID", node.ID(), "newport", port, "addresses", addrs)
	} else {
		node, err = startNodeWithIdentityInternal(privateKey, port)
		if err != nil {
			logger.Error("start p2p node with identity failed", "error", err.Error())
			return nil, err
		}
		// print the node's PeerInfo in multiaddr format
		peerInfo := peerstore.AddrInfo{
			ID:    node.ID(),
			Addrs: node.Addrs(),
		}
		addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
		if err != nil {
			return nil, err
		}
		logger.Info("started p2p node with identity", "ID", peerInfo.ID, "addresses", addrs, "newport", port)
	}
	p2pctx.Node = &node

	logger.Info("adding p2p protocol", "protocolId", protocolId)

	node.SetStreamHandler(protocol.ID(protocolId), handleStream)
	err = connectGossipSub(goContextParent, logger, p2pctx, protocolId)
	if err != nil {
		return nil, err
	}
	return node, nil
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

func connectAndListenPeerInternal(
	goContextParent context.Context,
	goRoutineGroup *errgroup.Group,
	logger log.Logger,
	protocolId, peeraddrstr string,
	listenPeerStream func(stream network.Stream, peeraddrstr string),
) (network.Stream, error) {
	p2pctx, err := GetP2PContext(goContextParent)
	if err != nil {
		logger.Error("p2pcontext not found")
		return nil, err
	}
	if p2pctx.Node == nil {
		return nil, fmt.Errorf(ERROR_NODE_NOT_INSTANTIATED)
	}
	stream, err := connectPeerInternal(*p2pctx.Node, protocolId, peeraddrstr)
	if err != nil {
		logger.Info("connect to peer failed", "peer", peeraddrstr, "protocol_id", protocolId, "error", err.Error())
		// try again
		stream, err = connectPeerInternal(*p2pctx.Node, protocolId, peeraddrstr)
		if err != nil {
			logger.Info("connect to peer failed", "peer", peeraddrstr, "protocol_id", protocolId, "error", err.Error())
			return nil, err
		}
	}
	p2pctx.AddPeer(protocolId, peeraddrstr, stream)

	goRoutineGroup.Go(func() error {
		intervalEnded := make(chan bool, 1)
		defer close(intervalEnded)
		go func(logger log.Logger, p2pctx_ *P2PContext) {
			logger.Info("goroutine peer connect started", "peer", peeraddrstr)
			defer logger.Info("goroutine peer connect finished", "peer", peeraddrstr)

			stream, found := p2pctx_.GetPeer(protocolId, peeraddrstr)
			if !found {
				logger.Debug("stream not found: ", "peer", peeraddrstr)
				intervalEnded <- true
			}
			listenPeerStream(stream, peeraddrstr)
		}(logger, p2pctx)

		select {
		case <-goContextParent.Done():
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
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil
	}

	msgReq := ContractMessage{
		Msg:             req.Msg,
		ContractAddress: req.Contract,
		SenderAddress:   req.Sender,
	}

	// make sure peers are connected
	for _, peer := range req.Peers {
		stream, found := p2pctx.GetPeer(req.ProtocolId, peer)
		if !found {
			stream, err = connectAndListenPeerInternal(ctx.Context.GoContextParent, ctx.Context.GoRoutineGroup, ctx.Logger, req.ProtocolId, peer, ctx.listenPeerStream)
			if err != nil {
				ctx.Logger.Error("connect to peer failed", "peer", peer, "error", err.Error())
			}
		}
		if stream != nil {
			err := sendMessageToPeersInternal(stream, msgReq)
			if err != nil {
				if err.Error() == ERROR_STREAM_RESET {
					// we just remove the stream from the mapping
					// if later needed, it will try to reconnect
					p2pctx.DeletePeer(req.ProtocolId, peer)
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
	p2pctx, err := GetP2PContext(ctx.Context.GoContextParent)
	if err != nil {
		ctx.Logger.Error("p2pcontext not found")
		return nil, err
	}

	// if we found the room, we unsubscribe and then resubscribe
	cr, found := p2pctx.GetChatRoom(protocolId, topic)
	if found && cr != nil {
		return cr, nil
		// ctx.Logger.Info("chat room connection found, resubscribing", "topic", topic, "protocolID", protocolId)
		// cr.Unsubscribe()
		// err := p2pctx.DeleteChatRoom(protocolId, topic)
		// ctx.Logger.Debug("failed to delete chat room connection", "error", err.Error(), "topic", topic, "protocolID", protocolId)
	}
	err = connectGossipSub(ctx.Context.GoContextParent, ctx.Logger, p2pctx, protocolId)
	if err != nil {
		ctx.Logger.Debug("connectGossipSub failed", "error", err.Error(), "topic", topic, "protocolID", protocolId)
		return nil, err
	}
	if p2pctx.Node == nil {
		return nil, fmt.Errorf(ERROR_NODE_NOT_INSTANTIATED)
	}
	cr, err = connectChatRoomInternal(ctx, p2pctx, *p2pctx.Node, protocolId, topic)
	if err != nil {
		ctx.Logger.Debug("connectChatRoomInternal failed", "error", err.Error(), "topic", topic, "protocolID", protocolId)
		return nil, err
	}

	ctx.Context.GoRoutineGroup.Go(func() error {
		intervalEnded := make(chan bool, 1)
		defer close(intervalEnded)
		go func(ctx_ *Context, p2pctx_ *P2PContext) {
			ctx_.Logger.Info("room connection started", "topic", topic)
			defer ctx_.Logger.Info("room connection successful", "topic", topic)

			cr, found := p2pctx_.GetChatRoom(protocolId, topic)
			if !found {
				ctx_.Logger.Debug("chat room not found: ", "topic", topic)
				intervalEnded <- true
			}
			err := listenChatRoomStream(cr)
			if err != nil {
				ctx_.Logger.Error("could not connect to room", "topic", topic, "error", err)
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

func connectGossipSub(
	goContextParent context.Context,
	logger log.Logger,
	p2pctx *P2PContext,
	protocolId string,
) error {
	// ctx := context.Background()
	if p2pctx.PubSub == nil {
		if p2pctx.Node == nil {
			return fmt.Errorf(ERROR_NODE_NOT_INSTANTIATED)
		}
		// create a new PubSub service using the GossipSub router
		// we can use pubsub.NewJSONTracer to trace messages
		ps, err := pubsub.NewGossipSub(goContextParent, *p2pctx.Node)
		if err != nil {
			return err
		}
		p2pctx.PubSub = ps
	}

	// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
	// This lets us automatically discover peers on the same LAN and connect to them.
	if p2pctx.Mdns == nil {
		if p2pctx.Node == nil {
			return fmt.Errorf(ERROR_NODE_NOT_INSTANTIATED)
		}
		// setup local mDNS discovery
		s := mdns.NewMdnsService(*p2pctx.Node, DiscoveryServiceTag, &DiscoveryNotifee{Node: *p2pctx.Node, Logger: logger, Context: goContextParent})
		err := s.Start()
		if err != nil {
			return err
		}
		p2pctx.Mdns = s
	}
	return nil
}

func connectChatRoomInternal(ctx *Context, p2pctx *P2PContext, node host.Host, protocolId string, topic string) (*ChatRoom, error) {
	cr, err := JoinChatRoom(ctx, p2pctx.PubSub, node.ID(), protocolId, topic)
	if err != nil {
		return nil, err
	}
	p2pctx.AddChatRoom(protocolId, topic, cr)
	return cr, nil
}

func listenChatRoomStream(cr *ChatRoom) error {
	go readDataChatRoomStd(cr)
	return nil
}

func readDataChatRoomStd(cr *ChatRoom) {
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
		SenderAddress:   req.Sender,
	}
	msgbz, err := json.Marshal(msgReq)
	if err != nil {
		return err
	}
	ctx.Logger.Debug("p2p publishing msg", "msg", string(msgbz), "topic", req.ProtocolId, "topic", req.Topic)
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
