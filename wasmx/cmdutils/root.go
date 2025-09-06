package cmdutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	pruningtypes "cosmossdk.io/store/pruning/types"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	sdksigning "cosmossdk.io/x/tx/signing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	tmcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	tmcfg "github.com/cometbft/cometbft/config"
	tmcli "github.com/cometbft/cometbft/libs/cli"

	// this line is used by starport scaffolding # root/moduleImport

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	mapi "github.com/loredanacirstea/wasmx/apictx"
	app "github.com/loredanacirstea/wasmx/app"
	mcodec "github.com/loredanacirstea/wasmx/codec"
	mcfg "github.com/loredanacirstea/wasmx/config"
	mctx "github.com/loredanacirstea/wasmx/context"
	appencoding "github.com/loredanacirstea/wasmx/encoding"
	"github.com/loredanacirstea/wasmx/multichain"
	server "github.com/loredanacirstea/wasmx/server"
	serverconfig "github.com/loredanacirstea/wasmx/server/config"
	cosmosmodtypes "github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	"github.com/loredanacirstea/wasmx/x/network/vmp2p"
	"github.com/loredanacirstea/wasmx/x/vmhttpserver"
	"github.com/loredanacirstea/wasmx/x/vmkv"
	"github.com/loredanacirstea/wasmx/x/vmsql"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// NewRootCmd creates a new root command for a Cosmos SDK application
func NewRootCmd(wasmVmMeta memc.IWasmVmMeta, defaultNodeHome string, initializeDb func(rootDir string, backendType dbm.BackendType) (dbm.DB, error)) (*cobra.Command, appencoding.EncodingConfig) {
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	chainId := mcfg.MYTHOS_CHAIN_ID_TESTNET
	chainCfg, err := mcfg.GetChainConfig(chainId)
	if err != nil {
		panic(err)
	}
	encodingConfig := appencoding.MakeEncodingConfig(chainCfg, app.GetCustomSigners())
	addrcodec := mcodec.MustUnwrapAccBech32Codec(encodingConfig.TxConfig.SigningContext().AddressCodec())
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(cosmosmodtypes.AccountRetriever{AddressCodec: addrcodec}).
		WithHomeDir(defaultNodeHome).
		WithViper("")
	logger := log.NewNopLogger()
	appOpts := multichain.DefaultAppOptions{}
	g, goctx, _ := multichain.GetTestCtx(logger, true)
	goctx = wasmxtypes.ContextWithBackgroundProcesses(goctx)
	goctx = vmp2p.WithP2PEmptyContext(goctx)
	goctx = networktypes.ContextWithMultiChainContext(g, goctx, logger)
	goctx, _ = mcfg.WithMultiChainAppEmpty(goctx)
	goctx, _ = mctx.WithExecutionMetaInfoEmpty(goctx)
	goctx, _ = mctx.WithTimeoutGoroutinesInfoEmpty(goctx)
	goctx, _ = wasmxtypes.WithSystemBootstrap(goctx)
	goctx = vmsql.WithSqlEmptyContext(goctx)
	goctx = vmkv.WithKvDbEmptyContext(goctx)
	goctx = vmimap.WithImapEmptyContext(goctx)
	goctx = vmsmtp.WithSmtpEmptyContext(goctx)
	goctx = vmhttpserver.WithHttpServerEmptyContext(goctx)
	wasmVmMeta.InitWasmRuntime(goctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", goctx)
	appOpts.Set(flags.FlagHome, tempDir(defaultNodeHome))
	appOpts.Set(flags.FlagChainID, chainId)
	appOpts.Set(sdkserver.FlagPruning, pruningtypes.PruningOptionDefault)
	baseappOptions := mcfg.DefaultBaseappOptions(appOpts)
	tempOpts := simtestutil.NewAppOptionsWithFlagHome(tempDir(defaultNodeHome))
	tempApp := app.NewApp(
		chainId,
		logger,
		dbm.NewMemDB(),
		nil, true, make(map[int64]bool, 0),
		cast.ToString(tempOpts.Get(flags.FlagHome)),
		cast.ToUint(tempOpts.Get(sdkserver.FlagInvCheckPeriod)), chainCfg, encodingConfig, nil, appOpts,
		wasmVmMeta,
		// tempBaseappOptions...,
		// baseapp.SetChainID(mcfg.MYTHOS_CHAIN_ID_TESTNET),
		baseappOptions...,
	)
	defer tempApp.Teardown()
	rootCmd := &cobra.Command{
		Use:   mcfg.Name + "d",
		Short: "Start mythos node",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}
			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL.
			enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
			txConfigOpts := tx.ConfigOptions{
				EnabledSignModes:           enabledSignModes,
				TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(initClientCtx),
			}
			txConfigWithTextual, err := tx.NewTxConfigWithOptions(
				codec.NewProtoCodec(encodingConfig.InterfaceRegistry),
				txConfigOpts,
			)
			if err != nil {
				return err
			}
			initClientCtx = initClientCtx.WithTxConfig(txConfigWithTextual)

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := serverconfig.AppConfig()
			customTMConfig := initTendermintConfig()

			serverCtx, err := InterceptConfigsAndCreateContext(cmd, customAppTemplate, customAppConfig, customTMConfig)
			if err != nil {
				return err
			}
			logger = server.NewDefaultLogger(cmd.Flags())
			serverCtx.Logger = logger.With(log.ModuleKey, "server")
			return sdkserver.SetCmdServerContext(cmd, serverCtx)
		},
	}

	apictx := &mapi.APICtx{
		GoRoutineGroup:  g,
		GoContextParent: goctx,
		SvrCtx:          &sdkserver.Context{},
		ClientCtx:       initClientCtx,
	}

	initRootCmd(wasmVmMeta, rootCmd, encodingConfig, tempApp.BasicModuleManager, g, goctx, apictx, initClientCtx, defaultNodeHome, initializeDb)

	// add keyring to autocli opts
	autoCliOpts := tempApp.AutoCliOpts()
	autoCliOpts.Keyring, _ = keyring.NewAutoCLIKeyring(initClientCtx.Keyring)
	autoCliOpts.ClientCtx = initClientCtx

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        strings.ReplaceAll(mcfg.Name, "-", ""),
		flags.FlagKeyringBackend: "test",
	})

	return rootCmd, encodingConfig
}

