package multichain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"

	mapi "github.com/loredanacirstea/wasmx/api"
	mcfg "github.com/loredanacirstea/wasmx/config"
	mctx "github.com/loredanacirstea/wasmx/context"
	menc "github.com/loredanacirstea/wasmx/encoding"
	srvconfig "github.com/loredanacirstea/wasmx/server/config"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
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

type MockApiCtx struct{}

func (ac *MockApiCtx) SetMultiapp(app *mcfg.MultiChainApp) {
	// ac.Multiapp = app
}

func (*MockApiCtx) BuildConfigs(
	chainId string,
	chainCfg *menc.ChainConfig,
	ports mctx.NodePorts,
) (mcfg.MythosApp, *sdkserver.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error) {
	return nil, nil, client.Context{}, nil, nil, nil, fmt.Errorf("ApiCtx.BuildConfigs not implemented")
}

func (*MockApiCtx) StartChainApis(
	chainId string,
	chainCfg *menc.ChainConfig,
	ports mctx.NodePorts,
) (mcfg.MythosApp, *sdkserver.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error) {
	return nil, nil, client.Context{}, nil, nil, nil, fmt.Errorf("ApiCtx.StartChainApis not implemented")
}

func CreateMockAppCreator(wasmVmMeta memc.IWasmVmMeta, appCreatorFactory NewAppCreator, homeDir string, getDB func(dbpath string) dbm.DB) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	// level := "x/wasmx:debug,*:info"
	level := "error"
	filter, _ := ParseLogLevel(level)
	logger := log.NewLogger(
		os.Stderr,
		// log.LevelOption(1), // info=1
		log.FilterOption(filter),
		log.TimeFormatOption(time.RFC3339),
	)
	// logger := log.NewNopLogger()
	var db dbm.DB
	if getDB != nil {
		db = getDB(homeDir)
	} else {
		db = dbm.NewMemDB()
	}
	appOpts := DefaultAppOptions{}
	appOpts.Set(flags.FlagHome, homeDir)
	// we set this so it does not try to read a genesis file
	appOpts.Set(flags.FlagChainID, mcfg.MYTHOS_CHAIN_ID_TESTNET)
	appOpts.Set(sdkserver.FlagInvCheckPeriod, 5)
	appOpts.Set(sdkserver.FlagUnsafeSkipUpgrades, 0)
	appOpts.Set(sdkserver.FlagMinGasPrices, "")
	appOpts.Set(sdkserver.FlagPruning, pruningtypes.PruningOptionDefault)
	g, goctx, _ := GetTestCtx(logger, true)

	srvCtx := sdkserver.NewDefaultContext()
	srvCtx.Config.RootDir = homeDir
	srvCfg := srvconfig.DefaultConfig()
	srvCfg.TestingModeDisableStateSync = true
	apictx := &mapi.APICtx{
		GoRoutineGroup:  g,
		GoContextParent: goctx,
		SvrCtx:          srvCtx,
		ClientCtx:       client.Context{},
		SrvCfg:          *srvCfg,
		TndCfg:          cmtcfg.DefaultConfig(),
	}
	return appCreatorFactory(wasmVmMeta, logger, db, nil, appOpts, g, goctx, apictx)
}

func CreateNoLoggerAppCreator(wasmVmMeta memc.IWasmVmMeta, appCreatorFactory NewAppCreator, homeDir string) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
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
	return appCreatorFactory(wasmVmMeta, logger, db, nil, appOpts, g, goctx, &MockApiCtx{})
}

func CreateNoLoggerAppCreatorTemp(wasmVmMeta memc.IWasmVmMeta, appCreatorFactory NewAppCreator, index int64) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	tempNodeHome := filepath.Join(userHomeDir, fmt.Sprintf(".mythostmp_%d", index))
	return CreateNoLoggerAppCreator(wasmVmMeta, appCreatorFactory, tempNodeHome)
}

func GetTestCtx(logger log.Logger, block bool) (*errgroup.Group, context.Context, context.CancelFunc) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	sdkserver.ListenForQuitSignals(g, block, cancelFn, logger)
	return g, ctx, cancelFn
}
