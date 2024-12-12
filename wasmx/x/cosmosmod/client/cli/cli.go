package cli

import (
	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"

	bank "github.com/loredanacirstea/wasmx/x/cosmosmod/client/cli/bank"
	gov "github.com/loredanacirstea/wasmx/x/cosmosmod/client/cli/gov"
	staking "github.com/loredanacirstea/wasmx/x/cosmosmod/client/cli/staking"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"

	"github.com/loredanacirstea/wasmx/multichain"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func NewTxCmd(wasmVmMeta memc.IWasmVmMeta, valAddrCodec, ac address.Codec, appCreator multichain.NewAppCreator) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Cosmos transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		bank.NewTxCmd(ac),
		staking.NewTxCmd(wasmVmMeta, valAddrCodec, ac, appCreator),
		gov.NewTxCmd(ac),
	)

	return txCmd
}

func GetQueryCmd(ac address.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Cosmos query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		bank.GetQueryCmd(ac),
		gov.GetCustomQueryCmd(ac),
	)

	return txCmd
}
