package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxcli "mythos/v1/x/wasmx/client/cli"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

const (
	flagAmount = "amount"
	FlagFrom   = "from"
)

func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "multichain",
		Short:                      "multichain transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		MultiChainTxExecuteCmd(),
	)

	return txCmd
}

func GetQueryCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "multichain",
		Short:                      "multichain query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdQueryMultiChainCall(),
		// GetCmdGetMultiChainContractState(),
	)

	return txCmd
}

// NewMultiChainTxExecuteCmd returns a CLI command handler for creating a MsgMultiChainWrap transaction.
func MultiChainTxExecuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute [contract_addr_bech32] [json_encoded_send_args] --amount [coins,optional]",
		Short:   "Execute a command on a wasm contract",
		Aliases: []string{"run", "call", "exec", "ex", "e"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := parseExecuteArgs(args[0], args[1], clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseExecuteArgs(contractAddr string, execMsg string, sender sdk.AccAddress, flags *flag.FlagSet) (types.MsgMultiChainWrap, error) {
	amountStr, err := flags.GetString(flagAmount)
	if err != nil {
		return types.MsgMultiChainWrap{}, fmt.Errorf("amount: %s", err)
	}
	chainId, err := flags.GetString(sdkflags.FlagChainID)
	if err != nil {
		return types.MsgMultiChainWrap{}, fmt.Errorf("amount: %s", err)
	}

	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return types.MsgMultiChainWrap{}, err
	}
	msgbz, err := wasmxMsgWrap(execMsg)
	if err != nil {
		return types.MsgMultiChainWrap{}, err
	}

	wasmxExecMsg := &wasmxtypes.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: contractAddr,
		Funds:    amount,
		Msg:      msgbz,
	}
	wasmxExecAny, err := codectypes.NewAnyWithValue(wasmxExecMsg)
	if err != nil {
		return types.MsgMultiChainWrap{}, err
	}

	return types.MsgMultiChainWrap{
		MultiChainId: chainId,
		Sender:       sender.String(),
		Data:         wasmxExecAny,
	}, nil
}

func wasmxMsgWrap(jsonmsg string) ([]byte, error) {
	msg := wasmxtypes.WasmxExecutionMessage{Data: []byte(jsonmsg)}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return msgbz, nil
}

func GetCmdQueryMultiChainCall() *cobra.Command {
	decoder := wasmxcli.NewArgDecoder(wasmxcli.AsciiDecodeString)
	cmd := &cobra.Command{
		Use:   "call [bech32_address] [query]",
		Short: "Calls contract with given address with query data and prints the returned result",
		Long:  "Calls contract with given address with query data and prints the returned result",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			if args[1] == "" {
				return errors.New("query data must not be empty")
			}

			queryData, err := decoder.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("decode query: %s", err)
			}
			if !json.Valid(queryData) {
				return errors.New("query data must be json")
			}
			msgbz, err := wasmxMsgWrap(args[1])
			if err != nil {
				return fmt.Errorf("wrap query data %s", err)
			}
			sender := clientCtx.GetFromAddress().String()
			if sender == "" {
				sender = args[0]
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ContractCall(
				context.Background(),
				&types.QueryContractCallRequest{
					Sender:    sender,
					Address:   args[0],
					QueryData: msgbz,
				},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	decoder.RegisterFlags(cmd.PersistentFlags(), "query argument")
	sdkflags.AddQueryFlagsToCmd(cmd)
	f := cmd.Flags()
	if cmd.Flag(FlagFrom) == nil { // avoid flag redefinition when it's already been added by AutoCLI
		f.String(FlagFrom, "", "Name or address of private key with which to sign")
	}
	sdkflags.AddKeyringFlags(f)
	return cmd
}
