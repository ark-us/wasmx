package keeper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/cors"

	gws "github.com/gorilla/websocket"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	// "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdkconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"

	address "cosmossdk.io/core/address"
	// runapp "github.com/cosmos/cosmos-sdk/runtime"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection

	abci "github.com/cometbft/cometbft/abci/types"
	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/libs/bytes"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cometjsonserver "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	cmttypes "github.com/cometbft/cometbft/types"

	mconfig "mythos/v1/config"
	"mythos/v1/server/config"
	networkconfig "mythos/v1/x/network/server/config"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var NETWORK_GAS_LIMIT = uint64(1000000000)

// Config is an gRPC server configuration.
type Config struct {
	MaxOpenConnections int
}

type MythosApp interface {
	GetNetworkKeeper() *Keeper
	GetActionExecutor() *ActionExecutor
	GetGoContextParent() context.Context
	GetGoRoutineGroup() *errgroup.Group
	GetMultiChainApp() (*mconfig.MultiChainApp, error)
	GetBaseApp() *baseapp.BaseApp
	GetCLessKey(storeKey string) *storetypes.ConsensuslessStoreKey
	AddressCodec() address.Codec

	// baseapp
	Query(context.Context, *abci.RequestQuery) (*abci.ResponseQuery, error)
	GRPCQueryRouter() *baseapp.GRPCQueryRouter
	MsgServiceRouter() *baseapp.MsgServiceRouter
}

type ABCIClientI interface {
	// Reading from abci app
	ABCIInfo(context.Context) (*rpctypes.ResultABCIInfo, error)
	ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*rpctypes.ResultABCIQuery, error)
	ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes,
		opts rpcclient.ABCIQueryOptions) (*rpctypes.ResultABCIQuery, error)

	// Writing to abci app
	BroadcastTxCommit(context.Context, cmttypes.Tx) (*rpctypes.ResultBroadcastTxCommit, error)
	BroadcastTxAsync(context.Context, cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error)
	BroadcastTxSync(context.Context, cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error)
}

