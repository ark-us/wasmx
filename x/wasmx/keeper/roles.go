package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

// role => contractAddress
// label => role
// contractAddress => label

// RegisterRole registers a contract
func (k *Keeper) RegisterRole(
	ctx sdk.Context,
	chainId string,
	role string,
	label string,
	contractAddress sdk.AccAddress,
) error {
	roleObj := &types.Role{
		Role:            role,
		Label:           label,
		ContractAddress: contractAddress.String(),
	}
	k.SetContractAddressByRole(ctx, chainId, role, contractAddress)
	k.SetRoleByLabel(ctx, chainId, roleObj)
	k.SetRoleLabelByContract(ctx, chainId, contractAddress, label)
	// TODO replace the previous role? if a role cannot hold 2 contracts?
	// e.g. consensus
	return nil
}

// DeregisterRole deregisters a contract
func (k *Keeper) DeregisterRole(
	ctx sdk.Context,
	chainId string,
	contractAddress sdk.AccAddress,
) error {
	return fmt.Errorf("DeregisterRole not implemented")
}

func (k *Keeper) GetAddressOrRole(ctx sdk.Context, chainId string, addressOrRole string) (sdk.AccAddress, error) {
	addr, found := k.GetContractAddressByRole(ctx, chainId, addressOrRole)
	if found {
		return addr, nil
	}
	role := k.GetRoleByLabel(ctx, chainId, addressOrRole)
	if role != nil {
		contractAddr, err := sdk.AccAddressFromBech32(addressOrRole)
		if err != nil {
			return nil, err
		}
		return contractAddr, nil
	}
	contractAddr, err := sdk.AccAddressFromBech32(addressOrRole)
	if err != nil {
		return nil, err
	}
	return contractAddr, nil
}

// GetContractAddressByRole
func (k *Keeper) GetContractAddressByRole(ctx sdk.Context, chainId string, role string) (sdk.AccAddress, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRolePrefix(chainId, role))
	if bz == nil {
		return nil, false
	}
	return sdk.AccAddress(bz), true
}

// SetContractAddressByRole
func (k *Keeper) SetContractAddressByRole(ctx sdk.Context, chainId string, role string, contractAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRolePrefix(chainId, role), contractAddress.Bytes())
}

// GetRoleByLabel
func (k *Keeper) GetRoleByLabel(ctx sdk.Context, chainId string, label string) *types.Role {
	store := ctx.KVStore(k.storeKey)
	var role types.Role
	bz := store.Get(types.GetRoleLabelPrefix(chainId, label))
	if bz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(bz, &role)
	return &role
}

// SetRoleByLabel
func (k *Keeper) SetRoleByLabel(ctx sdk.Context, chainId string, role *types.Role) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleLabelPrefix(chainId, role.Label), k.cdc.MustMarshal(role))
}

// GetRoleLabelByContract
func (k *Keeper) GetRoleLabelByContract(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRoleContractPrefix(chainId, contractAddress))
	return string(bz)
}

// SetRoleLabelByContract
func (k *Keeper) SetRoleLabelByContract(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress, label string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleContractPrefix(chainId, contractAddress), []byte(label))
}

// GetRoleByContractAddress
func (k *Keeper) GetRoleByContractAddress(ctx sdk.Context, chainId string, contractAddress sdk.AccAddress) string {
	label := k.GetRoleLabelByContract(ctx, chainId, contractAddress)
	if label == "" {
		return ""
	}
	role := k.GetRoleByLabel(ctx, chainId, label)
	if role == nil {
		return ""
	}
	return role.Role
}
