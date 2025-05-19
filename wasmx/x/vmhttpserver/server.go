package vmhttpserver

import (
	"context"
	"net"
	"net/http"

	cfg "github.com/loredanacirstea/wasmx/config"
	"github.com/loredanacirstea/wasmx/x/websrv/server/config"
	"github.com/rs/cors"
	"golang.org/x/net/netutil"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"

	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// StartWebsrv starts the web server
func StartWebsrv(
	coreHandler wasmxtypes.WasmxCosmosHandler,
	parentCtx context.Context,
	goRoutineGroup *errgroup.Group,
	logger log.Logger,
	cfg *WebsrvConfig,
	actionExecutor cfg.ActionExecutor,
	senderAddress string,
) (*http.Server, *WebsrvServer, chan struct{}, error) {
	logger.Info("starting websrv web server ")
	websrvServer := NewWebsrvServer(coreHandler, parentCtx, logger, cfg, actionExecutor, senderAddress)
	mux := http.NewServeMux()

	// TODO Oauth ?
	// if cfg.EnableOAuth {
	// 	logger.Info("starting websrv oauth2 server " + cfg.Address)
	// 	websrvServer.InitOauth2(mux, path.Join(clientCtx.HomeDir, dirname))
	// }
	mux.HandleFunc("/", websrvServer.Route)

	handlerWithCors := cors.Default()
	// TODO better cors
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
		return nil, nil, nil, err
	}

	goRoutineGroup.Go(func() error {
		// httpSrv, httpSrvDone, err
		err := startWebServerGoRoutine(httpSrv, ln, parentCtx, logger, cfg, httpSrvDone)
		if err != nil {
			logger.Error(err.Error())
		}
		return err
	})
	return httpSrv, websrvServer, httpSrvDone, nil
}

func startWebServerGoRoutine(
	httpSrv *http.Server,
	ln net.Listener,
	parentCtx context.Context,
	logger log.Logger,
	cfg *WebsrvConfig,
	httpSrvDone chan struct{},
) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Info("Starting Websrv server", "address", cfg.Address)
		if err := httpSrv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				logger.Info("closing Websrv", "message", err.Error())
				close(httpSrvDone)
				return
			}
			logger.Error("failed to serve Websrv", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case <-parentCtx.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the websrv server.
		logger.Info("stopping websrv web server...", "address", cfg.Address)
		err := httpSrv.Close()
		if err != nil {
			logger.Error("stopping websrv web server error: ", err.Error())
		}
		close(errCh)
		close(httpSrvDone)
		return nil
	case err := <-errCh:
		logger.Error("failed to boot websrv server", "error", err.Error())
		return err
	}
}

func Listen(addr string, cfg *WebsrvConfig) (net.Listener, error) {
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