// NewGRPCServer returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server. See StartGRPCServer.
func NewGRPCServer(
	ctx context.Context,
	svrCtx *server.Context,
	clientCtx client.Context,
	cfg *config.Config,
	app servertypes.Application,
	privValidator *pvm.FilePV,
	nodeKey *p2p.NodeKey,
	genesisDocProvider types.GenesisDocProvider,
	metricsProvider node.MetricsProvider,
) (*grpc.Server, *ABCIClient, error) {
	grpccfg := cfg.GRPC
	maxSendMsgSize := grpccfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = sdkconfig.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := grpccfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = sdkconfig.DefaultGRPCMaxRecvMsgSize
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	_, err := RegisterGRPCServer(ctx, svrCtx, clientCtx, cfg, app, grpcSrv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register grpc server: %w", err)
	}
	bapp, err := GetBaseApp(app)
	if err != nil {
		return nil, nil, err
	}
	mythosapp, err := GetMythosApp(app)
	if err != nil {
		return nil, nil, err
	}
	// networkServer := &msgServer{
	// 	Keeper: mythosapp.GetNetworkKeeper(),
	// 	// App:    bapp,
	// }

	logger := svrCtx.Logger.With("module", "network")
	client := NewABCIClient(mythosapp, bapp, logger, mythosapp.GetNetworkKeeper(), svrCtx.Config, cfg, mythosapp.GetActionExecutor())
	clientCtx = clientCtx.WithClient(client)

	multiapp, err := mythosapp.GetMultiChainApp()
	if err != nil {
		return nil, nil, err
	}

	// init all chains first and start them afterwards
	// InitChain runs multiple contract executions that are not under ActionExecutor control; starting the chains while InitChain is not finished will start delayed executions that will intersect with InitChain executions
	for _, chainId := range multiapp.ChainIds {
		iapp, _ := multiapp.GetApp(chainId)
		app, ok := iapp.(MythosApp)
		if !ok {
			return nil, nil, fmt.Errorf("cannot get MythosApp")
		}
		bapp := app.GetBaseApp()
		sapp, ok := app.(servertypes.Application)
		if !ok {
			return nil, nil, fmt.Errorf("cannot get servertypes.Application")
		}

		if bapp.LastBlockHeight() == 0 {
			networkServer := &msgServer{
				Keeper: app.GetNetworkKeeper(),
			}

			_, err := initChain(svrCtx, clientCtx, cfg, sapp, privValidator, nodeKey, genesisDocProvider, chainId, metricsProvider, networkServer)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	for _, chainId := range multiapp.ChainIds {
		iapp, _ := multiapp.GetApp(chainId)
		app, ok := iapp.(MythosApp)
		if !ok {
			return nil, nil, fmt.Errorf("cannot get MythosApp")
		}
		networkServer := &msgServer{
			Keeper: app.GetNetworkKeeper(),
		}

		// start the node
		err = StartNode(app, logger, networkServer)
		if err != nil {
			return nil, nil, err
		}
	}

	// TODO fix config - should I register multiple servers?

	// Reflection allows consumers to build dynamic clients that can write to any
	// Cosmos SDK application without relying on application packages at compile
	// time.
	err = reflection.Register(grpcSrv, reflection.Config{
		SigningModes: func() map[string]int32 {
			supportedModes := clientCtx.TxConfig.SignModeHandler().SupportedModes()
			modes := make(map[string]int32, len(supportedModes))
			for _, m := range supportedModes {
				modes[m.String()] = (int32)(m)
			}

			return modes
		}(),
		ChainID:           clientCtx.ChainID,
		SdkConfig:         sdk.GetConfig(),
		InterfaceRegistry: clientCtx.InterfaceRegistry,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register reflection service: %w", err)
	}

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	return grpcSrv, client, nil
}

// StartGRPCClient dials the gRPC server using protoAddr and returns a new
// BroadcastAPIClient.
func StartGRPCClient(protoAddr string) types.MsgClient {
	// TODO time limit grpc request
	// ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	// conn, err := grpc.DialContext(ctx, protoAddr, grpc.WithInsecure(), grpc.WithContextDialer(dialerFunc))
	conn, err := grpc.Dial(protoAddr, grpc.WithInsecure(), grpc.WithContextDialer(dialerFunc))
	if err != nil {
		panic(err)
	}
	return types.NewMsgClient(conn)
}

func dialerFunc(_ context.Context, addr string) (net.Conn, error) {
	return cmtnet.Connect(addr)
}

// var tstoreprefix = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40}
// var bzkey = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

// RegisterGRPCServer registers gRPC services directly with the gRPC server.
func RegisterGRPCServer(
	ctx context.Context,
	svrCtx *server.Context,
	clientCtx client.Context,
	cfg *config.Config,
	sapp servertypes.Application,
	server *grpc.Server,
) (*msgServer, error) {
	// app, err := GetBaseApp(sapp)
	// if err != nil {
	// 	return nil, err
	// }
	mythosapp, err := GetMythosApp(sapp)
	if err != nil {
		return nil, err
	}

	// sapp.RegisterGRPCServer(server)
	// return nil, nil

	// TODO introduce chain id in grpc, to make it available for all subchain
	actionExecutor := mythosapp.GetActionExecutor()

	// Define an interceptor for all gRPC queries: this interceptor will create
	// a new sdk.Context, and pass it into the query handler.
	interceptor := func(grpcCtx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// If there's some metadata in the context, retrieve it.
		md, ok := metadata.FromIncomingContext(grpcCtx)
		if !ok {
			return nil, status.Error(codes.Internal, "unable to retrieve metadata")
		}

		// Get height header from the request context, if present.
		var height int64
		if heightHeaders := md.Get(grpctypes.GRPCBlockHeightHeader); len(heightHeaders) == 1 {
			height, err = strconv.ParseInt(heightHeaders[0], 10, 64)
			if err != nil {
				return nil, errorsmod.Wrapf(
					sdkerrors.ErrInvalidRequest,
					"Baseapp.RegisterGRPCServer: invalid height header %q: %v", grpctypes.GRPCBlockHeightHeader, err)
			}
			if err := checkNegativeHeight(height); err != nil {
				return nil, err
			}
		}

		cb := func(goctx context.Context) (any, error) {
			sdkCtx := sdk.UnwrapSDKContext(goctx)
			// Add relevant gRPC headers
			if height == 0 {
				height = sdkCtx.BlockHeight() // If height was not set in the request, set it to the latest
			}

			md = metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))
			if err = grpc.SetHeader(goctx, md); err != nil {
				svrCtx.Logger.Error("failed to set gRPC header", "err", err)
			}
			return handler(goctx, req)
		}
		return actionExecutor.Execute(grpcCtx, height, cb)
	}

	// NewMsgServerImpl
	handler := &msgServer{
		Keeper: mythosapp.GetNetworkKeeper(),
		// App:    app,
	}

	desc := types.Network_Msg_serviceDesc
	newMethods := make([]grpc.MethodDesc, len(desc.Methods))
	for i, method := range desc.Methods {
		methodHandler := method.Handler
		newMethods[i] = grpc.MethodDesc{
			MethodName: method.MethodName,
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
				return methodHandler(srv, ctx, dec, grpcmiddleware.ChainUnaryServer(
					grpcrecovery.UnaryServerInterceptor(),
					interceptor,
				))
			},
		}
	}

	newDesc := &grpc.ServiceDesc{
		ServiceName: desc.ServiceName,
		HandlerType: desc.HandlerType,
		Methods:     newMethods,
		Streams:     desc.Streams,
		Metadata:    desc.Metadata,
	}

	server.RegisterService(newDesc, handler)
	return handler, nil
}

