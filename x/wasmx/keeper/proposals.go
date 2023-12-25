package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterRoleHandler registers a web route and its smart contract controller
func (k *Keeper) RegisterRoleHandler(
	ctx sdk.Context,
	role string,
	label string,
	contractAddressBech32 string,
) error {
	contractAddress, err := sdk.AccAddressFromBech32(contractAddressBech32)
	if err != nil {
		return err
	}
	return k.RegisterRole(ctx, role, label, contractAddress)
}

// DeregisterRoleHandler deregisters a web route
func (k *Keeper) DeregisterRoleHandler(
	ctx sdk.Context,
	contractAddressBech32 string,
) error {
	contractAddress, err := sdk.AccAddressFromBech32(contractAddressBech32)
	if err != nil {
		return err
	}
	return k.DeregisterRole(ctx, contractAddress)
}
