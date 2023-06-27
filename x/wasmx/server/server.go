package server

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/netutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"

	ethlog "github.com/ethereum/go-ethereum/log"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"mythos/v1/server/config"
	"mythos/v1/x/wasmx/rpc"
	jsonrpcconfig "mythos/v1/x/wasmx/server/config"
)

// StartJsonRpc starts the web server
func StartJsonRpc(
	ctx *server.Context,
	clientCtx client.Context,
	tmRPCAddr,
	tmEndpoint string,
	cfgAll *config.Config,
) (*http.Server, chan struct{}, error) {
	cfg := cfgAll.JsonRpc
	ctx.Logger.Info("starting JSON-RPC server ", cfg.Address)

	tmWsClient := ConnectTmWS(tmRPCAddr, tmEndpoint, ctx.Logger)
	logger := ctx.Logger.With("module", "geth")
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

	apis := rpc.GetRPCAPIs(ctx, clientCtx, tmWsClient, allowUnprotectedTxs, rpcAPIArr)

	for _, api := range apis {
		if err := rpcServer.RegisterName(api.Namespace, api.Service); err != nil {
			ctx.Logger.Error(
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

	// server := NewJsonRpcServer(ctx, ctx.Logger, clientCtx, &cfg)
	// router := http.NewServeMux()
	// router.HandleFunc("/", server.ServeHTTP)
	// handlerWithCors := cors.AllowAll()

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

	errCh := make(chan error)
	go func() {
		ctx.Logger.Info("Starting JSON-RPC server", "address", cfg.Address)
		if err := httpSrv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				close(httpSrvDone)
				return
			}

			ctx.Logger.Error("failed to start JSON-RPC server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		ctx.Logger.Error("failed to boot JSON-RPC server", "error", err.Error())
		return nil, nil, err
	case <-time.After(types.ServerStartTime): // assume JSON RPC server started successfully
	}

	return httpSrv, httpSrvDone, nil
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
