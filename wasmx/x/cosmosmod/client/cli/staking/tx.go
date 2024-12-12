package staking

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	cli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
	"github.com/loredanacirstea/wasmx/v1/multichain"
	memc "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm/memory/common"
)

// default values
var (
	DefaultTokens = sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	FlagMemo      = "memo"
)

// NewTxCmd returns a root CLI command handler for all x/staking transaction commands.
func NewTxCmd(wasmVmMeta memc.IWasmVmMeta, valAddrCodec, ac address.Codec, appCreator multichain.NewAppCreator) *cobra.Command {
	stakingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingTxCmd.AddCommand(
		NewCreateValidatorCmd(wasmVmMeta, valAddrCodec, ac, appCreator),
		NewEditValidatorCmd(valAddrCodec, ac),
		NewDelegateCmd(valAddrCodec, ac),
		NewRedelegateCmd(valAddrCodec, ac),
		NewUnbondCmd(valAddrCodec, ac),
		NewCancelUnbondingDelegation(valAddrCodec, ac),
	)

	return stakingTxCmd
}

// NewCreateValidatorCmd returns a CLI command handler for creating a MsgCreateValidator transaction.
func NewCreateValidatorCmd(wasmVmMeta memc.IWasmVmMeta, valAddrCodec address.Codec, ac address.Codec, appFactory multichain.NewAppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator [path/to/validator.json]",
		Short: "create new validator initialized with a self-delegation to it",
		Args:  cobra.ExactArgs(1),
		Long:  `Create a new validator initialized with a self-delegation by submitting a JSON file with the new validator details.`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %s tx staking create-validator path/to/validator.json --from keyname

Where validator.json contains:

{
	"pubkey": {"type_url":"/cosmos.crypto.ed25519.PubKey","value":"eyJrZXkiOiJqZXFaQU1mSHpZS1cyT0IraDlFWktiRW9mVEhETHduWG1sWSt5bXFkdk1jPSJ9"},
	"amount": "1000000stake",
	"moniker": "myvalidator",
	"identity": "optional identity signature (ex. UPort or Keybase)",
	"website": "validator's (optional) website",
	"security": "validator's (optional) security contact email",
	"details": "validator's (optional) details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
}

where we can get the pubkey using "%s tendermint show-validator"
`, version.AppName, version.AppName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			mcctx, err := multichain.MultiChainCtxByChainId(clientCtx, cmd.Flags(), []signing.CustomGetSigner{})
			if err != nil {
				return err
			}

			chainId := mcctx.ClientCtx.ChainID
			customValCdc := mcodec.NewValBech32Codec(mcctx.Config.Bech32PrefixValAddr, mcodec.NewAddressPrefixedFromVal)
			customValCodec := mcodec.MustUnwrapValBech32Codec(customValCdc)

			_, appCreator := createMockAppCreator(wasmVmMeta, appFactory, 0)
			chainapp := appCreator(chainId, mcctx.Config)

			txf, err := tx.NewFactoryCLI(mcctx.ClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			validator, err := parseAndValidateValidatorJSON(chainapp.AppCodec(), args[0])
			if err != nil {
				return fmt.Errorf("could not parse validator JSON file: %s", err.Error())
			}

			txf, msg, err := newBuildCreateValidatorMsg(mcctx.ClientCtx, customValCodec, mcctx.CustomAddrCodec, txf, cmd.Flags(), validator)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(mcctx.ClientCtx, txf, msg)
		},
	}

	// cmd.Flags().String(cli.FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	// cmd.Flags().String(cli.FlagNodeID, "", "The node's ID")
	cmd.Flags().String(FlagMemo, "", "The node's networking address")
	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

// NewEditValidatorCmd returns a CLI command handler for creating a MsgEditValidator transaction.
func NewEditValidatorCmd(valAddrCodec address.Codec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			moniker, _ := cmd.Flags().GetString(cli.FlagEditMoniker)
			identity, _ := cmd.Flags().GetString(cli.FlagIdentity)
			website, _ := cmd.Flags().GetString(cli.FlagWebsite)
			security, _ := cmd.Flags().GetString(cli.FlagSecurityContact)
			details, _ := cmd.Flags().GetString(cli.FlagDetails)
			description := types.NewDescription(moniker, identity, website, security, details)

			var newRate *math.LegacyDec

			commissionRate, _ := cmd.Flags().GetString(cli.FlagCommissionRate)
			if commissionRate != "" {
				rate, err := math.LegacyNewDecFromStr(commissionRate)
				if err != nil {
					return fmt.Errorf("invalid new commission rate: %v", err)
				}

				newRate = &rate
			}

			var newMinSelfDelegation *math.Int

			minSelfDelegationString, _ := cmd.Flags().GetString(cli.FlagMinSelfDelegation)
			if minSelfDelegationString != "" {
				msb, ok := math.NewIntFromString(minSelfDelegationString)
				if !ok {
					return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
				}

				newMinSelfDelegation = &msb
			}

			valAddr, err := valAddrCodec.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgEditValidator(valAddr, description, newRate, newMinSelfDelegation)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(flagSetDescriptionEdit())
	cmd.Flags().AddFlagSet(flagSetCommissionUpdate())
	cmd.Flags().AddFlagSet(cli.FlagSetMinSelfDelegation())
	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewDelegateCmd returns a CLI command handler for creating a MsgDelegate transaction.
func NewDelegateCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate [validator-addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Delegate liquid tokens to a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delegate an amount of liquid coins to a validator from your wallet.

Example:
$ %s tx staking delegate cosmosvalopers1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegate(delAddr, args[0], amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewRedelegateCmd returns a CLI command handler for creating a MsgBeginRedelegate transaction.
func NewRedelegateCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "Redelegate illiquid tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redelegate an amount of illiquid staking tokens from one validator to another.

Example:
$ %s tx staking redelegate cosmosvalopers1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj cosmosvalopers1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100stake --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[1])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgBeginRedelegate(delAddr, args[0], args[1], amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewUnbondCmd returns a CLI command handler for creating a MsgUndelegate transaction.
func NewUnbondCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "unbond [validator-addr] [amount]",
		Short: "Unbond shares from a validator",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unbond an amount of bonded shares from a validator.

Example:
$ %s tx staking unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgUndelegate(delAddr, args[0], amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

// NewCancelUnbondingDelegation returns a CLI command handler for creating a MsgCancelUnbondingDelegation transaction.
func NewCancelUnbondingDelegation(valAddrCodec, ac address.Codec) *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "cancel-unbond [validator-addr] [amount] [creation-height]",
		Short: "Cancel unbonding delegation and delegate back to the validator",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel Unbonding Delegation and delegate back to the validator.

Example:
$ %s tx staking cancel-unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake 2 --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Example: fmt.Sprintf(`$ %s tx staking cancel-unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake 2 --from mykey`,
			version.AppName, bech32PrefixValAddr),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			creationHeight, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return errorsmod.Wrap(fmt.Errorf("invalid height: %d", creationHeight), "invalid height")
			}

			msg := types.NewMsgCancelUnbondingDelegation(delAddr, args[0], creationHeight, amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	multichain.AddMultiChainFlagsToCmd(cmd)

	return cmd
}

func newBuildCreateValidatorMsg(clientCtx client.Context, customValCodec mcodec.ValBech32Codec, valAc address.Codec, txf tx.Factory, fs *flag.FlagSet, val validator) (tx.Factory, *types.MsgCreateValidator, error) {
	valAddr := customValCodec.BytesToValAddressPrefixed(clientCtx.GetFromAddress())

	description := types.NewDescription(
		val.Moniker,
		val.Identity,
		val.Website,
		val.Security,
		val.Details,
	)

	var pkAny *codectypes.Any
	if val.PubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(val.PubKey); err != nil {
			return txf, nil, err
		}
	}

	msg := &types.MsgCreateValidator{
		Description:       description,
		ValidatorAddress:  valAddr.String(),
		Pubkey:            pkAny,
		Value:             val.Amount,
		Commission:        val.CommissionRates,
		MinSelfDelegation: val.MinSelfDelegation,
	}

	if err := msg.Validate(valAc); err != nil {
		return txf, nil, err
	}

	memo, _ := fs.GetString(FlagMemo)
	if memo != "" {
		txf = txf.WithMemo(memo)
	}

	return txf, msg, nil
}
