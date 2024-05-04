package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	address "cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"mythos/v1/multichain"
	"mythos/v1/x/wasmx/ioutils"
	"mythos/v1/x/wasmx/types"
)

const (
	flagAmount                    = "amount"
	flagLabel                     = "label"
	flagSource                    = "code-source-url"
	flagBuilder                   = "builder"
	flagCodeHash                  = "code-hash"
	flagAdmin                     = "admin"
	flagNoAdmin                   = "no-admin"
	flagFixMsg                    = "fix-msg"
	flagRunAs                     = "run-as"
	flagInstantiateByEverybody    = "instantiate-everybody"
	flagInstantiateNobody         = "instantiate-nobody"
	flagInstantiateByAddress      = "instantiate-only-address"
	flagInstantiateByAnyOfAddress = "instantiate-anyof-addresses"
	flagUnpinCode                 = "unpin-code"
	flagAllowedMsgKeys            = "allow-msg-keys"
	flagAllowedRawMsgs            = "allow-raw-msgs"
	flagExpiration                = "expiration"
	flagMaxCalls                  = "max-calls"
	flagMaxFunds                  = "max-funds"
	flagAllowAllMsgs              = "allow-all-messages"
	flagNoTokenTransfer           = "no-token-transfer" //nolint:gosec
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(ac address.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	txCmd.AddCommand(
		StoreCodeCmd(ac),
		// DeployCmd(),
		InstantiateContractCmd(ac),
		InstantiateContract2Cmd(ac),
		ExecuteContractCmd(ac),
		NewRegisterRoleProposalCmd(ac),
		NewDeregisterRoleProposalCmd(ac),
	)
	return txCmd
}

// StoreCodeCmd will upload code to be reused.
func StoreCodeCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "store [wasm file]",
		Short:   "Upload a wasm binary",
		Aliases: []string{"upload", "st", "s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx, addrCodec, customAddrCodec, err := multichain.MultiChainCtx(ac, clientCtx)
			if err != nil {
				return err
			}
			msg, err := parseStoreCodeArgs(addrCodec, args[0], clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			// TODO use ValidateWithAddress with codec
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			fromAddr := customAddrCodec.BytesToAccAddressPrefixed(clientCtx.GetFromAddress())
			msgMultiChain, err := multichain.MultiChainWrap(clientCtx, &msg, fromAddr)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgMultiChain)
		},
		SilenceUsage: true,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseStoreCodeArgs(addrCodec address.Codec, file string, sender sdk.AccAddress, flags *flag.FlagSet) (types.MsgStoreCode, error) {
	wasm, err := os.ReadFile(file)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	// gzip the wasm file
	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)

		if err != nil {
			return types.MsgStoreCode{}, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return types.MsgStoreCode{}, fmt.Errorf("invalid input file. Use wasm binary or gzip")
	}

	senderstr, err := addrCodec.BytesToString(sender)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	msg := types.MsgStoreCode{
		Sender:   senderstr,
		ByteCode: wasm,
	}
	return msg, nil
}

// // DeployCmd will upload code to be reused.
// func DeployCmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:     "deploy-code [hex-bytecode]",
// 		Short:   "Upload and deploy interpreted bytecode, like EVM bytecode",
// 		Aliases: []string{"deploy-evm"},
// 		Args:    cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			clientCtx, err := client.GetClientTxContext(cmd)
// 			if err != nil {
// 				return err
// 			}

// 			sender := clientCtx.GetFromAddress()
// 			evmByteCode, err := hex.DecodeString(args[0])
// 			if err != nil {
// 				return err
// 			}

// 			msg := types.MsgDeploy{
// 				Sender:   sender.String(), // TODO replace this
// 				ByteCode: evmByteCode,
// 			}

// 			if err = msg.ValidateBasic(); err != nil {
// 				return err
// 			}
// 			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
// 		},
// 		SilenceUsage: true,
// 	}

// 	flags.AddTxFlagsToCmd(cmd)
// 	return cmd
// }

