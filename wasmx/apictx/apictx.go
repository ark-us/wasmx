package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdkserverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/telemetry"

	srvconfig "github.com/loredanacirstea/wasmx/server/config"

	networkgrpc "github.com/loredanacirstea/wasmx/x/network/keeper"

	mcfg "github.com/loredanacirstea/wasmx/config"
	mctx "github.com/loredanacirstea/wasmx/context"
	menc "github.com/loredanacirstea/wasmx/encoding"
)

type APICtx struct {
	GoRoutineGroup         *errgroup.Group
	GoContextParent        context.Context
	SvrCtx                 *server.Context
	ClientCtx              client.Context
	SrvCfg                 srvconfig.Config
	TndCfg                 *cmtcfg.Config
	AppCreator             func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp
	MetricsProvider        node.MetricsProvider
	Metrics                *telemetry.Metrics
	Multiapp               *mcfg.MultiChainApp
	StartGrpcServer        func(ctx context.Context, g *errgroup.Group, config sdkserverconfig.GRPCConfig, clientCtx client.Context, svrCtx *server.Context, app servertypes.Application) (*grpc.Server, client.Context, error)
	StartAPIServer         func(ctx context.Context, g *errgroup.Group, svrCfg sdkserverconfig.Config, clientCtx client.Context, svrCtx *server.Context, app servertypes.Application, grpcSrv *grpc.Server, metrics *telemetry.Metrics) error
	StartNetworkGRPCServer func(
		goCtxParent context.Context,
		g *errgroup.Group,
		app servertypes.Application,
		csvrCtx *server.Context,
		cclientCtx client.Context,
		cmsrvconfig *srvconfig.Config,
		metricsProvider node.MetricsProvider,
		rpcClient client.CometRPC,
	)
	StartJsonRPCServer func(goCtxParent context.Context, g *errgroup.Group, app servertypes.Application, csvrCtx *server.Context, cclientCtx client.Context, cmsrvconfig *srvconfig.Config, ctndcfg *cmtcfg.Config, chainId string, chainCfg *menc.ChainConfig)
	StartWebsrvServer  func(goCtxParent context.Context, g *errgroup.Group, csvrCtx *server.Context, cclientCtx client.Context, cmsrvconfig *srvconfig.Config)
}

func NewAPICtx(
	g *errgroup.Group,
	ctx context.Context,
	svrCtx *server.Context,
	clientCtx client.Context,
	msrvconfig srvconfig.Config,
	tndcfg *cmtcfg.Config,
	appCreator func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp,
	metricsProvider node.MetricsProvider,
	metrics *telemetry.Metrics,
	multiapp *mcfg.MultiChainApp,
	startGrpcServer func(ctx context.Context, g *errgroup.Group, config sdkserverconfig.GRPCConfig, clientCtx client.Context, svrCtx *server.Context, app servertypes.Application) (*grpc.Server, client.Context, error),
	startAPIServer func(ctx context.Context, g *errgroup.Group, svrCfg sdkserverconfig.Config, clientCtx client.Context, svrCtx *server.Context, app servertypes.Application, grpcSrv *grpc.Server, metrics *telemetry.Metrics) error,
	startNetworkGRPCServer func(
		goCtxParent context.Context,
		g *errgroup.Group,
		app servertypes.Application,
		csvrCtx *server.Context,
		cclientCtx client.Context,
		cmsrvconfig *srvconfig.Config,
		metricsProvider node.MetricsProvider,
		rpcClient client.CometRPC,
	),
	startJsonRPCServer func(goCtxParent context.Context, g *errgroup.Group, app servertypes.Application, csvrCtx *server.Context, cclientCtx client.Context, cmsrvconfig *srvconfig.Config, ctndcfg *cmtcfg.Config, chainId string, chainCfg *menc.ChainConfig),
	startWebsrvServer func(goCtxParent context.Context, g *errgroup.Group, csvrCtx *server.Context, cclientCtx client.Context, cmsrvconfig *srvconfig.Config),
) *APICtx {
	return &APICtx{
		GoRoutineGroup:         g,
		GoContextParent:        ctx,
		SvrCtx:                 svrCtx,
		ClientCtx:              clientCtx,
		SrvCfg:                 msrvconfig,
		TndCfg:                 tndcfg,
		AppCreator:             appCreator,
		MetricsProvider:        metricsProvider,
		Metrics:                metrics,
		Multiapp:               multiapp,
		StartGrpcServer:        startGrpcServer,
		StartAPIServer:         startAPIServer,
		StartNetworkGRPCServer: startNetworkGRPCServer,
		StartJsonRPCServer:     startJsonRPCServer,
		StartWebsrvServer:      startWebsrvServer,
	}
}

