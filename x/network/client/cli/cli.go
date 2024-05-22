package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	address "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"

	// "github.com/cosmos/cosmos-sdk/client/flags"

	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	"mythos/v1/multichain"
	"mythos/v1/x/network/types"
	wasmxcli "mythos/v1/x/wasmx/client/cli"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

const (
	flagAmount = "amount"
	FlagFrom   = "from"
)

type appwithTxConfig interface {
	TxConfig() client.TxConfig
}

func GetTxCmd(appCreator multichain.NewAppCreator) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "multichain",
		Short:                      "multichain transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		MultiChainTxExecuteCmd(),
		RegisterNewSubChain(),
		RegisterSubChainValidator(appCreator),
		InitializeSubChain(),
	)

	return txCmd
}

func GetQueryCmd(appCreator multichain.NewAppCreator) *cobra.Command {
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
		GetCmdQuerySubChain(),
		GetCmdQuerySubChainConfig(),
		GetCmdQuerySubChainIds(),
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
			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}

			amountStr, err := cmd.Flags().GetString(flagAmount)
			if err != nil {
				return fmt.Errorf("amount: %s", err)
			}
			amount, err := sdk.ParseCoinsNormalized(amountStr)
			if err != nil {
				return err
			}

			return sendMultiChainExecution(
				clientCtx,
				cmd.Flags(),
				customAddrCodec,
				args[0],
				clientCtx.GetFromAddress(),
				[]byte(args[1]),
				amount,
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func RegisterNewSubChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-subchain [chain_base_name] [denom_unit] [decimals] [level_index] [initial_balance]",
		Short: "Register subchain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register subchain

Example:
$ %s tx network register-subchain mythos myt 18 1 "10000000000" --chain-id="level0_1000-1"

		`, version.AppName)),
		Aliases: []string{},
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}

			decimals, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}
			levelIndex, err := strconv.Atoi(args[3])
			if err != nil {
				return err
			}
			initialBalance, err := math.LegacyNewDecFromStr(args[4])
			if err != nil {
				return err
			}
			msgbz, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{RegisterDefaultSubChain: &wasmxtypes.RegisterDefaultSubChainRequest{
				ChainBaseName:  args[0],
				DenomUnit:      args[1],
				Decimals:       uint32(decimals),
				LevelIndex:     uint32(levelIndex),
				InitialBalance: initialBalance.BigInt(),
			}})
			if err != nil {
				return err
			}
			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := customAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}

			return sendMultiChainExecution(
				clientCtx,
				cmd.Flags(),
				customAddrCodec,
				contractAddr,
				clientCtx.GetFromAddress(),
				msgbz,
				nil,
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func RegisterSubChainValidator(appCreator multichain.NewAppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-subchain-validator [subchain_id] [path/to/validator.json]",
		Short: "Register subchain validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register subchain validator

Example:
$ %s tx network register-subchain-validator mythos_7000-1 path/to/validator.json --chain-id="level0_1000-1"

Where validator.json contains:

{
	"moniker": "myvalidator",
	"identity": "optional identity signature (ex. UPort or Keybase)",
	"website": "validator's (optional) website",
	"security": "validator's (optional) security contact email",
	"details": "validator's (optional) details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
	"amount": "1000000000",
}

		`, version.AppName)),
		Aliases: []string{},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}
			flags := cmd.Flags()
			subChainId := args[0]
			chainId, err := flags.GetString(sdkflags.FlagChainID)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			validMsg, err := parseAndValidateValidatorJSON(args[1])
			if err != nil {
				return err
			}
			genTxData, err := getGenTxData(clientCtx, flags, customAddrCodec, sender, subChainId, validMsg)
			if err != nil {
				return err
			}
			signedGenTxData, err := signGenTxData(appCreator, clientCtx, flags, customAddrCodec, chainId, subChainId, genTxData, sender)
			if err != nil {
				return err
			}

			msgbz, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{RegisterSubChainValidator: &wasmxtypes.RegisterSubChainValidatorRequest{
				ChainId: subChainId,
				GenTx:   signedGenTxData,
			}})
			if err != nil {
				return err
			}
			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := customAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}

			return sendMultiChainExecution(
				clientCtx,
				cmd.Flags(),
				customAddrCodec,
				contractAddr,
				sender,
				msgbz,
				nil,
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func InitializeSubChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-subchain [subchain_id]",
		Short: "Initialize a registered subchain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register subchain

Example:
$ %s tx network init-subchain level1_1000-1 --chain-id="level0_1000-1"

		`, version.AppName)),
		Aliases: []string{},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}

			msgbz, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{InitSubChain: &wasmxtypes.InitSubChainRequest{
				ChainId: args[0],
			}})
			if err != nil {
				return err
			}
			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := customAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}
			return sendMultiChainExecution(
				clientCtx,
				cmd.Flags(),
				customAddrCodec,
				contractAddr,
				clientCtx.GetFromAddress(),
				msgbz,
				nil,
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	return cmd
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

			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}

			_, err = customAddrCodec.StringToBytes(args[0])
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
			sender, err := customAddrCodec.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return fmt.Errorf("sender: %s", err)
			}
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

func GetCmdQuerySubChainIds() *cobra.Command {
	decoder := wasmxcli.NewArgDecoder(wasmxcli.AsciiDecodeString)
	cmd := &cobra.Command{
		Use:   "subchains",
		Short: "Show all registered subchain ids (initialized or not)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Show all subchain ids

Example:
$ %s query network subchains --chain-id="level0_1000-1"

		`, version.AppName)),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}

			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := customAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}

			from, _ := cmd.Flags().GetString(sdkflags.FlagFrom)
			sender, _, _, _ := client.GetFromFields(clientCtx, clientCtx.Keyring, from)

			querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetSubChainIds: &wasmxtypes.QueryGetSubChainIdsRequest{}})
			if err != nil {
				return err
			}

			res, err := sendMultiChainQuery(
				clientCtx,
				cmd.Flags(),
				customAddrCodec,
				contractAddr,
				sender,
				querymsg,
				nil,
				nil,
			)
			if err != nil {
				return err
			}
			data2, err := decodeQueryResponse(res.Data)
			if err != nil {
				return err
			}
			fmt.Println(string(data2))
			return nil
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

