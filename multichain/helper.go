package multichain

import (
	"context"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"

	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
)

// DefaultAppOptions is a stub implementing AppOptions
type DefaultAppOptions map[string]interface{}

// Get implements AppOptions
func (m DefaultAppOptions) Get(key string) interface{} {
	v, ok := m[key]
	if !ok {
		return interface{}(nil)
	}

	return v
}

func (m DefaultAppOptions) Set(key string, value interface{}) {
	m[key] = value
}

func CreateMockAppCreator(appCreatorFactory NewAppCreator, homeDir string) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	// level := "network:debug,wasmx:debug,*:info"
	// filter, _ := log.ParseLogLevel(level)
	// logger := log.NewLogger(
	// 	os.Stderr,
	// 	log.LevelOption(1), // info=1
	// 	log.FilterOption(filter),
	// )
	logger := log.NewNopLogger()

	db := dbm.NewMemDB()
	appOpts := DefaultAppOptions{}
	appOpts.Set(flags.FlagHome, homeDir)
	// we set this so it does not try to read a genesis file
	appOpts.Set(flags.FlagChainID, mcfg.MYTHOS_CHAIN_ID_TESTNET)
	appOpts.Set(sdkserver.FlagInvCheckPeriod, 5)
	appOpts.Set(sdkserver.FlagUnsafeSkipUpgrades, 0)
	appOpts.Set(sdkserver.FlagMinGasPrices, "")
	appOpts.Set(sdkserver.FlagPruning, pruningtypes.PruningOptionDefault)
	g, goctx, _ := GetTestCtx(logger, true)
	return appCreatorFactory(logger, db, nil, appOpts, g, goctx)
}

func GetTestCtx(logger log.Logger, block bool) (*errgroup.Group, context.Context, context.CancelFunc) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	sdkserver.ListenForQuitSignals(g, block, cancelFn, logger)
	return g, ctx, cancelFn
}
