package types

import (
	"wasmx/v1/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
}