func loadGenDoc(genesisDocProvider types.GenesisDocProvider, chainId string) (*cmttypes.GenesisDoc, error) {
	genDoc, err := genesisDocProvider(chainId)
	if err != nil {
		return nil, err
	}

	err = genDoc.ValidateAndComplete()
	if err != nil {
		return nil, fmt.Errorf("error in genesis doc: %w", err)
	}
	// save genesis doc to prevent a certain class of user errors (e.g. when it
	// was changed, accidentally or not). Also good for audit trail.
	// TODO - save in contract
	// if err := saveGenesisDoc(stateDB, genDoc); err != nil {
	// 	return sm.State{}, nil, err
	// }
	return genDoc, nil
}

func initChain(
	svrCtx *server.Context,
	clientCtx client.Context,
	cfgAll *config.Config,
	app servertypes.Application,
	privValidator *pvm.FilePV,
	nodeKey *p2p.NodeKey,
	genesisDocProvider types.GenesisDocProvider,
	chainId string,
	metricsProvider node.MetricsProvider,
	networkServer *msgServer,
) (*abci.ResponseInitChain, error) {
	consensusLogger := svrCtx.Logger.With("module", "consensus")
	// cfg := svrCtx.Config

	res, err := app.Info(RequestInfo)
	if err != nil {
		return nil, fmt.Errorf("error calling Info: %v", err)
	}

	blockHeight := res.LastBlockHeight
	if blockHeight < 0 {
		return nil, fmt.Errorf("got a negative last block height (%d) from the app", blockHeight)
	}
	appHash := res.LastBlockAppHash

	consensusLogger.Info("ABCI Handshake App Info",
		"height", blockHeight,
		"hash", fmt.Sprintf("%X", appHash),
		"software-version", res.Version,
		"protocol-version", res.AppVersion,
	)

	// check if network contract exists
	genDoc, err := loadGenDoc(genesisDocProvider, chainId)
	if err != nil {
		return nil, err
	}

	validators := make([]*cmttypes.Validator, len(genDoc.Validators))
	for i, val := range genDoc.Validators {
		validators[i] = cmttypes.NewValidator(val.PubKey, val.Power)
	}
	validatorSet := cmttypes.NewValidatorSet(validators)
	nextVals := cmttypes.TM2PB.ValidatorUpdates(validatorSet)
	pbparams := genDoc.ConsensusParams.ToProto()
	mythosapp, err := GetMythosApp(app)
	if err != nil {
		return nil, err
	}

	req := &abci.RequestInitChain{
		Time:            genDoc.GenesisTime,
		ChainId:         genDoc.ChainID,
		InitialHeight:   genDoc.InitialHeight,
		ConsensusParams: &pbparams,
		Validators:      nextVals,
		AppStateBytes:   genDoc.AppState,
	}
	resInit, err := app.InitChain(req)
	if err != nil {
		return nil, err
	}

	// scfg := svrCtx.Config
	freq := &abci.RequestFinalizeBlock{
		Height: req.InitialHeight,
		Time:   req.Time,
	}
	_, err = app.FinalizeBlock(freq)
	if err != nil {
		return nil, err
	}

	_, err = app.Commit()
	if err != nil {
		return nil, err
	}

	// If the app returned validators or consensus params, update the state.
	if len(resInit.AppHash) > 0 {
		appHash = resInit.AppHash
	}
	validatorsUpdates := resInit.Validators
	if len(validatorsUpdates) == 0 {
		validatorsUpdates = req.Validators
	}
	if len(validatorsUpdates) == 0 {
		// If validator set is not set in genesis and still empty after InitChain, exit.
		return nil, fmt.Errorf("validator set is nil in genesis and still empty after InitChain")
	}
	consensusParams := *genDoc.ConsensusParams
	if resInit.ConsensusParams != nil {
		consensusParams = consensusParams.Update(resInit.ConsensusParams)
	}
	pubKey, err := privValidator.GetPubKey()
	if err != nil {
		return nil, err
	}
	privKey := privValidator.Key.PrivKey

	// peers := strings.Split(svrCtx.Config.P2P.PersistentPeers, ",")
	peers := strings.Split(cfgAll.Network.Ips, ",")

	err = InitConsensusContract(mythosapp, consensusLogger, cfgAll.Network, networkServer, appHash, &consensusParams, res.AppVersion, pubKey, privKey, peers)
	if err != nil {
		return nil, err
	}

	return resInit, nil
}

