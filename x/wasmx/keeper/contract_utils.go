package keeper

import (
	"encoding/binary"
	"strings"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

func (k *Keeper) autoIncrementID(ctx sdk.Context, chainId string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastCodeIDKey(chainId)
	bz := store.Get(key)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(key, bz)
	return id
}

// PeekAutoIncrementID reads the current value without incrementing it.
func (k *Keeper) PeekAutoIncrementID(ctx sdk.Context, chainId string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastCodeIDKey(chainId)
	bz := store.Get(key)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k *Keeper) importAutoIncrementID(ctx sdk.Context, chainId string, val uint64) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastCodeIDKey(chainId)
	if store.Has(key) {
		return sdkerr.Wrapf(types.ErrDuplicate, "autoincrement id: %s", string(key))
	}
	bz := sdk.Uint64ToBigEndian(val)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) ContractStore(ctx sdk.Context, storageType types.ContractStorageType, prefixStoreKey []byte) prefix.Store {
	// default must be core consensus
	storageKey := k.storeKey // types.ContractStorageType_CoreConsensus
	if storageType == types.ContractStorageType_MetaConsensus {
		storageKey = k.metaConsKey
	} else if storageType == types.ContractStorageType_SingleConsensus {
		storageKey = k.singleConsKey
	} else if storageType == types.ContractStorageType_Transient {
		storageKey = k.tKey
	} else if storageType == types.ContractStorageType_Memory {
		storageKey = k.memKey
	}
	return prefix.NewStore(ctx.KVStore(storageKey), prefixStoreKey)
}

func (k *Keeper) GetCode(checksum types.Checksum, deps []string) (types.WasmCode, error) {
	return k.wasmvm.GetCode(checksum, deps)
}

func (k *Keeper) ContractInstance(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress) (types.ContractInfo, types.CodeInfo, []byte, error) {
	store := ctx.KVStore(k.storeKey)

	contractBz := store.Get(types.GetContractAddressKey(chainId, contractAddress))
	if contractBz == nil {
		return types.ContractInfo{}, types.CodeInfo{}, nil, sdkerr.Wrap(types.ErrNotFound, "contract")
	}
	var contractInfo types.ContractInfo
	k.cdc.MustUnmarshal(contractBz, &contractInfo)

	codeInfoBz := store.Get(types.GetCodeKey(chainId, contractInfo.CodeId))
	if codeInfoBz == nil {
		return contractInfo, types.CodeInfo{}, nil, sdkerr.Wrap(types.ErrNotFound, "code info")
	}
	var codeInfo types.CodeInfo
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	prefixStoreKey := types.GetContractStorePrefix(chainId, contractAddress)
	return contractInfo, codeInfo, prefixStoreKey, nil
}

func (k *Keeper) GetContractInfo(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress) *types.ContractInfo {
	store := ctx.KVStore(k.storeKey)
	var contract types.ContractInfo
	contractBz := store.Get(types.GetContractAddressKey(chainId, contractAddress))
	if contractBz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(contractBz, &contract)
	return &contract
}

func (k *Keeper) HasContractInfo(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetContractAddressKey(chainId, contractAddress))
}

// storeContractInfo persists the ContractInfo. No secondary index updated here.
func (k *Keeper) storeContractInfo(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress, contract *types.ContractInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetContractAddressKey(chainId, contractAddress), k.cdc.MustMarshal(contract))
}

// StoreContractInfo persists the ContractInfo. No secondary index updated here.
func (k *Keeper) StoreContractInfo(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress, contract *types.ContractInfo) {
	k.storeContractInfo(ctx, chainId, contractAddress, contract)
}

func (k *Keeper) IterateContractInfo(ctx sdk.Context, chainId string, cb func(sdk.AccAddress, types.ContractInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetContractAddressRootKey(chainId))
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var contract types.ContractInfo
		k.cdc.MustUnmarshal(iter.Value(), &contract)
		// cb returns true to stop early
		if cb(iter.Key(), contract) {
			break
		}
	}
}

// IterateContractState iterates through all elements of the key value store for the given contract address and passes
// them to the provided callback function. The callback method can return true to abort early.
func (k *Keeper) IterateContractState(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress, cb func(key, value []byte) bool) {
	prefixStoreKey := types.GetContractStorePrefix(chainId, contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if cb(iter.Key(), iter.Value()) {
			break
		}
	}
}

func (k *Keeper) ImportContractState(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress, models []types.ContractStorage) error {
	prefixStoreKey := types.GetContractStorePrefix(chainId, contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	for _, model := range models {
		if model.Value == nil {
			model.Value = []byte{}
		}
		if prefixStore.Has(model.Key) {
			return sdkerr.Wrapf(types.ErrDuplicate, "duplicate key: %x", model.Key)
		}
		prefixStore.Set(model.Key, model.Value)
	}
	return nil
}

func (k *Keeper) MigrateContractStateByStorageType(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress, sourceStorage types.ContractStorageType, targetStorage types.ContractStorageType) error {
	prefixStoreKey := types.GetContractStorePrefix(chainId, contractAddress)
	prefixStoreSource := k.ContractStore(ctx, sourceStorage, prefixStoreKey)
	prefixStoreTarget := k.ContractStore(ctx, targetStorage, prefixStoreKey)
	iter := prefixStoreSource.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		prefixStoreTarget.Set(iter.Key(), iter.Value())
		prefixStoreSource.Delete(iter.Key())
	}
	return nil
}

func (k *Keeper) GetCodeInfo(ctx sdk.Context, chainId string, codeID uint64) *types.CodeInfo {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(chainId, codeID))
	if codeInfoBz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	return &codeInfo
}

func (k *Keeper) containsCodeInfo(ctx sdk.Context, chainId string, codeID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetCodeKey(chainId, codeID))
}

func (k *Keeper) IterateCodeInfos(ctx sdk.Context, chainId string, cb func(uint64, types.CodeInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetCodeRootKey(chainId))
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var c types.CodeInfo
		k.cdc.MustUnmarshal(iter.Value(), &c)
		// cb returns true to stop early
		if cb(binary.BigEndian.Uint64(iter.Key()), c) {
			return
		}
	}
}

func (k *Keeper) GetByteCode(ctx sdk.Context, chainId string, codeID uint64) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(chainId, codeID))
	if codeInfoBz == nil {
		return nil, nil
	}
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	if types.HasInterpreterDep(codeInfo.Deps) {
		return codeInfo.InterpretedBytecodeRuntime, nil
	}
	return k.wasmvm.GetCode(codeInfo.CodeHash, codeInfo.Deps)
}

// IsPinnedCode returns true when codeID is pinned in wasmvm cache
func (k *Keeper) IsPinnedCode(ctx sdk.Context, chainId string, codeID uint64) bool {
	// store := ctx.KVStore(k.storeKey)
	// return store.Has(types.GetPinnedCodeIndexPrefix(codeID))
	return false
}

func (k *Keeper) CanCallSystemContract(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress) bool {
	// is EOA
	if !k.HasContractInfo(ctx, chainId, contractAddress) {
		return true
	}
	// is system contract
	if types.IsSystemAddress(contractAddress) {
		return true
	}
	return false
}

func RequireNotSystemContract(contractAddress sdk.AccAddress, deps []string) error {
	for _, dep := range deps {
		if strings.Contains(dep, types.SYS_VM_EXPORT) && !types.IsSystemAddress(contractAddress) {
			return sdkerr.Wrap(types.ErrUnauthorizedAddress, "invalid address for system contracts")
		}
	}
	return nil
}