// InstantiateContractCmd will instantiate a contract from previously uploaded code.
func InstantiateContractCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instantiate [code_id_int64] [json_encoded_init_args] --label [text] --admin [address,optional] --amount [coins,optional] ",
		Short: "Instantiate a wasm contract",
		Long: fmt.Sprintf(`Creates a new instance of an uploaded wasm code with the given 'constructor' message.
Each contract instance has a unique address assigned.
Example:
$ %s tx wasmx instantiate 1 '{"foo":"bar"}' --admin="$(%s keys show mykey -a)" \
  --from mykey --amount="100amyt" --label "local0.1.0"
`, version.AppName, version.AppName),
		Aliases: []string{"start", "init", "inst", "i"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx, addrCodec, customAddrCodec, err := multichain.MultiChainCtx(ac, clientCtx)
			if err != nil {
				return err
			}

			msg, err := parseInstantiateArgs(addrCodec, args[0], args[1], clientCtx.Keyring, clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			fromAddr := customAddrCodec.BytesToAccAddressPrefixed(clientCtx.GetFromAddress())
			msgMultiChain, err := multichain.MultiChainWrap(clientCtx, msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgMultiChain)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(flagLabel, "", "A human-readable name for this contract in lists")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// InstantiateContract2Cmd will instantiate a contract from previously uploaded code with predicable address generated
func InstantiateContract2Cmd(ac address.Codec) *cobra.Command {
	decoder := NewArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use: "instantiate2 [code_id_int64] [json_encoded_init_args] [salt] --label [text] --admin [address,optional] --amount [coins,optional] " +
			"--fix-msg [bool,optional]",
		Short: "Instantiate a wasm contract with predictable address",
		Long: fmt.Sprintf(`Creates a new instance of an uploaded wasm code with the given 'constructor' message.
Each contract instance has a unique address assigned. They are assigned automatically but in order to have predictable addresses
for special use cases, the given 'salt' argument and '--fix-msg' parameters can be used to generate a custom address.

Predictable address example (also see '%s query wasm build-address -h'):
$ %s tx wasmx instantiate2 1 '{"foo":"bar"}' $(echo -n "testing" | xxd -ps) --admin="$(%s keys show mykey -a)" \
  --from mykey --amount="100ustake" --label "local0.1.0" \
   --fix-msg
`, version.AppName, version.AppName, version.AppName),
		Aliases: []string{"start", "init", "inst", "i"},
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx, addrCodec, customAddrCodec, err := multichain.MultiChainCtx(ac, clientCtx)
			if err != nil {
				return err
			}
			salt, err := decoder.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("salt: %w", err)
			}
			fixMsg, err := cmd.Flags().GetBool(flagFixMsg)
			if err != nil {
				return fmt.Errorf("fix msg: %w", err)
			}
			data, err := parseInstantiateArgs(addrCodec, args[0], args[1], clientCtx.Keyring, clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			msg := &types.MsgInstantiateContract2{
				Sender: data.Sender,
				CodeId: data.CodeId,
				Label:  data.Label,
				Msg:    data.Msg,
				Funds:  data.Funds,
				Salt:   salt,
				FixMsg: fixMsg,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			fromAddr := customAddrCodec.BytesToAccAddressPrefixed(clientCtx.GetFromAddress())
			msgMultiChain, err := multichain.MultiChainWrap(clientCtx, msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgMultiChain)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(flagLabel, "", "A human-readable name for this contract in lists")
	decoder.RegisterFlags(cmd.PersistentFlags(), "salt")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseInstantiateArgs(addrCodec address.Codec, rawCodeID, initMsg string, kr keyring.Keyring, sender sdk.AccAddress, flags *flag.FlagSet) (*types.MsgInstantiateContract, error) {
	// get the id of the code to instantiate
	codeID, err := strconv.ParseUint(rawCodeID, 10, 64)
	if err != nil {
		return nil, err
	}

	amountStr, err := flags.GetString(flagAmount)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	label, err := flags.GetString(flagLabel)
	if err != nil {
		return nil, fmt.Errorf("label: %s", err)
	}
	if label == "" {
		return nil, errors.New("label is required on all contracts")
	}
	msgbz, err := wasmxMsgWrap(initMsg)
	if err != nil {
		return nil, err
	}

	senderstr, err := addrCodec.BytesToString(sender)
	if err != nil {
		return nil, err
	}

	// build and sign the transaction, then broadcast to Tendermint
	msg := types.MsgInstantiateContract{
		Sender: senderstr,
		CodeId: codeID,
		Label:  label,
		Funds:  amount,
		Msg:    msgbz,
	}
	return &msg, nil
}

// ExecuteContractCmd will instantiate a contract from previously uploaded code.
func ExecuteContractCmd(ac address.Codec) *cobra.Command {
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
			addrCodec := clientCtx.InterfaceRegistry.SigningContext().AddressCodec()
			msg, err := parseExecuteArgs(addrCodec, args[0], args[1], clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			clientCtx, addrCodec, customAddrCodec, err := multichain.MultiChainCtx(ac, clientCtx)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			fromAddr := customAddrCodec.BytesToAccAddressPrefixed(clientCtx.GetFromAddress())
			msgMultiChain, err := multichain.MultiChainWrap(clientCtx, &msg, fromAddr)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgMultiChain)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseExecuteArgs(addrCodec address.Codec, contractAddr string, execMsg string, sender sdk.AccAddress, flags *flag.FlagSet) (types.MsgExecuteContract, error) {
	amountStr, err := flags.GetString(flagAmount)
	if err != nil {
		return types.MsgExecuteContract{}, fmt.Errorf("amount: %s", err)
	}

	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return types.MsgExecuteContract{}, err
	}
	msgbz, err := wasmxMsgWrap(execMsg)
	if err != nil {
		return types.MsgExecuteContract{}, err
	}
	senderstr, err := addrCodec.BytesToString(sender)
	if err != nil {
		return types.MsgExecuteContract{}, err
	}
	return types.MsgExecuteContract{
		Sender:   senderstr,
		Contract: contractAddr,
		Funds:    amount,
		Msg:      msgbz,
	}, nil
}

func wasmxMsgWrap(jsonmsg string) ([]byte, error) {
	msg := types.WasmxExecutionMessage{Data: []byte(jsonmsg)}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return msgbz, nil
}