func (ac *APICtx) SetMultiapp(app *mcfg.MultiChainApp) {
	ac.Multiapp = app
}

func (ac *APICtx) BuildConfigs(
	chainId string,
	chainCfg *menc.ChainConfig,
	ports mctx.NodePorts,
) (mcfg.MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error) {
	cmsrvconfig, ctndcfg, err := cloneConfigs(&ac.SrvCfg, ac.TndCfg)
	if err != nil {
		return nil, nil, ac.ClientCtx, nil, nil, nil, err
	}

	// TODO extract the port base ; define port family range in the toml files
	cmsrvconfig.Config.API.Address = strings.Replace(cmsrvconfig.Config.API.Address, "1317", fmt.Sprintf("%d", ports.CosmosRestApi), 1)
	cmsrvconfig.Config.GRPC.Address = strings.Replace(cmsrvconfig.Config.GRPC.Address, "9090", fmt.Sprintf("%d", ports.CosmosGrpc), 1)
	cmsrvconfig.Network.Address = strings.Replace(cmsrvconfig.Network.Address, "8090", fmt.Sprintf("%d", ports.WasmxNetworkGrpc), 1)
	cmsrvconfig.Websrv.Address = strings.Replace(cmsrvconfig.Websrv.Address, "9999", fmt.Sprintf("%d", ports.WebsrvWebServer), 1)
	cmsrvconfig.JsonRpc.Address = strings.Replace(cmsrvconfig.JsonRpc.Address, "8545", fmt.Sprintf("%d", ports.EvmJsonRpc), 1)
	cmsrvconfig.JsonRpc.WsAddress = strings.Replace(cmsrvconfig.JsonRpc.WsAddress, "8546", fmt.Sprintf("%d", ports.EvmJsonRpcWs), 1)
	ctndcfg.RPC.ListenAddress = strings.Replace(ctndcfg.RPC.ListenAddress, "26657", fmt.Sprintf("%d", ports.TendermintRpc), 1)

	var mythosapp mcfg.MythosApp
	found := false
	iapp, err := ac.Multiapp.GetApp(chainId)
	if err == nil {
		mythosapp, found = iapp.(mcfg.MythosApp)
	}
	if !found {
		mythosapp = ac.AppCreator(chainId, chainCfg)
	}
	bapp := mythosapp.GetBaseApp()
	bapp.Logger().Info("starting chain api servers and clients", "chain_id", chainId)

	cclientCtx := ac.ClientCtx.WithChainID(chainId)

	csvrCtx := &server.Context{
		Viper:  ac.SvrCtx.Viper,
		Logger: ac.SvrCtx.Logger.With("chain_id", chainId),
		Config: ctndcfg,
	}

	rpcClient := networkgrpc.NewABCIClient(mythosapp, bapp, csvrCtx.Logger, mythosapp.GetNetworkKeeper(), csvrCtx.Config, cmsrvconfig, mythosapp.GetActionExecutor().(*networkgrpc.ActionExecutor))
	cclientCtx = cclientCtx.WithClient(rpcClient)

	mythosapp.SetServerConfig(cmsrvconfig)
	mythosapp.SetTendermintConfig(ctndcfg)
	mythosapp.SetRpcClient(rpcClient)
	mythosapp.NonDeterministicSetNodePorts(ports)

	return mythosapp, csvrCtx, cclientCtx, cmsrvconfig, ctndcfg, rpcClient, nil
}

