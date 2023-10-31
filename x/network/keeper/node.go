package keeper

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/cometbft/cometbft/node"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"mythos/v1/server/config"
)

func StartGRPCServer(
	svrCtx *server.Context,
	clientCtx client.Context,
	ctx context.Context,
	GRPCAddr string,
	cfgAll *config.Config,
	app servertypes.Application,
	tmNode *node.Node,
) (*grpc.Server, chan struct{}, error) {

	ln, err := Listen(GRPCAddr)
	if err != nil {
		return nil, nil, err
	}

	logger := svrCtx.Logger.With("module", "network")

	grpcServer, err := NewGRPCServer(svrCtx, clientCtx, cfgAll.GRPC, app, tmNode)
	if err != nil {
		return nil, nil, err
	}

	httpSrvDone := make(chan struct{}, 1)

	errCh := make(chan error)
	go func() {
		svrCtx.Logger.Info("Starting network server", "address", GRPCAddr)
		if err := grpcServer.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				close(httpSrvDone)
				return
			}

			svrCtx.Logger.Error("failed to start JSON-RPC server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the JSON-RPC server.
		logger.Info("stopping JSON-RPC server...", "address", GRPCAddr)
		grpcServer.GracefulStop()

		return grpcServer, httpSrvDone, nil
	case err := <-errCh:
		svrCtx.Logger.Error("failed to boot JSON-RPC server", "error", err.Error())
		return nil, nil, err
	}
}

func Listen(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	// if cfg.MaxOpenConnections > 0 {
	// 	ln = netutil.LimitListener(ln, cfg.MaxOpenConnections)
	// }
	return ln, err
}

// func (n *Node) startRPC() ([]net.Listener, error) {
// 	// env, err := n.ConfigureRPC()
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	listenAddrs := splitAndTrimEmpty(n.config.RPC.ListenAddress, ",", " ")
// 	routes := env.GetRoutes()

// 	if n.config.RPC.Unsafe {
// 		env.AddUnsafeRoutes(routes)
// 	}

// 	config := rpcserver.DefaultConfig()
// 	config.MaxBodyBytes = n.config.RPC.MaxBodyBytes
// 	config.MaxHeaderBytes = n.config.RPC.MaxHeaderBytes
// 	config.MaxOpenConnections = n.config.RPC.MaxOpenConnections
// 	// If necessary adjust global WriteTimeout to ensure it's greater than
// 	// TimeoutBroadcastTxCommit.
// 	// See https://github.com/tendermint/tendermint/issues/3435
// 	if config.WriteTimeout <= n.config.RPC.TimeoutBroadcastTxCommit {
// 		config.WriteTimeout = n.config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
// 	}

// 	// we may expose the rpc over both a unix and tcp socket
// 	listeners := make([]net.Listener, len(listenAddrs))
// 	for i, listenAddr := range listenAddrs {
// 		mux := http.NewServeMux()
// 		rpcLogger := n.Logger.With("module", "rpc-server")
// 		wmLogger := rpcLogger.With("protocol", "websocket")
// 		wm := rpcserver.NewWebsocketManager(routes,
// 			rpcserver.OnDisconnect(func(remoteAddr string) {
// 				err := n.eventBus.UnsubscribeAll(context.Background(), remoteAddr)
// 				if err != nil && err != cmtpubsub.ErrSubscriptionNotFound {
// 					wmLogger.Error("Failed to unsubscribe addr from events", "addr", remoteAddr, "err", err)
// 				}
// 			}),
// 			rpcserver.ReadLimit(config.MaxBodyBytes),
// 			rpcserver.WriteChanCapacity(n.config.RPC.WebSocketWriteBufferSize),
// 		)
// 		wm.SetLogger(wmLogger)
// 		mux.HandleFunc("/websocket", wm.WebsocketHandler)
// 		rpcserver.RegisterRPCFuncs(mux, routes, rpcLogger)
// 		listener, err := rpcserver.Listen(
// 			listenAddr,
// 			config.MaxOpenConnections,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		var rootHandler http.Handler = mux
// 		if n.config.RPC.IsCorsEnabled() {
// 			corsMiddleware := cors.New(cors.Options{
// 				AllowedOrigins: n.config.RPC.CORSAllowedOrigins,
// 				AllowedMethods: n.config.RPC.CORSAllowedMethods,
// 				AllowedHeaders: n.config.RPC.CORSAllowedHeaders,
// 			})
// 			rootHandler = corsMiddleware.Handler(mux)
// 		}
// 		if n.config.RPC.IsTLSEnabled() {
// 			go func() {
// 				if err := rpcserver.ServeTLS(
// 					listener,
// 					rootHandler,
// 					n.config.RPC.CertFile(),
// 					n.config.RPC.KeyFile(),
// 					rpcLogger,
// 					config,
// 				); err != nil {
// 					n.Logger.Error("Error serving server with TLS", "err", err)
// 				}
// 			}()
// 		} else {
// 			go func() {
// 				if err := rpcserver.Serve(
// 					listener,
// 					rootHandler,
// 					rpcLogger,
// 					config,
// 				); err != nil {
// 					n.Logger.Error("Error serving server", "err", err)
// 				}
// 			}()
// 		}

// 		listeners[i] = listener
// 	}

// 	// we expose a simplified api over grpc for convenience to app devs
// 	grpcListenAddr := n.config.RPC.GRPCListenAddress
// 	if grpcListenAddr != "" {
// 		config := rpcserver.DefaultConfig()
// 		config.MaxBodyBytes = n.config.RPC.MaxBodyBytes
// 		config.MaxHeaderBytes = n.config.RPC.MaxHeaderBytes
// 		// NOTE: GRPCMaxOpenConnections is used, not MaxOpenConnections
// 		config.MaxOpenConnections = n.config.RPC.GRPCMaxOpenConnections
// 		// If necessary adjust global WriteTimeout to ensure it's greater than
// 		// TimeoutBroadcastTxCommit.
// 		// See https://github.com/tendermint/tendermint/issues/3435
// 		if config.WriteTimeout <= n.config.RPC.TimeoutBroadcastTxCommit {
// 			config.WriteTimeout = n.config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
// 		}
// 		listener, err := rpcserver.Listen(grpcListenAddr, config.MaxOpenConnections)
// 		if err != nil {
// 			return nil, err
// 		}
// 		go func() {
// 			//nolint:staticcheck // SA1019: core_grpc.StartGRPCClient is deprecated: A new gRPC API will be introduced after v0.38.
// 			// if err := grpccore.StartGRPCServer(env, listener); err != nil {
// 			if err := grpccore.StartGRPCServer(listener); err != nil {
// 				n.Logger.Error("Error starting gRPC server", "err", err)
// 			}
// 		}()
// 		listeners = append(listeners, listener)

// 	}

// 	return listeners, nil
// }
