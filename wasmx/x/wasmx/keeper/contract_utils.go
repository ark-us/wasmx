package keeper

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

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

func (k *Keeper) storeNewCodeInfo(ctx sdk.Context, codeInfo types.CodeInfo) (uint64, error) {
	contractAddr := k.GetCodeRegistryAddress(ctx)
	databz, err := json.Marshal(&codeInfo)
	if err != nil {
		return 0, err
	}
	msg := fmt.Sprintf(`{"NewCodeInfo":{"code_info":%s}}`, string(databz))
	data, err := k.internalExecute(ctx, contractAddr, msg)
	if err != nil {
		return 0, err
	}
	var res types.QueryLastCodeIdResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return 0, err
	}
	return res.CodeId, nil
}

func (k *Keeper) storeCodeInfo(ctx sdk.Context, codeId uint64, codeInfo types.CodeInfo) error {
	contractAddr := k.GetCodeRegistryAddress(ctx)
	databz, err := json.Marshal(&codeInfo)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf(`{"SetCodeInfo":{"code_id":%d,"code_info":%s}}`, codeId, string(databz))
	_, err = k.internalExecute(ctx, contractAddr, msg)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keeper) storeContractInfo(ctx sdk.Context, addr mcodec.AccAddressPrefixed, value types.ContractInfo) error {
	contractAddr := k.GetCodeRegistryAddress(ctx)
	databz, err := json.Marshal(&value)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf(`{"SetContractInfo":{"address":"%s","contract_info":%s}}`, base64.StdEncoding.EncodeToString(addr.Bytes()), string(databz))
	_, err = k.internalExecute(ctx, contractAddr, msg)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keeper) GetLastCodeId(ctx sdk.Context) (codeId uint64, err error) {
	contractAddr := k.GetCodeRegistryAddress(ctx)
	msg := `{"GetLastCodeId":{}}`
	// Note! role contract should not have any other depedencies aside from the host import interface
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		return 0, err
	}
	var res types.QueryLastCodeIdResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return 0, err
	}
	return res.CodeId, nil
}

func (k *Keeper) GetNextCodeId(ctx sdk.Context) (codeId uint64, err error) {
	codeId, err = k.GetLastCodeId(ctx)
	if err != nil {
		return 0, err
	}
	return codeId + 1, nil
}

func (k *Keeper) GetCodeInfo(ctx sdk.Context, codeId uint64) (*types.CodeInfo, error) {
	cache := k.GetSystemBootstrap(ctx)
	if cache == nil {
		return nil, fmt.Errorf("cannot find system bootstrap data")
	}
	if cache.CodeRegistryId == codeId {
		return cache.CodeRegistryCodeInfo, nil
	}

	registryAddr := cache.CodeRegistryAddress
	msg := fmt.Sprintf(`{"GetCodeInfo":{"code_id":%d}}`, codeId)
	// Note! registry contract should not have any other depedencies aside from the host import interface
	data, err := k.internalQuery(ctx, registryAddr, msg)
	if err != nil {
		return nil, err
	}
	var res types.GetCodeInfoResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	if res.CodeInfo == nil {
		return nil, nil
	}
	return res.CodeInfo, nil
}

func (k *Keeper) GetContractInfo(ctx sdk.Context, address mcodec.AccAddressPrefixed) (*types.ContractInfo, error) {
	cache := k.GetSystemBootstrap(ctx)
	if cache == nil {
		return nil, fmt.Errorf("cannot find system bootstrap data")
	}
	registryAddr := cache.CodeRegistryAddress
	if registryAddr.String() == address.String() {
		return cache.CodeRegistryContractInfo, nil
	}

	msg := fmt.Sprintf(`{"GetContractInfo":{"address":"%s"}}`, base64.StdEncoding.EncodeToString(address.Bytes()))
	// Note! registry contract should not have any other depedencies aside from the host import interface
	data, err := k.internalQuery(ctx, registryAddr, msg)
	if err != nil {
		return nil, err
	}
	var res types.GetContractInfoResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	if res.ContractInfo == nil {
		return nil, nil
	}
	return res.ContractInfo, nil
}

func (k *Keeper) GetCode(checksum types.Checksum, deps []string) (types.WasmCode, error) {
	return k.wasmvm.GetCode(checksum, deps)
}

func (k *Keeper) ContractInstance(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) (*types.ContractInfo, *types.CodeInfo, []byte, error) {
	prefixStoreKey := types.GetContractStorePrefix(contractAddress.Bytes())

	cache := k.GetSystemBootstrap(ctx)
	if cache == nil {
		return nil, nil, nil, fmt.Errorf("cannot find system bootstrap data")
	}
	registryAddr := cache.CodeRegistryAddress
	if registryAddr.String() == contractAddress.String() {
		return cache.CodeRegistryContractInfo, cache.CodeRegistryCodeInfo, prefixStoreKey, nil
	}

	msg := fmt.Sprintf(`{"GetContractInstance":{"address":"%s"}}`, base64.StdEncoding.EncodeToString(contractAddress.Bytes()))
	// Note! role contract should not have any other depedencies aside from the host import interface
	data, err := k.internalQuery(ctx, registryAddr, msg)
	if err != nil {
		return nil, nil, nil, err
	}
	var res types.ContractInstance
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, nil, err
	}
	return res.ContractInfo, res.CodeInfo, prefixStoreKey, nil
}

