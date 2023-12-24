package server

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/netutil"
	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	ethlog "github.com/ethereum/go-ethereum/log"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	rpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"

	"mythos/v1/server/config"
	"mythos/v1/x/wasmx/rpc"
	jsonrpcconfig "mythos/v1/x/wasmx/server/config"
)

// StartJsonRpc starts the web server
func StartJsonRpc(
	svrCtx *server.Context,
	clientCtx client.Context,
	ctx context.Context,
	tmRPCAddr,
	tmEndpoint string,
	cfgAll *config.Config,
) (*http.Server, chan struct{}, error) {
	cfg := cfgAll.JsonRpc
	svrCtx.Logger.Info("starting JSON-RPC server ", cfg.Address)

	// tmWsClient := ConnectTmWS(tmRPCAddr, tmEndpoint, svrCtx.Logger)
	var tmWsClient *rpcclient.WSClient

	logger := svrCtx.Logger.With("module", "geth")
	ethlog.Root().SetHandler(ethlog.FuncHandler(func(r *ethlog.Record) error {
		switch r.Lvl {
		case ethlog.LvlTrace, ethlog.LvlDebug:
			logger.Debug(r.Msg, r.Ctx...)
		case ethlog.LvlInfo, ethlog.LvlWarn:
			logger.Info(r.Msg, r.Ctx...)
		case ethlog.LvlError, ethlog.LvlCrit:
			logger.Error(r.Msg, r.Ctx...)
		}
		return nil
	}))

	rpcServer := ethrpc.NewServer()

	allowUnprotectedTxs := cfg.AllowUnprotectedTxs
	rpcAPIArr := cfg.API

	apis := rpc.GetRPCAPIs(svrCtx, clientCtx, ctx, tmWsClient, allowUnprotectedTxs, rpcAPIArr)

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