// initTendermintConfig helps to override default Tendermint Config values.
// return tmcfg.DefaultConfig if no custom configuration is required for the application.
func initTendermintConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()
	return cfg
}

// MigrationMap is a map of SDK versions to their respective genesis migration functions.
var MigrationMap = genutiltypes.MigrationMap{}

func initRootCmd(
	wasmVmMeta memc.IWasmVmMeta,
	rootCmd *cobra.Command,
	encodingConfig appencoding.EncodingConfig,
	basicManager module.BasicManager,
	g *errgroup.Group,
	ctx context.Context,
	apictx mcfg.APICtxI,
	clientCtx client.Context,
	defaultNodeHome string,
	initializeDb func(rootDir string, backendType dbm.BackendType) (dbm.DB, error),
) {
	gentxModule := basicManager[genutiltypes.ModuleName].(genutil.AppModuleBasic)

	a := appCreator{
		wasmVmMeta,
		encodingConfig,
		g,
		ctx,
		apictx,
		clientCtx,
		rootCmd,
	}

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, defaultNodeHome),
		genutilcli.CollectGenTxsCmd(cosmosmodtypes.GenesisBalancesIterator{}, defaultNodeHome, gentxModule.GenTxValidator, encodingConfig.TxConfig.SigningContext().ValidatorAddressCodec()),
		genutilcli.MigrateGenesisCmd(MigrationMap),
		genutilcli.GenTxCmd(
			basicManager,
			encodingConfig.TxConfig,
			cosmosmodtypes.GenesisBalancesIterator{},
			defaultNodeHome,
			encodingConfig.TxConfig.SigningContext().ValidatorAddressCodec(),
		),
		genutilcli.ValidateGenesisCmd(basicManager),
		AddGenesisAccountCmd(defaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		NewTestnetCmd(wasmVmMeta, basicManager, cosmosmodtypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(a.newApp, defaultNodeHome),
		snapshot.Cmd(a.newApp),
	)

	// add server commands
	server.AddCommands(
		rootCmd,
		defaultNodeHome,
		a.newApp,
		a.appExport,
		addModuleInitFlags,
		wasmVmMeta,
		initializeDb,
	)
	extendUnsafeResetAllCmd(rootCmd)

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		sdkserver.StatusCommand(),
		genesisCommand(encodingConfig.TxConfig, basicManager, defaultNodeHome),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)
}

