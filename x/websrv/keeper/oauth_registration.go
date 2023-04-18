package keeper

import (
	"encoding/binary"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmxtypes "mythos/v1/x/wasmx/types"
	"mythos/v1/x/websrv/types"
)

func (k Keeper) GetAddressToClients(ctx sdk.Context, owner sdk.AccAddress) ([]uint64, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetKeyOauthAddressToClientsPrefix(owner))
	if len(bz) == 0 {
		return []uint64{}, nil
	}
	var clientIds []uint64
	err := json.Unmarshal(bz, &clientIds)
	if err != nil {
		return nil, err
	}
	return clientIds, nil
}

func (k Keeper) SetNewClientId(ctx sdk.Context, owner sdk.AccAddress, clientId uint64) error {
	store := ctx.KVStore(k.storeKey)
	clientIds, _ := k.GetAddressToClients(ctx, owner)
	if len(clientIds) > 100 {
		return sdkerrors.Wrapf(types.ErrOAuthTooManyClientsRegistered, "%d already registered", len(clientIds))
	}
	clientIds = append(clientIds, clientId)
	bz, err := json.Marshal(clientIds)
	if err != nil {
		return err
	}
	store.Set(types.GetKeyOauthAddressToClientsPrefix(owner), bz)
	return nil
}

func (k Keeper) DeleteClientIdFromOwner(ctx sdk.Context, owner sdk.AccAddress, clientId uint64) error {
	store := ctx.KVStore(k.storeKey)

	clientIds, _ := k.GetAddressToClients(ctx, owner)
	var newClientIds []uint64
	for _, id := range clientIds {
		if id != clientId {
			newClientIds = append(newClientIds, id)
		}
	}
	if len(newClientIds) == 0 {
		store.Delete(types.GetKeyOauthAddressToClientsPrefix(owner))
		return nil
	}
	bz, err := json.Marshal(newClientIds)
	if err != nil {
		return err
	}
	store.Set(types.GetKeyOauthAddressToClientsPrefix(owner), bz)
	return nil
}

func (k Keeper) GetClientIdToInfo(ctx sdk.Context, clientId uint64) (*types.OauthClientInfo, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetKeyOauthClientIdToInfoPrefix(clientId))
	if len(bz) == 0 {
		return nil, nil
	}
	var info types.OauthClientInfo
	err := k.cdc.Unmarshal(bz, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (k Keeper) SetClientIdToInfo(ctx sdk.Context, clientId uint64, info types.OauthClientInfo) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&info)
	if err != nil {
		return err
	}
	store.Set(types.GetKeyOauthClientIdToInfoPrefix(clientId), bz)
	return nil
}

func (k Keeper) DeleteClientIdToInfo(ctx sdk.Context, clientId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetKeyOauthClientIdToInfoPrefix(clientId))
}

func (k Keeper) autoIncrementClientId(ctx sdk.Context, lastIdKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(lastIdKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(lastIdKey, bz)
	return id
}

// PeekAutoIncrementClientId reads the current value without incrementing it.
func (k Keeper) PeekAutoIncrementClientId(ctx sdk.Context, lastIdKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(lastIdKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k Keeper) importAutoIncrementClientId(ctx sdk.Context, lastIdKey []byte, val uint64) error {
	store := ctx.KVStore(k.storeKey)
	if store.Has(lastIdKey) {
		return sdkerrors.Wrapf(wasmxtypes.ErrDuplicate, "autoincrement id: %s", string(lastIdKey))
	}
	bz := sdk.Uint64ToBigEndian(val)
	store.Set(lastIdKey, bz)
	return nil
}

// GetOauthClients
func (k Keeper) GetOauthClients(ctx sdk.Context) (infos []types.OauthClientInfo) {
	k.IterateOauthClients(ctx, func(contract types.OauthClientInfo) bool {
		infos = append(infos, contract)
		return false
	})
	return
}

// IterateOauthClients
// When the callback returns true, the loop is aborted early.
func (k Keeper) IterateOauthClients(ctx sdk.Context, cb func(types.OauthClientInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyOauthClientIdToInfoPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		var value types.OauthClientInfo
		k.cdc.MustUnmarshal(iter.Value(), &value)
		if cb(value) {
			return
		}
	}
}
