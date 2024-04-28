package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	mcfg "mythos/v1/config"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// EthSigVerificationDecorator validates an ethereum signatures
type EthSigVerificationDecorator struct {
	wasmxKeeper WasmxKeeperI
}

// NewEthSigVerificationDecorator creates a new EthSigVerificationDecorator
func NewEthSigVerificationDecorator(wasmxKeeper WasmxKeeperI) EthSigVerificationDecorator {
	return EthSigVerificationDecorator{wasmxKeeper: wasmxKeeper}
}

// AnteHandle validates checks that the registered chain id is the same as the one on the message, and
// that the signer address matches the one defined on the message.
// It's not skipped for RecheckTx, because it set `From` address which is critical from other ante handler to work.
// Failure in RecheckTx will prevent tx to be included into block, especially when CheckTx succeed, in which case user
// won't see the error message.
func (esvd EthSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// TODO
	// chainID := esvd.evmKeeper.ChainID()
	// evmParams := esvd.evmKeeper.GetParams(ctx)
	// chainCfg := evmParams.GetChainConfig()
	// ethCfg := chainCfg.EthereumConfig(chainID)
	// blockNum := big.NewInt(ctx.BlockHeight())
	// signer := ethtypes.MakeSigner(ethCfg, blockNum)
	chainID, err := wasmxtypes.ParseChainID(ctx.ChainID())
	if err != nil {
		return ctx, errorsmod.Wrapf(
			errortypes.ErrInvalidChainID,
			"could not parse chainId: %s",
			err.Error(),
		)
	}
	ethSigner := ethtypes.LatestSignerForChainID(chainID)

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*wasmxtypes.MsgExecuteEth)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*wasmxtypes.MsgExecuteEth)(nil))
		}

		// TODO
		// allowUnprotectedTxs := evmParams.GetAllowUnprotectedTxs()
		allowUnprotectedTxs := true
		ethTx := msgEthTx.AsTransaction()
		if !allowUnprotectedTxs && !ethTx.Protected() {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrNotSupported,
				"rejected unprotected Ethereum transaction. Please EIP155 sign your transaction to protect it against replay-attacks")
		}

		// Don't check signature on a simulation, because it may be sent by JSON-RPC
		// through the gas estimation endpoint
		if simulate && msgEthTx.Sender != "" {
			sender, err := esvd.wasmxKeeper.AddressCodec().StringToBytes(msgEthTx.Sender)
			if err != nil {
				return ctx, errorsmod.Wrapf(err, "simulate transaction has invalid sender")
			}
			aliasAddr, found := esvd.wasmxKeeper.GetAlias(ctx, sender)
			if found {
				aliasAddrStr, err := esvd.wasmxKeeper.AddressCodec().BytesToString(aliasAddr)
				if err != nil {
					return ctx, errorsmod.Wrapf(err, "alias: %s", mcfg.ERRORMSG_ACC_TOSTRING)
				}
				msgEthTx.Sender = aliasAddrStr
			}
			continue
		}

		_ethSigner := ethSigner
		if !ethTx.Protected() {
			_ethSigner = ethtypes.HomesteadSigner{}
		}

		sender, err := msgEthTx.GetSignerFromSignature(_ethSigner)
		if err != nil {
			return ctx, err
		}

		senderStr, err := esvd.wasmxKeeper.AddressCodec().BytesToString(sender)
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "sender: %s", mcfg.ERRORMSG_ACC_TOSTRING)
		}

		// set up the sender to the transaction field if not already
		if msgEthTx.Sender != senderStr {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrUnauthorized,
				"eth transaction signer does not match with Sender : %s",
				err.Error(),
			)
		}

		aliasAddr, found := esvd.wasmxKeeper.GetAlias(ctx, sender)
		if found {
			aliasAddrStr, err := esvd.wasmxKeeper.AddressCodec().BytesToString(aliasAddr)
			if err != nil {
				return ctx, errorsmod.Wrapf(err, "alias: %s", mcfg.ERRORMSG_ACC_TOSTRING)
			}
			msgEthTx.Sender = aliasAddrStr
		}
	}

	return next(ctx, tx, simulate)
}