func GetCmdQuerySubChainConfig() *cobra.Command {
	decoder := wasmxcli.NewArgDecoder(wasmxcli.AsciiDecodeString)
	cmd := &cobra.Command{
		Use:   "subchain [subchain_id]",
		Short: "Show subchain configuration by id",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Show subchain configuration by id

Example:
$ %s query network subchain level1_1000-1 --chain-id="level0_1000-1"

		`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}

			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := customAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}
			querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetSubChainConfigById: &wasmxtypes.QueryGetSubChainRequest{
				ChainId: args[0],
			}})
			if err != nil {
				return err
			}

			from, _ := cmd.Flags().GetString(sdkflags.FlagFrom)
			sender, _, _, _ := client.GetFromFields(clientCtx, clientCtx.Keyring, from)

			res, err := sendMultiChainQuery(
				clientCtx,
				cmd.Flags(),
				customAddrCodec,
				contractAddr,
				sender,
				querymsg,
				nil,
				nil,
			)
			if err != nil {
				return err
			}
			data2, err := decodeQueryResponse(res.Data)
			if err != nil {
				return err
			}
			fmt.Println(string(data2))
			return nil
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

func GetCmdQuerySubChain() *cobra.Command {
	decoder := wasmxcli.NewArgDecoder(wasmxcli.AsciiDecodeString)
	cmd := &cobra.Command{
		Use:   "subchain-data [subchain_id]",
		Short: "Show subchain data by id",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Show subchain data by id

Example:
$ %s query network subchain-data level1_1000-1 --chain-id="level0_1000-1"

		`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			clientCtx, customAddrCodec, err := multichain.MultiChainCustomCtx(clientCtx)
			if err != nil {
				return err
			}

			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := customAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}
			querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetSubChainById: &wasmxtypes.QueryGetSubChainRequest{
				ChainId: args[0],
			}})
			if err != nil {
				return err
			}

			from, _ := cmd.Flags().GetString(sdkflags.FlagFrom)
			sender, _, _, _ := client.GetFromFields(clientCtx, clientCtx.Keyring, from)

			res, err := sendMultiChainQuery(
				clientCtx,
				cmd.Flags(),
				customAddrCodec,
				contractAddr,
				sender,
				querymsg,
				nil,
				nil,
			)
			if err != nil {
				return err
			}
			data2, err := decodeQueryResponse(res.Data)
			if err != nil {
				return err
			}
			fmt.Println(string(data2))
			return nil
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

