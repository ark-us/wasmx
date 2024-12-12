package server

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/version"

	tmcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"

	multichain "github.com/loredanacirstea/wasmx/v1/multichain"
	memc "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm/memory/common"
)

// AddCommands adds server commands
func AddCommands(
	rootCmd *cobra.Command,
	defaultNodeHome string,
	appCreator types.AppCreator,
	appExport types.AppExporter,
	addStartFlags types.ModuleInitFlags,
	wasmVmMeta memc.IWasmVmMeta,
) {
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		sdkserver.ShowNodeIDCmd(),
		sdkserver.ShowValidatorCmd(),
		sdkserver.ShowAddressCmd(),
		sdkserver.VersionCmd(),
		tmcmd.ResetAllCmd,
		tmcmd.ResetStateCmd,
	)

	startCmd := StartCmd(wasmVmMeta, appCreator, defaultNodeHome)
	addStartFlags(startCmd)

	rootCmd.AddCommand(
		startCmd,
		tendermintCmd,
		sdkserver.ExportCmd(appExport, defaultNodeHome),
		version.NewVersionCommand(),
		NewRollbackCmd(appCreator, defaultNodeHome),
		NewResetChainData(appCreator, defaultNodeHome, wasmVmMeta),
	)
}

// GetServerContextFromCmd returns a Context from a command or an empty Context
// if it has not been set.
func GetServerContextFromCmd(cmd *cobra.Command) *sdkserver.Context {
	if v := cmd.Context().Value(sdkserver.ServerContextKey); v != nil {
		serverCtxPtr := v.(*sdkserver.Context)
		return serverCtxPtr
	}

	return NewDefaultContext(cmd.Flags())
}

func NewDefaultContext(flags *pflag.FlagSet) *sdkserver.Context {
	return sdkserver.NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		NewDefaultLogger(flags),
	)
}

func NewDefaultLogger(flags *pflag.FlagSet) log.Logger {
	// format, err := flags.GetString(sdkflags.FlagLogFormat)
	level, err := flags.GetString(sdkflags.FlagLogLevel)
	if err != nil && level == "" {
		level = "info"
	}
	filter, _ := multichain.ParseLogLevel(level)
	return log.NewLogger(
		os.Stdout,
		log.FilterOption(filter),
		log.TimeFormatOption(time.RFC3339),
	)
}
