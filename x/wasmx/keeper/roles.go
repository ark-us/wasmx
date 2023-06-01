package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

// RegisterRole registers a contract
func (k Keeper) RegisterRole(
	ctx sdk.Context,
	role string,
	label string,
	contractAddress sdk.AccAddress,
) error {
	roleObj := &types.Role{
		Role:            role,
		Label:           label,
		ContractAddress: contractAddress.String(),
	}
	k.SetRoleByLabel(ctx, roleObj)
	k.SetRoleLabelByContract(ctx, contractAddress, label)
	// cache the new value
	k.SystemDepFromLabel(ctx, label)
	return nil
}

// DeregisterRole deregisters a contract
func (k Keeper) DeregisterRole(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
) error {
	// TODO remove from k.systemDepsByLabelCache
	return fmt.Errorf("DeregisterRole not implemented")
}

// GetRoleByLabel
func (k Keeper) GetRoleByLabel(ctx sdk.Context, label string) *types.Role {
	store := ctx.KVStore(k.storeKey)
	var role types.Role
	bz := store.Get(types.GetRoleLabelPrefix(label))
	if bz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(bz, &role)
	return &role
}

// SetRoleByLabel
func (k Keeper) SetRoleByLabel(ctx sdk.Context, role *types.Role) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleLabelPrefix(role.Label), k.cdc.MustMarshal(role))
}

// GetRoleLabelByContract
func (k Keeper) GetRoleLabelByContract(ctx sdk.Context, contractAddress sdk.AccAddress) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRoleContractPrefix(contractAddress))
	return string(bz)
}

// SetRoleLabelByContract
func (k Keeper) SetRoleLabelByContract(ctx sdk.Context, contractAddress sdk.AccAddress, label string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleContractPrefix(contractAddress), []byte(label))
}
