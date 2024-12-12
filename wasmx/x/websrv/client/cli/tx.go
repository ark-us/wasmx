package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	sdkaddress "cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	address "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/loredanacirstea/wasmx/multichain"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/websrv/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// this line is used by starport scaffolding # 1
	cmd.AddCommand(
		NewRegisterRouteProposalCmd(ac),
		NewDeregisterRouteProposalCmd(ac),
		NewRegisterOauthClientCmd(ac),
		NewDeregisterOauthClientCmd(ac),
		NewEditOauthClientCmd(ac),
	)

	return cmd
}

// NewRegisterRouteProposalCmd returns a CLI command handler for registering a
// webserver route smart contract handler
func NewRegisterRouteProposalCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-route [url_path] [contract_address]",
		Args:    cobra.ExactArgs(2),
		Short:   "Submit a governance proposal to register a URL path and the contract address handling it.",
		Long:    "Submit a governance proposal to register a URL path and the contract address handling it.",
		Example: fmt.Sprintf("$ %s tx gov submit-legacy-proposal register-route <route> <contract-address> --from=<key_or_address>", version.AppName),

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

			path := args[0]
			contractAddress := args[1]
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			authority, err := mcctx.CustomAddrCodec.BytesToString(sdk.AccAddress(address.Module(wasmxtypes.ROLE_GOVERNANCE)))
			if err != nil {
				return err
			}

			content := &types.MsgRegisterRoute{Authority: authority, Title: title, Description: description, Path: path, ContractAddress: contractAddress}

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
	return cmd
}

// NewDeregisterRouteProposalCmd returns a CLI command handler for registering a
// deregistration of a webserver route smart contract handler
func NewDeregisterRouteProposalCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deregister-route [url_path] [contract_address]",
		Args:    cobra.ExactArgs(2),
		Short:   "Submit a governance proposal to deregister a URL path and the contract address handling it.",
		Long:    "Submit a governance proposal to deregister a URL path and the contract address handling it.",
		Example: fmt.Sprintf("$ %s tx gov submit-legacy-proposal deregister-route <route> <contract-address> --from=<key_or_address>", version.AppName),

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

			path := args[0]
			contractAddress := args[1]
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			authority, err := mcctx.CustomAddrCodec.BytesToString(sdk.AccAddress(address.Module(wasmxtypes.ROLE_GOVERNANCE)))
			if err != nil {
				return err
			}

			content := &types.MsgDeregisterRoute{Authority: authority, Title: title, Description: description, Path: path, ContractAddress: contractAddress}

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
	return cmd
}

// NewRegisterOauthClientCmd returns a CLI command handler for registering a
// oauth client
func NewRegisterOauthClientCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-oauth-client [domain-url]",
		Args:    cobra.ExactArgs(1),
		Short:   "Register an oauth client by supplying a domain URL.",
		Long:    "Register an oauth client by supplying a domain URL.",
		Example: fmt.Sprintf("$ %s tx %s register-oauth-client \"https://oauthdebugger.com\" --from=<key_or_address>", version.AppName, types.ModuleName),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			domain := args[0]

			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			msg := &types.MsgRegisterOAuthClient{
				Owner:  fromAddr.String(),
				Domain: domain,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewEditOauthClientCmd returns a CLI command handler for editing a
// oauth client
func NewEditOauthClientCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit-oauth-client [client-id] [domain-url]",
		Args:    cobra.ExactArgs(2),
		Short:   "Edit an oauth client by supplying a new domain URL.",
		Long:    "Edit an oauth client by supplying a new domain URL.",
		Example: fmt.Sprintf("$ %s tx %s edit-oauth-client <client-id>  <domain-url> --from=<key_or_address>", version.AppName, types.ModuleName),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			clientId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			domain := args[1]
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			msg := &types.MsgEditOAuthClient{
				Owner:    fromAddr.String(),
				ClientId: uint64(clientId),
				Domain:   domain,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewDeregisterOauthClientCmd returns a CLI command handler for deregistering a
// oauth client
func NewDeregisterOauthClientCmd(ac sdkaddress.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deregister-oauth-client [client-id]",
		Args:    cobra.ExactArgs(1),
		Short:   "Deregister an oauth client by client ID.",
		Long:    "Deregister an oauth client by client ID.",
		Example: fmt.Sprintf("$ %s tx %s deregister-oauth-client <client-id> --from=<key_or_address>", version.AppName, types.ModuleName),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			clientId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			fromAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())

			msg := &types.MsgDeregisterOAuthClient{
				Owner:    fromAddr.String(),
				ClientId: uint64(clientId),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			msgMultiChain, err := mcctx.MultiChainWrap(msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(mcctx.ClientCtx, cmd.Flags(), msgMultiChain)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
