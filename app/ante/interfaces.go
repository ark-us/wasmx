package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmxKeeperI interface {
	GetAlias(ctx sdk.Context, addr sdk.AccAddress) (sdk.AccAddress, bool)
}
