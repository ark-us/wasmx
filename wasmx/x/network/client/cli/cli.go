package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	address "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"

	// "github.com/cosmos/cosmos-sdk/client/flags"

	mcodec "wasmx/v1/codec"
	mcfg "wasmx/v1/config"
	menc "wasmx/v1/encoding"
	"wasmx/v1/multichain"
	"wasmx/v1/x/network/types"
	wasmxcli "wasmx/v1/x/wasmx/client/cli"
	wasmxtypes "wasmx/v1/x/wasmx/types"
	memc "wasmx/v1/x/wasmx/vm/memory/common"
)

const (
	flagAmount = "amount"
	FlagFrom   = "from"
)

type appwithTxConfig interface {
	TxConfig() client.TxConfig
	LegacyAmino() *codec.LegacyAmino
	AppCodec() codec.Codec
	InterfaceRegistry() codectypes.InterfaceRegistry
}

func GetTxCmd(wasmVmMeta memc.IWasmVmMeta, appCreator multichain.NewAppCreator) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "multichain",
		Short:                      "multichain transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AtomicTxExecuteCmd(wasmVmMeta, appCreator),
		RegisterNewSubChain(),
		RegisterSubChainValidator(wasmVmMeta, appCreator),
		InitializeSubChain(),
		RegisterLobbyGenTx(wasmVmMeta, appCreator),
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
		GetCmdQuerySubChainValidators(),
	)

	return txCmd
}

