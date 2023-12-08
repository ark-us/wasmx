package keeper

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var NETWORK_GAS_LIMIT = uint64(1000000000)

// Config is an gRPC server configuration.
type Config struct {
	MaxOpenConnections int
}

type MythosApp interface {
	GetNetworkKeeper() Keeper
	// GetTKey(storeKey string) *storetypes.TransientStoreKey
	// GetMKey(storeKey string) *storetypes.MemoryStoreKey
	GetCLessKey(storeKey string) *storetypes.ConsensuslessStoreKey
}

type BaseApp interface {
	Name() string
	ChainID() string
	CreateQueryContext(height int64, prove bool) (sdk.Context, error)
	CommitMultiStore() storetypes.CommitMultiStore
	GetContextForCheckTx(txBytes []byte) sdk.Context
	GetContextForFinalizeBlock(txBytes []byte) sdk.Context
	NewUncachedContext(isCheckTx bool, header cmtproto.Header) sdk.Context
	LastBlockHeight() int64
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
	cfg config.GRPCConfig,
	app servertypes.Application,
	privValidator cmttypes.PrivValidator,
	nodeKey *p2p.NodeKey,
	genesisDocProvider node.GenesisDocProvider,
	metricsProvider node.MetricsProvider,
) (*grpc.Server, *ABCIClient, error) {
	maxSendMsgSize := cfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = config.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := cfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = config.DefaultGRPCMaxRecvMsgSize
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	networkServer, err := RegisterGRPCServer(ctx, svrCtx, clientCtx, app, grpcSrv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register grpc server: %w", err)
	}

	bapp, ok := app.(BaseApp)
	if !ok {
		return nil, nil, fmt.Errorf("failed to get BaseApp from server Application")
	}
	// mythosapp, ok := bapp.(MythosApp)
	// if !ok {
	// 	return nil, fmt.Errorf("failed to get MythosApp from server Application")
	// }
	logger := svrCtx.Logger.With("module", "network")

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

	client := NewABCIClient(bapp, logger, networkServer)

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
	sapp servertypes.Application,
	server *grpc.Server,
) (*msgServer, error) {
	app, ok := sapp.(BaseApp)
	if !ok {
		return nil, fmt.Errorf("failed to get BaseApp from server Application")
	}
	mythosapp, ok := app.(MythosApp)
	if !ok {
		return nil, fmt.Errorf("failed to get MythosApp from server Application")
	}

	// Define an interceptor for all gRPC queries: this interceptor will create
	// a new sdk.Context, and pass it into the query handler.
	interceptor := func(grpcCtx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		fmt.Println("-----NETWORK REQUEST---")
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

		// TODO - use this for grpc queries
		sdkCtx_, ctxcachems, err := CreateQueryContext(app, svrCtx.Logger, height, false)
		if err != nil {
			return nil, err
		}
		sdkCtx, commitCacheCtx := sdkCtx_.CacheContext()
		// sdkCtx := sdkCtx_

		// Add relevant gRPC headers
		if height == 0 {
			height = sdkCtx.BlockHeight() // If height was not set in the request, set it to the latest
		}

		// Attach the sdk.Context into the gRPC's context.Context.
		grpcCtx = context.WithValue(grpcCtx, sdk.SdkContextKey, sdkCtx)

		md = metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))
		if err = grpc.SetHeader(grpcCtx, md); err != nil {
			svrCtx.Logger.Error("failed to set gRPC header", "err", err)
		}
		fmt.Println("-----NETWORK REQUEST-before handler--")

		hresp, err := handler(grpcCtx, req)
		if err != nil {
			return hresp, err
		}

		commitCacheCtx()

		// commit original context
		origtstore := ctxcachems.GetStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
		origtstore.(storetypes.CacheWrap).Write()

		origtstore2 := ctxcachems.GetKVStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
		origtstore2.CacheWrap().Write()

		cms := app.CommitMultiStore()
		origtstore3 := cms.GetCommitKVStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
		origtstore3.Commit()

		fmt.Println("-----NETWORK REQUEST--COMMITED-")

		return hresp, nil
	}

	handler := &msgServer{
		Keeper: mythosapp.GetNetworkKeeper(),
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

func checkNegativeHeight(height int64) error {
	if height < 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot query with height < 0; please provide a valid height")
	}

	return nil
}