func (ac *APICtx) StartChainApis(
	chainId string,
	chainCfg *menc.ChainConfig,
	ports mctx.NodePorts,
) (mcfg.MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error) {
	mythosapp, csvrCtx, cclientCtx, cmsrvconfig, ctndcfg, rpcClient, err := ac.BuildConfigs(chainId, chainCfg, ports)
	if err != nil {
		return nil, nil, ac.ClientCtx, nil, nil, nil, err
	}
	app := mythosapp.(servertypes.Application)

	// Add the tx service to the gRPC router. We only need to register this
	// service if API or gRPC or JSONRPC is enabled, and avoid doing so in the general
	// case, because it spawns a new local tendermint RPC client.
	// if cmsrvconfig.API.Enable || cmsrvconfig.GRPC.Enable || cmsrvconfig.Websrv.Enable || cmsrvconfig.JsonRpc.Enable {
	// Re-assign for making the client available below do not use := to avoid
	// shadowing the clientCtx variable.
	mythosapp.GetBaseApp().Logger().Info("registering chain services", "chain_id", chainId)
	app.RegisterTxService(cclientCtx)
	app.RegisterTendermintService(cclientCtx)
	app.RegisterNodeService(cclientCtx, cmsrvconfig.Config)
	// }

	if ac.StartNetworkGRPCServer != nil {
		ac.StartNetworkGRPCServer(
			ac.GoContextParent,
			ac.GoRoutineGroup,
			app,
			csvrCtx,
			cclientCtx,
			cmsrvconfig,
			ac.MetricsProvider,
			rpcClient,
		)
	}

	var grpcSrv *grpc.Server
	if ac.StartGrpcServer != nil {
		grpcSrv, cclientCtx, err = ac.StartGrpcServer(ac.GoContextParent, ac.GoRoutineGroup, cmsrvconfig.Config.GRPC, cclientCtx, csvrCtx, app)
		if err != nil {
			return nil, nil, ac.ClientCtx, nil, nil, nil, err
		}
	}
	if ac.StartAPIServer != nil {
		err = ac.StartAPIServer(ac.GoContextParent, ac.GoRoutineGroup, cmsrvconfig.Config, cclientCtx, csvrCtx, app, grpcSrv, ac.Metrics)
		if err != nil {
			return nil, nil, ac.ClientCtx, nil, nil, nil, err
		}
	}
	if ac.StartJsonRPCServer != nil {
		ac.StartJsonRPCServer(ac.GoContextParent, ac.GoRoutineGroup, app, csvrCtx, cclientCtx, cmsrvconfig, ctndcfg, chainId, chainCfg)
	}
	if ac.StartWebsrvServer != nil {
		ac.StartWebsrvServer(ac.GoContextParent, ac.GoRoutineGroup, csvrCtx, cclientCtx, cmsrvconfig)
	}
	return mythosapp, csvrCtx, cclientCtx, cmsrvconfig, ctndcfg, rpcClient, nil
}

func cloneConfigs(msrvconfig *srvconfig.Config, tndcfg *cmtcfg.Config) (*srvconfig.Config, *cmtcfg.Config, error) {
	newmsrvconfig := &srvconfig.Config{}
	newtndcfg := &cmtcfg.Config{}
	msrvconfigbz, err := json.Marshal(msrvconfig)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(msrvconfigbz, newmsrvconfig)
	if err != nil {
		return nil, nil, err
	}

	tndcfgbz, err := json.Marshal(tndcfg)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(tndcfgbz, newtndcfg)
	if err != nil {
		return nil, nil, err
	}
	return newmsrvconfig, newtndcfg, nil
}
