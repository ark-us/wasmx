package app

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	dbm "github.com/cosmos/cosmos-db"
	baseapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	"mythos/v1/x/network/vmp2p"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type AppOptions interface {
	Get(string) interface{}
	Set(key string, value any)
}

// newApp creates a new Cosmos SDK app
func AppCreator(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts AppOptions,
	g *errgroup.Group,
	ctx context.Context,
) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	ctx = wasmxtypes.ContextWithBackgroundProcesses(ctx)
	ctx = vmp2p.WithP2PEmptyContext(ctx)
	ctx, bapps := mcfg.WithMultiChainAppEmpty(ctx)
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
		fmt.Println("---appCreator newApp--")

		encodingConfig := menc.MakeEncodingConfig(chainCfg)
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
			encodingConfig,
			minGasPrices,
			appOpts,
			baseappOptions...,
		)
		bapps.SetApp(chainId, app)
		return app
	}
	bapps.SetAppCreator(appCreator)

	return bapps, appCreator
}
