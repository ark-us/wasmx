package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	config "github.com/loredanacirstea/wasmx/config"
	mcfg "github.com/loredanacirstea/wasmx/config"
	mctx "github.com/loredanacirstea/wasmx/context"
	appencoding "github.com/loredanacirstea/wasmx/encoding"
	menc "github.com/loredanacirstea/wasmx/encoding"
	"github.com/loredanacirstea/wasmx/multichain"
	srvconfig "github.com/loredanacirstea/wasmx/server/config"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// DefaultTestingAppInit defines the IBC application used for testing
var DefaultTestingAppInit func(wasmVmMeta memc.IWasmVmMeta, chainId string, chainCfg *appencoding.ChainConfig, index int32) (ibctesting.TestingApp, map[string]json.RawMessage) = SetupTestingApp

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

// Setup initializes a new Mythos. A Nop logger is set in Mythos.
func SetupApp(
	wasmVmMeta memc.IWasmVmMeta,
	isCheckTx bool,
) *App {
	chainId := config.MYTHOS_CHAIN_ID_TEST
	chainCfg, err := config.GetChainConfig(chainId)
	if err != nil {
		panic(err)
	}
	_, appCreator := multichain.CreateMockAppCreator(wasmVmMeta, NewAppCreator, DefaultNodeHome)
	iapp := appCreator(chainId, chainCfg)
	app := iapp.(*App)

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
func SetupTestingApp(wasmVmMeta memc.IWasmVmMeta, chainID string, chainCfg *appencoding.ChainConfig, index int32) (ibctesting.TestingApp, map[string]json.RawMessage) {
	_, appCreator := multichain.CreateMockAppCreator(wasmVmMeta, NewAppCreator, DefaultNodeHome+strconv.Itoa(int(index)))
	iapp := appCreator(chainID, chainCfg)
	app := iapp.(*App)
	return app, app.DefaultGenesis()
}

// NewTestNetworkFixture returns a new simapp AppConstructor for network simulation tests
func NewTestNetworkFixture(wasmVmMeta memc.IWasmVmMeta) func() network.TestFixture {
	return func() network.TestFixture {
		dir, err := os.MkdirTemp("", "mythos")
		if err != nil {
			panic(fmt.Sprintf("failed creating temporary directory: %v", err))
		}
		defer os.RemoveAll(dir)

		db := dbm.NewMemDB()
		logger := log.NewNopLogger()
		chainId := config.MYTHOS_CHAIN_ID_TEST
		appOpts := multichain.DefaultAppOptions{}
		appOpts.Set(flags.FlagHome, DefaultNodeHome)
		appOpts.Set(flags.FlagChainID, chainId)
		appOpts.Set(sdkserver.FlagInvCheckPeriod, 5)
		appOpts.Set(sdkserver.FlagUnsafeSkipUpgrades, 0)
		appOpts.Set(sdkserver.FlagMinGasPrices, "")
		appOpts.Set(sdkserver.FlagPruning, pruningtypes.PruningOptionDefault)
		g, goctx, _ := multichain.GetTestCtx(logger, true)

		chainCfg, err := config.GetChainConfig(chainId)
		if err != nil {
			panic(err)
		}
		_, appCreator := NewAppCreator(wasmVmMeta, logger, db, nil, appOpts, g, goctx, &multichain.MockApiCtx{})
		iapp := appCreator(chainId, chainCfg)
		app := iapp.(*App)

		appCtr := func(val network.ValidatorI) servertypes.Application {
			chainId := val.GetCtx().Viper.GetString(flags.FlagChainID)
			chainCfg, err := config.GetChainConfig(chainId)
			if err != nil {
				panic(err)
			}
			gasPricesStr := val.GetAppConfig().MinGasPrices
			// appOpts := simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir)
			appOpts.Set(flags.FlagHome, val.GetCtx().Config.RootDir)
			appOpts.Set(flags.FlagChainID, chainId)
			appOpts.Set(sdkserver.FlagMinGasPrices, gasPricesStr)
			appOpts.Set(flags.FlagHome, val.GetCtx().Config.RootDir)
			appOpts.Set(sdkserver.FlagPruning, val.GetAppConfig().Pruning)
			// bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),

			_, appCreator := NewAppCreator(wasmVmMeta, val.GetCtx().Logger, db, nil, appOpts, g, goctx, &multichain.MockApiCtx{})
			iapp := appCreator(chainId, chainCfg)
			app := iapp.(*App)
			return app
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
}

func NopStartChainApis(
	chainId string,
	chainCfg *menc.ChainConfig,
	ports mctx.NodePorts,
) (mcfg.MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error) {
	return nil, nil, client.Context{}, nil, nil, nil, nil
}
