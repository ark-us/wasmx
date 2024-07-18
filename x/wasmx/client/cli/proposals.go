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

	"mythos/v1/multichain"
	"mythos/v1/x/wasmx/types"
)

// NewRegisterRoleProposalCmd returns a CLI command handler for registering a
// role contract handler
func NewRegisterRoleProposalCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-role [role] [role_label] [contract_address]",
		Args:    cobra.ExactArgs(3),
		Short:   "Submit a governance proposal to register a contract for a system role.",
		Long:    "Submit a governance proposal to register a contract for a system role.",
		Example: fmt.Sprintf("$ %s tx wasmx register-role <role> <role_label> <contract-address> --title=<title> --description=<description> --deposit=<deposit> --from=<key_or_address>", version.AppName),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
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

			role := args[0]
			label := args[1]
			contractAddress := args[2]

			authority, err := mcctx.CustomAddrCodec.BytesToString(sdk.AccAddress(address.Module(types.ROLE_GOVERNANCE)))
			if err != nil {
				return err
			}
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			content := &types.MsgRegisterRole{Authority: authority, Title: title, Description: description, Role: role, Label: label, ContractAddress: contractAddress}

			msg, err := gov1.NewMsgSubmitProposal([]sdk.Msg{content}, deposit, fromAddr.String(), "", title, description, false)
			if err != nil {
				return err
			}

			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
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

// NewDeregisterRoleProposalCmd returns a CLI command handler for registering a
// deregistration of a webserver route smart contract handler
func NewDeregisterRoleProposalCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deregister-role [contract_address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a governance proposal to deregister a role contract handler.",
		Long:    "Submit a governance proposal to deregister a role contract handler.",
		Example: fmt.Sprintf("$ %s tx gov submit-legacy-proposal deregister-role <contract-address> --from=<key_or_address>", version.AppName),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
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

			contractAddress := args[0]

			authority, err := mcctx.CustomAddrCodec.BytesToString(sdk.AccAddress(address.Module(types.ROLE_GOVERNANCE)))
			if err != nil {
				return err
			}
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			content := &types.MsgDeregisterRole{Authority: authority, Title: title, Description: description, ContractAddress: contractAddress}

			msg, err := gov1.NewMsgSubmitProposal([]sdk.Msg{content}, deposit, fromAddr.String(), "", title, description, false)
			if err != nil {
				return err
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgMultiChain)
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
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}
