package types

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmxkeeper "mythos/v1/x/wasmx/keeper"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// WasmxKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type WasmxKeeper interface {
	CloneWithStoreKey(storeKey storetypes.StoreKey, memKey storetypes.StoreKey) wasmxkeeper.Keeper
	Query(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, dependencies []string) ([]byte, error)
}
