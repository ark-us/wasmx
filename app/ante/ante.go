package ante

import (
	"fmt"
	"runtime/debug"
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdklog "cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"mythos/v1/crypto/ethsecp256k1"
	networktypes "mythos/v1/x/network/types"
)

const (
	Secp256k1VerifyCost uint64 = 21000
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(cdc codec.Codec, txConfig client.TxConfig, options HandlerOptions) (sdk.AnteHandler, error) {
	if err := options.validate(); err != nil {
		return nil, err
	}

	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		defer Recover(ctx.Logger(), &err)

		txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case networktypes.TypeURL_ExtensionOptionEthereumTx:
					anteHandler = newEthAnteHandler(options)
				case networktypes.TypeURL_ExtensionOptionAtomicMultiChainTx:
					ext := opts[0].GetCachedValue().(*networktypes.ExtensionOptionAtomicMultiChainTx)
					if !slices.Contains(ext.ChainIds, ext.LeaderChainId) {
						return ctx, errorsmod.Wrapf(
							errortypes.ErrInvalidRequest,
							"rejecting tx with ExtensionOptionAtomicMultiChainTx: leader chain id is not included in the chain ids array",
						)
					}
					if !slices.Contains(ext.ChainIds, ctx.ChainID()) {
						return ctx, errorsmod.Wrapf(
							errortypes.ErrInvalidRequest,
							"rejecting tx with ExtensionOptionAtomicMultiChainTx: not for current chain",
						)
					}
					if len(ext.ChainIds) > 2 {
						return ctx, errorsmod.Wrapf(
							errortypes.ErrInvalidRequest,
							"rejecting tx with ExtensionOptionAtomicMultiChainTx: atomic tx running on more than 2 chains",
						)
					}
					// CheckTx - only the AnteHandler is executed
					// all the rest execute the msgs, so the subtx antehandlers are executed then
					isCheck := ctx.ExecMode() == sdk.ExecModeCheck || ctx.ExecMode() == sdk.ExecModeReCheck
					if !isCheck {
						return ctx, nil
					}
					// we decompose the atomic tx & validate each subtx
					// we only expect one message per atomic tx
					if len(tx.GetMsgs()) != 1 {
						return ctx, errorsmod.Wrapf(
							errortypes.ErrInvalidRequest,
							"rejecting tx with ExtensionOptionAtomicMultiChainTx: requires 1 MsgExecuteAtomicTxRequest message",
						)
					}
					msg := tx.GetMsgs()[0]
					atomicTx, ok := msg.(*networktypes.MsgExecuteAtomicTxRequest)
					if !ok {
						return ctx, errorsmod.Wrapf(
							errortypes.ErrInvalidRequest,
							"rejecting tx with ExtensionOptionAtomicMultiChainTx: requires MsgExecuteAtomicTxRequest message",
						)
					}
					anteHandler, _ := NewAnteHandler(cdc, txConfig, options)
					for _, subtxbz := range atomicTx.Txs {
						subtx, err := txConfig.TxDecoder()(subtxbz)
						if err != nil {
							return ctx, err
						}

						// we check that the chainId of the subtx is included in the atomic wrapper
						subtxWithExtensions, ok := subtx.(authante.HasExtensionOptionsTx)
						if !ok {
							return ctx, errorsmod.Wrapf(
								errortypes.ErrInvalidRequest,
								"rejecting subtx with no extension options",
							)
						}
						subopts := subtxWithExtensions.GetExtensionOptions()
						if len(subopts) == 0 {
							return ctx, errorsmod.Wrapf(
								errortypes.ErrInvalidRequest,
								"rejecting subtx with no ExtensionOptionMultiChainTx",
							)
						}
						extsubtx := subopts[0].GetCachedValue().(*networktypes.ExtensionOptionMultiChainTx)
						if !slices.Contains(ext.ChainIds, extsubtx.ChainId) {
							return ctx, errorsmod.Wrapf(
								errortypes.ErrInvalidRequest,
								"rejecting subtx with chain_id not included in ExtensionOptionAtomicMultiChainTx",
							)
						}

						ctx, err = anteHandler(ctx, subtx, sim)
						if err != nil {
							return ctx, err
						}
					}
					return ctx, nil
				case networktypes.TypeURL_ExtensionOptionMultiChainTx:
					ext := opts[0].GetCachedValue().(*networktypes.ExtensionOptionMultiChainTx)
					if ctx.ChainID() == ext.ChainId {
						if len(opts) > 1 && opts[1].GetTypeUrl() == networktypes.TypeURL_ExtensionOptionEthereumTx {
							anteHandler = newEthAnteHandler(options)
						} else {
							anteHandler = newCosmosAnteHandler(options)
						}
					} else {
						// we skip antehandler verification on transactions that are not meant to be ran on this chain
						return ctx, nil
					}
				default:
					return ctx, errorsmod.Wrapf(
						errortypes.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s", typeURL,
					)
				}

				return anteHandler(ctx, tx, sim)
			}
		}

		// handle as totally normal Cosmos SDK tx
		switch tx.(type) {
		case sdk.Tx:
			anteHandler = newCosmosAnteHandler(options)
		default:
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}, nil
}

func Recover(logger sdklog.Logger, err *error) {
	if r := recover(); r != nil {
		*err = errorsmod.Wrapf(errortypes.ErrPanic, "%v", r)

		if e, ok := r.(error); ok {
			logger.Error(
				"ante handler panicked",
				"error", e,
				"stack trace", string(debug.Stack()),
			)
		} else {
			logger.Error(
				"ante handler panicked",
				"recover", fmt.Sprintf("%v", r),
			)
		}
	}
}

var _ authante.SignatureVerificationGasConsumer = DefaultSigVerificationGasConsumer

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	pubkey := sig.PubKey
	switch pubkey := pubkey.(type) {
	case *ethsecp256k1.PubKey:
		meter.ConsumeGas(Secp256k1VerifyCost, "ante verify: eth_secp256k1")
		return nil

	case multisig.PubKey:
		// Multisig keys
		multisignature, ok := sig.Data.(*signing.MultiSignatureData)
		if !ok {
			return fmt.Errorf("expected %T, got, %T", &signing.MultiSignatureData{}, sig.Data)
		}
		return ConsumeMultisignatureVerificationGas(meter, multisignature, pubkey, params, sig.Sequence)

	default:
		return authante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}

// ConsumeMultisignatureVerificationGas consumes gas from a GasMeter for verifying a multisig pubkey signature
func ConsumeMultisignatureVerificationGas(
	meter storetypes.GasMeter, sig *signing.MultiSignatureData, pubkey multisig.PubKey,
	params authtypes.Params, accSeq uint64,
) error {
	size := sig.BitArray.Count()
	sigIndex := 0

	for i := 0; i < size; i++ {
		if !sig.BitArray.GetIndex(i) {
			continue
		}
		sigV2 := signing.SignatureV2{
			PubKey:   pubkey.GetPubKeys()[i],
			Data:     sig.Signatures[sigIndex],
			Sequence: accSeq,
		}
		err := DefaultSigVerificationGasConsumer(meter, sigV2, params)
		if err != nil {
			return err
		}
		sigIndex++
	}

	return nil
}