func AtomicTxExecuteCmd(wasmVmMeta memc.IWasmVmMeta, appFactory multichain.NewAppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "atomic [path/to/atomictx.json] [comma_separated_chainIds]",
		Short:   "Execute an atomic transaction between chains",
		Aliases: []string{},
		Args:    cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Execute an atomic cross-chain transaction on multiple chains.

Example:
$ %s tx multichain atomic path/to/atomictx.json --from mykey

Where atomictx.json contains:

{

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

			subchainId, err := cmd.Flags().GetString(sdkflags.FlagChainID)
			if err != nil {
				return fmt.Errorf("subchainId: %s", err)
			}
			_, appCreator := createMockAppCreator(wasmVmMeta, appFactory, 0)
			isubchainapp := appCreator(subchainId, mcctx.Config)
			subchainapp := isubchainapp.(appwithTxConfig)
			subtxconfig := subchainapp.TxConfig()

			registryId, err := cmd.Flags().GetString(multichain.FlagRegistryChainId)
			if err != nil {
				return fmt.Errorf("registry chainId: %s", err)
			}

			chainIds := []string{}
			if len(args) > 1 {
				ids := strings.Split(strings.TrimSpace(args[1]), ",")
				for _, id := range ids {
					chainIds = append(chainIds, strings.TrimSpace(id))
				}
			}

			sender := mcctx.ClientCtx.GetFromAddress()
			mcwrapsInternal, err := parseAndValidateAtomicTxContentJSON(args[0])
			if err != nil {
				return err
			}
			txcount := len(mcwrapsInternal)
			chainConfigs := make([]menc.ChainConfig, txcount)
			txs := make([][]byte, txcount)
			for i, msg := range mcwrapsInternal {
				chainId := msg.MultiChainId
				if !slices.Contains(chainIds, chainId) {
					chainIds = append(chainIds, chainId)
				}

				chainConfig, err := multichain.GetSubChainConfig(mcctx.ClientCtx, chainId, registryId)
				if err != nil {
					return err
				}
				chainConfigs[i] = *chainConfig

				_, subappCreator := createMockAppCreator(wasmVmMeta, appFactory, i+1)
				isubchainapp_ := subappCreator(chainId, chainConfig)
				subchainapp_ := isubchainapp_.(appwithTxConfig)
				subtxconfig_ := subchainapp_.TxConfig()

				customCdc := mcodec.NewAccBech32Codec(chainConfig.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
				chainAddrCodec := mcodec.MustUnwrapAccBech32Codec(customCdc)

				msgmc, err := prepareMultiChainExecution(
					mcctx.ClientCtx,
					cmd.Flags(),
					chainAddrCodec,
					msg.Contract,
					sender,
					[]byte(msg.MsgJson),
					nil,
				)
				if err != nil {
					return err
				}

				// prepare tx: add multichain extension options & sign
				txBuilder, txf, err := GenerateTxCLI(mcctx.ClientCtx, cmd.Flags(), msgmc)
				if err != nil {
					return err
				}
				txBuilder, err = msgmc.SetExtensionOptions(txBuilder, chainId, int32(i), int32(txcount))
				if err != nil {
					return err
				}
				txBuilder, err = SignTx(mcctx.ClientCtx, txf, txBuilder)
				if err != nil {
					return err
				}
				txBytes, err := subtxconfig_.TxEncoder()(txBuilder.GetTx())
				if err != nil {
					return err
				}
				txs[i] = txBytes
			}

			leaderChainId := types.GetLeaderChain(chainIds)
			atomictx := &types.MsgExecuteAtomicTxRequest{
				Txs:    txs,
				Sender: sender,
			}
			// prepare atomic tx: add atomic transaction extension options & broadcast unsigned
			txBuilder := subtxconfig.NewTxBuilder()
			err = txBuilder.SetMsgs(atomictx)
			if err != nil {
				return err
			}
			txBuilder, err = atomictx.SetExtensionOptions(txBuilder, chainIds, leaderChainId)
			if err != nil {
				return err
			}
			txBytes, err := subtxconfig.TxEncoder()(txBuilder.GetTx())
			if err != nil {
				return err
			}

			res, err := mcctx.ClientCtx.BroadcastTx(txBytes)
			if err != nil {
				return err
			}

			return mcctx.ClientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

type internalMsgMultiChainWrap struct {
	MultiChainId string `json:"multi_chain_id"`
	MsgJson      string `json:"msg_json"`
	Contract     string `json:"contract"`
}

func parseAndValidateAtomicTxContentJSON(path string) ([]internalMsgMultiChainWrap, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var v []internalMsgMultiChainWrap
	err = json.Unmarshal(contents, &v)
	if err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return nil, fmt.Errorf("must have at least 1 MsgMultiChainWrap")
	}
	for i, msg := range v {
		if msg.Contract == "" {
			return nil, fmt.Errorf("must specify contract address: %d", i)
		}
		if msg.MultiChainId == "" {
			return nil, fmt.Errorf("must specify subchain id: %d", i)
		}
		if len(msg.MsgJson) == 0 {
			return nil, fmt.Errorf("must have non empty message: %d", i)
		}
	}
	return v, nil
}

func RegisterNewSubChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-subchain [chain_base_name] [denom_unit] [decimals] [level_index] [initial_balance]",
		Short: "Register subchain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register subchain

Example:
$ %s tx multichain register-subchain mythos myt 18 1 "10000000000" --chain-id="level0_1000-1"

		`, version.AppName)),
		Aliases: []string{},
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
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
			contractAddr, err := mcctx.CustomAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}

			return sendMultiChainExecution(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
				contractAddr,
				mcctx.ClientCtx.GetFromAddress(),
				msgbz,
				nil,
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func RegisterSubChainValidator(wasmVmMeta memc.IWasmVmMeta, appCreator multichain.NewAppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-subchain-validator [subchain_id] [path/to/validator.json]",
		Short: "Register subchain validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register subchain validator

Example:
$ %s tx multichain register-subchain-validator mythos_7000-1 path/to/validator.json --chain-id="level0_1000-1"

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
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}
			flags := cmd.Flags()
			subChainId := args[0]
			chainId, err := flags.GetString(sdkflags.FlagChainID)
			if err != nil {
				return err
			}
			sender := mcctx.ClientCtx.GetFromAddress()

			validMsg, err := parseAndValidateValidatorJSON(args[1])
			if err != nil {
				return err
			}
			fmt.Println("---sender--", sender)
			genTxData, err := getGenTxData(mcctx.ClientCtx, flags, mcctx.CustomAddrCodec, "consensus", sender, subChainId, validMsg)
			if err != nil {
				return err
			}
			signedGenTxData, err := signGenTxData(wasmVmMeta, appCreator, mcctx.ClientCtx, flags, mcctx.CustomAddrCodec, chainId, subChainId, genTxData, sender)
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
			contractAddr, err := mcctx.CustomAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}

			return sendMultiChainExecution(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
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
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func RegisterLobbyGenTx(wasmVmMeta memc.IWasmVmMeta, appCreator multichain.NewAppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-subchain-gentx [path/to/validator.json]",
		Short: "Register subchain gentx for lobby contract deployed on the provided subchain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register subchain gentx

Example:
$ %s tx multichain register-subchain-gentx path/to/validator.json --chain-id="level0_1000-1"

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
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}
			flags := cmd.Flags()
			chainId, err := flags.GetString(sdkflags.FlagChainID)
			if err != nil {
				return err
			}
			sender := mcctx.ClientCtx.GetFromAddress()

			validMsg, err := parseAndValidateValidatorJSON(args[0])
			if err != nil {
				return err
			}
			genTxData, err := getGenTxData(mcctx.ClientCtx, flags, mcctx.CustomAddrCodec, "lobby", sender, chainId, validMsg)
			if err != nil {
				return err
			}

			signedGenTxData, err := signGenTxDataLobby(wasmVmMeta, appCreator, mcctx.ClientCtx, flags, mcctx.CustomAddrCodec, chainId, genTxData, sender)
			if err != nil {
				return err
			}

			paramBz, err := json.Marshal(&wasmxtypes.ActionParam{Key: "gentx", Value: base64.StdEncoding.EncodeToString(signedGenTxData)})
			if err != nil {
				return err
			}
			msgbz := []byte(fmt.Sprintf(`{"run":{"event":{"type":"addGenTx","params":[%s]}}}`, string(paramBz)))

			lobbyAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_LOBBY)
			contractAddr, err := mcctx.CustomAddrCodec.BytesToString(lobbyAddress)
			if err != nil {
				return err
			}

			return sendMultiChainExecution(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
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
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func InitializeSubChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-subchain [subchain_id]",
		Short: "Initialize a registered subchain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register subchain

Example:
$ %s tx multichain init-subchain level1_1000-1 --chain-id="level0_1000-1"

		`, version.AppName)),
		Aliases: []string{},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
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
			contractAddr, err := mcctx.CustomAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}
			return sendMultiChainExecution(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
				contractAddr,
				mcctx.ClientCtx.GetFromAddress(),
				msgbz,
				nil,
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	sdkflags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)
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

			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
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

			res, err := sendMultiChainQuery(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
				args[0],
				mcctx.ClientCtx.GetFromAddress(),
				queryData,
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
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func GetCmdQuerySubChainValidators() *cobra.Command {
	decoder := wasmxcli.NewArgDecoder(wasmxcli.AsciiDecodeString)
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Show all registered validator address for registered subchains (initialized or not)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Show all subchain validator addresses

Example:
$ %s query multichain validators --chain-id="level0_1000-1"

		`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}
			sunchainId := args[0]

			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := mcctx.CustomAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}

			from, _ := cmd.Flags().GetString(sdkflags.FlagFrom)
			sender, _, _, _ := client.GetFromFields(mcctx.ClientCtx, mcctx.ClientCtx.Keyring, from)

			querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetValidatorAddressesByChainId: &wasmxtypes.QueryValidatorAddressesByChainIdRequest{ChainId: sunchainId}})
			if err != nil {
				return err
			}

			res, err := sendMultiChainQuery(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
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
	multichain.AddMultiChainFlagsToCmd(cmd)
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
$ %s query multichain subchains --chain-id="level0_1000-1"

		`, version.AppName)),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := mcctx.CustomAddrCodec.BytesToString(registryAddress)
			if err != nil {
				return err
			}

			from, _ := cmd.Flags().GetString(sdkflags.FlagFrom)
			sender, _, _, _ := client.GetFromFields(mcctx.ClientCtx, mcctx.ClientCtx.Keyring, from)

			querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetSubChainIds: &wasmxtypes.QueryGetSubChainIdsRequest{}})
			if err != nil {
				return err
			}

			res, err := sendMultiChainQuery(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
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
	multichain.AddMultiChainFlagsToCmd(cmd)
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
$ %s query multichain subchain level1_1000-1 --chain-id="level0_1000-1"

		`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			from, _ := cmd.Flags().GetString(sdkflags.FlagFrom)
			sender, _, _, _ := client.GetFromFields(mcctx.ClientCtx, mcctx.ClientCtx.Keyring, from)

			data2, err := getSubChainConfigBz(mcctx.ClientCtx, cmd.Flags(), mcctx.CustomAddrCodec, args[0], sender)
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
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func getSubChainConfig(clientCtx client.Context, flags *flag.FlagSet, customAddrCodec mcodec.AccBech32Codec, subchainId string, sender sdk.AccAddress) (*menc.ChainConfig, error) {
	chainConfigBz, err := getSubChainConfigBz(clientCtx, flags, customAddrCodec, subchainId, sender)
	if err != nil {
		return nil, err
	}
	var chainConfig menc.ChainConfig
	err = json.Unmarshal(chainConfigBz, &chainConfig)
	if err != nil {
		return nil, err
	}
	return &chainConfig, nil
}

func getSubChainConfigBz(clientCtx client.Context, flags *flag.FlagSet, customAddrCodec mcodec.AccBech32Codec, subchainId string, sender sdk.AccAddress) ([]byte, error) {
	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	contractAddr, err := customAddrCodec.BytesToString(registryAddress)
	if err != nil {
		return nil, err
	}
	querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetSubChainConfigById: &wasmxtypes.QueryGetSubChainRequest{
		ChainId: subchainId,
	}})
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
	return data2, nil
}

func GetCmdQuerySubChain() *cobra.Command {
	decoder := wasmxcli.NewArgDecoder(wasmxcli.AsciiDecodeString)
	cmd := &cobra.Command{
		Use:   "subchain-data [subchain_id]",
		Short: "Show subchain data by id",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Show subchain data by id

Example:
$ %s query multichain subchain-data level1_1000-1 --chain-id="level0_1000-1"

		`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
			contractAddr, err := mcctx.CustomAddrCodec.BytesToString(registryAddress)
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
			sender, _, _, _ := client.GetFromFields(mcctx.ClientCtx, mcctx.ClientCtx.Keyring, from)

			res, err := sendMultiChainQuery(
				mcctx.ClientCtx,
				cmd.Flags(),
				mcctx.CustomAddrCodec,
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
	multichain.AddMultiChainFlagsToCmd(cmd)
	return cmd
}

func sendMultiChainExecution(
	clientCtx client.Context,
	flags *flag.FlagSet,
	addrCodec address.Codec,
	contractAddr string, sender sdk.AccAddress,
	execMsg []byte, amount sdk.Coins,
) error {
	mmsg, err := prepareMultiChainExecution(clientCtx, flags, addrCodec, contractAddr, sender, execMsg, amount)
	if err != nil {
		return err
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, flags, mmsg)
}

func prepareMultiChainExecution(
	clientCtx client.Context,
	flags *flag.FlagSet,
	addrCodec address.Codec,
	contractAddr string, sender sdk.AccAddress,
	execMsg []byte, amount sdk.Coins,
) (*types.MsgMultiChainWrap, error) {
	chainId, err := flags.GetString(sdkflags.FlagChainID)
	if err != nil {
		return nil, fmt.Errorf("chainId: %s", err)
	}
	msg := wasmxtypes.WasmxExecutionMessage{Data: execMsg}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	senderAddr, err := addrCodec.BytesToString(sender)
	if err != nil {
		return nil, err
	}
	wasmxExecMsg := &wasmxtypes.MsgExecuteContract{
		Sender:   senderAddr,
		Contract: contractAddr,
		Funds:    amount,
		Msg:      msgbz,
	}
	wasmxExecAny, err := codectypes.NewAnyWithValue(wasmxExecMsg)
	if err != nil {
		return nil, err
	}

	mmsg := types.MsgMultiChainWrap{
		MultiChainId: chainId,
		Sender:       senderAddr,
		Data:         wasmxExecAny,
	}
	if err := mmsg.ValidateBasic(); err != nil {
		return nil, err
	}
	return &mmsg, nil
}

func sendMultiChainContractStore(
	clientCtx client.Context,
	flags *flag.FlagSet,
	addrCodec mcodec.AccBech32Codec,
	sender sdk.AccAddress,
	registryChainId string,
	pathToWasm string,
) error {
	// registry encoding
	chainId, err := flags.GetString(sdkflags.FlagChainID)
	if err != nil {
		return fmt.Errorf("chainId: %s", err)
	}

	registryConfig, err := mcfg.GetChainConfig(registryChainId)
	if err != nil {
		return err
	}
	customCdc := mcodec.NewAccBech32Codec(registryConfig.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
	registryCodec := mcodec.MustUnwrapAccBech32Codec(customCdc)
	// registry encoding END

	config, err := getSubChainConfig(clientCtx, flags, registryCodec, chainId, sender)
	if err != nil {
		return err
	}
	newaddrCondec := mcodec.NewAccBech32Codec(config.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)

	wasmxMsg, err := wasmxcli.ParseStoreCodeArgs(newaddrCondec, pathToWasm, sender, flags)
	if err != nil {
		return err
	}
	wasmxExecAny, err := codectypes.NewAnyWithValue(&wasmxMsg)
	if err != nil {
		return err
	}

	mmsg := types.MsgMultiChainWrap{
		MultiChainId: chainId,
		Sender:       wasmxMsg.Sender,
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
		return nil, fmt.Errorf("chainId: %s", err)
	}
	if !json.Valid(queryData) {
		return nil, errors.New("query data must be json")
	}
	msg := wasmxtypes.WasmxExecutionMessage{Data: queryData}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if len(sender) == 0 {
		sender = multichain.EMPTY_BYTES20
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
	contractAddress string,
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
	querymsg := []byte(fmt.Sprintf(`{"execute":{"action": {"type": "buildGenTx", "params": [%s],"event":null}}}`, string(paramBz)))

	res, err := sendMultiChainQuery(
		clientCtx,
		flags,
		customAddrCodec,
		contractAddress,
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
	wasmVmMeta memc.IWasmVmMeta,
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
	return signGenTxDataInternal(wasmVmMeta, appCreatorFactory, clientCtx, flags, subChainId, genTxData, sender, subchainConfig)
}

type ChainConfigData struct {
	ChainId     string           `json:"chain_id"`
	ChainConfig menc.ChainConfig `json:"chain_config"`
}

func signGenTxDataLobby(
	wasmVmMeta memc.IWasmVmMeta,
	appCreatorFactory multichain.NewAppCreator,
	clientCtx client.Context,
	flags *flag.FlagSet,
	customAddrCodec mcodec.AccBech32Codec,
	chainId string,
	genTxData []byte,
	sender sdk.AccAddress,
) ([]byte, error) {
	querymsg := []byte(`{"execute":{"action":{"type": "getConfigData","params":[],"event":null}}}`)
	lobbyAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_LOBBY)
	contractAddr, err := customAddrCodec.BytesToString(lobbyAddress)
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

	var subchainConfig ChainConfigData
	err = json.Unmarshal(data2, &subchainConfig)
	if err != nil {
		return nil, err
	}
	return signGenTxDataInternal(wasmVmMeta, appCreatorFactory, clientCtx, flags, subchainConfig.ChainId, genTxData, sender, subchainConfig.ChainConfig)
}

func signGenTxDataInternal(
	wasmVmMeta memc.IWasmVmMeta,
	appCreatorFactory multichain.NewAppCreator,
	clientCtx client.Context,
	flags *flag.FlagSet,
	subChainId string,
	genTxData []byte,
	sender sdk.AccAddress,
	subchainConfig menc.ChainConfig,
) ([]byte, error) {

	_, appCreator := createMockAppCreator(wasmVmMeta, appCreatorFactory, 0)
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

	newClientCtx := clientCtx.
		WithChainID(subChainId).
		WithTxConfig(subtxconfig).
		WithCodec(subchainapp.AppCodec()).
		WithInterfaceRegistry(subchainapp.InterfaceRegistry()).
		WithLegacyAmino(subchainapp.LegacyAmino())
		// WithAccountRetriever(cosmosmodtypes.AccountRetriever{AddressCodec: addrcodec}).

	txf, err := tx.NewFactoryCLI(clientCtx, flags)
	if err != nil {
		return nil, err
	}
	txbuilder, err = signTxWithSignerData(newClientCtx, txf, flags, txbuilder, subtxconfig, sender, subChainId, 0, 0)
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
	return signTxWithSignerData(clientCtx, txf, flags, txbuilder, txconfig, sender, subChainId, txf.AccountNumber(), txf.Sequence())
}

func signTxWithSignerData(
	clientCtx client.Context,
	txf tx.Factory,
	flags *flag.FlagSet,
	txbuilder client.TxBuilder,
	txconfig client.TxConfig,
	sender sdk.AccAddress,
	subChainId string,
	accountNumber uint64,
	sequence uint64,
) (client.TxBuilder, error) {
	k, err := txf.Keybase().Key(clientCtx.FromName)
	if err != nil {
		return nil, err
	}

	pubKey, err := k.GetPubKey()
	if err != nil {
		return nil, err
	}

	sigV2 := txsigning.SignatureV2{
		PubKey: pubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode(txconfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: sequence,
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
		AccountNumber: accountNumber,
		Sequence:      sequence,
		PubKey:        pubKey,
		Address:       subchainSender,
	}
	signMode := txsigning.SignMode(txconfig.SignModeHandler().DefaultMode())

	bytesToSign, err := authsigning.GetSignBytesAdapter(clientCtx.CmdContext, txconfig.SignModeHandler(), signMode, signerData, txbuilder.GetTx())
	if err != nil {
		return nil, err
	}

	sigBytes, _, err := txf.Keybase().Sign(clientCtx.FromName, bytesToSign, signMode)
	if err != nil {
		return nil, err
	}
	// Construct the SignatureV2 struct
	sigData := txsigning.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}

	sigV2 = txsigning.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: sequence,
	}
	err = txbuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}
	return txbuilder, nil
}

func createMockAppCreator(wasmVmMeta memc.IWasmVmMeta, appCreatorFactory multichain.NewAppCreator, index int) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	tempNodeHome := filepath.Join(userHomeDir, fmt.Sprintf(".mythostmp_%d", index))
	return multichain.CreateNoLoggerAppCreator(wasmVmMeta, appCreatorFactory, tempNodeHome)
}