func (k *Keeper) HasContractInfo(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) bool {
	data, err := k.GetContractInfo(ctx, contractAddress)
	if err != nil {
		return false
	}
	if data == nil {
		return false
	}
	return true
}

func (k *Keeper) IterateContractInfos(ctx sdk.Context, cb func(sdk.AccAddress, types.ContractInfo) bool) {
	registryAddr := k.GetCodeRegistryAddress(ctx)
	msg := `{"GetContractInfoPrefix":{}}`
	infoPrefix, err := k.internalQuery(ctx, registryAddr, msg)
	if err != nil {
		panic(err)
	}

	prefixStoreKey := types.GetContractStorePrefix(registryAddr.Bytes())
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	iter := prefixStore.Iterator(infoPrefix, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, found := bytes.CutPrefix(iter.Key(), infoPrefix)
		if !found {
			// when prefix is no longer found, stop early
			return
		}
		if len(key) == 0 {
			continue
		}
		var c types.ContractInfo
		err := json.Unmarshal(iter.Value(), &c)
		if err != nil {
			panic(sdkerr.Wrapf(err, "code iteration: cannot unmarshal CodeInfo: %s", hex.EncodeToString(iter.Value())))
		}
		// cb returns true to stop early
		if cb(key, c) {
			return
		}
	}
}

// IterateContractState iterates through all elements of the key value store for the given contract address and passes
// them to the provided callback function. The callback method can return true to abort early.
func (k *Keeper) IterateContractState(ctx sdk.Context, contractAddress sdk.AccAddress, cb func(key, value []byte) bool) {
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if cb(iter.Key(), iter.Value()) {
			break
		}
	}
}

func (k *Keeper) ImportContractState(ctx sdk.Context, contractAddress sdk.AccAddress, storageType types.ContractStorageType, models []types.ContractStoragePB) error {
	if len(models) == 0 {
		return nil
	}
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := k.ContractStore(ctx, storageType, prefixStoreKey)
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

func (k *Keeper) containsCodeInfo(ctx sdk.Context, codeID uint64) bool {
	codeInfo, err := k.GetCodeInfo(ctx, codeID)
	if err != nil {
		return false
	}
	if codeInfo == nil {
		return false
	}
	return true
}

func (k *Keeper) IterateCodeInfos(ctx sdk.Context, cb func(uint64, types.CodeInfo) bool) {
	registryAddr := k.GetCodeRegistryAddress(ctx)
	// msg := `{"GetCodeInfoPrefix":{}}`
	// codeInfoPrefix, err := k.internalQuery(ctx, registryAddr, msg)
	// if err != nil {
	// 	panic(err)
	// }
	// we need to hardcode the prefix here, otherwise we cannot apply state sync snapshots
	codeInfoPrefix := types.KeyCodePrefix

	prefixStoreKey := types.GetContractStorePrefix(registryAddr.Bytes())
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	iter := prefixStore.Iterator(codeInfoPrefix, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// either this or use storetypes.PrefixEndBytes(codeInfoPrefix) for end prefix
		key, found := bytes.CutPrefix(iter.Key(), codeInfoPrefix)
		if !found {
			// when prefix is no longer found, stop early
			return
		}
		if len(key) == 0 {
			continue
		}
		var c types.CodeInfo
		err := json.Unmarshal(iter.Value(), &c)
		if err != nil {
			panic(sdkerr.Wrapf(err, "code iteration: cannot unmarshal CodeInfo: %s", hex.EncodeToString(iter.Value())))
		}
		codeId := binary.BigEndian.Uint64(key)
		// cb returns true to stop early
		if cb(codeId, c) {
			return
		}
	}
}

func (k *Keeper) GetByteCode(ctx sdk.Context, codeID uint64) ([]byte, error) {
	codeInfo, err := k.GetCodeInfo(ctx, codeID)
	if err != nil {
		return nil, err
	}
	if codeInfo == nil {
		return nil, nil
	}
	// it can be a utf8 interpreter, handled in wasmvm.GetCode
	if len(codeInfo.InterpretedBytecodeRuntime) > 0 && types.HasInterpreterDep(codeInfo.Deps) {
		return codeInfo.InterpretedBytecodeRuntime, nil
	}
	return k.wasmvm.GetCode(types.Checksum(codeInfo.CodeHash), codeInfo.Deps)
}

// IsPinnedCode returns true when codeID is pinned in wasmvm cache
func (k *Keeper) IsPinnedCode(ctx sdk.Context, codeID uint64) bool {
	// store := ctx.KVStore(k.storeKey)
	// return store.Has(types.GetPinnedCodeIndexPrefix(codeID))
	return false
}

func (k *Keeper) CanCallSystemContract(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) bool {
	// is EOA
	if !k.HasContractInfo(ctx, contractAddress) {
		return true
	}
	// is system contract
	if types.IsSystemAddress(contractAddress.Bytes()) {
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
