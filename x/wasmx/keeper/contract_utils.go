package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/x/wasmx/types"
)

func (k *Keeper) autoIncrementID(ctx sdk.Context, lastIDKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(lastIDKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(lastIDKey, bz)
	return id
}

// PeekAutoIncrementID reads the current value without incrementing it.
func (k *Keeper) PeekAutoIncrementID(ctx sdk.Context, lastIDKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(lastIDKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k *Keeper) importAutoIncrementID(ctx sdk.Context, lastIDKey []byte, val uint64) error {
	store := ctx.KVStore(k.storeKey)
	if store.Has(lastIDKey) {
		return sdkerr.Wrapf(types.ErrDuplicate, "autoincrement id: %s", string(lastIDKey))
	}
	bz := sdk.Uint64ToBigEndian(val)
	store.Set(lastIDKey, bz)
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

func (k *Keeper) ContractInstance(ctx sdk.Context, contractAddress sdk.AccAddress) (types.ContractInfo, types.CodeInfo, []byte, error) {
	store := ctx.KVStore(k.storeKey)

	contractBz := store.Get(types.GetContractAddressKey(contractAddress))
	if contractBz == nil {
		return types.ContractInfo{}, types.CodeInfo{}, nil, sdkerr.Wrap(types.ErrNotFound, "contract")
	}
	var contractInfo types.ContractInfo
	k.cdc.MustUnmarshal(contractBz, &contractInfo)

	codeInfoBz := store.Get(types.GetCodeKey(contractInfo.CodeId))
	if codeInfoBz == nil {
		return contractInfo, types.CodeInfo{}, nil, sdkerr.Wrap(types.ErrNotFound, "code info")
	}
	var codeInfo types.CodeInfo
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	return contractInfo, codeInfo, prefixStoreKey, nil
}

func (k *Keeper) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *types.ContractInfo {
	store := ctx.KVStore(k.storeKey)
	var contract types.ContractInfo
	contractBz := store.Get(types.GetContractAddressKey(contractAddress))
	if contractBz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(contractBz, &contract)
	return &contract
}

func (k *Keeper) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetContractAddressKey(contractAddress))
}

// storeContractInfo persists the ContractInfo. No secondary index updated here.
func (k *Keeper) storeContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress, contract *types.ContractInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetContractAddressKey(contractAddress), k.cdc.MustMarshal(contract))
}

// StoreContractInfo persists the ContractInfo. No secondary index updated here.
func (k *Keeper) StoreContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress, contract *types.ContractInfo) {
	k.storeContractInfo(ctx, contractAddress, contract)
}

func (k *Keeper) IterateContractInfo(ctx sdk.Context, cb func(sdk.AccAddress, types.ContractInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyContractPrefix)
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

func (k *Keeper) ImportContractState(ctx sdk.Context, contractAddress sdk.AccAddress, models []types.ContractStorage) error {
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
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

func (k *Keeper) MigrateContractStateByStorageType(ctx sdk.Context, contractAddress sdk.AccAddress, sourceStorage types.ContractStorageType, targetStorage types.ContractStorageType) error {
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
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

func (k *Keeper) GetCodeInfo(ctx sdk.Context, codeID uint64) *types.CodeInfo {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(codeID))
	if codeInfoBz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	return &codeInfo
}

func (k *Keeper) containsCodeInfo(ctx sdk.Context, codeID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetCodeKey(codeID))
}

func (k *Keeper) IterateCodeInfos(ctx sdk.Context, cb func(uint64, types.CodeInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyCodePrefix)
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

func (k *Keeper) GetByteCode(ctx sdk.Context, codeID uint64) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(codeID))
	if codeInfoBz == nil {
		return nil, nil
	}
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	if types.HasInterpreterDep(codeInfo.Deps) {
		return codeInfo.InterpretedBytecodeRuntime, nil
	}
	return k.wasmvm.GetCode(codeInfo.CodeHash, codeInfo.Deps)
}

// TransferCoins transfers coins from source to destination account when coin send was enabled for them and the recipient
// is not in the blocked address list.
func (k *Keeper) TransferCoins(parentCtx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error {
	em := sdk.NewEventManager()
	ctx := parentCtx.WithEventManager(em)
	// TODO do we do blocked addresses here?
	// if err := k.IsSendEnabledCoins(ctx, amount...); err != nil {
	// 	return err
	// }
	// if k.BlockedAddr(toAddr) {
	// 	return sdkerr.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", toAddr.String())
	// }

	err := k.SendCoins(ctx, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	for _, e := range em.Events() {
		if e.Type == sdk.EventTypeMessage { // skip messages as we talk to the keeper directly
			continue
		}
		parentCtx.EventManager().EmitEvent(e)
	}
	return nil
}

func (k *Keeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, coins sdk.Coins) error {
	aliasFrom, found := k.GetAlias(ctx, fromAddr)
	if found {
		fromAddr = aliasFrom
	}
	aliasTo, found := k.GetAlias(ctx, toAddr)
	if found {
		toAddr = aliasTo
	}
	bankAddress, err := k.GetAddressOrRole(ctx, types.ROLE_BANK)
	msg := banktypes.NewMsgSend(fromAddr, toAddr, coins)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SendCoins":%s}`, string(bankmsgbz)))
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return err
	}
	_, err = k.execute(ctx, bankAddress, bankAddress, execmsg, nil, nil)
	return err
}

func (k *Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	aliasAddr, found := k.GetAlias(ctx, addr)
	if found {
		addr = aliasAddr
	}
	bankAddress, err := k.GetAddressOrRole(ctx, types.ROLE_BANK)
	if err != nil {
		return sdk.Coin{}, err
	}
	msg := banktypes.NewQueryBalanceRequest(addr, denom)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return sdk.Coin{}, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetBalance":%s}`, string(bankmsgbz)))
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return sdk.Coin{}, err
	}
	resp, err := k.query(ctx, bankAddress, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil, false)
	if err != nil {
		return sdk.Coin{}, err
	}
	var response banktypes.QueryBalanceResponse
	err = k.cdc.UnmarshalJSON(resp.Data, &response)
	if err != nil {
		return sdk.Coin{}, err
	}
	return *response.Balance, nil
}

func (k *Keeper) AllBalances(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	aliasAddr, found := k.GetAlias(ctx, addr)
	if found {
		addr = aliasAddr
	}
	bankAddress, err := k.GetAddressOrRole(ctx, types.ROLE_BANK)
	if err != nil {
		return nil, err
	}
	msg := banktypes.NewQueryAllBalancesRequest(addr, nil, false)
	msgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return nil, err
	}
	resp, err := k.query(ctx, bankAddress, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil, false)
	if err != nil {
		return nil, err
	}
	var contractResp types.ContractResponse
	err = json.Unmarshal(resp.Data, &contractResp)
	if err != nil {
		return nil, err
	}
	var response banktypes.QueryAllBalancesResponse
	err = k.cdc.UnmarshalJSON(contractResp.Data, &response)
	if err != nil {
		return nil, err
	}
	return response.Balances, nil
}

// IsPinnedCode returns true when codeID is pinned in wasmvm cache
func (k *Keeper) IsPinnedCode(ctx sdk.Context, codeID uint64) bool {
	// store := ctx.KVStore(k.storeKey)
	// return store.Has(types.GetPinnedCodeIndexPrefix(codeID))
	return false
}

func (k *Keeper) CanCallSystemContract(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	// is EOA
	if !k.HasContractInfo(ctx, contractAddress) {
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
