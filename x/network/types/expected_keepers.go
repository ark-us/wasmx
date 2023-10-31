package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// WasmxKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, dependencies []string) ([]byte, error)
}