func sendMultiChainExecution(
	clientCtx client.Context,
	flags *flag.FlagSet,
	addrCodec address.Codec,
	contractAddr string, sender sdk.AccAddress,
	execMsg []byte, amount sdk.Coins,
) error {
	chainId, err := flags.GetString(sdkflags.FlagChainID)
	if err != nil {
		return fmt.Errorf("chainId: %s", err)
	}
	msg := wasmxtypes.WasmxExecutionMessage{Data: execMsg}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	senderAddr, err := addrCodec.BytesToString(sender)
	if err != nil {
		return err
	}
	wasmxExecMsg := &wasmxtypes.MsgExecuteContract{
		Sender:   senderAddr,
		Contract: contractAddr,
		Funds:    amount,
		Msg:      msgbz,
	}
	wasmxExecAny, err := codectypes.NewAnyWithValue(wasmxExecMsg)
	if err != nil {
		return err
	}

	mmsg := types.MsgMultiChainWrap{
		MultiChainId: chainId,
		Sender:       senderAddr,
		Data:         wasmxExecAny,
	}
	if err := mmsg.ValidateBasic(); err != nil {
		return err
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, flags, &mmsg)
}

func sendMultiChainQuery(
	clientCtx client.Context,
	flags *flag.FlagSet,
	addrCodec address.Codec,
	contractAddr string, sender sdk.AccAddress,
	queryData []byte, amount sdk.Coins,
	dependencies []string,
) (*types.QueryMultiChainResponse, error) {
	chainId, err := flags.GetString(sdkflags.FlagChainID)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	if !json.Valid(queryData) {
		return nil, errors.New("query data must be json")
	}
	msg := wasmxtypes.WasmxExecutionMessage{Data: queryData}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	senderAddr, err := addrCodec.BytesToString(sender)
	if err != nil {
		return nil, err
	}
	queryClient := types.NewQueryClient(clientCtx)

	query := wasmxtypes.QuerySmartContractCallRequest{
		Sender:       senderAddr,
		Address:      contractAddr,
		QueryData:    msgbz,
		Funds:        amount,
		Dependencies: dependencies,
	}
	bz, err := query.Marshal()
	if err != nil {
		return nil, err
	}

	req := &abci.RequestQuery{Data: bz, Path: "/mythos.wasmx.v1.Query/SmartContractCall"}
	reqbz, err := req.Marshal()
	if err != nil {
		return nil, err
	}

	res, err := queryClient.QueryMultiChain(
		context.Background(),
		&types.QueryMultiChainRequest{
			MultiChainId: chainId,
			QueryData:    reqbz,
		},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func wasmxMsgWrap(jsonmsg string) ([]byte, error) {
	msg := wasmxtypes.WasmxExecutionMessage{Data: []byte(jsonmsg)}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return msgbz, nil
}

func parseAndValidateValidatorJSON(path string) (*stakingtypes.MsgCreateValidator, error) {
	type internalVal struct {
		Amount              string `json:"amount"`
		Moniker             string `json:"moniker"`
		Identity            string `json:"identity,omitempty"`
		Website             string `json:"website,omitempty"`
		Security            string `json:"security,omitempty"`
		Details             string `json:"details,omitempty"`
		CommissionRate      string `json:"commission-rate"`
		CommissionMaxRate   string `json:"commission-max-rate"`
		CommissionMaxChange string `json:"commission-max-change-rate"`
		MinSelfDelegation   string `json:"min-self-delegation"`
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var v internalVal
	err = json.Unmarshal(contents, &v)
	if err != nil {
		return nil, err
	}

	if v.Amount == "" {
		return nil, fmt.Errorf("must specify amount of coins to bond")
	}

	amount, ok := math.NewIntFromString(v.Amount)
	if !ok {
		return nil, fmt.Errorf("failed to parse amount")
	}

	if v.Moniker == "" {
		return nil, fmt.Errorf("must specify the moniker name")
	}

	commissionRates, err := buildCommissionRates(v.CommissionRate, v.CommissionMaxRate, v.CommissionMaxChange)
	if err != nil {
		return nil, err
	}

	if v.MinSelfDelegation == "" {
		return nil, fmt.Errorf("must specify minimum self delegation")
	}
	minSelfDelegation, ok := math.NewIntFromString(v.MinSelfDelegation)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
	}

	return &stakingtypes.MsgCreateValidator{
		Value: sdk.Coin{Denom: "faked", Amount: amount},
		Description: stakingtypes.Description{
			Moniker:         v.Moniker,
			Identity:        v.Identity,
			Website:         v.Website,
			SecurityContact: v.Security,
			Details:         v.Details,
		},
		Commission:        commissionRates,
		MinSelfDelegation: minSelfDelegation,
	}, nil
}

func buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr string) (commission stakingtypes.CommissionRates, err error) {
	if rateStr == "" || maxRateStr == "" || maxChangeRateStr == "" {
		return commission, errors.New("must specify all validator commission parameters")
	}

	rate, err := math.LegacyNewDecFromStr(rateStr)
	if err != nil {
		return commission, err
	}

	maxRate, err := math.LegacyNewDecFromStr(maxRateStr)
	if err != nil {
		return commission, err
	}

	maxChangeRate, err := math.LegacyNewDecFromStr(maxChangeRateStr)
	if err != nil {
		return commission, err
	}

	commission = stakingtypes.NewCommissionRates(rate, maxRate, maxChangeRate)

	return commission, nil
}

func getGenTxData(
	clientCtx client.Context,
	flags *flag.FlagSet,
	customAddrCodec mcodec.AccBech32Codec,
	sender sdk.AccAddress,
	subChainId string,
	validMsg *stakingtypes.MsgCreateValidator,
) ([]byte, error) {
	genTxInfo, err := json.Marshal(&wasmxtypes.QueryBuildGenTxRequest{
		ChainId: subChainId,
		Msg:     *validMsg,
	})
	if err != nil {
		return nil, err
	}
	paramBz, err := json.Marshal(&wasmxtypes.ActionParam{Key: "message", Value: string(genTxInfo)})
	if err != nil {
		return nil, err
	}
	querymsg := []byte(fmt.Sprintf(`{"query":{"action": {"type": "buildGenTx", "params": [%s],"event":null}}}`, string(paramBz)))
	level0Addr := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_LEVEL0)
	level0AddrStr, err := customAddrCodec.BytesToString(level0Addr)
	if err != nil {
		return nil, err
	}
	res, err := sendMultiChainQuery(
		clientCtx,
		flags,
		customAddrCodec,
		level0AddrStr,
		sender,
		querymsg,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return decodeQueryResponse(res.Data)
}

func decodeQueryResponse(resp []byte) ([]byte, error) {
	var data wasmxtypes.QuerySmartContractCallResponse
	err := data.Unmarshal(resp)
	if err != nil {
		return nil, err
	}
	var data2 wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(data.Data, &data2)
	if err != nil {
		return nil, err
	}
	return data2.Data, nil
}

func signGenTxData(
	appCreatorFactory multichain.NewAppCreator,
	clientCtx client.Context,
	flags *flag.FlagSet,
	customAddrCodec mcodec.AccBech32Codec,
	chainId string,
	subChainId string,
	genTxData []byte,
	sender sdk.AccAddress,
) ([]byte, error) {
	querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetSubChainConfigById: &wasmxtypes.QueryGetSubChainRequest{
		ChainId: subChainId,
	}})
	if err != nil {
		return nil, err
	}
	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	contractAddr, err := customAddrCodec.BytesToString(registryAddress)
	if err != nil {
		return nil, err
	}

	res, err := sendMultiChainQuery(
		clientCtx,
		flags,
		customAddrCodec,
		contractAddr,
		sender,
		querymsg,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	data2, err := decodeQueryResponse(res.Data)
	if err != nil {
		return nil, err
	}

	var subchainConfig menc.ChainConfig
	err = json.Unmarshal(data2, &subchainConfig)
	if err != nil {
		return nil, err
	}

	_, appCreator := createMockAppCreator(appCreatorFactory)
	isubchainapp := appCreator(subChainId, &subchainConfig)
	subchainapp := isubchainapp.(appwithTxConfig)

	subtxconfig := subchainapp.TxConfig()
	sdktx, err := subtxconfig.TxJSONDecoder()(genTxData)
	if err != nil {
		return nil, err
	}

	txbuilder, err := subtxconfig.WrapTxBuilder(sdktx)
	if err != nil {
		return nil, err
	}
	txbuilder, err = signTx(clientCtx, flags, txbuilder, subtxconfig, sender, subChainId)
	if err != nil {
		return nil, err
	}

	valSdkTx := txbuilder.GetTx()
	txjsonbz, err := subtxconfig.TxJSONEncoder()(valSdkTx)
	if err != nil {
		return nil, err
	}
	return txjsonbz, nil
}