type PrivKey interface {
	Bytes() []byte
}

func InitConsensusContract(
	mythosapp MythosApp,
	consensusLogger log.Logger,
	cfgNetwork networkconfig.NetworkConfig,
	networkServer MsgServerInternal,
	appHash []byte,
	consensusParams *cmttypes.ConsensusParams,
	appVersion uint64,
	pubKey crypto.PubKey,
	privKey PrivKey,
	peers []string,
) error {
	version := types.Version{
		Software: "",
		Consensus: types.Consensus{
			App:   appVersion,
			Block: 0,
		},
	}
	// TODO ?
	// version.Consensus.App = consensusParams.Version.App

	initChainSetup := &types.InitChainSetup{
		ChainID:         mythosapp.GetBaseApp().ChainID(),
		ConsensusParams: consensusParams,
		AppHash:         appHash,
		// We update the last results hash with the empty hash, to conform with RFC-6962.
		LastResultsHash:  merkle.HashFromByteSlices(nil),
		Version:          version,
		ValidatorAddress: pubKey.Address(), // hexbytes
		ValidatorPrivKey: privKey.Bytes(),  // consensus privkey
		ValidatorPubKey:  pubKey.Bytes(),   // consensus pubkey
		Peers:            peers,
	}

	// TODO check if app block height is same as network block height
	// https://github.com/cometbft/cometbft/blob/9cccc8c463f204b210b2a290c2066445188dc681/internal/consensus/replay.go#L360

	// setup the consensus contract
	err := SetupNode(mythosapp, cfgNetwork, consensusLogger, networkServer, initChainSetup)
	if err != nil {
		return err
	}
	return nil
}

