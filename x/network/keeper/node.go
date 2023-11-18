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
	cfgAll *config.Config,
	app servertypes.Application,
	tmNode *node.Node,
) (*grpc.Server, chan struct{}, error) {
	GRPCAddr := cfgAll.Network.Address
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
				svrCtx.Logger.Error("Closing network server", "address", GRPCAddr, err.Error())
				close(httpSrvDone)
				return
			}

			svrCtx.Logger.Error("failed to start network GRPC server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the GRPC server.
		logger.Info("stopping network GRPC server...", "address", GRPCAddr)
		grpcServer.GracefulStop()

		return grpcServer, httpSrvDone, nil
	case err := <-errCh:
		svrCtx.Logger.Error("failed to bootnetwork GRPC server", "error", err.Error())
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
