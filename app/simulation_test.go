package app_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	app "mythos/v1/app"
	mcfg "mythos/v1/config"
	appencoding "mythos/v1/encoding"
	networkkeeper "mythos/v1/x/network/keeper"
	networkvm "mythos/v1/x/network/vm"
	wasmxtypes "mythos/v1/x/wasmx/types"
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

	encoding := appencoding.MakeEncodingConfig(cfg)

	appOpts := app.DefaultAppOptions{}
	g, goctx, _ := app.GetTestCtx(logger, true)
	goctx = wasmxtypes.ContextWithBackgroundProcesses(goctx)
	goctx = networkvm.WithP2PEmptyContext(goctx)
	goctx, bapps := mcfg.WithMultiChainAppEmpty(goctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", goctx)

	actionExecutor := networkkeeper.NewActionExecutor(bapps, logger)

	app := app.NewApp(
		actionExecutor,
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		appOpts,
	)

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
