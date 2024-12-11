package gov

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	cli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"wasmx/v1/multichain"
)

// NewTxCmd returns the transaction commands for this module
// governance ModuleClient is slightly different from other ModuleClients in that
// it contains a slice of legacy "proposal" child commands. These commands are respective
// to the proposal type handlers that are implemented in other modules but are mounted
// under the governance CLI (eg. parameter change proposals).
func NewTxCmd(ac address.Codec) *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Governance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	govTxCmd.AddCommand(
		NewCmdDeposit(ac),
		NewCmdVote(ac),
		NewCmdWeightedVote(ac),
		NewCmdSubmitProposal(ac),
		cli.NewCmdDraftProposal(),
		NewCmdCancelProposal(ac),
	)

	return govTxCmd
}

// NewCmdSubmitProposal implements submitting a proposal transaction command.
func NewCmdSubmitProposal(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [path/to/proposal.json]",
		Short: "Submit a proposal along with some messages, metadata and deposit",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a proposal along with some messages, metadata and deposit.
They should be defined in a JSON file.

Example:
$ %s tx gov submit-proposal path/to/proposal.json

Where proposal.json contains:

{
  // array of proto-JSON-encoded sdk.Msgs
  "messages": [
    {
      "@type": "/cosmos.bank.v1beta1.MsgSend",
      "from_address": "cosmos1...",
      "to_address": "cosmos1...",
      "amount":[{"denom": "stake","amount": "10"}]
    }
  ],
  // metadata can be any of base64 encoded, raw text, stringified json, IPFS link to json
  // see below for example metadata
  "metadata": "4pIMOgIGx1vZGU=",
  "deposit": "10stake",
  "title": "My proposal",
  "summary": "A short summary of my proposal",
  "expedited": false
}

metadata example:
{
	"title": "",
	"authors": [""],
	"summary": "",
	"details": "",
	"proposal_forum_url": "",
	"vote_option_context": "",
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			proposal, msgs, deposit, err := parseSubmitProposal(mcctx.ClientCtx.Codec, args[0])
			if err != nil {
				return err
			}
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			msg, err := v1.NewMsgSubmitProposal(msgs, deposit, fromAddr.String(), proposal.Metadata, proposal.Title, proposal.Summary, proposal.Expedited)
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewCmdCancelProposal implements submitting a cancel proposal transaction command.
func NewCmdCancelProposal(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-proposal [proposal-id]",
		Short:   "Cancel governance proposal before the voting period ends. Must be signed by the proposal creator.",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`$ %s tx gov cancel-proposal 1 --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			msg := v1.NewMsgCancelProposal(proposalID, fromAddr.String())
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

// NewCmdSubmitLegacyProposal implements submitting a proposal transaction command.
// Deprecated: please use NewCmdSubmitProposal instead.
func NewCmdSubmitLegacyProposal(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-legacy-proposal",
		Short: "Submit a legacy proposal along with an initial deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a legacy proposal along with an initial deposit.
Proposal title, description, type and deposit can be given directly or through a proposal JSON file.

Example:
$ %s tx gov submit-legacy-proposal --proposal="path/to/proposal.json" --from mykey

Where proposal.json contains:

{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "type": "Text",
  "deposit": "10test"
}

Which is equivalent to:

$ %s tx gov submit-legacy-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10test" --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			proposal, err := parseSubmitLegacyProposal(cmd.Flags())
			if err != nil {
				return fmt.Errorf("failed to parse proposal: %w", err)
			}

			amount, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			content, ok := v1beta1.ContentFromProposalType(proposal.Title, proposal.Description, proposal.Type)
			if !ok {
				return fmt.Errorf("failed to create proposal content: unknown proposal type %s", proposal.Type)
			}
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			msg, err := v1beta1.NewMsgSubmitProposal(content, amount, mcctx.ClientCtx.GetFromAddress())
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(cli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(cli.FlagProposalType, "", "The proposal Type")
	cmd.Flags().String(cli.FlagDeposit, "", "The proposal deposit")
	cmd.Flags().String(cli.FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)")
	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewCmdDeposit implements depositing tokens for an active proposal.
func NewCmdDeposit(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit tokens for an active proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a deposit for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov deposit 1 10stake --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			// Get amount of coins
			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := &v1.MsgDeposit{
				ProposalId: proposalID,
				Depositor:  fromAddr.String(),
				Amount:     amount,
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewCmdVote implements creating a new vote command.
func NewCmdVote(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [option]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a vote for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov vote 1 yes --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// Find out which vote option user chose
			byteVoteOption, err := v1.VoteOptionFromString(govutils.NormalizeVoteOption(args[1]))
			if err != nil {
				return err
			}

			metadata, err := cmd.Flags().GetString(cli.FlagMetadata)
			if err != nil {
				return err
			}

			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			// Build vote message and run basic validation
			msg := &v1.MsgVote{
				ProposalId: proposalID,
				Voter:      fromAddr.String(),
				Option:     byteVoteOption,
				Metadata:   metadata,
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}

	cmd.Flags().String(cli.FlagMetadata, "", "Specify metadata of the vote")
	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewCmdWeightedVote implements creating a new weighted vote command.
func NewCmdWeightedVote(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weighted-vote [proposal-id] [weighted-options]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a vote for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov weighted-vote 1 yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05 --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// Figure out which vote options user chose
			options, err := v1.WeightedVoteOptionsFromString(govutils.NormalizeWeightedVoteOptions(args[1]))
			if err != nil {
				return err
			}

			metadata, err := cmd.Flags().GetString(cli.FlagMetadata)
			if err != nil {
				return err
			}

			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			// Build vote message and run basic validation
			msg := &v1.MsgVoteWeighted{
				ProposalId: proposalID,
				Voter:      fromAddr.String(),
				Options:    options,
				Metadata:   metadata,
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}

	cmd.Flags().String(cli.FlagMetadata, "", "Specify metadata of the weighted vote")
	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}
