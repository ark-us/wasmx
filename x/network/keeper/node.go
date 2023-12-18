package keeper

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	cmttypes "github.com/cometbft/cometbft/types"

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
	privValidator cmttypes.PrivValidator,
	nodeKey *p2p.NodeKey,
	genesisDocProvider node.GenesisDocProvider,
	metricsProvider node.MetricsProvider,
) (*grpc.Server, client.CometRPC, chan struct{}, error) {
	GRPCAddr := cfgAll.Network.Address
	ln, err := Listen(GRPCAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	logger := svrCtx.Logger.With("module", "network")

	// TODO we are starting the raft protocol before the grpc server is running; we should start it after
	grpcServer, rpcClient, err := NewGRPCServer(ctx, svrCtx, clientCtx, cfgAll, app, privValidator, nodeKey, genesisDocProvider, metricsProvider)
	if err != nil {
		return nil, nil, nil, err
	}

	err = StartRPC(svrCtx, ctx, app, rpcClient, svrCtx.Logger, cfgAll)
	if err != nil {
		return nil, nil, nil, err
	}

	httpSrvDone := make(chan struct{}, 1)

	errCh := make(chan error)
	go func() {
		svrCtx.Logger.Info("Starting network GRPC server", "address", GRPCAddr)
		if err := grpcServer.Serve(ln); err != nil {
			fmt.Println("---err--", err)
			if err == http.ErrServerClosed {
				svrCtx.Logger.Info("Closing network GRPC server", "address", GRPCAddr, err.Error())
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

		return grpcServer, rpcClient, httpSrvDone, nil
	case err := <-errCh:
		svrCtx.Logger.Error("failed to boot network GRPC server", "error", err.Error())
		return nil, nil, nil, err
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
