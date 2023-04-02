package server

import (
	"net"
	"net/http"
	"time"

	"github.com/rs/cors"
	"golang.org/x/net/netutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"

	"wasmx/x/websrv/server/config"
)

// StartJSONRPC starts the JSON-RPC server
func StartWebsrv(
	ctx *server.Context,
	clientCtx client.Context,
	cfg *config.WebsrvConfig,
) (*http.Server, chan struct{}, error) {
	ctx.Logger.Info("starting websrv web server at ", cfg.Address)
	websrvServer := NewWebsrvServer(ctx, ctx.Logger, clientCtx, cfg)

	mux := http.NewServeMux()

	if cfg.EnableOAuth {
		ctx.Logger.Info("starting websrv oauth2 server at ", cfg.Address)
		websrvServer.InitOauth2(mux)
	}
	mux.HandleFunc("/", websrvServer.Route)

	handlerWithCors := cors.Default()
	// if len(cfg.CORSAllowedOrigins) == 1 && cfg.CORSAllowedOrigins[0] == "*" {
	handlerWithCors = cors.AllowAll()
	// }

	httpSrv := &http.Server{
		Addr:    cfg.Address,
		Handler: handlerWithCors.Handler(mux),
		// TODO
		// ReadHeaderTimeout: cfg.HTTPTimeout,
		// ReadTimeout:       cfg.HTTPTimeout,
		// WriteTimeout:      cfg.HTTPTimeout,
		// IdleTimeout:       cfg.HTTPIdleTimeout,
	}
	httpSrvDone := make(chan struct{}, 1)

	ln, err := Listen(httpSrv.Addr, cfg)
	if err != nil {
		return nil, nil, err
	}

	errCh := make(chan error)
	go func() {
		ctx.Logger.Info("Starting Websrv server", "address", cfg.Address)
		if err := httpSrv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				close(httpSrvDone)
				return
			}

			ctx.Logger.Error("failed to start Websrv server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		ctx.Logger.Error("failed to boot Websrv server", "error", err.Error())
		return nil, nil, err
	case <-time.After(types.ServerStartTime): // assume JSON RPC server started successfully
	}

	return httpSrv, httpSrvDone, nil
}

func Listen(addr string, cfg *config.WebsrvConfig) (net.Listener, error) {
	if addr == "" {
		addr = ":" + config.DefaultWebsrvPort
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