// createQueryContext creates a new sdk.Context for a query, taking as args
// the block height and whether the query needs a proof or not.
func CreateQueryContext(app BaseApp, logger log.Logger, height int64, prove bool) (sdk.Context, storetypes.CacheMultiStore, error) {
	if err := checkNegativeHeight(height); err != nil {
		return sdk.Context{}, nil, err
	}

	cms := app.CommitMultiStore()
	qms := cms.(storetypes.MultiStore)

	lastBlockHeight := qms.LatestVersion()
	if lastBlockHeight == 0 {
		return sdk.Context{}, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidHeight, "%s is not ready; please wait for first block", app.Name())
	}

	if height > lastBlockHeight {
		return sdk.Context{}, nil,
			errorsmod.Wrap(
				sdkerrors.ErrInvalidHeight,
				"cannot query with height in the future; please provide a valid height",
			)
	}

	// when a client did not provide a query height, manually inject the latest
	if height == 0 {
		height = lastBlockHeight
	}

	fmt.Println("-CreateQueryContext-", height)

	if height <= 1 && prove {
		return sdk.Context{}, nil,
			errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			)
	}

	// cacheMS, err := qms.CacheMultiStoreWithVersion(height)
	// if err != nil {
	// 	return sdk.Context{}, nil,CacheMultiStoreWithVersion
	// 		errorsmod.Wrapf(
	// 			sdkerrors.ErrInvalidRequest,
	// 			"failed to load state at height %d; %s (latest height: %d)", height, err, lastBlockHeight,
	// 		)
	// }
	cacheMS := qms.CacheMultiStore()

	// tmpctx, err := app.CreateQueryContext(height, false)
	// if err != nil {
	// 	return sdk.Context{}, nil, err
	// }
	fmt.Println("-----NETWORK REQUEST-before GetContextForFinalizeBlock--")
	// tmpctx := app.GetContextForFinalizeBlock(make([]byte, 0))
	// tmpctx := app.GetContextForCheckTx(make([]byte, 0))

	// header := cmtproto.Header{ChainID: req.ChainId, Time: req.Time}

	// TODO fixme!!!
	header := cmtproto.Header{
		ChainID:            app.ChainID(),
		Height:             10,
		Time:               time.Now().UTC(),
		ProposerAddress:    []byte("proposer"),
		NextValidatorsHash: []byte("proposer"),
		// AppHash:            app.LastCommitID().Hash,
		// Version: tmversion.Consensus{
		// 	Block: version.BlockProtocol,
		// },
		// LastBlockId: tmproto.BlockID{
		// 	Hash: tmhash.Sum([]byte("block_id")),
		// 	PartSetHeader: tmproto.PartSetHeader{
		// 		Total: 11,
		// 		Hash:  tmhash.Sum([]byte("partset_header")),
		// 	},
		// },
		// AppHash:            tmhash.Sum([]byte("app")),
		// DataHash:           tmhash.Sum([]byte("data")),
		// EvidenceHash:       tmhash.Sum([]byte("evidence")),
		// ValidatorsHash:     tmhash.Sum([]byte("validators")),
		// NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		// ConsensusHash:      tmhash.Sum([]byte("consensus")),
		// LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}
	tmpctx := app.NewUncachedContext(false, header)

	// branch the commit multi-store for safety
	ctx := sdk.NewContext(cacheMS, tmpctx.BlockHeader(), true, logger).
		WithMinGasPrices(nil).
		WithBlockHeight(height).
		WithGasMeter(storetypes.NewGasMeter(NETWORK_GAS_LIMIT))

	if height != lastBlockHeight {
		rms, ok := app.CommitMultiStore().(*rootmulti.Store)
		if ok {
			cInfo, err := rms.GetCommitInfo(height)
			if cInfo != nil && err == nil {
				ctx = ctx.WithBlockTime(cInfo.Timestamp)
			}
		}
	}
	return ctx, cacheMS, nil
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
	grpccfg config.GRPCConfig,
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

	// // Only set the version if there is no existing state.
	// if h.initialState.LastBlockHeight == 0 {
	// 	h.initialState.Version.Consensus.App = res.AppVersion
	// }

	// check if network contract exists
	// check block is 0

	genDoc, err := loadGenDoc(genesisDocProvider)
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
	fmt.Println("--InitialHeight--", genDoc.InitialHeight)
	bapp, ok := app.(BaseApp)
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

	// TODO resInit.Validators
	scfg := svrCtx.Config
	// scfg.StateSync.
	fmt.Println("--scfg.PrivValidatorListenAddr--", scfg.PrivValidatorListenAddr)
	fmt.Println("--scfg--", scfg)
	fmt.Println("--scfg.P2P.Seeds--", scfg.P2P.Seeds)
	fmt.Println("--scfg.P2P.ExternalAddress--", scfg.P2P.ExternalAddress)
	fmt.Println("--scfg.P2P.PersistentPeers--", scfg.P2P.PersistentPeers)

	// TODO - valdiators
	// https://github.com/cometbft/cometbft/blob/9cccc8c463f204b210b2a290c2066445188dc681/spec/abci/abci%2B%2B_methods.md#initchain
	// validators := resInit.Validators || req.Validators

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

	// TODO
	// https://github.com/cometbft/cometbft/blob/9cccc8c463f204b210b2a290c2066445188dc681/internal/consensus/replay.go#L326
	// appHash = resInit.AppHash
	// if len(res.AppHash) > 0 {
	// 	state.AppHash = res.AppHash
	// }

	// setup the raft machine

	err = setupNode(bapp, consensusLogger, networkServer)
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

func setupNode(bapp BaseApp, logger log.Logger, networkServer *msgServer) error {
	mythosapp, ok := bapp.(MythosApp)
	if !ok {
		return fmt.Errorf("failed to get MythosApp from server Application")
	}

	// setup network
	// resInit

	// contractAddress := sdk.MustAccAddressFromBech32("mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy")
	// msg :=
	// respp, err := mythosapp.GetNetworkKeeper().ExecuteContract(ctx, contractAddress, contractAddress, msg, nil, nil)

	height := bapp.LastBlockHeight()

	cms := bapp.CommitMultiStore()
	qms := cms.(storetypes.MultiStore)
	ctxcachems := qms.CacheMultiStore()
	// TODO fixme - who should commit first block?
	header := cmtproto.Header{
		ChainID:            bapp.ChainID(),
		Height:             height,
		Time:               time.Now().UTC(),
		ProposerAddress:    []byte("proposer"),
		NextValidatorsHash: []byte("proposer"),
		// AppHash:            bapp.LastCommitID().Hash,
		// Version: tmversion.Consensus{
		// 	Block: version.BlockProtocol,
		// },
		// LastBlockId: tmproto.BlockID{
		// 	Hash: tmhash.Sum([]byte("block_id")),
		// 	PartSetHeader: tmproto.PartSetHeader{
		// 		Total: 11,
		// 		Hash:  tmhash.Sum([]byte("partset_header")),
		// 	},
		// },
		// AppHash:            tmhash.Sum([]byte("app")),
		// DataHash:           tmhash.Sum([]byte("data")),
		// EvidenceHash:       tmhash.Sum([]byte("evidence")),
		// ValidatorsHash:     tmhash.Sum([]byte("validators")),
		// NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		// ConsensusHash:      tmhash.Sum([]byte("consensus")),
		// LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}
	tmpctx := bapp.NewUncachedContext(false, header)

	// branch the commit multi-store for safety
	sdkCtx_ := sdk.NewContext(ctxcachems, tmpctx.BlockHeader(), true, logger).
		WithMinGasPrices(nil).
		WithBlockHeight(height).
		WithGasMeter(storetypes.NewGasMeter(NETWORK_GAS_LIMIT))

	// sdkCtx_, ctxcachems, err := CreateQueryContext(bapp, logger, height, false)
	// fmt.Println("--StartInterval--CreateQueryContext", err)
	// if err != nil {
	// 	logger.Error("failed to create query context", "err", err)
	// 	return nil, err
	// }
	sdkCtx, commitCacheCtx := sdkCtx_.CacheContext()

	goCtx := context.Background()
	goCtx = context.WithValue(goCtx, sdk.SdkContextKey, sdkCtx)

	// TODO ips!

	msg := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"0"},{"key":"nodeIPs","value":"[\"%s\"]"}]}}}`, "localhost:9080"))
	rresp, err := networkServer.ExecuteContract(goCtx, &types.MsgExecuteContract{
		Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Msg:      msg,
	})
	fmt.Println("--ExecuteContract setupNode--", rresp, err)
	if err != nil {
		return err
	}

	// make node a candidate
	msg = []byte(`{"run":{"event": {"type": "change", "params": []}}}`)
	rresp, err = networkServer.ExecuteContract(goCtx, &types.MsgExecuteContract{
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
	rresp, err = networkServer.ExecuteContract(goCtx, &types.MsgExecuteContract{
		Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Msg:      msg,
	})
	fmt.Println("--ExecuteContract leader--", rresp, err)
	if err != nil {
		return err
	}

	commitCacheCtx()
	origtstore := ctxcachems.GetStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
	origtstore.(storetypes.CacheWrap).Write()

	return nil
}

type ABCIClient struct {
	bapp          BaseApp
	networkServer *msgServer
	logger        log.Logger
}

func NewABCIClient(bapp BaseApp, logger log.Logger, networkServer *msgServer) *ABCIClient {
	return &ABCIClient{
		bapp:          bapp,
		networkServer: networkServer,
		logger:        logger,
	}
}

func (c *ABCIClient) ABCIInfo(context.Context) (*rpctypes.ResultABCIInfo, error) {
	fmt.Println("--ABCIInfo--")
	return nil, nil
}

func (c *ABCIClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*rpctypes.ResultABCIQuery, error) {
	fmt.Println("--ABCIQuery--")
	return nil, nil
}

func (c *ABCIClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts rpcclient.ABCIQueryOptions) (*rpctypes.ResultABCIQuery, error) {
	fmt.Println("--ABCIQueryWithOptions--")
	return nil, nil
}

// func (c *ABCIClient) Simulate(_ context.Context, tx cmttypes.SimulateRequest) (*rpctypes.SimulateResponse, error) {
// 	fmt.Println("--BroadcastTxCommit--")
// 	return nil, nil
// }

func (c *ABCIClient) BroadcastTxCommit(_ context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("--BroadcastTxCommit--")
	return nil, nil
}

func (c *ABCIClient) BroadcastTxAsync(_ context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error) {
	fmt.Println("--BroadcastTxAsync--")
	// TODO use ctx from params?

	sdkCtx, commitCacheCtx, ctxcachems, err := c.prepareCtx()

	goCtx := context.Background()
	goCtx = context.WithValue(goCtx, sdk.SdkContextKey, sdkCtx)

	msg := []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(tx)))
	rresp, err := c.networkServer.ExecuteContract(goCtx, &types.MsgExecuteContract{
		Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Msg:      msg,
	})
	fmt.Println("--ExecuteContract BroadcastTxAsync--", rresp, err)
	if err != nil {
		return nil, err
	}

	err = c.commitCtx(sdkCtx, commitCacheCtx, ctxcachems)
	if err != nil {
		return nil, err
	}

	return &rpctypes.ResultBroadcastTx{Hash: tx.Hash()}, nil
}

func (c *ABCIClient) BroadcastTxSync(ctx context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error) {
	fmt.Println("--BroadcastTxSync--")
	// TODO fixme
	return c.BroadcastTxAsync(ctx, tx)
}

func (c *ABCIClient) prepareCtx() (sdk.Context, func(), storetypes.CacheMultiStore, error) {
	fmt.Println("--prepareCtx--")
	height := c.bapp.LastBlockHeight()

	cms := c.bapp.CommitMultiStore()
	qms := cms.(storetypes.MultiStore)
	ctxcachems := qms.CacheMultiStore()
	// TODO fixme - who should commit first block?
	header := cmtproto.Header{
		ChainID:            c.bapp.ChainID(),
		Height:             10,
		Time:               time.Now().UTC(),
		ProposerAddress:    []byte("proposer"),
		NextValidatorsHash: []byte("proposer"),
		// AppHash:            app.LastCommitID().Hash,
		// Version: tmversion.Consensus{
		// 	Block: version.BlockProtocol,
		// },
		// LastBlockId: tmproto.BlockID{
		// 	Hash: tmhash.Sum([]byte("block_id")),
		// 	PartSetHeader: tmproto.PartSetHeader{
		// 		Total: 11,
		// 		Hash:  tmhash.Sum([]byte("partset_header")),
		// 	},
		// },
		// AppHash:            tmhash.Sum([]byte("app")),
		// DataHash:           tmhash.Sum([]byte("data")),
		// EvidenceHash:       tmhash.Sum([]byte("evidence")),
		// ValidatorsHash:     tmhash.Sum([]byte("validators")),
		// NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		// ConsensusHash:      tmhash.Sum([]byte("consensus")),
		// LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}
	tmpctx := c.bapp.NewUncachedContext(false, header)

	// branch the commit multi-store for safety
	sdkCtx_ := sdk.NewContext(ctxcachems, tmpctx.BlockHeader(), true, c.logger).
		WithMinGasPrices(nil).
		WithBlockHeight(height).
		WithGasMeter(storetypes.NewGasMeter(NETWORK_GAS_LIMIT))

	// sdkCtx_, ctxcachems, err := CreateQueryContext(bapp, logger, height, false)
	// fmt.Println("--StartInterval--CreateQueryContext", err)
	// if err != nil {
	// 	logger.Error("failed to create query context", "err", err)
	// 	return nil, err
	// }
	sdkCtx, commitCacheCtx := sdkCtx_.CacheContext()

	return sdkCtx, commitCacheCtx, ctxcachems, nil

}

func (c *ABCIClient) commitCtx(sdkCtx sdk.Context, commitCacheCtx func(), ctxcachems storetypes.CacheMultiStore) error {
	fmt.Println("--commitCtx--")
	mythosapp, ok := c.bapp.(MythosApp)
	if !ok {
		return fmt.Errorf("failed to get MythosApp from server Application")
	}

	commitCacheCtx()
	origtstore := ctxcachems.GetStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
	origtstore.(storetypes.CacheWrap).Write()
	return nil
}

func (c *ABCIClient) Validators(ctx context.Context, height *int64, page, perPage *int) (*rpctypes.ResultValidators, error) {
	fmt.Println("--Validators--")
	return nil, nil
}

func (c *ABCIClient) Status(context.Context) (*rpctypes.ResultStatus, error) {
	fmt.Println("--Status--")
	return nil, nil
}

func (c *ABCIClient) Block(ctx context.Context, height *int64) (*rpctypes.ResultBlock, error) {
	fmt.Println("--Block--")
	return nil, nil
}

func (c *ABCIClient) BlockByHash(ctx context.Context, hash []byte) (*rpctypes.ResultBlock, error) {
	fmt.Println("--BlockByHash--")
	return nil, nil
}

func (c *ABCIClient) BlockResults(ctx context.Context, height *int64) (*rpctypes.ResultBlockResults, error) {
	fmt.Println("--BlockResults--")
	return nil, nil
}

func (c *ABCIClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*rpctypes.ResultBlockchainInfo, error) {
	fmt.Println("--BlockchainInfo--")
	return nil, nil
}

func (c *ABCIClient) Commit(ctx context.Context, height *int64) (*rpctypes.ResultCommit, error) {
	fmt.Println("--Commit--")
	return nil, nil
}

func (c *ABCIClient) Tx(ctx context.Context, hash []byte, prove bool) (*rpctypes.ResultTx, error) {
	fmt.Println("--Tx--")
	return nil, nil
}

func (c *ABCIClient) TxSearch(
	ctx context.Context,
	query string,
	prove bool,
	page, perPage *int,
	orderBy string,
) (*rpctypes.ResultTxSearch, error) {
	fmt.Println("--TxSearch--")
	return nil, nil
}

func (c *ABCIClient) BlockSearch(
	ctx context.Context,
	query string,
	page, perPage *int,
	orderBy string,
) (*rpctypes.ResultBlockSearch, error) {
	fmt.Println("--BlockSearch--")
	return nil, nil
}
