package app

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	cmtcfg "github.com/cometbft/cometbft/config"

	dbm "github.com/cosmos/cosmos-db"
	baseapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcfg "mythos/v1/config"
	mctx "mythos/v1/context"
	menc "mythos/v1/encoding"
	multichain "mythos/v1/multichain"
	srvconfig "mythos/v1/server/config"
	networktypes "mythos/v1/x/network/types"
	"mythos/v1/x/network/vmp2p"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// newApp creates a new Cosmos SDK app
func NewAppCreator(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts multichain.AppOptions,
	g *errgroup.Group,
	ctx context.Context,
	startChainAPIs func(string, *menc.ChainConfig, mctx.NodePorts) (mcfg.MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error),
) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	ctx = wasmxtypes.ContextWithBackgroundProcesses(ctx)
	ctx = vmp2p.WithP2PEmptyContext(ctx)
	ctx = networktypes.ContextWithMultiChainContext(g, ctx, logger)
	ctx, bapps := mcfg.WithMultiChainAppEmpty(ctx)
	ctx, _ = mctx.WithExecutionMetaInfoEmpty(ctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", ctx)

	baseappOptions := mcfg.DefaultBaseappOptions(appOpts)

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(sdkserver.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	gasPricesStr := cast.ToString(appOpts.Get(sdkserver.FlagMinGasPrices))
	gasPrices, err := sdk.ParseDecCoins(gasPricesStr)
	if err != nil {
		panic(fmt.Sprintf("invalid minimum gas prices: %v", err))
	}
	minGasAmount := math.LegacyNewDec(0)
	if len(gasPrices) > 0 {
		minGasAmount = gasPrices[0].Amount
	}

	appCreator := func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp {

		encodingConfig := menc.MakeEncodingConfig(chainCfg, GetCustomSigners())
		minGasPrices := sdk.NewDecCoins(sdk.NewDecCoin(chainCfg.BaseDenom, minGasAmount.RoundInt()))

		baseappOptions[1] = baseapp.SetMinGasPrices(minGasPrices.String())
		baseappOptions[len(baseappOptions)-1] = baseapp.SetChainID(chainId)

		app := NewApp(
			logger,
			db,
			traceStore,
			true,
			skipUpgradeHeights,
			cast.ToString(appOpts.Get(flags.FlagHome)),
			cast.ToUint(appOpts.Get(sdkserver.FlagInvCheckPeriod)),
			chainCfg,
			encodingConfig,
			minGasPrices,
			appOpts,
			baseappOptions...,
		)
		bapps.SetApp(chainId, app)
		return app
	}
	bapps.SetAppCreator(appCreator)
	bapps.SetStartAPIs(startChainAPIs)

	return bapps, appCreator
}
