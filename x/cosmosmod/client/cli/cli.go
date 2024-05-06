package cli

import (
	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"

	bank "mythos/v1/x/cosmosmod/client/cli/bank"
	gov "mythos/v1/x/cosmosmod/client/cli/gov"
	staking "mythos/v1/x/cosmosmod/client/cli/staking"
	"mythos/v1/x/cosmosmod/types"
)

func NewTxCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Cosmos transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		bank.NewTxCmd(ac),
		staking.NewTxCmd(valAddrCodec, ac),
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