func StartNode(mythosapp MythosApp, logger log.Logger, networkServer MsgServerInternal) error {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		msg := []byte(fmt.Sprintf(`{"RunHook":{"hook":"%s","data":""}}`, wasmxtypes.HOOK_START_NODE))
		res, err := networkServer.ExecuteContract(ctx, &types.MsgExecuteContract{
			Sender:   wasmxtypes.ROLE_HOOKS_NONC,
			Contract: wasmxtypes.ROLE_HOOKS_NONC,
			Msg:      msg,
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	actionExecutor := mythosapp.GetActionExecutor()

	_, err := actionExecutor.Execute(mythosapp.GetGoContextParent(), mythosapp.GetBaseApp().LastBlockHeight(), cb)
	if err != nil {
		return err
	}
	return nil
}

func SetupNode(mythosapp MythosApp, netcfg networkconfig.NetworkConfig, logger log.Logger, networkServer MsgServerInternal, initChainSetup *types.InitChainSetup) error {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)

		initbz, err := json.Marshal(initChainSetup)
		if err != nil {
			return nil, err
		}
		initData := base64.StdEncoding.EncodeToString(initbz)

		// TODO node IPS!!!
		msg := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"%d"},{"key":"initChainSetup","value":"%s"}]}}}`, netcfg.Id, initData))
		res, err := networkServer.ExecuteContract(ctx, &types.MsgExecuteContract{
			Sender:   wasmxtypes.ROLE_CONSENSUS,
			Contract: wasmxtypes.ROLE_CONSENSUS,
			Msg:      msg,
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	actionExecutor := mythosapp.GetActionExecutor()
	_, err := actionExecutor.Execute(mythosapp.GetGoContextParent(), mythosapp.GetBaseApp().LastBlockHeight(), cb)
	if err != nil {
		return err
	}
	return nil
}

type WebsocketManager struct {
	gws.Upgrader

	// funcMap       map[string]*RPCFunc
	logger        log.Logger
	wsConnOptions []func(*wsConnection)
}

type wsConnection struct {
	baseConn   *gws.Conn
	remoteAddr string
	// funcMap    map[string]*RPCFunc
}

func StartRPC(svrCtx *server.Context, ctx context.Context, app servertypes.Application, networkWrap *ABCIClient, logger log.Logger, cfg *config.Config) error {
	// listenAddrs := splitAndTrimEmpty(n.config.RPC.ListenAddress, ",", " ")
	listenAddr := svrCtx.Config.RPC.ListenAddress

	env := Environment{app: app, networkWrap: networkWrap, serverConfig: svrCtx.Config, config: cfg}
	routes := env.GetRoutes()
	wm := WebsocketManager{logger: logger}
	rpcLogger := servercmtlog.CometLoggerWrapper{Logger: logger.With("module", "rpc-server")}
	mux := http.NewServeMux()
	mux.HandleFunc("/websocket", wm.WebsocketHandler)
	cometjsonserver.RegisterRPCFuncs(mux, routes, rpcLogger)

	// mux.HandleFunc("/health", makeHTTPHandler(rpcFunc, logger))

	var rootHandler http.Handler = mux
	if svrCtx.Config.RPC.IsCorsEnabled() {
		rootHandler = addCORSHandler(svrCtx.Config.RPC, mux)
	}
	// return rootHandler

	config := cometjsonserver.DefaultConfig()
	// config.MaxBodyBytes = cfg.API.MaxBodyBytes
	// config.MaxHeaderBytes = cfg.API.MaxHeaderBytes
	// config.MaxOpenConnections = cfg.API.MaxOpenConnections

	listener, err := cometjsonserver.Listen(
		listenAddr,
		int(config.MaxOpenConnections),
	)
	if err != nil {
		return err
	}

	// httpSrvDone := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	go func() {
		if err := cometjsonserver.Serve(
			listener,
			rootHandler,
			rpcLogger,
			config,
		); err != nil {
			if err == http.ErrServerClosed {
				svrCtx.Logger.Info("Closing network RPC server", "address", listenAddr, err.Error())
				return
			}
			logger.Error("Error serving RPC network server", "err", err.Error())
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the GRPC server.
		logger.Info("stopping network RPC server...", "address", listenAddr)
		close(errCh)
		return nil
	case err := <-errCh:
		svrCtx.Logger.Error("failed to boot network RPC server", "error", err.Error())
		return err
	}
}

func addCORSHandler(rpcConfig *cmtconfig.RPCConfig, h http.Handler) http.Handler {
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: rpcConfig.CORSAllowedOrigins,
		AllowedMethods: rpcConfig.CORSAllowedMethods,
		AllowedHeaders: rpcConfig.CORSAllowedHeaders,
	})
	h = corsMiddleware.Handler(h)
	return h
}

func (wm *WebsocketManager) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := wm.Upgrade(w, r, nil)
	if err != nil {
		// TODO - return http error
		wm.logger.Error("Failed to upgrade connection", "err", err)
		return
	}
	defer func() {
		if err := wsConn.Close(); err != nil {
			wm.logger.Error("Failed to close connection", "err", err)
		}
	}()

	// register connection
	// con := newWSConnection(wsConn, wm.funcMap, wm.wsConnOptions...)
	con := &wsConnection{
		// funcMap:    wm.funcMap,
		remoteAddr: wsConn.RemoteAddr().String(),
		baseConn:   wsConn,
	}

	// con.SetLogger(wm.logger.With("remote", wsConn.RemoteAddr()))
	wm.logger.Info("New websocket connection", "remote", con.remoteAddr)
	// err = con.Start() // BLOCKING
	// if err != nil {
	// 	wm.logger.Error("Failed to start connection", "err", err)
	// 	return
	// }
	// if err := con.Stop(); err != nil {
	// 	wm.logger.Error("error while stopping connection", "error", err)
	// }
}

func GetBaseApp(app servertypes.Application) (types.BaseApp, error) {
	bapp, ok := app.(types.BaseApp)
	if !ok {
		return nil, fmt.Errorf("failed to get App from server Application")
	}
	// _, ok = app.(*baseapp.BaseApp)
	// if !ok {
	// 	return nil, fmt.Errorf("failed to get App as baseapp pointer from server Application")
	// }
	// _, ok = app.(*runapp.App)
	// if !ok {
	// 	return nil, fmt.Errorf("failed to get App as runapp pointer from server Application")
	// }
	return bapp, nil
}

func GetMythosApp(app servertypes.Application) (MythosApp, error) {
	mythosapp, ok := app.(MythosApp)
	if !ok {
		return nil, fmt.Errorf("failed to get MythosApp from server Application")
	}
	return mythosapp, nil
}
