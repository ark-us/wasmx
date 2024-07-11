package keeper

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

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

	// "github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdkconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"

	// runapp "github.com/cosmos/cosmos-sdk/runtime"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection

	abci "github.com/cometbft/cometbft/abci/types"
	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/bytes"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cometjsonserver "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	cmttypes "github.com/cometbft/cometbft/types"

	mcfg "mythos/v1/config"
	"mythos/v1/server/config"
	networkserver "mythos/v1/x/network/server"
	"mythos/v1/x/network/types"
)

var NETWORK_GAS_LIMIT = uint64(1000000000)

// Config is an gRPC server configuration.
type Config struct {
	MaxOpenConnections int
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
	genesisDocProvider mcfg.GenesisDocProvider,
	metricsProvider node.MetricsProvider,
	client *ABCIClient,
) (*grpc.Server, error) {
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
		return nil, fmt.Errorf("failed to register grpc server: %w", err)
	}
	clientCtx = clientCtx.WithClient(client)

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
		return nil, fmt.Errorf("failed to register reflection service: %w", err)
	}

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	return grpcSrv, nil
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
) (MsgServerInternal, error) {
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

	handler := NewMsgServerImpl(mythosapp.GetNetworkKeeper().(*Keeper))

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

func loadGenDoc(genesisDocProvider mcfg.GenesisDocProvider, chainId string) (*cmttypes.GenesisDoc, error) {
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

func InitChain(
	logger log.Logger,
	cfgAll *config.Config,
	app servertypes.Application,
	privValidator *pvm.FilePV,
	nodeKey *p2p.NodeKey,
	genesisDocProvider mcfg.GenesisDocProvider,
	chainId string,
	networkServer mcfg.NetworkKeeper,
) (*abci.ResponseInitChain, error) {
	consensusLogger := logger.With("subchain_id", chainId)

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
	resInit, res, err := networkserver.InitChainAndCommitBlock(app, req, consensusLogger)
	if err != nil {
		return nil, err
	}
	appHash := res.LastBlockAppHash
	if len(resInit.AppHash) > 0 {
		appHash = resInit.AppHash
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

	ipsMapForChain := ""
	chainips := strings.Split(cfgAll.Network.Ips, ";")
	for _, chainip := range chainips {
		ips := strings.Split(chainip, ":")
		if ips[0] == chainId {
			ipsMapForChain = ips[1]
			break
		}
	}
	peers := strings.Split(ipsMapForChain, ",")

	currentIdStr := "0"
	nodeids := strings.Split(cfgAll.Network.Id, ";")
	for _, nodeid := range nodeids {
		chainIdPair := strings.Split(nodeid, ":")
		if len(chainIdPair) == 1 {
			currentIdStr = chainIdPair[0]
		} else if len(chainIdPair) > 1 && chainIdPair[0] == chainId {
			currentIdStr = chainIdPair[1]
			break
		}
	}
	currentId, err := strconv.Atoi(currentIdStr)
	if err != nil {
		return nil, err
	}

	err = networkserver.InitConsensusContract(mythosapp, consensusLogger, networkServer, appHash, &consensusParams, res.AppVersion, pubKey.Address(), pubKey.Bytes(), privKey.Bytes(), int32(currentId), peers)
	if err != nil {
		return nil, err
	}

	return resInit, nil
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

func GetBaseApp(app servertypes.Application) (mcfg.BaseApp, error) {
	bapp, ok := app.(mcfg.BaseApp)
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

func GetMythosApp(app servertypes.Application) (mcfg.MythosApp, error) {
	mythosapp, ok := app.(mcfg.MythosApp)
	if !ok {
		return nil, fmt.Errorf("failed to get MythosApp from server Application")
	}
	return mythosapp, nil
}
