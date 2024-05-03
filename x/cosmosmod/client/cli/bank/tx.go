package bank

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	cli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	appencoding "mythos/v1/encoding"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
)

// NewTxCmd returns a root CLI command handler for all x/bank transaction commands.
func NewTxCmd(ac address.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSendTxCmd(ac),
		NewMultiSendTxCmd(ac),
	)

	return txCmd
}

// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from_key_or_address] [to_address] [amount]",
		Short: "Send funds from one account to another.",
		Long: `Send funds from one account to another.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			chainId := clientCtx.ChainID
			mcfg.SetGlobalChainConfig(chainId)
			config, err := mcfg.GetChainConfig(chainId)
			addrCodec := mcodec.MustUnwrapAccBech32Codec(ac)

			customEncoding := appencoding.MakeEncodingConfig(config)

			customCdc := mcodec.NewAccBech32Codec(config.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
			customAddrCodec := mcodec.MustUnwrapAccBech32Codec(customCdc)

			clientCtx = clientCtx.WithAccountRetriever(cosmosmodtypes.NewAccountRetriever(chainId, customAddrCodec))
			// TODO clientCtx.WithCodec(customCdc) ??

			clientCtx = clientCtx.
				WithCodec(customEncoding.Marshaler).
				WithInterfaceRegistry(customEncoding.InterfaceRegistry).
				WithTxConfig(customEncoding.TxConfig).
				WithLegacyAmino(customEncoding.Amino)

			toAddr_, err := addrCodec.StringToAccAddressPrefixed(args[1])
			if err != nil {
				return err
			}
			toAddr := customAddrCodec.BytesToAccAddressPrefixed(toAddr_.Bytes())
			fromAddr := customAddrCodec.BytesToAccAddressPrefixed(clientCtx.GetFromAddress())


			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			if len(coins) == 0 {
				return fmt.Errorf("invalid coins")
			}

			msg := &types.MsgSend{
				FromAddress: fromAddr.String(),
				ToAddress:   toAddr.String(),
				Amount:      coins,
			}

			msgAny, err := codectypes.NewAnyWithValue(msg)
			if err != nil {
				return err
			}
			// TODO only if diff chain than context chainId
			msgMultiChain := &networktypes.MsgMultiChainWrap{
				MultiChainId: chainId,
				Sender:       fromAddr.String(),
				Data:         msgAny,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgMultiChain)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// For a better UX this command is limited to send funds from one account to two or more accounts.
func NewMultiSendTxCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send [from_key_or_address] [to_address_1 to_address_2 ...] [amount]",
		Short: "Send funds from one account to two or more accounts.",
		Long: `Send funds from one account to two or more accounts.
By default, sends the [amount] to each address of the list.
Using the '--split' flag, the [amount] is split equally between the addresses.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address] and
separate addresses with space.
When using '--dry-run' a key name cannot be used, only a bech32 address.`,
		Example: fmt.Sprintf("%s tx bank multi-send cosmos1... cosmos1... cosmos1... cosmos1... 10stake", version.AppName),
		Args:    cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[len(args)-1])
			if err != nil {
				return err
			}

			if coins.IsZero() {
				return fmt.Errorf("must send positive amount")
			}

			split, err := cmd.Flags().GetBool(cli.FlagSplit)
			if err != nil {
				return err
			}

			totalAddrs := sdkmath.NewInt(int64(len(args) - 2))
			// coins to be received by the addresses
			sendCoins := coins
			if split {
				sendCoins = coins.QuoInt(totalAddrs)
			}

			var output []types.Output
			for _, arg := range args[1 : len(args)-1] {
				toAddr, err := ac.StringToBytes(arg)
				if err != nil {
					return err
				}

				output = append(output, types.NewOutput(toAddr, sendCoins))
			}

			// amount to be send from the from address
			var amount sdk.Coins
			if split {
				// user input: 1000stake to send to 3 addresses
				// actual: 333stake to each address (=> 999stake actually sent)
				amount = sendCoins.MulInt(totalAddrs)
			} else {
				amount = coins.MulInt(totalAddrs)
			}

			msg := types.NewMsgMultiSend(types.NewInput(clientCtx.FromAddress, amount), output)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(cli.FlagSplit, false, "Send the equally split token amount to each address")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
