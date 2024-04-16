package keeper

import (
	"mythos/v1/x/network/types"
)

var _ types.QueryServer = &Keeper{}
