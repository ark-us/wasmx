package server

import (
	"context"
	"net"
	"net/http"
	"path"

	"github.com/rs/cors"
	"golang.org/x/net/netutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/loredanacirstea/wasmx/x/websrv/server/config"
)

var dirname = "oauth"

// StartWebsrv starts the web server
func StartWebsrv(
	svrCtx *server.Context,
	clientCtx client.Context,
	ctx context.Context,
	cfg *config.WebsrvConfig,
) (*http.Server, chan struct{}, error) {
	svrCtx.Logger.Info("starting websrv web server " + cfg.Address)
	websrvServer := NewWebsrvServer(svrCtx, svrCtx.Logger, clientCtx, ctx, cfg)
	mux := http.NewServeMux()

	if cfg.EnableOAuth {
		svrCtx.Logger.Info("starting websrv oauth2 server " + cfg.Address)
		websrvServer.InitOauth2(mux, path.Join(clientCtx.HomeDir, dirname))
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

	errCh := make(chan error, 1)
	go func() {
		svrCtx.Logger.Info("Starting Websrv server", "address", cfg.Address)
		if err := httpSrv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				svrCtx.Logger.Info("closing Websrv", "message", err.Error())
				close(httpSrvDone)
				return
			}
			svrCtx.Logger.Error("failed to serve Websrv", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the websrv server.
		svrCtx.Logger.Info("stopping websrv web server...", "address", cfg.Address)
		httpSrv.Close()
		close(errCh)
		return httpSrv, httpSrvDone, nil
	case err := <-errCh:
		svrCtx.Logger.Error("failed to boot websrv server", "error", err.Error())
		return nil, nil, err
	}
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
