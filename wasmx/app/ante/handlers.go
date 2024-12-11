package ante

import (
	mcodec "wasmx/v1/codec"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	// "github.com/cosmos/cosmos-sdk/x/auth/ante"
	circuitante "cosmossdk.io/x/circuit/ante"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper, EVM Keeper and Fee Market Keeper.
type HandlerOptions struct {
	// Cosmos SDK HandlerOptions
	AccountKeeper          AccountKeeper
	BankKeeper             authtypes.BankKeeper
	ExtensionOptionChecker sdkante.ExtensionOptionChecker
	FeegrantKeeper         sdkante.FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	TxFeeChecker           TxFeeChecker
	CircuitKeeper          circuitante.CircuitBreaker

	WasmxKeeper WasmxKeeperI

	// TODO needed?
	IBCKeeper *ibckeeper.Keeper
	// EvmKeeper              EVMKeeper
	MaxTxGasWanted    uint64
	DisabledAuthzMsgs []string

	AccBech32Codec mcodec.AccBech32Codec
}

func (options HandlerOptions) validate() error {
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.CircuitKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "circuit keeper is required for AnteHandler")
	}
	// if options.FeeMarketKeeper == nil {
	// 	return errorsmod.Wrap(errortypes.ErrLogic, "fee market keeper is required for AnteHandler")
	// }
	// if options.EvmKeeper == nil {
	// 	return errorsmod.Wrap(errortypes.ErrLogic, "evm keeper is required for AnteHandler")
	// }
	return nil
}

func newEthAnteHandler(options HandlerOptions) sdk.AnteHandler {
	anteDecorators := []sdk.AnteDecorator{
		sdkante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		// TODO?
		// sdkante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		// sdkante.NewValidateBasicDecorator(),
		// sdkante.NewTxTimeoutHeightDecorator(),
		// sdkante.NewValidateMemoDecorator(options.AccountKeeper),
		// sdkante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// sdkante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		sdkante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		// NewValidateSigCountDecorator
		sdkante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		NewEthSigVerificationDecorator(options.WasmxKeeper),
		// NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		// sdkante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}

// 	RejectMessagesDecorator{}, // reject MsgEthereumTxs
// 	// disable the Msg types that cannot be included on an authz.MsgExec msgs field
// 	NewAuthzLimiterDecorator(options.DisabledAuthzMsgs),
// 	NewMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
// 	ante.NewValidateMemoDecorator(options.AccountKeeper),
// 	ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
// 	NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),

func newCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	anteDecorators := []sdk.AnteDecorator{
		sdkante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		circuitante.NewCircuitBreakerDecorator(options.CircuitKeeper),
		// TODO restrict ethereumtx extension, but not atomictx extensions
		// sdkante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		sdkante.NewValidateBasicDecorator(),
		sdkante.NewTxTimeoutHeightDecorator(),
		sdkante.NewValidateMemoDecorator(options.AccountKeeper),
		sdkante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker, options.AccBech32Codec),
		sdkante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		sdkante.NewValidateSigCountDecorator(options.AccountKeeper),
		sdkante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler, options.AccBech32Codec),
		sdkante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}
