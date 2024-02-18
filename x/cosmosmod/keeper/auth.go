package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (k Keeper) SetParams(ctx sdk.Context, params authtypes.Params) error {
	k.Logger(ctx).Error("Auth.Params not implemented")
	return nil
}

func (k Keeper) GetParams(ctx sdk.Context) (params authtypes.Params) {
	k.Logger(ctx).Error("Auth.GetParams not implemented")
	return params
}
