package multichain

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"

	mcfg "mythos/v1/config"
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
	apictx mcfg.APICtxI,
) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp)
