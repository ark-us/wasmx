package keeper

import (
	_ "embed"
	"mythos/v1/x/wasmx/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"mythos/v1/x/wasmx/ewasm"
	precompiles "mythos/v1/x/wasmx/ewasm/contracts"
)

func (k Keeper) BootstrapSystemContracts(
	ctx sdk.Context,
	bootstrapAccountAddr sdk.AccAddress,
	contracts []types.SystemContract,
	compiledFolderPath string,
) error {
	for _, contract := range contracts {
		err := k.ActivateEmbeddedSystemContract(ctx, bootstrapAccountAddr, contract, compiledFolderPath)
		if err != nil {
			return sdkerrors.Wrap(err, "bootstrap")
		}
	}
	return nil
}

// ActivateEmbeddedSystemContract
func (k Keeper) ActivateEmbeddedSystemContract(
	ctx sdk.Context,
	bootstrapAccountAddr sdk.AccAddress,
	contract types.SystemContract,
	compiledFolderPath string,
) error {
	wasmbin := precompiles.GetPrecompileByLabel(contract.Label)
	return k.ActivateSystemContract(ctx, bootstrapAccountAddr, contract, wasmbin, compiledFolderPath)
}

// ActivateSystemContract
func (k Keeper) ActivateSystemContract(
	ctx sdk.Context,
	bootstrapAccountAddr sdk.AccAddress,
	contract types.SystemContract,
	wasmbin []byte,
	compiledFolderPath string,
) error {
	k.SetSystemContract(ctx, contract)
	var codeID uint64
	var err error

	if contract.Native {
		codeID = k.autoIncrementID(ctx, types.KeyLastCodeID)
		codeInfo := types.NewCodeInfo([]byte(contract.Address), bootstrapAccountAddr, nil, contract.Metadata)
		k.storeCodeInfo(ctx, codeID, codeInfo)
	} else {
		codeID, _, err = k.Create(ctx, bootstrapAccountAddr, wasmbin, contract.Metadata)
		if err != nil {
			return sdkerrors.Wrap(err, "store system contract: "+contract.Label)
		}
	}

	if contract.Pinned {
		if err := k.PinCode(ctx, codeID, compiledFolderPath); err != nil {
			return sdkerrors.Wrap(err, "pin system contract: "+contract.Label)
		}
	}

	contractAddress := ewasm.AccAddressFromHex(contract.Address)
	if contract.Native {
		contractInfo := types.NewContractInfo(codeID, bootstrapAccountAddr, contract.Label)
		k.storeContractInfo(ctx, contractAddress, &contractInfo)
	} else {
		_, err = k.instantiateWithAddress(
			ctx,
			codeID,
			bootstrapAccountAddr,
			contractAddress,
			contract.InitMessage,
			contract.Label,
			nil,
		)
		if err != nil {
			return sdkerrors.Wrap(err, "instantiate system contract: "+contract.Label)
		}
	}
	k.Logger(ctx).Info("activated system contract", contract.Label, "address", contract.Address, "code_id", codeID)
	return nil
}

// SetSystemContract
func (k Keeper) SetSystemContract(ctx sdk.Context, contract types.SystemContract) {
	addr := ewasm.AccAddressFromHex(contract.Address)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSystemContract)
	bz := k.cdc.MustMarshal(&contract)
	prefixStore.Set(addr.Bytes(), bz)
}

// GetSystemContracts
func (k Keeper) GetSystemContracts(ctx sdk.Context) (contracts []types.SystemContract) {
	k.IterateSystemContracts(ctx, func(contract types.SystemContract) bool {
		contracts = append(contracts, contract)
		return false
	})
	return
}

// IterateSystemContracts
// When the callback returns true, the loop is aborted early.
func (k Keeper) IterateSystemContracts(ctx sdk.Context, cb func(types.SystemContract) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSystemContract)
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
