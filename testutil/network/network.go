package network

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	app "mythos/v1/app"
	mcodec "mythos/v1/codec"
	config "mythos/v1/config"
	appencoding "mythos/v1/encoding"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	"mythos/v1/x/network/vmp2p"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type (
	Network = network.Network
	Config  = network.Config
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() network.Config {
	chainId := config.MYTHOS_CHAIN_ID_TEST
	chainCfg, err := config.GetChainConfig(chainId)
	if err != nil {
		panic(err)
	}
	encoding := appencoding.MakeEncodingConfig(chainCfg)
	logger := log.NewNopLogger()

	appOpts := app.DefaultAppOptions{}
	appOpts.Set(flags.FlagHome, tempDir())
	appOpts.Set(sdkserver.FlagInvCheckPeriod, 1)
	g, goctx, _ := app.GetTestCtx(logger, true)
	goctx = wasmxtypes.ContextWithBackgroundProcesses(goctx)
	goctx = vmp2p.WithP2PEmptyContext(goctx)
	goctx, _ = config.WithMultiChainAppEmpty(goctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", goctx)

	tempApp := app.NewApp(
		logger,
		dbm.NewMemDB(),
		nil, true, make(map[int64]bool, 0),
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(sdkserver.FlagInvCheckPeriod)), encoding, nil, appOpts)

	addrcodec := mcodec.MustUnwrapAccBech32Codec(encoding.TxConfig.SigningContext().AddressCodec())

	return network.Config{
		Codec:             encoding.Marshaler,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  cosmosmodtypes.AccountRetriever{AddressCodec: addrcodec},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			gasPricesStr := val.GetAppConfig().MinGasPrices
			gasPrices, err := sdk.ParseDecCoins(gasPricesStr)
			if err != nil {
				panic(fmt.Sprintf("invalid minimum gas prices: %v", err))
			}

			return app.NewApp(
				val.GetCtx().Logger, dbm.NewMemDB(), nil, true, map[int64]bool{}, val.GetCtx().Config.RootDir, 0,
				encoding,
				gasPrices,
				simtestutil.EmptyAppOptions{},
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
				baseapp.SetMinGasPrices(gasPricesStr),
			)
		},
		GenesisState:    tempApp.BasicModuleManager.DefaultGenesis(encoding.Marshaler),
		TimeoutCommit:   2 * time.Second,
		ChainID:         chainId,
		NumValidators:   1,
		BondDenom:       config.BondBaseDenom,
		MinGasPrices:    fmt.Sprintf("1%s", config.BaseDenom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy: pruningtypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
	}
}

// Logger is a network logger interface that exposes testnet-level Log() methods for an in-process testing network
// This is not to be confused with logging that may happen at an individual node or validator level
type Logger interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

var (
	_ Logger = (*testing.T)(nil)
	_ Logger = (*CLILogger)(nil)
)

type CLILogger struct {
	cmd *cobra.Command
}

func (s CLILogger) Log(args ...interface{}) {
	s.cmd.Println(args...)
}

func (s CLILogger) Logf(format string, args ...interface{}) {
	s.cmd.Printf(format, args...)
}

func NewCLILogger(cmd *cobra.Command) CLILogger {
	return CLILogger{cmd}
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", "mythos")
	if err != nil {
		dir = app.DefaultNodeHome
	}
	defer os.RemoveAll(dir)

	return dir
}
