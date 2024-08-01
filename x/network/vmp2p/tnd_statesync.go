package vmp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "cosmossdk.io/log"
	abcicli "github.com/cometbft/cometbft/abci/client"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/node"
	cmtp2p "github.com/cometbft/cometbft/p2p"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	cmtversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/proxy"
	sm "github.com/cometbft/cometbft/state"
	statesync "github.com/cometbft/cometbft/statesync"
	"golang.org/x/sync/errgroup"

	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"

	// "github.com/cosmos/gogoproto/proto"

	mcfg "mythos/v1/config"
	networktypes "mythos/v1/x/network/types"
)

type StateSyncContext struct {
	ProtocolId        string
	AbciClient        abcicli.Client
	StateSyncReactor  *statesync.Reactor
	BcReactor         *MockBlockSyncReactor
	StateSyncProvider statesync.StateProvider
	StateStore        *StateStore
	StateSyncGenesis  sm.State
	Sw                *cmtp2p.Switch
	Peer              *Peer
	P2pctx            *P2PContext
	OnReceive         func(chID byte, msgBytes []byte)
}

func startStateSyncRequest(
	goContextParent context.Context,
	sdklogger log.Logger,
	interfaceRegistry types.InterfaceRegistry,
	jsonCdc codec.JSONCodec,
	ctndcfg *cmtcfg.Config,
	chainId string,
	app mcfg.MythosApp,
	rpcClient client.CometRPC,
	p2pctx *P2PContext,
	protocolId string,
	peeraddress string,
	stream network.Stream,
) error {
	if p2pctx.ssctx != nil {
		return fmt.Errorf("state sync process ongoing, cannot start another state sync process")
	}

	ssctx, err := InitializeStateSync(goContextParent, sdklogger, interfaceRegistry, jsonCdc, ctndcfg, chainId, app.GetBaseApp(), rpcClient, p2pctx, protocolId, peeraddress, stream)
	if err != nil {
		return err
	}

	err = resetStoresToVersion0(app)
	if err != nil {
		return err
	}

	err = node.StartStateSync(ssctx.StateSyncReactor, ssctx.BcReactor, ssctx.StateSyncProvider, ctndcfg.StateSync, ssctx.StateStore, nil, ssctx.StateSyncGenesis)
	if err != nil {
		return fmt.Errorf("failed to start state sync: %w", err)
	}
	return nil
}

func startStateSyncResponse(
	goContextParent context.Context,
	sdklogger log.Logger,
	interfaceRegistry types.InterfaceRegistry,
	jsonCdc codec.JSONCodec,
	ctndcfg *cmtcfg.Config,
	chainId string,
	app mcfg.MythosApp,
	rpcClient client.CometRPC,
	p2pctx *P2PContext,
	protocolId string,
	peeraddress string,
	stream network.Stream,
) error {
	fmt.Println("---startStateSyncResponse--")
	if p2pctx.ssctx != nil {
		return fmt.Errorf("state sync process ongoing, cannot start another state sync process")
	}

	_, err := InitializeStateSync(goContextParent, sdklogger, interfaceRegistry, jsonCdc, ctndcfg, chainId, app.GetBaseApp(), rpcClient, p2pctx, protocolId, peeraddress, stream)
	if err != nil {
		return err
	}
	return nil
}

type StateSyncP2PCtx struct {
	GoContextParent        context.Context
	Logger                 log.Logger
	HandleStateSyncMessage func(netmsg P2PMessage, contractAddress string, senderAddress string)
}

func (c *StateSyncP2PCtx) listenPeerStream(stream network.Stream, peeraddrstr string) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.GoContextParent, c.Logger, rw, stream.ID(), peeraddrstr, c.handleContractMessage)
	c.Logger.Debug("Connected to:", peeraddrstr)
}

func (c *StateSyncP2PCtx) handleStream(stream network.Stream) {
	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.GoContextParent, c.Logger, rw, stream.ID(), STREAM_MAIN, c.handleContractMessage)
}

func (c *StateSyncP2PCtx) handleContractMessage(msgbz []byte, frompeer string) {
	var msg ContractMessage
	err := json.Unmarshal(msgbz, &msg)
	if err != nil {
		c.Logger.Debug(fmt.Sprintf("ContractMessage unmarshal failed: %s; err: %s", string(msgbz), err.Error()))
	}
	netmsg := P2PMessage{
		Message:   msg.Msg,
		Timestamp: time.Now(),
		RoomId:    "",
		Sender:    NodeInfo{Ip: frompeer},
	}
	c.HandleStateSyncMessage(netmsg, msg.ContractAddress, msg.SenderAddress)
}

func InitializeStateSyncWithPeer(
	goContextParent context.Context,
	goRoutineGroup *errgroup.Group,
	sdklogger log.Logger,
	ctndcfg *cmtcfg.Config,
	chainId string,
	app mcfg.MythosApp,
	rpcClient client.CometRPC,
	protocolId string,
	peeraddress string,
	privateKey []byte,
	port string,
) (*StateSyncContext, error) {
	p2pctx, err := GetP2PContext(goContextParent)
	if err != nil {
		return nil, err
	}

	ssp2pctx := &StateSyncP2PCtx{
		GoContextParent: goContextParent,
		Logger:          sdklogger,
	}

	_, err = startNodeWithIdentityAndGossip(goContextParent, p2pctx, sdklogger, privateKey, port, protocolId, ssp2pctx.handleStream)
	if err != nil {
		return nil, err
	}

	var stream network.Stream
	for {
		stream, err = connectAndListenPeerInternal(goContextParent, goRoutineGroup, sdklogger, protocolId, peeraddress, ssp2pctx.listenPeerStream)
		if err == nil && stream != nil {
			break
		}
		time.Sleep(time.Second * 5)
	}

	ssctx, err := InitializeStateSync(goContextParent, sdklogger, app.InterfaceRegistry(), app.JSONCodec(), ctndcfg, chainId, app.GetBaseApp(), rpcClient, p2pctx, protocolId, peeraddress, stream)
	if err != nil {
		return nil, err
	}
	ssp2pctx.HandleStateSyncMessage = ssctx.handleStateSyncMessage
	return ssctx, nil
}

