package keeper

import (
	"xwasm/x/xwasm/types"
)

var _ types.QueryServer = Keeper{}
