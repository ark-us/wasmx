package cli

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"wasmx/v1/multichain"
	"wasmx/v1/x/wasmx/keeper"
	"wasmx/v1/x/wasmx/types"
)

var (
	FlagFrom = "from"
)

func GetQueryCmd(queryRoute string, ac address.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	queryCmd.AddCommand(
		GetCmdGetContractCall(ac),
		GetCmdListCode(ac),
		GetCmdListContractByCode(ac),
		GetCmdQueryCode(ac),
		GetCmdQueryCodeInfo(ac),
		GetCmdGetContractInfo(ac),
		GetCmdGetContractState(ac),
		GetCmdLibVersion(),
		GetCmdQueryParams(ac),
		GetCmdBuildAddress(ac),
	)
	return queryCmd
}

// GetCmdLibVersion gets current wasmedge version.
func GetCmdLibVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wasmedge-version",
		Short:   "Get wasmedge version",
		Long:    "Get wasmedge version",
		Aliases: []string{"lib-version"},
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := types.LibWasmEdgeVersion()
			fmt.Println(version)
			return nil
		},
		SilenceUsage: true,
	}
	return cmd
}

// GetCmdBuildAddress build a contract address
func GetCmdBuildAddress(ac address.Codec) *cobra.Command {
	decoder := NewArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use:     "build-address [code-hash] [creator-address] [salt-hex-encoded] [json_encoded_init_args (required when set as fixed)]",
		Short:   "build contract address",
		Aliases: []string{"address"},
		Args:    cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}
			codeHash, err := hex.DecodeString(args[0])
			if err != nil {
				return fmt.Errorf("code-hash: %s", err)
			}
			creator_, err := mcctx.CustomAddrCodec.StringToAddressPrefixedUnsafe(args[1])
			if err != nil {
				return fmt.Errorf("creator: %s", err)
			}
			creator := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(creator_.Bytes())
			salt, err := hex.DecodeString(args[2])
			switch {
			case err != nil:
				return fmt.Errorf("salt: %s", err)
			case len(salt) == 0:
				return errors.New("empty salt")
			}

			addr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(keeper.EwasmBuildContractAddressPredictable(creator.Bytes(), salt, codeHash))

			cmd.Println(addr.String())
			return nil
		},
		SilenceUsage: true,
	}
	decoder.RegisterFlags(cmd.PersistentFlags(), "salt")
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

// GetCmdListCode lists all wasm code uploaded
func GetCmdListCode(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-code",
		Short:   "List all wasm bytecode on the chain",
		Long:    "List all wasm bytecode on the chain",
		Aliases: []string{"list-codes", "codes", "lco"},
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}
			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.Codes(
				context.Background(),
				&types.QueryCodesRequest{
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}
			return mcctx.ClientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "list codes")
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

