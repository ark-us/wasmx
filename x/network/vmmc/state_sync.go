package vmmc

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	cmtcfg "github.com/cometbft/cometbft/config"
	pvm "github.com/cometbft/cometbft/privval"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"

	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	vmp2p "mythos/v1/x/network/vmp2p"
)

func StartStateSyncWithChainId(ctx *Context, req StateSyncRequestMsg) error {
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return err
	}
	apictx := multichainapp.APICtx

	mythosapp, _, _, _, ctndcfg, rpcClient, err := apictx.BuildConfigs(req.ChainId, &req.ChainConfig, req.NodePorts)
	if err != nil {
		return err
	}

	mythosapp.NonDeterministicSetNodePortsInitial(req.InitialPorts)

	privValidator := pvm.LoadOrGenFilePV(ctndcfg.PrivValidatorKeyFile(), ctndcfg.PrivValidatorStateFile())

	ctndcfg.StateSync.Enable = req.StatesyncConfig.Enable
	ctndcfg.StateSync.TempDir = req.StatesyncConfig.TempDir
	ctndcfg.StateSync.ChunkFetchers = req.StatesyncConfig.ChunkFetchers
	ctndcfg.StateSync.RPCServers = req.StatesyncConfig.RpcServers
	ctndcfg.StateSync.TrustHeight = req.StatesyncConfig.TrustHeight
	ctndcfg.StateSync.TrustHash = req.StatesyncConfig.TrustHash
	ctndcfg.StateSync.TrustPeriod = time.Millisecond * time.Duration(req.StatesyncConfig.TrustPeriod)
	ctndcfg.StateSync.DiscoveryTime = time.Millisecond * time.Duration(req.StatesyncConfig.DiscoveryTime)
	ctndcfg.StateSync.ChunkRequestTimeout = time.Millisecond * time.Duration(req.StatesyncConfig.ChunkRequestTimeout)

	return vmp2p.StartStateSyncWithChainId(
		ctx.GoContextParent,
		ctx.GoRoutineGroup,
		ctx.Logger(ctx.GetContext()),
		ctndcfg,
		req.ChainId,
		req.ChainConfig,
		mythosapp,
		rpcClient,
		req.ProtocolId,
		req.PeerAddress,
		privValidator.Key.PrivKey.Bytes(),
		fmt.Sprintf(`%d`, req.NodePorts.WasmxNetworkP2P),
	)
}

func InitializeStateSyncProvider(
	goContextParent context.Context,
	goRoutineGroup *errgroup.Group,
	sdklogger log.Logger,
	ctndcfg *cmtcfg.Config,
	chainId string,
	chainCfg menc.ChainConfig,
	app mcfg.MythosApp,
	rpcClient client.CometRPC,
	protocolId string,
	port string,
) {
	privValidator := pvm.LoadOrGenFilePV(ctndcfg.PrivValidatorKeyFile(), ctndcfg.PrivValidatorStateFile())

	go func() {
		err := vmp2p.InitializeStateSyncProvider(goContextParent, goRoutineGroup, sdklogger, ctndcfg, chainId, chainCfg, app, rpcClient, mcfg.GetStateSyncProtocolId(chainId), privValidator.Key.PrivKey.Bytes(), port)
		if err != nil {
			sdklogger.Error("InitializeStateSyncProvider", "error", err.Error())
		}
	}()
}
