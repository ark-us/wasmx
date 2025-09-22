package server

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mapp "github.com/loredanacirstea/wasmx/app"
	"github.com/loredanacirstea/wasmx/multichain"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// NewRollbackCmd creates a command to rollback block, and multistore state by one height.
func NewRollbackCmd(appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	// var removeBlock bool
	var blockHeightPtr *int64

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
			baseapp := app.(*mapp.App)
			height := baseapp.LastBlockHeight()
			newheight := height - 1
			blockHeight := int64(0)
			if blockHeightPtr != nil {
				blockHeight = *blockHeightPtr
			}
			if blockHeight == int64(0) {
				blockHeight = height
			}
			if blockHeight < height {
				return fmt.Errorf("can only roll back last block: %d", height)
			}
			if blockHeight == height {
				fmt.Printf("rolling back to version: %d \n", newheight)
				// rollback the multistore
				if err := app.CommitMultiStore().RollbackToVersion(newheight); err != nil {
					return fmt.Errorf("failed to rollback to version: %w", err)
				}
			}

			height = baseapp.LastBlockHeight()
			hash := baseapp.LastCommitID().Hash
			fmt.Printf("Rolled back state to height %d and hash %X", height, hash)

			_, goctx := getCtx(ctx, true)
			cb := func(goctx context.Context) (any, error) {
				ctx := sdk.UnwrapSDKContext(goctx)
				msg := []byte(fmt.Sprintf(`{"execute":{"action": {"type": "rollback", "params": [{"key":"height","value":"%d"}],"event":null}}}`, blockHeight))
				execmsg := &networktypes.MsgExecuteContract{
					Sender:   wasmxtypes.ROLE_CONSENSUS,
					Contract: wasmxtypes.ROLE_CONSENSUS,
					Msg:      msg,
				}
				res, err := baseapp.NetworkKeeper.ExecuteContractInternal(ctx, execmsg)
				if err != nil {
					return nil, err
				}
				return res, nil
			}

			_, err = baseapp.GetActionExecutor().Execute(goctx, height, sdk.ExecModeFinalize, cb)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	// cmd.Flags().BoolVar(&removeBlock, "hard", false, "remove last block as well as state")
	blockHeightPtr = cmd.Flags().Int64("height", 0, "block height to remove, optional")
	cmd.Flags().String(flags.FlagChainID, "testnet", "Specify Chain ID for sending Tx")
	cmd.Flags().String(multichain.FlagRegistryChainId, "", "multichain registry chain id")
	return cmd
}