// genesisCommand builds genesis-related `simd genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(txConfig client.TxConfig, basicManager module.BasicManager, defaultNodeHome string, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.Commands(txConfig, basicManager, defaultNodeHome)

	for _, subCmd := range cmds {
		cmd.AddCommand(subCmd)
	}
	return cmd
}

// queryCommand returns the sub-command to send queries to the app
func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.QueryEventForTxCmd(),
		sdkserver.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		sdkserver.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		sdkserver.QueryBlockResultsCmd(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// txCommand returns the sub-command to send transactions to the app
func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	// this line is used by starport scaffolding # root/arguments
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			f.Value.Set(val)
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

type appCreator struct {
	wasmVmMeta     memc.IWasmVmMeta
	encodingConfig appencoding.EncodingConfig
	g              *errgroup.Group
	ctx            context.Context
	apictx         mcfg.APICtxI
	clientCtx      client.Context
	cmd            *cobra.Command
}

// newApp creates a new Cosmos SDK app
func (a appCreator) newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	_, appCreator := app.NewAppCreator(a.wasmVmMeta, logger, db, traceStore, appOpts.(multichain.AppOptions), a.g, a.ctx, a.apictx)

	chainId := mcfg.GetChainId(appOpts)
	registryId := cast.ToString(appOpts.Get(multichain.FlagRegistryChainId))
	_, _, config, err := multichain.MultiChainCtx(a.clientCtx, []sdksigning.CustomGetSigner{}, chainId, registryId)
	if err != nil {
		panic(err)
	}
	mapp := appCreator(chainId, config)
	return mapp.(*app.App)
}

// appExport creates a new simapp (optionally at a given height)
func (a appCreator) appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(sdkserver.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	chainId := mcfg.GetChainId(appOpts)
	gasPricesStr := cast.ToString(appOpts.Get(sdkserver.FlagMinGasPrices))
	gasPrices, err := sdk.ParseDecCoins(gasPricesStr)
	if err != nil {
		panic(fmt.Sprintf("invalid minimum gas prices: %v", err))
	}
	minGasAmount := math.LegacyNewDec(0)
	if len(gasPrices) > 0 {
		minGasAmount = gasPrices[0].Amount
	}
	chainCfg, err := mcfg.GetChainConfig(chainId)
	if err != nil {
		panic(err)
	}
	minGasPrices := sdk.NewDecCoins(sdk.NewDecCoin(chainCfg.BaseDenom, minGasAmount.RoundInt()))

	app := app.NewApp(
		chainId,
		logger,
		db,
		traceStore,
		height == -1, // -1: no height provided
		map[int64]bool{},
		homePath,
		uint(1),
		chainCfg,
		a.encodingConfig,
		minGasPrices,
		appOpts,
		a.wasmVmMeta,
	)

	if height != -1 {
		if err := app.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// extendUnsafeResetAllCmd - also clear wasm dir
func extendUnsafeResetAllCmd(rootCmd *cobra.Command) {
	unsafeResetCmd := tmcmd.ResetAllCmd.Use
	for _, branchCmd := range rootCmd.Commands() {
		if branchCmd.Use != "tendermint" {
			continue
		}
		for _, cmd := range branchCmd.Commands() {
			if cmd.Use == unsafeResetCmd {
				serverRunE := cmd.RunE
				cmd.RunE = func(cmd *cobra.Command, args []string) error {
					if err := serverRunE(cmd, args); err != nil {
						return nil
					}
					serverCtx := server.GetServerContextFromCmd(cmd)
					return os.RemoveAll(filepath.Join(serverCtx.Config.RootDir, wasmxtypes.ContractsDir))
				}
				return
			}
		}
	}
}

var tempDir = func(defaultNodeHome string) string {
	dir, err := os.MkdirTemp("", "mythos")
	if err != nil {
		dir = defaultNodeHome
	}
	defer os.RemoveAll(dir)

	return dir
}