func InitializeStateSync(
	goContextParent context.Context,
	sdklogger log.Logger,
	interfaceRegistry types.InterfaceRegistry,
	jsonCdc codec.JSONCodec,
	ctndcfg *cmtcfg.Config,
	chainId string,
	bapp *baseapp.BaseApp,
	rpcClient client.CometRPC,
	p2pctx *P2PContext,
	protocolId string,
	peeraddress string,
	stream network.Stream,
) (*StateSyncContext, error) {
	// TODO store peer address, to be checked when we receive state sync messages
	// add custom handler

	peeraddr, err := multiaddr.NewMultiaddr(peeraddress)
	if err != nil {
		return nil, err
	}
	peerInfo, err := peerstore.AddrInfoFromP2pAddr(peeraddr)
	if err != nil {
		return nil, err
	}
	peer := NewPeer(peeraddress, stream, peerInfo, protocolId, p2pctx)

	logger := servercmtlog.CometLoggerWrapper{Logger: sdklogger}
	bcReactor := &MockBlockSyncReactor{}
	metricsProvider := node.DefaultMetricsProvider(ctndcfg.Instrumentation)
	_, p2pMetrics, _, _, proxyMetrics, _, ssMetrics := metricsProvider(chainId)

	cmtApp := server.NewCometABCIWrapper(bapp)

	abciClient := abcicli.NewLocalClient(new(cmtsync.Mutex), cmtApp)
	stateSyncReactor := statesync.NewReactor(
		*ctndcfg.StateSync,
		proxy.NewAppConnSnapshot(abciClient, proxyMetrics),
		proxy.NewAppConnQuery(abciClient, proxyMetrics),
		ssMetrics,
	)
	stateSyncReactor.SetLogger(logger.With("module", "statesync"))

	// state, genDoc, err := node.LoadStateFromDBOrGenesisDocProvider(stateDB, genesisDocProvider(chainId))

	res, err := bapp.Info(networktypes.RequestInfo)
	if err != nil {
		return nil, err
	}

	// TODO get the current state from contract
	stateSyncGenesis := sm.State{
		ChainID: chainId,
		Version: cmtstate.Version{
			Software: "",
			Consensus: cmtversion.Consensus{
				App:   res.AppVersion,
				Block: 0,
			},
		},
		InitialHeight: 1,
	}

	transport := &Transport{}
	peerFilters := make([]cmtp2p.PeerFilterFunc, 0)
	// TODO
	nodeInfo := cmtp2p.DefaultNodeInfo{}
	nodeKey := &cmtp2p.NodeKey{}

	sw := cmtp2p.NewSwitch(
		ctndcfg.P2P,
		transport,
		cmtp2p.WithMetrics(p2pMetrics),
		cmtp2p.SwitchPeerFilters(peerFilters...),
	)
	sw.SetLogger(logger)
	// sw.AddReactor("BLOCKSYNC", bcReactor)
	sw.AddReactor("STATESYNC", stateSyncReactor)
	sw.SetNodeInfo(nodeInfo)
	sw.SetNodeKey(nodeKey)

	stateSyncReactor.SetSwitch(sw)
	stateSyncReactor.AddPeer(peer)
	sw.AddPeer(peer)
	// services are stopped at StateStore.Bootstrap
	err = stateSyncReactor.Start()
	if err != nil {
		return nil, err
	}

	var stateSyncProvider statesync.StateProvider
	stateStore := StateStore{
		ChainId:           chainId,
		GoContextParent:   goContextParent,
		Logger:            sdklogger,
		InterfaceRegistry: interfaceRegistry,
		JsonCdc:           jsonCdc,
		StateSyncReactor:  stateSyncReactor,
		Sw:                sw,
	}

	ssctx := &StateSyncContext{
		AbciClient:        abciClient,
		StateSyncReactor:  stateSyncReactor,
		BcReactor:         bcReactor,
		StateSyncProvider: stateSyncProvider,
		StateStore:        &stateStore,
		StateSyncGenesis:  stateSyncGenesis,
		Peer:              peer,
		P2pctx:            p2pctx,
		ProtocolId:        protocolId,
		OnReceive:         cmtp2p.CreateOnReceive(peer, sw.GetReactorsByCh(), sw.GetMsgTypeByChID()),
	}
	p2pctx.AddCustomHandler(CUSTOM_HANDLER_STATESYNC, ssctx.handleStateSyncMessage)
	p2pctx.ssctx = ssctx
	return ssctx, nil
}

func (ssctx *StateSyncContext) handleStateSyncMessage(netmsg P2PMessage, contractAddress string, senderAddress string) {
	err := ssctx.handleStateSyncMessageWithError(netmsg)
	if err != nil {
		ssctx.Sw.Logger.Debug("handling statesync message failed", "error", err.Error())
	}
}

func (ssctx *StateSyncContext) handleStateSyncMessageWithError(netmsg P2PMessage) error {
	var msgwrap WrapMsg
	err := json.Unmarshal(netmsg.Message, &msgwrap)
	if err != nil {
		return err
	}
	ssctx.OnReceive(msgwrap.ChannelID, msgwrap.Msg)
	return nil
}

func resetStoresToVersion0(app mcfg.MythosApp) error {
	err := app.GetBaseApp().ResetStores()
	if err != nil {
		return err
	}

	// mapp.DebugDb()
	return nil
}
