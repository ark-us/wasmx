package keeper

import (
	"context"
	"fmt"
	"net"
	"strconv"

	// "golang.org/x/net/context"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/reflection"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	errorsmod "cosmossdk.io/errors"

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

	dbm "github.com/cosmos/cosmos-db"

	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/node"

	"mythos/v1/x/network/types"
)

// Config is an gRPC server configuration.
//
// Deprecated: A new gRPC API will be introduced after v0.38.
type Config struct {
	MaxOpenConnections int
}

type MythosApp interface {
	GetNetworkKeeper() Keeper
	GetDB() dbm.DB
}

type BaseApp interface {
	CreateQueryContext(height int64, prove bool) (sdk.Context, error)
}

// // GetGRPCServer
// func GetGRPCServer(
// // env *core.Environment,
// ) *grpc.Server {
// 	grpcServer := grpc.NewServer()
// 	types.RegisterMsgServer(grpcServer, &msgServer{})
// 	// Register reflection service on gRPC server.
// 	reflection.Register(grpcServer)
// 	return grpcServer
// }

// NewGRPCServer returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server. See StartGRPCServer.
// func NewGRPCServer(clientCtx client.Context, app types.Application, cfg config.GRPCConfig) (*grpc.Server, error) {
func NewGRPCServer(
	svrCtx *server.Context,
	clientCtx client.Context,
	cfg config.GRPCConfig,
	app servertypes.Application,
	tmNode *node.Node,
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

	// app.RegisterGRPCServer(grpcSrv)
	err := RegisterGRPCServer(svrCtx, clientCtx, tmNode, app, grpcSrv)
	if err != nil {
		return nil, fmt.Errorf("failed to register grpc server: %w", err)
	}
	// mythosapp, ok := app.(MythosApp)
	// if !ok {
	// 	return nil, fmt.Errorf("failed to get MythosApp from server Application")
	// }
	// types.RegisterMsgServer(grpcSrv, &msgServer{Keeper: mythosapp.GetNetworkKeeper()})

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
//
// Deprecated: A new gRPC API will be introduced after v0.38.
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

// RegisterGRPCServer registers gRPC services directly with the gRPC server.
func RegisterGRPCServer(svrCtx *server.Context, clientCtx client.Context, tmNode *node.Node, sapp servertypes.Application, server *grpc.Server) error {
	app, ok := sapp.(BaseApp)
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

		// Create the sdk.Context. Passing false as 2nd arg, as we can't
		// actually support proofs with gRPC right now.
		sdkCtx, err := app.CreateQueryContext(height, false)
		if err != nil {
			return nil, err
		}

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

		return handler(grpcCtx, req)
	}

	mythosapp, ok := app.(MythosApp)
	if !ok {
		return fmt.Errorf("failed to get MythosApp from server Application")
	}
	handler := &msgServer{Keeper: mythosapp.GetNetworkKeeper(), DB: mythosapp.GetDB(), ClientCtx: clientCtx, TmNode: tmNode}
	fmt.Println("---before RegisterMsgServer")
	// types.RegisterMsgServer(server, handler)
	fmt.Println("---after RegisterMsgServer")

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

	// infos := server.GetServiceInfo()
	// fmt.Println("---infos", infos)

	// for key, val := range infos {
	// 	fmt.Println("--- key, val", key, val)
	// 	// newDesc := &grpc.ServiceDesc{
	// 	// 	ServiceName: key,
	// 	// 	HandlerType: desc.HandlerType,
	// 	// 	Methods:     newMethods,
	// 	// 	Streams:     desc.Streams,
	// 	// 	Metadata:    desc.Metadata,
	// 	// }
	// 	// server.RegisterService(newDesc, data.handler)
	// }
	// fmt.Println("--- interceptor", interceptor)

	// newDesc := &grpc.ServiceDesc{
	// 	ServiceName: desc.ServiceName,
	// 	HandlerType: desc.HandlerType,
	// 	Methods:     newMethods,
	// 	Streams:     desc.Streams,
	// 	Metadata:    desc.Metadata,
	// }
	// server.RegisterService(newDesc, data.handler)

	// Loop through all services and methods, add the interceptor, and register
	// the service.
	// for _, data := range app.GRPCQueryRouter().serviceData {
	// 	desc := data.serviceDesc
	// 	newMethods := make([]grpc.MethodDesc, len(desc.Methods))

	// 	for i, method := range desc.Methods {
	// 		methodHandler := method.Handler
	// 		newMethods[i] = grpc.MethodDesc{
	// 			MethodName: method.MethodName,
	// 			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	// 				return methodHandler(srv, ctx, dec, grpcmiddleware.ChainUnaryServer(
	// 					grpcrecovery.UnaryServerInterceptor(),
	// 					interceptor,
	// 				))
	// 			},
	// 		}
	// 	}

	// 	newDesc := &grpc.ServiceDesc{
	// 		ServiceName: desc.ServiceName,
	// 		HandlerType: desc.HandlerType,
	// 		Methods:     newMethods,
	// 		Streams:     desc.Streams,
	// 		Metadata:    desc.Metadata,
	// 	}

	// 	server.RegisterService(newDesc, data.handler)
	// }
	return nil
}

func checkNegativeHeight(height int64) error {
	if height < 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot query with height < 0; please provide a valid height")
	}

	return nil
}
