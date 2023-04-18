package keeper

import (
	"mythos/v1/x/websrv/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	for _, pair := range params.ParamSetPairs() {
		k.paramstore.GetIfExists(ctx, pair.Key, pair.Value)
	}
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetOauthClientRegistrationOnlyEId returns wether or not eIDs are required
// for registering oauth2 client applications
func (k Keeper) GetOauthClientRegistrationOnlyEId(ctx sdk.Context) (onlyEId bool) {
	k.paramstore.GetIfExists(ctx, types.ParamStoreKeyOauthClientRegistrationOnlyEId, &onlyEId)
	return onlyEId
}
