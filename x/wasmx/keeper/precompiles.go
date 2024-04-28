package keeper

import (
	_ "embed"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcfg "mythos/v1/config"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/precompiles"
)

func (k *Keeper) BootstrapSystemContracts(
	ctx sdk.Context,
	bootstrapAccountAddr sdk.AccAddress,
	contracts []types.SystemContract,
	compiledFolderPath string,
) error {
	for _, contract := range contracts {
		err := k.ActivateEmbeddedSystemContract(ctx, bootstrapAccountAddr, contract, compiledFolderPath)
		if err != nil {
			return sdkerr.Wrap(err, "bootstrap")
		}
	}
	return nil
}

// ActivateEmbeddedSystemContract
func (k *Keeper) ActivateEmbeddedSystemContract(
	ctx sdk.Context,
	bootstrapAccountAddr sdk.AccAddress,
	contract types.SystemContract,
	compiledFolderPath string,
) error {
	wasmbin := precompiles.GetPrecompileByLabel(contract.Label)
	return k.ActivateSystemContract(ctx, bootstrapAccountAddr, contract, wasmbin, compiledFolderPath)
}

// ActivateSystemContract
func (k *Keeper) ActivateSystemContract(
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
		codeID = k.autoIncrementID(ctx)
		codeInfo := types.NewCodeInfo([]byte(contract.Address), bootstrapAccountAddr, contract.Deps, contract.Metadata)
		k.storeCodeInfo(ctx, codeID, codeInfo)
	} else {
		codeID, _, err = k.Create(ctx, bootstrapAccountAddr, wasmbin, contract.Deps, contract.Metadata)
		if err != nil {
			return sdkerr.Wrap(err, "store system contract: "+contract.Label)
		}
	}

	if contract.Pinned {
		if err := k.PinCode(ctx, codeID, compiledFolderPath); err != nil {
			return sdkerr.Wrap(err, "pin system contract: "+contract.Label)
		}
	}
	// no address, we just need to create a code id
	if contract.Address == "" {
		k.Logger(ctx).Info("created system contract", "label", contract.Label, "code_id", codeID)
		return nil
	}

	contractAddress := types.AccAddressFromHex(contract.Address)
	// register role first, to be able to initialize the account keeper
	if contract.Role != "" {
		k.RegisterRole(ctx, contract.Role, contract.Label, contractAddress)
	}
	if contract.Native {
		contractInfo := types.NewContractInfo(codeID, bootstrapAccountAddr, nil, contract.InitMessage, contract.Label)
		k.storeContractInfo(ctx, contractAddress, &contractInfo)
	} else {
		_, err = k.instantiateWithAddress(
			ctx,
			codeID,
			bootstrapAccountAddr,
			contractAddress,
			contract.StorageType,
			contract.InitMessage,
			nil,
			contract.Label,
		)
		if err != nil {
			return sdkerr.Wrap(err, "instantiate system contract: "+contract.Label)
		}
	}
	contractAddressStr, err := k.AddressCodec().BytesToString(contractAddress)
	if err != nil {
		return sdkerr.Wrapf(err, "alias: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}

	k.Logger(ctx).Info("activated system contract", "label", contract.Label, "address", contractAddressStr, "hex_address", contract.Address, "code_id", codeID)
	return nil
}

// SetSystemContract
func (k *Keeper) SetSystemContract(ctx sdk.Context, contract types.SystemContract) {
	// for contracts where we just need the code id and are not deployed
	// TODO better, because these contracts will not be exported
	if contract.Address == "" {
		return
	}
	addr := types.AccAddressFromHex(contract.Address)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSystemContract)
	bz := k.cdc.MustMarshal(&contract)
	prefixStore.Set(addr.Bytes(), bz)
}

// GetSystemContracts
func (k *Keeper) GetSystemContracts(ctx sdk.Context) (contracts []types.SystemContract) {
	k.IterateSystemContracts(ctx, func(contract types.SystemContract) bool {
		contracts = append(contracts, contract)
		return false
	})
	return
}

// IterateSystemContracts
// When the callback returns true, the loop is aborted early.
func (k *Keeper) IterateSystemContracts(ctx sdk.Context, cb func(types.SystemContract) bool) {
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
