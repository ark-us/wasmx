package cli

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	// "strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/loredanacirstea/wasmx/v1/x/websrv/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group websrv queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdGetAllOauthClients(),
		CmdGetOauthClient(),
		CmdGetOauthClientsByOwner(),
	)
	// this line is used by starport scaffolding # 1

	return cmd
}

// CmdGetAllOauthClients lists all oauth clients registered
func CmdGetAllOauthClients() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "oauth-all-clients",
		Short:   "List all oauth clients that have been registered",
		Long:    "List all oauth clients that have been registered",
		Aliases: []string{"oauth-all", "oauth-clients"},
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf("$ %s query %s oauth-all-clients", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetAllOauthClients(
				context.Background(),
				&types.QueryGetAllOauthClientsRequest{
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "oauth clients")
	return cmd
}

// CmdGetOauthClient lists oauth client by client id
func CmdGetOauthClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "oauth-client",
		Short:   "List a registered oauth client by client id",
		Long:    "List a registered oauth client by client id",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("$ %s query %s oauth-client <client-id>", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			clientId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetOauthClient(
				context.Background(),
				&types.QueryGetOauthClientRequest{
					ClientId: uint64(clientId),
				},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "oauth clients")
	return cmd
}

// CmdGetOauthClientsByOwner lists all wasm code uploaded
func CmdGetOauthClientsByOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "oauth-client-ids-by-owner",
		Short:   "List all registered client ids by owner address",
		Long:    "List all registered client ids by owner address",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("$ %s query %s oauth-client-ids-by-owner <owner-address>", version.AppName, types.ModuleName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			addrCodec := clientCtx.InterfaceRegistry.SigningContext().AddressCodec()

			owner, err := addrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}
			ownerstr, err := clientCtx.InterfaceRegistry.SigningContext().AddressCodec().BytesToString(owner)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetOauthClientsByOwner(
				context.Background(),
				&types.QueryGetOauthClientsByOwnerRequest{
					Owner: ownerstr,
				},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "oauth clients")
	return cmd
}

// sdk ReadPageRequest expects binary but we encoded to base64 in our marshaller
func withPageKeyDecoded(flagSet *flag.FlagSet) *flag.FlagSet {
	encoded, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		panic(err.Error())
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err.Error())
	}
	err = flagSet.Set(flags.FlagPageKey, string(raw))
	if err != nil {
		panic(err.Error())
	}
	return flagSet
}
