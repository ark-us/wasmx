package keeper

import (
	"context"
	"fmt"
	"net"
	"strconv"

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

	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/node"

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
	GetTKey(storeKey string) *storetypes.TransientStoreKey
	GetMKey(storeKey string) *storetypes.MemoryStoreKey
}

type BaseApp interface {
	Name() string
	CreateQueryContext(height int64, prove bool) (sdk.Context, error)
	CommitMultiStore() storetypes.CommitMultiStore
	GetContextForCheckTx(txBytes []byte) sdk.Context
	GetContextForFinalizeBlock(txBytes []byte) sdk.Context
}

// NewGRPCServer returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server. See StartGRPCServer.
func NewGRPCServer(
	svrCtx *server.Context,
	clientCtx client.Context,
	cfg config.GRPCConfig,
	app servertypes.Application,
	tmNode *node.Node,
	createGoRoutine func(description string, fn func() error, gracefulStop func()) (chan struct{}, error),
) (*grpc.Server, error) {
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

	err := RegisterGRPCServer(svrCtx, clientCtx, tmNode, app, grpcSrv, createGoRoutine)
	if err != nil {
		return nil, fmt.Errorf("failed to register grpc server: %w", err)
	}

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
	// fmt.Println("---NewGRPCServer-END-")
	return grpcSrv, nil
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
	svrCtx *server.Context,
	clientCtx client.Context,
	tmNode *node.Node,
	sapp servertypes.Application,
	server *grpc.Server,
	createGoRoutine func(description string, fn func() error, gracefulStop func()) (chan struct{}, error),
) error {
	app, ok := sapp.(BaseApp)
	if !ok {
		return fmt.Errorf("failed to get BaseApp from server Application")
	}
	mythosapp, ok := app.(MythosApp)
	if !ok {
		return fmt.Errorf("failed to get MythosApp from server Application")
	}

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

		// fmt.Println("-----storage before-execution---")
		// bz, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
		// tstorer := app.CommitMultiStore().GetKVStore(mythosapp.GetMKey(wasmxtypes.MemStoreKey))
		// fmt.Println("-----GET-----0000000000000000000000000000000000000000000000000000000000000001", tstorer.Get(append(tstoreprefix, bz...)))
		// bz, _ = hex.DecodeString("b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6")
		// fmt.Println("------GET----b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6", tstorer.Get(append(tstoreprefix, bz...)))

		// TODO - use this for grpc queries
		sdkCtx_, ctxcachems, err := CreateQueryContext(app, svrCtx.Logger, height, false)
		if err != nil {
			return nil, err
		}
		sdkCtx, commitCacheCtx := sdkCtx_.CacheContext()

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

		hresp, err := handler(grpcCtx, req)
		if err != nil {
			return hresp, err
		}
		// commit changes

		// fmt.Println("----temp state")
		// fmt.Println("--ContractStore--storageKey--", wasmxtypes.MemStoreKey, mythosapp.GetMKey(wasmxtypes.MemStoreKey).Name())

		// bz, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
		// tstorer = sdkCtx.TransientStore(mythosapp.GetMKey(wasmxtypes.MemStoreKey))
		// fmt.Println("-----GET-----0000000000000000000000000000000000000000000000000000000000000001", tstorer.Get(append(tstoreprefix, bz...)))
		// bz, _ = hex.DecodeString("b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6")
		// fmt.Println("------GET----b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6", tstorer.Get(append(tstoreprefix, bz...)))

		// bz, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
		// tstorer = sdkCtx.KVStore(mythosapp.GetMKey(wasmxtypes.MemStoreKey))
		// fmt.Println("-----GET-----0000000000000000000000000000000000000000000000000000000000000001", tstorer.Get(append(tstoreprefix, bz...)))
		// bz, _ = hex.DecodeString("b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6")
		// fmt.Println("------GET----b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6", tstorer.Get(append(tstoreprefix, bz...)))

		commitCacheCtx()

		// we just committed to the query context
		// fmt.Println("----query context state")
		// tstore0 := sdkCtx_.MultiStore().GetKVStore(mythosapp.GetMKey(wasmxtypes.MemStoreKey))
		// bz, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
		// fmt.Println("-----GET-----0000000000000000000000000000000000000000000000000000000000000001", tstore0.Get(append(tstoreprefix, bz...)))
		// bz, _ = hex.DecodeString("b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6")
		// fmt.Println("------GET----b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6", tstore0.Get(append(tstoreprefix, bz...)))

		// commit original context
		origtstore := ctxcachems.GetStore(mythosapp.GetMKey(wasmxtypes.MemStoreKey))
		origtstore.(storetypes.CacheWrap).Write()

		// fmt.Println("----app state")
		// tstorer = app.CommitMultiStore().GetKVStore(mythosapp.GetMKey(wasmxtypes.MemStoreKey))
		// bz, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
		// fmt.Println("-----GET-----0000000000000000000000000000000000000000000000000000000000000001", tstorer.Get(append(tstoreprefix, bz...)))
		// bz, _ = hex.DecodeString("b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6")
		// fmt.Println("------GET----b10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6", tstorer.Get(append(tstoreprefix, bz...)))

		return hresp, nil
	}

	handler := &msgServer{
		Keeper:          mythosapp.GetNetworkKeeper(),
		TmNode:          tmNode,
		createGoRoutine: createGoRoutine,
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
	return nil
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

	if height <= 1 && prove {
		return sdk.Context{}, nil,
			errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			)
	}

	cacheMS, err := qms.CacheMultiStoreWithVersion(height)
	if err != nil {
		return sdk.Context{}, nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"failed to load state at height %d; %s (latest height: %d)", height, err, lastBlockHeight,
			)
	}

	tmpctx, err := app.CreateQueryContext(height, false)
	if err != nil {
		return sdk.Context{}, nil, err
	}

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
