package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	config "mythos/v1/config"
	appencoding "mythos/v1/encoding"
	networkkeeper "mythos/v1/x/network/keeper"
	networkvm "mythos/v1/x/network/vm"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// DefaultTestingAppInit defines the IBC application used for testing
var DefaultTestingAppInit func(chainId string, index int32) (ibctesting.TestingApp, map[string]json.RawMessage) = SetupTestingApp

// DefaultTestingConsensusParams defines the default Tendermint consensus params used in
// Mythos testing.
var DefaultTestingConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
		MaxBytes: 2_000_000,
		MaxGas:   30_000_000, // -1 no limit
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
	Version: &tmproto.VersionParams{
		App: 0,
	},
}

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

// Setup initializes a new Mythos. A Nop logger is set in Mythos.
func SetupApp(
	isCheckTx bool,
) *App {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	appOpts := DefaultAppOptions{}
	g, goctx, _ := GetTestCtx(logger, true)
	goctx = wasmxtypes.ContextWithBackgroundProcesses(goctx)
	goctx = networkvm.WithP2PEmptyContext(goctx)
	goctx, bapps := config.WithMultiChainAppEmpty(goctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", goctx)
	actionExecutor := networkkeeper.NewActionExecutor(bapps, logger)

	chainId := config.MYTHOS_CHAIN_ID_TEST
	chainCfg, err := config.GetChainConfig(chainId)
	if err != nil {
		panic(err)
	}

	app := NewApp(actionExecutor, log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, appencoding.MakeEncodingConfig(chainCfg), nil, appOpts)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := app.DefaultGenesis()

		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			&abci.RequestInitChain{
				ChainId:         chainId,
				Time:            time.Now().UTC(),
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultTestingConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// SetupTestingApp initializes the IBC-go testing application
func SetupTestingApp(chainID string, index int32) (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	chainCfg, err := config.GetChainConfig(chainID)
	if err != nil {
		panic(err)
	}

	// level := "network:debug,wasmx:debug,*:info"
	// filter, _ := log.ParseLogLevel(level)
	// logger := log.NewLogger(
	// 	os.Stderr,
	// 	log.LevelOption(1), // info=1
	// 	log.FilterOption(filter),
	// )
	logger := log.NewNopLogger()
	cfg := appencoding.MakeEncodingConfig(chainCfg)
	appOpts := DefaultAppOptions{}
	g, goctx, _ := GetTestCtx(logger, true)
	goctx = wasmxtypes.ContextWithBackgroundProcesses(goctx)
	goctx = networkvm.WithP2PEmptyContext(goctx)
	goctx, bapps := config.WithMultiChainAppEmpty(goctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", goctx)
	actionExecutor := networkkeeper.NewActionExecutor(bapps, logger)
	app := NewApp(
		actionExecutor,
		logger,
		db, nil, true, map[int64]bool{},
		DefaultNodeHome+strconv.Itoa(int(index)), 5, cfg, nil, appOpts,
		bam.SetChainID(chainID),
	)
	bapps.SetApp(chainID, app)
	return app, app.DefaultGenesis()
}

// NewTestNetworkFixture returns a new simapp AppConstructor for network simulation tests
func NewTestNetworkFixture() network.TestFixture {
	dir, err := os.MkdirTemp("", "mythos")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	defer os.RemoveAll(dir)

	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	appOpts := DefaultAppOptions{}
	g, goctx, _ := GetTestCtx(logger, true)
	goctx = wasmxtypes.ContextWithBackgroundProcesses(goctx)
	goctx = networkvm.WithP2PEmptyContext(goctx)
	goctx, bapps := config.WithMultiChainAppEmpty(goctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", goctx)
	actionExecutor := networkkeeper.NewActionExecutor(bapps, logger)
	chainCfg, err := config.GetChainConfig(config.MYTHOS_CHAIN_ID_TEST)
	if err != nil {
		panic(err)
	}

	app := NewApp(actionExecutor, logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, appencoding.MakeEncodingConfig(chainCfg), nil, appOpts)

	appCtr := func(val network.ValidatorI) servertypes.Application {
		// appOpts := simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir)
		appOpts.Set(flags.FlagHome, val.GetCtx().Config.RootDir)
		chainId := val.GetCtx().Viper.GetString(flags.FlagChainID)
		chainCfg, err := config.GetChainConfig(chainId)
		if err != nil {
			panic(err)
		}
		gasPricesStr := val.GetAppConfig().MinGasPrices
		gasPrices, err := sdk.ParseDecCoins(gasPricesStr)
		if err != nil {
			panic(fmt.Sprintf("invalid minimum gas prices: %v", err))
		}
		return NewApp(
			actionExecutor,
			val.GetCtx().Logger, dbm.NewMemDB(), nil, true, map[int64]bool{},
			DefaultNodeHome, 5, appencoding.MakeEncodingConfig(chainCfg),
			gasPrices,
			appOpts,
			bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			bam.SetMinGasPrices(gasPricesStr),
			bam.SetChainID(chainId),
		)
	}

	return network.TestFixture{
		AppConstructor: appCtr,
		GenesisState:   app.DefaultGenesis(),
		EncodingConfig: testutil.TestEncodingConfig{
			InterfaceRegistry: app.InterfaceRegistry(),
			Codec:             app.AppCodec(),
			TxConfig:          app.TxConfig(),
			Amino:             app.LegacyAmino(),
		},
	}
}

func GetTestCtx(logger log.Logger, block bool) (*errgroup.Group, context.Context, context.CancelFunc) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	server.ListenForQuitSignals(g, block, cancelFn, logger)
	return g, ctx, cancelFn
}
