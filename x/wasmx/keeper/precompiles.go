package keeper

import (
	_ "embed"
	"wasmx/x/wasmx/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"wasmx/x/wasmx/ewasm"
)

func (k Keeper) BootstrapEwasmPrecompiles(ctx sdk.Context, bootstrapAccountAddr sdk.AccAddress, precompiles []types.SystemContract) error {
	for _, precompile := range precompiles {
		err := k.ActivateEmbeddedPrecompile(ctx, bootstrapAccountAddr, precompile)
		if err != nil {
			return sdkerrors.Wrap(err, "bootstrap")
		}
	}
	return nil
}

// ActivateEmbeddedPrecompile
func (k Keeper) ActivateEmbeddedPrecompile(ctx sdk.Context, bootstrapAccountAddr sdk.AccAddress, precompile types.SystemContract) error {
	wasmbin := ewasm.GetPrecompileByLabel(precompile.Label)
	return k.ActivatePrecompile(ctx, bootstrapAccountAddr, precompile, wasmbin)
}

// ActivatePrecompile
func (k Keeper) ActivatePrecompile(ctx sdk.Context, bootstrapAccountAddr sdk.AccAddress, precompile types.SystemContract, wasmbin []byte) error {
	k.SetPrecompile(ctx, precompile)

	codeID, _, err := k.Create(ctx, bootstrapAccountAddr, wasmbin)
	if err != nil {
		return sdkerrors.Wrap(err, "store precompile: "+precompile.Label)
	}
	contractAddress := ewasm.AccAddressFromHex(precompile.Address)

	_, err = k.instantiateWithAddress(
		ctx,
		codeID,
		bootstrapAccountAddr,
		contractAddress,
		precompile.InitMessage,
		precompile.Label,
		nil,
	)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate precompile: "+precompile.Label)
	}
	// if err := k.PinCode(ctx, codeID); err != nil {
	// 	return sdkerrors.Wrap(err, "pin precompile: "+precompile.Label)
	// }
	k.Logger(ctx).Info("precompile", precompile.Label, "address", precompile.Address, "code_id", codeID)
	return nil
}

// SetPrecompile
func (k Keeper) SetPrecompile(ctx sdk.Context, precompile types.SystemContract) {
	addr := ewasm.AccAddressFromHex(precompile.Address)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPrecompile)
	bz := k.cdc.MustMarshal(&precompile)
	prefixStore.Set(addr.Bytes(), bz)
}

// GetPrecompiles
func (k Keeper) GetPrecompiles(ctx sdk.Context) (precompiles []types.SystemContract) {
	k.IteratePrecompiles(ctx, func(precompile types.SystemContract) bool {
		precompiles = append(precompiles, precompile)
		return false
	})
	return
}

// IteratePrecompiles
// When the callback returns true, the loop is aborted early.
func (k Keeper) IteratePrecompiles(ctx sdk.Context, cb func(types.SystemContract) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPrecompile)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		var value types.SystemContract
		k.cdc.MustUnmarshal(iter.Value(), &value)
		if cb(value) {
			return
		}
	}
}
