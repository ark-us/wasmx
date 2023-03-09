package keeper

import (
	"wasmx/x/websrv/types"
)

var _ types.QueryServer = Keeper{}
