package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"

	mapp "github.com/loredanacirstea/wasmx/v1/app"
	"github.com/loredanacirstea/wasmx/v1/multichain"
)

// NewRollbackCmd creates a command to rollback CometBFT and multistore state by one height.
func NewRollbackCmd(appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	var removeBlock bool

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "rollback Cosmos SDK and CometBFT state by one height",
		Long: `
A state rollback is performed to recover from an incorrect application state transition,
when CometBFT has persisted an incorrect app hash and is thus unable to make
progress. Rollback overwrites a state at height n with the state at height n - 1.
The application also rolls back to height n - 1. No blocks are removed, so upon
restarting CometBFT the transactions in block n will be re-executed against the
application.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := GetServerContextFromCmd(cmd)
			cfg := ctx.Config
			home := cfg.RootDir
			db, err := openDB(home, sdkserver.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}
			app := appCreator(ctx.Logger, db, nil, ctx.Viper)
			// TODO maybe rollback blocks too?
			// // rollback CometBFT state
			// height, hash, err := cmtcmd.RollbackState(ctx.Config, removeBlock)
			// if err != nil {
			// 	return fmt.Errorf("failed to rollback CometBFT state: %w", err)
			// }

			baseapp := app.(*mapp.App)
			height := baseapp.LastBlockHeight() - 1
			fmt.Printf("rolling back to version: %d \n", height)

			// rollback the multistore
			if err := app.CommitMultiStore().RollbackToVersion(height); err != nil {
				return fmt.Errorf("failed to rollback to version: %w", err)
			}

			height = baseapp.LastBlockHeight()
			hash := baseapp.LastCommitID().Hash
			fmt.Printf("Rolled back state to height %d and hash %X", height, hash)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().BoolVar(&removeBlock, "hard", false, "remove last block as well as state")
	cmd.Flags().String(flags.FlagChainID, "testnet", "Specify Chain ID for sending Tx")
	cmd.Flags().String(multichain.FlagRegistryChainId, "", "multichain registry chain id")
	return cmd
}
