package keeper

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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

	gws "github.com/gorilla/websocket"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdkconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"

	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection

	abci "github.com/cometbft/cometbft/abci/types"
	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/libs/bytes"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cometjsonserver "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	cmttypes "github.com/cometbft/cometbft/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"

	"mythos/v1/server/config"
	networkconfig "mythos/v1/x/network/server/config"
)

var NETWORK_GAS_LIMIT = uint64(1000000000)

// Config is an gRPC server configuration.
type Config struct {
	MaxOpenConnections int
}

type MythosApp interface {
	GetNetworkKeeper() *Keeper
	GetActionExecutor() *ActionExecutor
	// GetTKey(storeKey string) *storetypes.TransientStoreKey
	// GetMKey(storeKey string) *storetypes.MemoryStoreKey
	GetCLessKey(storeKey string) *storetypes.ConsensuslessStoreKey
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
	privValidator cmttypes.PrivValidator,
	nodeKey *p2p.NodeKey,
	genesisDocProvider node.GenesisDocProvider,
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

	networkServer, err := RegisterGRPCServer(ctx, svrCtx, clientCtx, cfg, app, grpcSrv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register grpc server: %w", err)
	}

	bapp, ok := app.(types.BaseApp)
	if !ok {
		return nil, nil, fmt.Errorf("failed to get BaseApp from server Application")
	}
	mythosapp, ok := bapp.(MythosApp)
	if !ok {
		return nil, nil, fmt.Errorf("failed to get MythosApp from server Application")
	}

	logger := svrCtx.Logger.With("module", "network")
	client := NewABCIClient(bapp, logger, mythosapp.GetNetworkKeeper(), mythosapp.GetActionExecutor())
	clientCtx = clientCtx.WithClient(client)

	// load genesis state
	// Run the InitChain logic
	// setup node ips
	if bapp.LastBlockHeight() == 0 {
		resInit, err := initChain(svrCtx, clientCtx, cfg, app, privValidator, nodeKey, genesisDocProvider, metricsProvider, networkServer)
		if err != nil {
			return nil, nil, err
		}
		fmt.Println("--resInit--", resInit)

		// err = setupNode(bapp, logger, networkServer)
		// if err != nil {
		// 	return nil, nil, err
		// }
	}
	// start the node
	err = startNode(svrCtx.Config, cfg.Network, bapp, logger, networkServer)
	if err != nil {
		return nil, nil, err
	}

	// ctx := sdk.UnwrapSDKContext(goCtx)
	// mythosapp.GetNetworkKeeper().wasmxKeeper.ContractInstance(ctx, contractAddress)

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
	app, ok := sapp.(types.BaseApp)
	if !ok {
		return nil, fmt.Errorf("failed to get BaseApp from server Application")
	}
	mythosapp, ok := app.(MythosApp)
	if !ok {
		return nil, fmt.Errorf("failed to get MythosApp from server Application")
	}

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

		fmt.Println("-----NETWORK REQUEST--height-", height)

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
		Keeper:         mythosapp.GetNetworkKeeper(),
		App:            app,
		ActionExecutor: actionExecutor,
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

func loadGenDoc(genesisDocProvider node.GenesisDocProvider) (*cmttypes.GenesisDoc, error) {
	genDoc, err := genesisDocProvider()
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
	privValidator cmttypes.PrivValidator,
	nodeKey *p2p.NodeKey,
	genesisDocProvider node.GenesisDocProvider,
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

	genDoc, err := loadGenDoc(genesisDocProvider)
	if err != nil {
		return nil, err
	}

	pubKey, err := privValidator.GetPubKey()
	if err != nil {
		return nil, err
	}
	validatorAddress := pubKey.Address()
	validators := make([]*cmttypes.Validator, len(genDoc.Validators))
	for i, val := range genDoc.Validators {
		validators[i] = cmttypes.NewValidator(val.PubKey, val.Power)
	}
	validatorSet := cmttypes.NewValidatorSet(validators)
	nextVals := cmttypes.TM2PB.ValidatorUpdates(validatorSet)
	pbparams := genDoc.ConsensusParams.ToProto()
	fmt.Println("--InitialHeight--", genDoc.InitialHeight)
	bapp, ok := app.(types.BaseApp)
	if !ok {
		return nil, fmt.Errorf("failed to get BaseApp from server Application")
	}
	fmt.Println("--app.LastBlockHeight()--", bapp.LastBlockHeight())
	req := &abci.RequestInitChain{
		Time:            genDoc.GenesisTime,
		ChainId:         genDoc.ChainID,
		InitialHeight:   genDoc.InitialHeight,
		ConsensusParams: &pbparams,
		Validators:      nextVals,
		AppStateBytes:   genDoc.AppState,
	}
	fmt.Println("--InitChain--", req)
	resInit, err := app.InitChain(req)
	if err != nil {
		return nil, err
	}
	fmt.Println("--resInit--", resInit)

	scfg := svrCtx.Config

	fmt.Println("--scfg.PrivValidatorListenAddr--", scfg.PrivValidatorListenAddr)
	fmt.Println("--scfg--", scfg)
	fmt.Println("--scfg.P2P.Seeds--", scfg.P2P.Seeds)
	fmt.Println("--scfg.P2P.ExternalAddress--", scfg.P2P.ExternalAddress)
	fmt.Println("--scfg.P2P.PersistentPeers--", scfg.P2P.PersistentPeers)
	fmt.Println("--app.LastBlockHeight()--", bapp.LastBlockHeight())

	freq := &abci.RequestFinalizeBlock{
		Height: req.InitialHeight,
		Time:   req.Time,
	}
	resFinalize, err := app.FinalizeBlock(freq)
	fmt.Println("--resFinalize--", resFinalize, err)

	resCommit, err := app.Commit()
	fmt.Println("--resCommit--", resCommit, err)
	if err != nil {
		return nil, err
	}

	fmt.Println("--app.LastBlockHeight()--", bapp.LastBlockHeight())

	// If the app returned validators or consensus params, update the state.

	appHash = resInit.AppHash
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
	vals, err := cmttypes.PB2TM.ValidatorUpdates(validatorsUpdates)
	if err != nil {
		return nil, err
	}
	consensusParams := *genDoc.ConsensusParams
	if resInit.ConsensusParams != nil {
		consensusParams = consensusParams.Update(resInit.ConsensusParams)
	}

	version := types.Version{
		Software: "",
		Consensus: types.Consensus{
			App:   res.AppVersion,
			Block: 0,
		},
	}
	// TODO ?
	// version.Consensus.App = consensusParams.Version.App

	initChainSetup := &types.InitChainSetup{
		ChainID:         bapp.ChainID(),
		ConsensusParams: &consensusParams,
		AppHash:         appHash,
		Validators:      vals,
		// We update the last results hash with the empty hash, to conform with RFC-6962.
		LastResultsHash:  merkle.HashFromByteSlices(nil),
		CurrentValidator: validatorAddress,
		Version:          version,
	}

	// TODO check if app block height is same as network block height
	// https://github.com/cometbft/cometbft/blob/9cccc8c463f204b210b2a290c2066445188dc681/internal/consensus/replay.go#L360

	// setup the raft machine
	err = setupNode(scfg, cfgAll.Network, bapp, consensusLogger, networkServer, initChainSetup)
	if err != nil {
		return nil, err
	}

	fmt.Println("--app.LastBlockHeight()--", bapp.LastBlockHeight())

	// freq = &abci.RequestFinalizeBlock{
	// 	Height: bapp.LastBlockHeight(),
	// 	Time:   req.Time,
	// }
	// resFinalize, err = app.FinalizeBlock(freq)
	// fmt.Println("--resFinalize--", resFinalize, err)

	// resCommit, err = app.Commit()
	// fmt.Println("--resCommit--", resCommit, err)
	// if err != nil {
	// 	return nil, err
	// }

	return resInit, nil
}

func startNode(scfg *cmtconfig.Config, netcfg networkconfig.NetworkConfig, bapp types.BaseApp, logger log.Logger, networkServer *msgServer) error {
	sdkCtx, commitCacheCtx, ctxcachems, err := CreateQueryContext(bapp, logger, bapp.LastBlockHeight(), false)
	if err != nil {
		return err
	}

	msg := []byte(`{"run":{"event": {"type": "start", "params": []}}}`)
	rresp, err := networkServer.ExecuteContract(sdkCtx, &types.MsgExecuteContract{
		Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Msg:      msg,
	})
	if err != nil {
		return err
	}

	err = commitCtx(bapp, sdkCtx, commitCacheCtx, ctxcachems)
	if err != nil {
		return err
	}
	return nil
}

func setupNode(scfg *cmtconfig.Config, netcfg networkconfig.NetworkConfig, bapp types.BaseApp, logger log.Logger, networkServer *msgServer, initChainSetup *types.InitChainSetup) error {

	// TODO do I need to replace CreateQueryContext with ActionExecutor?
	sdkCtx, commitCacheCtx, ctxcachems, err := CreateQueryContext(bapp, logger, bapp.LastBlockHeight(), false)
	if err != nil {
		return err
	}

	// TODO ips!

	initbz, err := json.Marshal(initChainSetup)
	if err != nil {
		return err
	}
	initData := base64.StdEncoding.EncodeToString(initbz)

	// peers are joined by comma
	nodeIps := strings.Split(netcfg.Ips, ",")
	peersbz, err := json.Marshal(nodeIps)
	if err != nil {
		return err
	}

	peers := string(peersbz)
	peers = strings.Replace(peers, `"`, `\"`, -1)
	fmt.Println("-nodeIPs peers-", peers)

	// TODO node IPS!!!
	msg := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"%d"},{"key":"nodeIPs","value":"%s"},{"key":"initChainSetup","value":"%s"}]}}}`, netcfg.Id, peers, initData))
	rresp, err := networkServer.ExecuteContract(sdkCtx, &types.MsgExecuteContract{
		Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Msg:      msg,
	})
	if err != nil {
		return err
	}
	if netcfg.Leader {
		// make node a candidate
		msg = []byte(`{"run":{"event": {"type": "change", "params": []}}}`)
		rresp, err = networkServer.ExecuteContract(sdkCtx, &types.MsgExecuteContract{
			Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Msg:      msg,
		})
		fmt.Println("--ExecuteContract candidate--", rresp, err)
		if err != nil {
			return err
		}

		// make node a leader
		msg = []byte(`{"run":{"event": {"type": "change", "params": []}}}`)
		rresp, err = networkServer.ExecuteContract(sdkCtx, &types.MsgExecuteContract{
			Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Msg:      msg,
		})
		if err != nil {
			return err
		}
	}

	err = commitCtx(bapp, sdkCtx, commitCacheCtx, ctxcachems)
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

	env := Environment{app: app, networkWrap: networkWrap}
	routes := env.GetRoutes()
	wm := WebsocketManager{logger: logger}
	// rpcLogger := logger.With("module", "rpc-server")
	rpcLogger := servercmtlog.CometLoggerWrapper{Logger: logger.With("module", "rpc-server")}
	mux := http.NewServeMux()
	mux.HandleFunc("/websocket", wm.WebsocketHandler)
	cometjsonserver.RegisterRPCFuncs(mux, routes, rpcLogger)

	// mux.HandleFunc("/health", makeHTTPHandler(rpcFunc, logger))

	var rootHandler http.Handler = mux
	// if rpcConfig.IsCorsEnabled() {
	// 	rootHandler = addCORSHandler(rpcConfig, mux)
	// }
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
	// errCh := make(chan error)

	go func() {
		if err := cometjsonserver.Serve(
			listener,
			rootHandler,
			rpcLogger,
			config,
		); err != nil {
			logger.Error("Error serving RPC network server", "err", err)
			// errCh <- err
		}
	}()

	// select {
	// case <-ctx.Done():
	// 	// The calling process canceled or closed the provided context, so we must
	// 	// gracefully stop the GRPC server.
	// 	logger.Info("stopping network GRPC server...", "address", GRPCAddr)

	// 	return grpcServer, rpcClient, httpSrvDone, nil
	// case err := <-errCh:
	// 	svrCtx.Logger.Error("failed to boot network GRPC server", "error", err.Error())
	// 	return nil, nil, nil, err
	// }

	return nil
}

// func RegisterRPCFuncs(mux *http.ServeMux, funcMap map[string]*RPCFunc, logger log.Logger) {
// 	// HTTP endpoints
// 	for funcName, rpcFunc := range funcMap {
// 		mux.HandleFunc("/"+funcName, makeHTTPHandler(rpcFunc, logger))
// 	}

// 	// JSONRPC endpoints
// 	mux.HandleFunc("/", handleInvalidJSONRPCPaths(makeJSONRPCHandler(funcMap, logger)))
// }

func health(w http.ResponseWriter, r *http.Request) {

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

type ABCIClient struct {
	bapp           types.BaseApp
	nk             types.WasmxWrapper
	logger         log.Logger
	actionExecutor *ActionExecutor
}

func NewABCIClient(bapp types.BaseApp, logger log.Logger, networkKeeper types.WasmxWrapper, actionExecutor *ActionExecutor) *ABCIClient {
	return &ABCIClient{
		bapp:           bapp,
		nk:             networkKeeper,
		logger:         logger,
		actionExecutor: actionExecutor,
	}
}

func (c *ABCIClient) ABCIInfo(context.Context) (*rpctypes.ResultABCIInfo, error) {
	fmt.Println("-network-ABCIInfo--")
	return nil, nil
}

func (c *ABCIClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*rpctypes.ResultABCIQuery, error) {
	fmt.Println("-network-ABCIQuery--")
	return nil, nil
}

func (c *ABCIClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts rpcclient.ABCIQueryOptions) (*rpctypes.ResultABCIQuery, error) {
	fmt.Println("-network-ABCIQueryWithOptions--")
	return nil, nil
}

// func (c *ABCIClient) Simulate(_ context.Context, tx cmttypes.SimulateRequest) (*rpctypes.SimulateResponse, error) {
// 	fmt.Println("--BroadcastTxCommit--")
// 	return nil, nil
// }

func (c *ABCIClient) BroadcastTxCommit(_ context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("-network-BroadcastTxCommit--")
	return nil, nil
}

func (c *ABCIClient) BroadcastTxAsync(_ context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error) {
	fmt.Println("-network-BroadcastTxAsync--")
	// TODO use ctx from params?

	cb := func(goctx context.Context) (any, error) {
		msg := []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(tx)))
		rresp, err := c.nk.ExecuteContract(sdk.UnwrapSDKContext(goctx), &types.MsgExecuteContract{
			Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Msg:      msg,
		})
		fmt.Println("-network-ExecuteContract BroadcastTxAsync--", rresp, err)
		if err != nil {
			return nil, err
		}
		return rresp, nil
	}
	resp, err := c.actionExecutor.Execute(context.Background(), c.bapp.LastBlockHeight(), cb)
	// TODO handle resp, err ?


	return &rpctypes.ResultBroadcastTx{Hash: tx.Hash()}, nil
}

func (c *ABCIClient) BroadcastTxSync(ctx context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error) {
	fmt.Println("-network-BroadcastTxSync--")

	return c.BroadcastTxAsync(ctx, tx)

	// TODO fixme

	// resCh := make(chan *abci.ResponseCheckTx, 1)
	// err := env.Mempool.CheckTx(tx, func(res *abci.ResponseCheckTx) {
	// 	select {
	// 	case <-ctx.Context().Done():
	// 	case resCh <- res:
	// 	}
	// }, mempl.TxInfo{})
	// if err != nil {
	// 	return nil, err
	// }

	// select {
	// case <-ctx.Context().Done():
	// 	return nil, fmt.Errorf("broadcast confirmation not received: %w", ctx.Context().Err())
	// case res := <-resCh:
	// 	return &ctypes.ResultBroadcastTx{
	// 		Code:      res.Code,
	// 		Data:      res.Data,
	// 		Log:       res.Log,
	// 		Codespace: res.Codespace,
	// 		Hash:      tx.Hash(),
	// 	}, nil
	// }
}

func (c *ABCIClient) Validators(ctx context.Context, height *int64, page, perPage *int) (*rpctypes.ResultValidators, error) {
	fmt.Println("-network-Validators--")
	return nil, nil
}

func (c *ABCIClient) Status(context.Context) (*rpctypes.ResultStatus, error) {
	fmt.Println("-network-Status--")
	return nil, nil
}

func (c *ABCIClient) Block(ctx context.Context, height *int64) (*rpctypes.ResultBlock, error) {
	fmt.Println("-network-Block--")

	// get indexed tx
	key := types.GetBlockKey(*height)
	resp, err := c.fsmQuery(key)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("block (%d) not found", height)
	}

	var logEntry types.LogEntry
	err = json.Unmarshal(resp.Data, &logEntry)
	if err != nil {
		return nil, err
	}

	var b abci.RequestProcessProposal
	err = json.Unmarshal(logEntry.Data, &b)
	if err != nil {
		return nil, err
	}

	// TODO fixme
	blockId := cmttypes.BlockID{
		Hash:          make([]byte, 0),
		PartSetHeader: cmttypes.PartSetHeader{},
	}

	lastCommit := cmttypes.Commit{
		Height:     *height,
		Round:      b.ProposedLastCommit.Round,
		BlockID:    blockId,
		Signatures: make([]cmttypes.CommitSig, 0),
	}

	evidence := make([]cmttypes.Evidence, 0)

	txs := make([]cmttypes.Tx, len(b.Txs))
	for i, tx := range b.Txs {
		txs[i] = cmttypes.Tx(tx)
	}
	result := rpctypes.ResultBlock{
		BlockID: blockId,
		Block:   cmttypes.MakeBlock(b.Height, txs, &lastCommit, evidence),
	}

	return &result, nil
}

func (c *ABCIClient) BlockByHash(ctx context.Context, hash []byte) (*rpctypes.ResultBlock, error) {
	fmt.Println("-network-BlockByHash--")
	return nil, nil
}

func (c *ABCIClient) BlockResults(ctx context.Context, height *int64) (*rpctypes.ResultBlockResults, error) {
	fmt.Println("-network-BlockResults--")
	return nil, nil
}

func (c *ABCIClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*rpctypes.ResultBlockchainInfo, error) {
	fmt.Println("-network-BlockchainInfo--")
	return nil, nil
}

func (c *ABCIClient) Commit(ctx context.Context, height *int64) (*rpctypes.ResultCommit, error) {
	fmt.Println("-network-Commit--")
	return nil, nil
}

func (c *ABCIClient) Tx(ctx context.Context, hash []byte, prove bool) (*rpctypes.ResultTx, error) {
	fmt.Println("-network-Tx--")

	// get indexed tx
	key := types.GetTxKey(hash)
	resp, err := c.fsmQuery(key)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("tx (%X) not found", hash)
	}

	var indexedTx types.IndexedTransaction
	err = json.Unmarshal(resp.Data, &indexedTx)
	if err != nil {
		return nil, err
	}

	// get block data
	key = types.GetBlockKey(indexedTx.Height)
	resp, err = c.fsmQuery(key)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("tx block (%d) not found", indexedTx.Height)
	}

	var logEntry types.LogEntry
	err = json.Unmarshal(resp.Data, &logEntry)
	if err != nil {
		return nil, err
	}

	var blockData abci.RequestProcessProposal
	err = json.Unmarshal(logEntry.Data, &blockData)
	if err != nil {
		return nil, err
	}

	var blockResultData abci.ResponseFinalizeBlock
	err = json.Unmarshal(logEntry.Result, &blockResultData)
	if err != nil {
		return nil, err
	}

	if len(blockResultData.TxResults) < int(indexedTx.Index) {
		return nil, fmt.Errorf("tx index (%d) not found in block %d", indexedTx.Index, indexedTx.Height)
	}

	var proof cmttypes.TxProof
	// TODO proof

	// this is a query, we do not commit anything
	return &rpctypes.ResultTx{
		Hash:     hash,
		Height:   indexedTx.Height,
		Index:    indexedTx.Index,
		TxResult: *blockResultData.TxResults[indexedTx.Index],
		Tx:       blockData.Txs[indexedTx.Index],
		Proof:    proof,
	}, nil
}

func (c *ABCIClient) TxSearch(
	ctx context.Context,
	query string,
	prove bool,
	page, perPage *int,
	orderBy string,
) (*rpctypes.ResultTxSearch, error) {
	fmt.Println("-network-TxSearch--")
	return nil, nil
}

func (c *ABCIClient) BlockSearch(
	ctx context.Context,
	query string,
	page, perPage *int,
	orderBy string,
) (*rpctypes.ResultBlockSearch, error) {
	fmt.Println("-network-BlockSearch--")
	return nil, nil
}

func (c *ABCIClient) fsmQuery(key string) (*wasmxtypes.ContractResponse, error) {
	cb := func(goctx context.Context) (any, error) {
		msg := []byte(fmt.Sprintf(`{"getContextValue":{"key":"%s"}}`, key))
		return c.nk.QueryContract(sdk.UnwrapSDKContext(goctx), &types.MsgQueryContract{
			Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
			Msg:      msg,
		})
	}
	qresp, err := c.actionExecutor.Execute(context.Background(), c.bapp.LastBlockHeight(), cb)
	if err != nil {
		return nil, err
	}
	rresp := qresp.(*types.MsgQueryContractResponse)

	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(rresp.Data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
