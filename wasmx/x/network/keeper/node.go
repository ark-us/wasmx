package keeper

import (
	"context"
	"net"
	"net/http"
	"time"

	_ "net/http/pprof" //nolint: gosec

	"google.golang.org/grpc"

	"github.com/cometbft/cometbft/node"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"wasmx/v1/server/config"
)

func StartGRPCServer(
	svrCtx *server.Context,
	clientCtx client.Context,
	ctx context.Context,
	cfgAll *config.Config,
	app servertypes.Application,
	metricsProvider node.MetricsProvider,
	rpcClient client.CometRPC,
) (*grpc.Server, error) {
	GRPCAddr := cfgAll.Network.Address
	ln, err := Listen(GRPCAddr)
	if err != nil {
		return nil, err
	}

	svrCtx.Logger = svrCtx.Logger.With(log.ModuleKey, "network")

	// TODO we are starting the consensus protocol before the grpc server is running; we should start it after
	grpcServer, err := NewGRPCServer(ctx, svrCtx, clientCtx, cfgAll, app, metricsProvider, rpcClient)
	if err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)

	go func() {
		err = StartRPC(svrCtx, ctx, app, rpcClient, svrCtx.Logger, cfgAll)
		if err != nil {
			svrCtx.Logger.Error("Failed to start network RPC server", "error", err.Error())
			errCh <- err
		}
	}()

	go func() {
		svrCtx.Logger.Info("Starting network GRPC server", "address", GRPCAddr)
		if err := grpcServer.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				svrCtx.Logger.Info("Closing network GRPC server", "address", GRPCAddr, err.Error())
				return
			}

			svrCtx.Logger.Error("failed to start network GRPC server", "error", err.Error())
			errCh <- err
		}
	}()

	var pprofSrv *http.Server
	// run pprof server if it is enabled
	if svrCtx.Config.RPC.IsPprofEnabled() {
		readHeaderTimeout := 10 * time.Second
		pprofSrv = &http.Server{
			Addr:              svrCtx.Config.RPC.PprofListenAddress,
			Handler:           nil,
			ReadHeaderTimeout: readHeaderTimeout,
		}
		go func() {
			if err := pprofSrv.ListenAndServe(); err != http.ErrServerClosed {
				// Error starting or closing listener:
				svrCtx.Logger.Error("pprof HTTP server ListenAndServe", "err", err)
			}
		}()
	}

	select {
	case <-ctx.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the GRPC server.
		svrCtx.Logger.Info("stopping network GRPC server...", "address", GRPCAddr)
		grpcServer.GracefulStop()

		if svrCtx.Config.RPC.IsPprofEnabled() {
			if err := pprofSrv.Shutdown(context.Background()); err != nil {
				svrCtx.Logger.Error("Pprof HTTP server Shutdown", "err", err)
			}
		}

		return grpcServer, nil
	case err := <-errCh:
		svrCtx.Logger.Error("failed to boot network GRPC server", "error", err.Error())
		return nil, err
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
