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
	role string,
	label string,
	contractAddress sdk.AccAddress,
) error {
	roleObj := &types.Role{
		Role:            role,
		Label:           label,
		ContractAddress: contractAddress.String(),
	}
	k.SetContractAddressByRole(ctx, role, contractAddress)
	k.SetRoleByLabel(ctx, roleObj)
	k.SetRoleLabelByContract(ctx, contractAddress, label)
	// TODO replace the previous role? if a role cannot hold 2 contracts?
	// e.g. consensus
	return nil
}

// DeregisterRole deregisters a contract
func (k *Keeper) DeregisterRole(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
) error {
	return fmt.Errorf("DeregisterRole not implemented")
}

func (k *Keeper) GetAddressOrRole(ctx sdk.Context, addressOrRole string) (sdk.AccAddress, error) {
	addr, found := k.GetContractAddressByRole(ctx, addressOrRole)
	if found {
		return addr, nil
	}
	role := k.GetRoleByLabel(ctx, addressOrRole)
	if role != nil {
		contractAddr, err := k.AddressCodec().StringToBytes(addressOrRole)
		if err != nil {
			return nil, err
		}
		return contractAddr, nil
	}
	contractAddr, err := k.AddressCodec().StringToBytes(addressOrRole)
	if err != nil {
		return nil, err
	}
	return contractAddr, nil
}

// GetContractAddressByRole
func (k *Keeper) GetContractAddressByRole(ctx sdk.Context, role string) (sdk.AccAddress, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRolePrefix(role))
	if bz == nil {
		return nil, false
	}
	return sdk.AccAddress(bz), true
}

// SetContractAddressByRole
func (k *Keeper) SetContractAddressByRole(ctx sdk.Context, role string, contractAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRolePrefix(role), contractAddress.Bytes())
}

// GetRoleByLabel
func (k *Keeper) GetRoleByLabel(ctx sdk.Context, label string) *types.Role {
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
func (k *Keeper) SetRoleByLabel(ctx sdk.Context, role *types.Role) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleLabelPrefix(role.Label), k.cdc.MustMarshal(role))
}

// GetRoleLabelByContract
func (k *Keeper) GetRoleLabelByContract(ctx sdk.Context, contractAddress sdk.AccAddress) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRoleContractPrefix(contractAddress))
	return string(bz)
}

// SetRoleLabelByContract
func (k *Keeper) SetRoleLabelByContract(ctx sdk.Context, contractAddress sdk.AccAddress, label string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleContractPrefix(contractAddress), []byte(label))
}

// GetRoleByContractAddress
func (k *Keeper) GetRoleByContractAddress(ctx sdk.Context, contractAddress sdk.AccAddress) string {
	label := k.GetRoleLabelByContract(ctx, contractAddress)
	if label == "" {
		return ""
	}
	role := k.GetRoleByLabel(ctx, label)
	if role == nil {
		return ""
	}
	return role.Role
}
