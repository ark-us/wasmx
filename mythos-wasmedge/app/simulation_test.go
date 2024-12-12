package app_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	app "github.com/loredanacirstea/wasmx/v1/app"
	mcfg "github.com/loredanacirstea/wasmx/v1/config"
	multichain "github.com/loredanacirstea/wasmx/v1/multichain"

	runtime "github.com/loredanacirstea/wasmx-wasmedge"
)

func init() {
	simcli.GetSimulatorFlags()
}

// BenchmarkSimulation run the chain simulation
// Running using starport command:
// `ignite chain simulate -v --numBlocks 200 --blockSize 50`
// Running as go benchmark test:
// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50 -Commit=true -Verbose=true -Enabled=true`
func BenchmarkSimulation(b *testing.B) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = mcfg.MYTHOS_CHAIN_ID_TEST
	cfg, err := mcfg.GetChainConfig(config.ChainID)
	require.NoError(b, err)
	simcli.FlagEnabledValue = true
	simcli.FlagCommitValue = true

	db, dir, logger, _, err := simtestutil.SetupSimulation(config, "goleveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(b, err, "simulation setup failed")

	b.Cleanup(func() {
		db.Close()
		err = os.RemoveAll(dir)
		require.NoError(b, err)
	})

	appOpts := multichain.DefaultAppOptions{}
	appOpts.Set(flags.FlagHome, app.DefaultNodeHome)
	appOpts.Set(sdkserver.FlagInvCheckPeriod, 0)
	appOpts.Set(sdkserver.FlagUnsafeSkipUpgrades, 0)
	appOpts.Set(sdkserver.FlagMinGasPrices, "")
	g, goctx, _ := multichain.GetTestCtx(logger, true)

	chainId := config.ChainID
	_, appCreator := app.NewAppCreator(runtime.WasmEdgeVmMeta{}, logger, db, nil, appOpts, g, goctx, &multichain.MockApiCtx{})
	iapp := appCreator(chainId, cfg)
	app := iapp.(*app.App)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simulationtypes.RandomAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(b, err)
	require.NoError(b, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}
