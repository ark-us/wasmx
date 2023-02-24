package keeper

import (
	"wasmx/x/wasmx/types"
)

var _ types.QueryServer = Keeper{}