func signTx(
	clientCtx client.Context,
	flags *flag.FlagSet,
	txbuilder client.TxBuilder,
	txconfig client.TxConfig,
	sender sdk.AccAddress,
	subChainId string,
) (client.TxBuilder, error) {
	txf, err := tx.NewFactoryCLI(clientCtx, flags)
	if err != nil {
		return nil, err
	}

	k, err := txf.Keybase().Key(clientCtx.FromName)
	if err != nil {
		return nil, err
	}

	pubKey, err := k.GetPubKey()
	if err != nil {
		return nil, err
	}

	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(txconfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: 0,
	}

	subchainSender, err := txconfig.SigningContext().AddressCodec().BytesToString(sender)
	if err != nil {
		return nil, err
	}

	err = txbuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       subChainId,
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        pubKey,
		Address:       subchainSender,
	}
	signMode := signing.SignMode(txconfig.SignModeHandler().DefaultMode())

	bytesToSign, err := authsigning.GetSignBytesAdapter(clientCtx.CmdContext, txconfig.SignModeHandler(), signMode, signerData, txbuilder.GetTx())
	if err != nil {
		return nil, err
	}

	sigBytes, _, err := txf.Keybase().Sign(clientCtx.FromName, bytesToSign, signMode)
	if err != nil {
		return nil, err
	}
	// Construct the SignatureV2 struct
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}
	sigV2 = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}
	err = txbuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}
	return txbuilder, nil
}

func createMockAppCreator(appCreatorFactory multichain.NewAppCreator) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	tempNodeHome := filepath.Join(userHomeDir, ".mythostmp")
	return multichain.CreateMockAppCreator(appCreatorFactory, tempNodeHome)
}
