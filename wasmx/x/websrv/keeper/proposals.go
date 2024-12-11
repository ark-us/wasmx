package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterRouteHandler registers a web route and its smart contract controller
func (k *Keeper) RegisterRouteHandler(
	ctx sdk.Context,
	path string,
	contractAddressBech32 string,
) error {
	contractAddress, err := k.AddressCodec().StringToBytes(contractAddressBech32)
	if err != nil {
		return err
	}
	k.RegisterRoute(ctx, path, contractAddress)
	return nil
}

// DeregisterRouteHandler deregisters a web route
func (k *Keeper) DeregisterRouteHandler(
	ctx sdk.Context,
	path string,
	contractAddressBech32 string,
) error {
	contractAddress, err := k.AddressCodec().StringToBytes(contractAddressBech32)
	if err != nil {
		return err
	}
	k.DeregisterRoute(ctx, path, contractAddress)
	return nil
}
