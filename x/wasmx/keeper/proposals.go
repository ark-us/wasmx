package keeper

import (
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterRoleHandler registers a core role and its smart contract controller
func (k *Keeper) RegisterRoleHandler(
	ctx sdk.Context,
	role string,
	label string,
	contractAddressBech32 string,
) error {
	contractAddress, err := sdk.AccAddressFromBech32(contractAddressBech32)
	if err != nil {
		return sdkerr.Wrap(err, "contract address")
	}
	// execute setup(prevcontract)
	prevContractBech32 := ""
	prevContract, found := k.GetContractAddressByRole(ctx, role)
	if found {
		prevContractBech32 = prevContract.String()
	}

	// inherit storage type from previous contract
	prevContractInfo := k.GetContractInfo(ctx, prevContract)
	if prevContractInfo == nil {
		return fmt.Errorf("previous contract info not found for role %s", prevContractBech32)
	}
	contractInfo := k.GetContractInfo(ctx, contractAddress)
	if contractInfo == nil {
		return fmt.Errorf("proposed contract info not found for role %s", contractAddressBech32)
	}
	if contractInfo.StorageType != prevContractInfo.StorageType {
		k.Logger(ctx).Info("migrating contract storage...", "address", contractAddress.String(), "source storage type", contractInfo.StorageType, "target storage type", prevContractInfo.StorageType)
		k.MigrateContractStateByStorageType(ctx, contractAddress, contractInfo.StorageType, prevContractInfo.StorageType)
		contractInfo.StorageType = prevContractInfo.StorageType
		k.Logger(ctx).Info("contract storage migrated", "address", contractAddress.String())
	}
	k.StoreContractInfo(ctx, contractAddress, contractInfo)
	k.RegisterRole(ctx, role, label, contractAddress)

	// we do not remove role from previous contract
	// e.g. for consensus, the previous contract may be a backup consensus
	// or if the new contract fails to set up, we revert to the previous one
	return nil
}

// DeregisterRoleHandler deregisters a core role contract
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
