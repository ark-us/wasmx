package keeper

import (
	"strings"
	"wasmx/v1/x/websrv/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) GetMostSpecificRouteToContract(ctx sdk.Context, route string) sdk.AccAddress {
	// "" "arg1" "arg2" ..
	parts := strings.Split(route, "/")
	// a hardcoded max of path parts
	if len(parts) > 100 {
		return nil
	}

	for {
		contractAddress := k.GetRouteToContract(ctx, route)
		if contractAddress != nil {
			return contractAddress
		}
		parts = parts[:len(parts)-1]
		partlen := len(parts)
		if partlen < 1 {
			break
		}
		if partlen > 1 {
			route = strings.Join(parts, "/")
		} else {
			route = "/"
		}
	}
	return nil
}

func (k *Keeper) RegisterRoute(ctx sdk.Context, route string, contractAddress sdk.AccAddress) {
	// If it was previously registered, remove that registration
	k.DeleteContractToRoute(ctx, contractAddress)
	k.DeleteRouteToContract(ctx, route)
	k.SetContractToRoute(ctx, contractAddress, route)
	k.SetRouteToContract(ctx, route, contractAddress)
}

func (k *Keeper) DeregisterRoute(ctx sdk.Context, route string, contractAddress sdk.AccAddress) {
	// If it was previously registered, remove that registration
	k.DeleteContractToRoute(ctx, contractAddress)
	k.DeleteRouteToContract(ctx, route)
}

func (k *Keeper) GetRouteToContract(ctx sdk.Context, route string) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	addressbz := store.Get(types.GetRouteToContractKey(route))
	if addressbz == nil {
		return nil
	}
	return sdk.AccAddress(addressbz)
}

func (k *Keeper) SetRouteToContract(ctx sdk.Context, route string, contractAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRouteToContractKey(route), contractAddress.Bytes())
}

func (k *Keeper) DeleteRouteToContract(ctx sdk.Context, route string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetRouteToContractKey(route))
}

func (k *Keeper) GetContractToRoute(ctx sdk.Context, contractAddress sdk.AccAddress) string {
	store := ctx.KVStore(k.storeKey)
	route := store.Get(types.GetContractToRouteKey(contractAddress))
	return string(route)
}

func (k *Keeper) SetContractToRoute(ctx sdk.Context, contractAddress sdk.AccAddress, route string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetContractToRouteKey(contractAddress), []byte(route))
}

func (k *Keeper) DeleteContractToRoute(ctx sdk.Context, contractAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetContractToRouteKey(contractAddress))
}