// GetCmdListContractByCode lists all wasm code uploaded for given code id
func GetCmdListContractByCode(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-contract-by-code [code_id]",
		Short:   "List wasm all bytecode on the chain for given code id",
		Long:    "List wasm all bytecode on the chain for given code id",
		Aliases: []string{"list-contracts-by-code", "list-contracts", "contracts", "lca"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			if codeID == 0 {
				return errors.New("empty code id")
			}

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.ContractsByCode(
				context.Background(),
				&types.QueryContractsByCodeRequest{
					CodeId:     codeID,
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}
			return mcctx.ClientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "list contracts by code")
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryCode returns the bytecode for a given contract
func GetCmdQueryCode(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "code [code_id] [output filename]",
		Short:   "Downloads wasm bytecode for given code id",
		Long:    "Downloads wasm bytecode for given code id",
		Aliases: []string{"source-code", "source"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.Code(
				context.Background(),
				&types.QueryCodeRequest{
					CodeId: codeID,
				},
			)
			if err != nil {
				return err
			}
			if len(res.Data) == 0 {
				return fmt.Errorf("contract not found")
			}

			fmt.Printf("Downloading wasm code to %s\n", args[1])
			return os.WriteFile(args[1], res.Data, 0o600)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryCodeInfo returns the code info for a given code id
func GetCmdQueryCodeInfo(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code-info [code_id]",
		Short: "Prints out metadata of a code id",
		Long:  "Prints out metadata of a code id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.Code(
				context.Background(),
				&types.QueryCodeRequest{
					CodeId: codeID,
				},
			)
			if err != nil {
				return err
			}
			if res.CodeInfo == nil {
				return fmt.Errorf("contract not found")
			}

			return mcctx.ClientCtx.PrintProto(res.CodeInfo)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetContractInfo gets details about a given contract
func GetCmdGetContractInfo(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contract [bech32_address]",
		Short:   "Prints out metadata of a contract given its address",
		Long:    "Prints out metadata of a contract given its address",
		Aliases: []string{"meta", "c"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			_, err = mcctx.CustomAddrCodec.StringToAddressPrefixedUnsafe(args[0])
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.ContractInfo(
				context.Background(),
				&types.QueryContractInfoRequest{
					Address: args[0],
				},
			)
			if err != nil {
				return err
			}
			return mcctx.ClientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetContractState dumps full internal state of a given contract
func GetCmdGetContractState(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "contract-state",
		Short:                      "Querying commands for the wasm module",
		Aliases:                    []string{"state", "cs", "s"},
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	cmd.AddCommand(
		GetCmdGetContractStateAll(ac),
		GetCmdGetContractStateRaw(ac),
	)
	return cmd
}

func GetCmdGetContractStateAll(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all [bech32_address]",
		Short: "Prints out all internal state of a contract given its address",
		Long:  "Prints out all internal state of a contract given its address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			contractAddr_, err := mcctx.CustomAddrCodec.StringToAddressPrefixedUnsafe(args[0])
			if err != nil {
				return err
			}
			contractAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(contractAddr_.Bytes())

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.AllContractState(
				context.Background(),
				&types.QueryAllContractStateRequest{
					Address:    contractAddr.String(),
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}
			return mcctx.ClientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "contract state")
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func GetCmdGetContractStateRaw(ac address.Codec) *cobra.Command {
	decoder := NewArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use:   "raw [bech32_address] [key]",
		Short: "Prints out internal state for key of a contract given its address",
		Long:  "Prints out internal state for of a contract given its address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			contractAddr_, err := mcctx.CustomAddrCodec.StringToAddressPrefixedUnsafe(args[0])
			if err != nil {
				return err
			}
			contractAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(contractAddr_.Bytes())
			queryData, err := decoder.DecodeString(args[1])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.RawContractState(
				context.Background(),
				&types.QueryRawContractStateRequest{
					Address:   contractAddr.String(),
					QueryData: queryData,
				},
			)
			if err != nil {
				return err
			}
			return mcctx.ClientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	decoder.RegisterFlags(cmd.PersistentFlags(), "key argument")
	flags.AddQueryFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func GetCmdGetContractCall(ac address.Codec) *cobra.Command {
	decoder := NewArgDecoder(AsciiDecodeString)
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
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			contractAddr_, err := mcctx.CustomAddrCodec.StringToAddressPrefixedUnsafe(args[0])
			if err != nil {
				return err
			}
			contractAddr := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(contractAddr_.Bytes())
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
			sender := mcctx.CustomAddrCodec.BytesToAccAddressPrefixed(mcctx.ClientCtx.GetFromAddress())
			if err != nil {
				return fmt.Errorf("sender to string: %s", err)
			}
			if len(sender.Bytes()) == 0 {
				sender = contractAddr
			}

			queryClient := types.NewQueryClient(mcctx.ClientCtx)
			res, err := queryClient.SmartContractCall(
				context.Background(),
				&types.QuerySmartContractCallRequest{
					Sender:    sender.String(),
					Address:   contractAddr.String(),
					QueryData: msgbz,
				},
			)
			if err != nil {
				return err
			}
			return mcctx.ClientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	decoder.RegisterFlags(cmd.PersistentFlags(), "query argument")
	flags.AddQueryFlagsToCmd(cmd)
	f := cmd.Flags()
	if cmd.Flag(FlagFrom) == nil { // avoid flag redefinition when it's already been added by AutoCLI
		f.String(FlagFrom, "", "Name or address of private key with which to sign")
	}
	flags.AddKeyringFlags(f)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

type argumentDecoder struct {
	// dec is the default decoder
	dec                func(string) ([]byte, error)
	asciiF, hexF, b64F bool
}

func NewArgDecoder(def func(string) ([]byte, error)) *argumentDecoder {
	return &argumentDecoder{dec: def}
}

func (a *argumentDecoder) RegisterFlags(f *flag.FlagSet, argName string) {
	f.BoolVar(&a.asciiF, "ascii", false, "ascii encoded "+argName)
	f.BoolVar(&a.hexF, "hex", false, "hex encoded "+argName)
	f.BoolVar(&a.b64F, "b64", false, "base64 encoded "+argName)
}

func (a *argumentDecoder) DecodeString(s string) ([]byte, error) {
	found := -1
	for i, v := range []*bool{&a.asciiF, &a.hexF, &a.b64F} {
		if !*v {
			continue
		}
		if found != -1 {
			return nil, errors.New("multiple decoding flags used")
		}
		found = i
	}
	switch found {
	case 0:
		return AsciiDecodeString(s)
	case 1:
		return hex.DecodeString(s)
	case 2:
		return base64.StdEncoding.DecodeString(s)
	default:
		return a.dec(s)
	}
}

func AsciiDecodeString(s string) ([]byte, error) {
	return []byte(s), nil
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

// GetCmdQueryParams implements a command to return the current wasm
// parameters.
func GetCmdQueryParams(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current wasm parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(mcctx.ClientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			return mcctx.ClientCtx.PrintProto(&res.Params)
		},
		SilenceUsage: true,
	}

	flags.AddQueryFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}
