package multichain

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"

	cmtcfg "github.com/cometbft/cometbft/config"

	srvconfig "mythos/v1/server/config"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	mcfg "mythos/v1/config"
	mctx "mythos/v1/context"
	menc "mythos/v1/encoding"
)

type AppOptions interface {
	Get(string) interface{}
	Set(key string, value any)
}

type NewAppCreator = func(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts AppOptions,
	g *errgroup.Group,
	ctx context.Context,
	startChainApis func(string, *menc.ChainConfig, mctx.NodePorts) (mcfg.MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, error),
) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp)
