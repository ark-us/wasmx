package ante

import (
	"fmt"

	mcodec "wasmx/v1/codec"

	"google.golang.org/protobuf/types/known/anypb"

	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// SigVerificationDecorator verifies all signatures for a tx and return an error if any are invalid. Note,
// the SigVerificationDecorator will not check signatures on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type SigVerificationDecorator struct {
	ak              AccountKeeper
	signModeHandler *txsigning.HandlerMap
	accBech32Codec  mcodec.AccBech32Codec
}

func NewSigVerificationDecorator(ak AccountKeeper, signModeHandler *txsigning.HandlerMap, accBech32Codec mcodec.AccBech32Codec) SigVerificationDecorator {
	return SigVerificationDecorator{
		ak:              ak,
		signModeHandler: signModeHandler,
		accBech32Codec:  accBech32Codec,
	}
}

// OnlyLegacyAminoSigners checks SignatureData to see if all
// signers are using SIGN_MODE_LEGACY_AMINO_JSON. If this is the case
// then the corresponding SignatureV2 struct will not have account sequence
// explicitly set, and we should skip the explicit verification of sig.Sequence
// in the SigVerificationDecorator's AnteHandler function.
func OnlyLegacyAminoSigners(sigData signing.SignatureData) bool {
	switch v := sigData.(type) {
	case *signing.SingleSignatureData:
		return v.SignMode == signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case *signing.MultiSignatureData:
		for _, s := range v.Signatures {
			if !OnlyLegacyAminoSigners(s) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (svd SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}

	signers, err := sigTx.GetSigners()
	if err != nil {
		return ctx, err
	}

	// check that signer length and signature length are the same
	if len(sigs) != len(signers) {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signers), len(sigs))
	}

	for i, sig := range sigs {
		sigPrefixed := svd.accBech32Codec.BytesToAccAddressPrefixed(signers[i])
		acc, err := svd.ak.GetAccountPrefixed(ctx, sigPrefixed)
		if err != nil {
			return ctx, err
		}

		// retrieve pubkey
		pubKey := acc.GetPubKey()
		if !simulate && pubKey == nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		// Check account sequence number.
		if sig.Sequence != acc.GetSequence() {
			return ctx, errorsmod.Wrapf(
				sdkerrors.ErrWrongSequence,
				"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
			)
		}

		// retrieve signer data
		genesis := ctx.BlockHeight() == 0
		chainID := ctx.ChainID()
		var accNum uint64
		if !genesis {
			accNum = acc.GetAccountNumber()
		}

		// no need to verify signatures on recheck tx
		if !simulate && !ctx.IsReCheckTx() {
			anyPk, _ := codectypes.NewAnyWithValue(pubKey)

			signerData := txsigning.SignerData{
				Address:       acc.GetAddressPrefixed().String(),
				ChainID:       chainID,
				AccountNumber: accNum,
				Sequence:      acc.GetSequence(),
				PubKey: &anypb.Any{
					TypeUrl: anyPk.TypeUrl,
					Value:   anyPk.Value,
				},
			}

			adaptableTx, ok := tx.(authsigning.V2AdaptableTx)
			if !ok {
				return ctx, fmt.Errorf("expected tx to implement V2AdaptableTx, got %T", tx)
			}
			txData := adaptableTx.GetSigningTxData()
			err = authsigning.VerifySignature(ctx, pubKey, signerData, sig.Data, svd.signModeHandler, txData)
			if err != nil {
				var errMsg string
				if OnlyLegacyAminoSigners(sig.Data) {
					// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
					// and therefore communicate sequence number as a potential cause of error.
					errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)", accNum, acc.GetSequence(), chainID)
				} else {
					errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s): (%s)", accNum, chainID, err.Error())
				}
				return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, errMsg)

			}
		}
	}

	return next(ctx, tx, simulate)
}
