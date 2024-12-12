package app

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcfg "github.com/loredanacirstea/wasmx/config"
	mctx "github.com/loredanacirstea/wasmx/context"
	menc "github.com/loredanacirstea/wasmx/encoding"
	multichain "github.com/loredanacirstea/wasmx/multichain"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	"github.com/loredanacirstea/wasmx/x/network/vmp2p"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// newApp creates a new Cosmos SDK app
func NewAppCreator(
	wasmVmMeta memc.IWasmVmMeta,
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts multichain.AppOptions,
	g *errgroup.Group,
	ctx context.Context,
	apictx mcfg.APICtxI,
) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	ctx = wasmxtypes.ContextWithBackgroundProcesses(ctx)
	ctx = vmp2p.WithP2PEmptyContext(ctx)
	ctx = networktypes.ContextWithMultiChainContext(g, ctx, logger)
	ctx, bapps := mcfg.WithMultiChainAppEmpty(ctx)
	ctx, _ = mctx.WithExecutionMetaInfoEmpty(ctx)
	ctx, _ = mctx.WithTimeoutGoroutinesInfoEmpty(ctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", ctx)

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

		appOpts.Set(flags.FlagChainID, chainId)
		appOpts.Set(sdkserver.FlagMinGasPrices, minGasPrices.String())
		appOpts.Set(sdkserver.FlagPruning, pruningtypes.PruningOptionDefault)
		baseappOptions := mcfg.DefaultBaseappOptions(appOpts)

		app := NewApp(
			chainId,
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
			wasmVmMeta,
			baseappOptions...,
		)
		bapps.SetApp(chainId, app)
		return app
	}
	bapps.SetAppCreator(appCreator)
	bapps.SetAPICtx(apictx)

	return bapps, appCreator
}
