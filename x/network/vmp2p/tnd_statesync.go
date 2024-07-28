package vmp2p

import (
	"encoding/json"
	"fmt"

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

	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"

	// "github.com/cosmos/gogoproto/proto"

	networktypes "mythos/v1/x/network/types"
)

type StateSyncContext struct {
	protocolId        string
	abciClient        abcicli.Client
	stateSyncReactor  *statesync.Reactor
	bcReactor         *MockBlockSyncReactor
	stateSyncProvider statesync.StateProvider
	stateStore        *StateStore
	stateSyncGenesis  sm.State
	sw                *cmtp2p.Switch
	peer              *Peer
	p2pctx            *P2PContext
	onReceive         func(chID byte, msgBytes []byte)
}

func startStateSyncRequest(sdklogger log.Logger, ctndcfg *cmtcfg.Config, chainId string, bapp *baseapp.BaseApp, rpcClient client.CometRPC, p2pctx *P2PContext, protocolId string, peeraddress string, stream network.Stream) error {
	fmt.Println("---startStateSync--")
	if p2pctx.ssctx != nil {
		return fmt.Errorf("state sync process ongoing, cannot start another state sync process")
	}

	ssctx, err := initializeStateSync(sdklogger, ctndcfg, chainId, bapp, rpcClient, p2pctx, protocolId, peeraddress, stream)
	if err != nil {
		return err
	}
	err = node.StartStateSync(ssctx.stateSyncReactor, ssctx.bcReactor, ssctx.stateSyncProvider, ctndcfg.StateSync, ssctx.stateStore, nil, ssctx.stateSyncGenesis)
	if err != nil {
		return fmt.Errorf("failed to start state sync: %w", err)
	}
	return nil
}

func startStateSyncResponse(sdklogger log.Logger, ctndcfg *cmtcfg.Config, chainId string, bapp *baseapp.BaseApp, rpcClient client.CometRPC, p2pctx *P2PContext, protocolId string, peeraddress string, stream network.Stream) error {
	fmt.Println("---startStateSyncResponse--")
	if p2pctx.ssctx != nil {
		return fmt.Errorf("state sync process ongoing, cannot start another state sync process")
	}
	_, err := initializeStateSync(sdklogger, ctndcfg, chainId, bapp, rpcClient, p2pctx, protocolId, peeraddress, stream)
	if err != nil {
		return err
	}
	return nil
}

func initializeStateSync(sdklogger log.Logger, ctndcfg *cmtcfg.Config, chainId string, bapp *baseapp.BaseApp, rpcClient client.CometRPC, p2pctx *P2PContext, protocolId string, peeraddress string, stream network.Stream) (*StateSyncContext, error) {
	// TODO store peer address, to be checked when we receive state sync messages
	// add custom handler
	fmt.Println("---startStateSync--")

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

	var stateSyncProvider statesync.StateProvider
	stateStore := StateStore{}
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

	// sw1 := &Switch{}

	// transport, peerFilters := createTransport(config, nodeInfo, nodeKey, proxyApp)
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

	// sw := *cmtp2p.Switch{
	// 	BaseService: *cmtservice.NewBaseService(logger, )
	// }
	stateSyncReactor.SetSwitch(sw)
	// stateSyncReactor.SetSwitch(sw1)

	stateSyncReactor.AddPeer(peer)
	sw.AddPeer(peer)
	err = stateSyncReactor.Start()
	if err != nil {
		return nil, err
	}

	ssctx := &StateSyncContext{
		abciClient:        abciClient,
		stateSyncReactor:  stateSyncReactor,
		bcReactor:         bcReactor,
		stateSyncProvider: stateSyncProvider,
		stateStore:        &stateStore,
		stateSyncGenesis:  stateSyncGenesis,
		peer:              peer,
		p2pctx:            p2pctx,
		protocolId:        protocolId,
		onReceive:         cmtp2p.CreateOnReceive(peer, sw.GetReactorsByCh(), sw.GetMsgTypeByChID()),
	}
	p2pctx.AddCustomHandler(CUSTOM_HANDLER_STATESYNC, ssctx.handleStateSyncMessage)
	p2pctx.ssctx = ssctx
	return ssctx, nil

	// TODO stop state sync
	// stateSyncReactor.Stop()
}

func (ssctx *StateSyncContext) handleStateSyncMessage(netmsg P2PMessage, contractAddress string, senderAddress string) {
	fmt.Println("-handleStateSyncMessage-")
	err := ssctx.handleStateSyncMessageWithError(netmsg)
	if err != nil {
		ssctx.sw.Logger.Debug("handling statesync message failed", "error", err.Error())
	}
}

func (ssctx *StateSyncContext) handleStateSyncMessageWithError(netmsg P2PMessage) error {
	fmt.Println("-handleStateSyncMessageWithError-", string(netmsg.Message))
	var msgwrap WrapMsg
	err := json.Unmarshal(netmsg.Message, &msgwrap)
	if err != nil {
		return err
	}
	fmt.Println("-handleStateSyncMessageWithError msgwrap-", msgwrap)

	ssctx.onReceive(msgwrap.ChannelID, msgwrap.Msg)
	fmt.Println("-handleStateSyncMessageWithError onReceive END-")

	// var msg proto.Message
	// err = proto.Unmarshal(msgwrap.Msg, msg)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("-handleStateSyncMessageWithError msg-", msg)
	// peermultiaddr := netmsg.Sender.Ip
	// peeraddr, err := multiaddr.NewMultiaddr(peermultiaddr)
	// if err != nil {
	// 	return err
	// }
	// peerInfo, err := peerstore.AddrInfoFromP2pAddr(peeraddr)
	// if err != nil {
	// 	return err
	// }
	// stream, found := ssctx.p2pctx.GetPeer(ssctx.protocolId, peermultiaddr)
	// if found {
	// 	stream.Close()
	// 	ssctx.p2pctx.DeletePeer(ssctx.protocolId, peermultiaddr)
	// 	ssctx.sw.Logger.Debug("p2p disconnect from peer", "protocolID", ssctx.protocolId, "peer", peermultiaddr)
	// }

	// peer := NewPeer(peermultiaddr, stream, peerInfo, ssctx.protocolId, ssctx.p2pctx)

	// e := cmtp2p.Envelope{
	// 	Src:       peer,
	// 	Message:   msg,
	// 	ChannelID: msgwrap.ChannelID,
	// }
	// ssctx.stateSyncReactor.Receive(e)
	return nil
}
