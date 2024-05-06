package bank

import (
	"context"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	mcodec "mythos/v1/codec"
	"mythos/v1/multichain"
	networktypes "mythos/v1/x/network/types"
)

func GetQueryCmd(ac address.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        banktypes.ModuleName,
		Short:                      "Bank query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	queryCmd.AddCommand(
		GetCmdGetBalance(ac),
	)
	return queryCmd
}

// GetCmdLibVersion gets current wasmedge version.
func GetCmdGetBalance(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance [owner_address] [denom]",
		Short: "Balance",
		Long:  `Balance`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			clientCtx, _, customAddrCodec, err := multichain.MultiChainCtx(ac, clientCtx)
			if err != nil {
				return err
			}
			ownerAddr_, err := customAddrCodec.StringToAddressPrefixedUnsafe(args[0])
			if err != nil {
				return err
			}
			ownerAddr := customAddrCodec.BytesToAccAddressPrefixed(ownerAddr_.Bytes())

			query := &banktypes.QueryBalanceRequest{
				Address: ownerAddr.String(),
				Denom:   args[1],
			}
			querybz, err := query.Marshal()
			if err != nil {
				return err
			}
			typeUrl := sdk.MsgTypeURL(query)
			queryPath, err := mcodec.GetPathFromType(typeUrl)
			if err != nil {
				return err
			}

			abciQuery := &abci.RequestQuery{Data: querybz, Path: queryPath}
			abciQueryBz, err := abciQuery.Marshal()
			if err != nil {
				return err
			}

			queryClient := networktypes.NewQueryClient(clientCtx)
			res, err := queryClient.QueryMultiChain(context.Background(), &networktypes.QueryMultiChainRequest{
				MultiChainId: clientCtx.ChainID,
				QueryData:    abciQueryBz,
			})
			if err != nil {
				return err
			}
			var response banktypes.QueryBalanceResponse
			err = response.Unmarshal(res.Data)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(&response)
		},
		SilenceUsage: true,
	}
	return cmd
}
