package server

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"golang.org/x/net/netutil"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	flags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"

	ethlog "github.com/ethereum/go-ethereum/log"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	rpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"

	menc "wasmx/v1/encoding"
	"wasmx/v1/server/config"
	"wasmx/v1/x/wasmx/rpc"
	jsonrpcconfig "wasmx/v1/x/wasmx/server/config"
)

// StartJsonRpc starts the web server
func StartJsonRpc(
	svrCtx *server.Context,
	clientCtx client.Context,
	ctx context.Context,
	tmRPCAddr,
	tmEndpoint string,
	cfgAll *config.Config,
	chainId string,
	chainConfig menc.ChainConfig,
) (*http.Server, chan struct{}, error) {
	cfg := cfgAll.JsonRpc
	svrCtx.Logger.Info("starting JSON-RPC server ", cfg.Address)

	// tmWsClient := ConnectTmWS(tmRPCAddr, tmEndpoint, svrCtx.Logger)
	var tmWsClient *rpcclient.WSClient

	logger := svrCtx.Logger.With(log.ModuleKey, "geth")
	logLevel := cast.ToString(svrCtx.Viper.Get(flags.FlagLogLevel))
	switch logLevel {
	case zerolog.DebugLevel.String():
		ethlog.SetDefault(ethlog.NewLogger(ethlog.NewTerminalHandlerWithLevel(os.Stderr, ethlog.LevelDebug, true)))
	case zerolog.InfoLevel.String():
		ethlog.SetDefault(ethlog.NewLogger(ethlog.NewTerminalHandlerWithLevel(os.Stderr, ethlog.LevelInfo, true)))
	case zerolog.ErrorLevel.String():
		ethlog.SetDefault(ethlog.NewLogger(ethlog.NewTerminalHandlerWithLevel(os.Stderr, ethlog.LevelError, true)))
	}

	rpcServer := ethrpc.NewServer()

	allowUnprotectedTxs := cfg.AllowUnprotectedTxs
	rpcAPIArr := cfg.API

	apis := rpc.GetRPCAPIs(svrCtx, clientCtx, ctx, tmWsClient, allowUnprotectedTxs, rpcAPIArr, chainId, chainConfig)

	for _, api := range apis {
		if err := rpcServer.RegisterName(api.Namespace, api.Service); err != nil {
			svrCtx.Logger.Error(
				"failed to register service in JSON RPC namespace",
				"namespace", api.Namespace,
				"service", api.Service,
			)
			return nil, nil, err
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/", rpcServer.ServeHTTP).Methods("POST")

	handlerWithCors := cors.Default()
	if cfgAll.API.EnableUnsafeCORS {
		handlerWithCors = cors.AllowAll()
	}

	httpSrv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithCors.Handler(router),
		ReadHeaderTimeout: cfg.HTTPTimeout,
		ReadTimeout:       cfg.HTTPTimeout,
		WriteTimeout:      cfg.HTTPTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}
	httpSrvDone := make(chan struct{}, 1)

	ln, err := Listen(httpSrv.Addr, &cfg)
	if err != nil {
		return nil, nil, err
	}

	errCh := make(chan error, 1)
	go func() {
		svrCtx.Logger.Info("Starting JSON-RPC server", "address", cfg.Address)
		if err := httpSrv.Serve(ln); err != nil {
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
		logger.Info("stopping JSON-RPC server...", "address", cfg.Address)
		httpSrv.Close()
		close(errCh)
		return httpSrv, httpSrvDone, nil
	case err := <-errCh:
		svrCtx.Logger.Error("failed to boot JSON-RPC server", "error", err.Error())
		return nil, nil, err
	}
}

func Listen(addr string, cfg *jsonrpcconfig.JsonRpcConfig) (net.Listener, error) {
	if addr == "" {
		addr = ":" + jsonrpcconfig.DefaultJsonRpcPort
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConnections > 0 {
		ln = netutil.LimitListener(ln, cfg.MaxOpenConnections)
	}
	return ln, err
}

func getCtx(svrCtx *server.Context, block bool) (*errgroup.Group, context.Context) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	server.ListenForQuitSignals(g, block, cancelFn, svrCtx.Logger)
	return g, ctx
}
