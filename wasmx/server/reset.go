package server

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	sdkdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/version"

	mapp "github.com/loredanacirstea/wasmx/app"
	mcfg "github.com/loredanacirstea/wasmx/config"
	"github.com/loredanacirstea/wasmx/multichain"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func NewResetChainData(appCreator types.AppCreator, defaultNodeHome string, wasmVmMeta memc.IWasmVmMeta) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-chain",
		Short: "Reset subchain data",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Reset subchain data

Example:
$ %s reset-chain --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1
`, version.AppName)),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			subChainId, err := cmd.Flags().GetString(sdkflags.FlagChainID)
			if err != nil {
				return fmt.Errorf("subchainId: %s", err)
			}

			ctx := GetServerContextFromCmd(cmd)
			cfg := ctx.Config
			home := cfg.RootDir
			db, err := openDB(home, sdkserver.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}
			g, goctx := getCtx(serverCtx, true)

			apictx := &APICtx{
				GoRoutineGroup:  g,
				GoContextParent: goctx,
				SvrCtx:          &sdkserver.Context{},
				ClientCtx:       clientCtx,
			}
			_, appCreator := mapp.NewAppCreator(wasmVmMeta, ctx.Logger, db, nil, ctx.Viper, g, goctx, apictx)

			chainCfg, _ := mcfg.GetChainConfig(mcfg.MYTHOS_CHAIN_ID_TESTNET)
			iapp := appCreator(subChainId, chainCfg)
			app := iapp.(*mapp.App)

			return resetStoresToVersion0(app, db)

		},
		SilenceUsage: true,
	}

	sdkflags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func resetStoresToVersion0(app mcfg.MythosApp, db sdkdb.DB) error {
	chainId := app.GetBaseApp().ChainID()
	cms := app.GetBaseApp().CommitMultiStore()
	fmt.Printf("resetting chain database: %s , version: %d \n", chainId, cms.LatestVersion())

	err := app.GetBaseApp().ResetStores()
	if err != nil {
		return err
	}

	// remove database entries for this chain
	// s/k:CHAIN_ID
	// s/latest/CHAIN_ID
	// s/CHAIN_ID
	chainIdBz := []byte(chainId)
	prefixlen := len(chainIdBz) + 15 // should account for any prefix + chain_id
	count := 0
	batch := db.NewBatch()
	defer func() {
		err = batch.Close()
	}()

	itr, _ := db.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		if prefixlen < len(key) {
			key = key[0:prefixlen]
		}
		if bytes.Contains(key, chainIdBz) {
			batch.Delete(itr.Key())
			count += 1
		}
	}
	err = batch.WriteSync()
	if err != nil {
		return err
	}

	cms.SetInitialVersion(0)

	app.GetBaseApp().Logger().Info("chain reset to version 0", "removed_db_keys", count, "latest_version", cms.LatestVersion())
	return nil
}
