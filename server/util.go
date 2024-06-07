package server

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/version"

	tmcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
)

// AddCommands adds server commands
func AddCommands(
	rootCmd *cobra.Command,
	defaultNodeHome string,
	appCreator types.AppCreator,
	appExport types.AppExporter,
	addStartFlags types.ModuleInitFlags,
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

	startCmd := StartCmd(appCreator, defaultNodeHome)
	addStartFlags(startCmd)

	rootCmd.AddCommand(
		startCmd,
		tendermintCmd,
		sdkserver.ExportCmd(appExport, defaultNodeHome),
		version.NewVersionCommand(),
		sdkserver.NewRollbackCmd(appCreator, defaultNodeHome),
	)
}

// GetServerContextFromCmd returns a Context from a command or an empty Context
// if it has not been set.
func GetServerContextFromCmd(cmd *cobra.Command) *sdkserver.Context {
	if v := cmd.Context().Value(sdkserver.ServerContextKey); v != nil {
		serverCtxPtr := v.(*sdkserver.Context)
		return serverCtxPtr
	}

	return NewDefaultContext()
}

func NewDefaultContext() *sdkserver.Context {
	return sdkserver.NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		NewDefaultLogger(),
	)
}

func NewDefaultLogger() log.Logger {
	return log.NewLogger(
		os.Stdout,
		log.TimeFormatOption(time.RFC3339),
	)
}
