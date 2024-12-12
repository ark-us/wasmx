package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	sdkaddress "cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"wasmx/v1/multichain"
	"wasmx/v1/x/wasmx/types"
	memc "wasmx/v1/x/wasmx/vm/memory/common"
)

// NewProposalExecuteContractCmd returns a CLI command handler for executing any contract (public or internal)
func NewProposalExecuteContractCmd(wasmVmMeta memc.IWasmVmMeta, ac sdkaddress.Codec, appCreator multichain.NewAppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "propose-execution [contract_address] [json_args]",
		Args:    cobra.ExactArgs(2),
		Short:   "Submit a governance proposal to execute any contract",
		Long:    "Submit a governance proposal to execute any contract",
		Example: fmt.Sprintf("$ %s tx wasmx propose-execution <contract-address> <json_args> --title=<title> --description=<description> --deposit=<deposit> --from=<key_or_address>", version.AppName),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainIdWithAppMsgs(wasmVmMeta, clientCtx, cmd.Flags(), []signing.CustomGetSigner{}, appCreator)
			if err != nil {
				return err
			}
			txf, err := tx.NewFactoryCLI(mcctx.ClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			toAddr_, err := mcctx.CustomAddrCodec.StringToAddressPrefixedUnsafe(args[0])
			if err != nil {
				return err
			}
			toAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(toAddr_.Bytes())

			authority, err := mcctx.CustomAddrCodec.BytesToString(sdk.AccAddress(address.Module(types.ROLE_GOVERNANCE)))
			if err != nil {
				return err
			}
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			msg, err := parseExecuteArgs(mcctx.CustomAddrCodec, toAddr, args[1], mcctx.ClientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			msg.Sender = authority

			proposal, err := gov1.NewMsgSubmitProposal([]sdk.Msg{&msg}, deposit, fromAddr.String(), "", title, description, false)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(mcctx.ClientCtx, txf, proposal)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1amyt", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}
