package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"mythos/v1/x/cosmosmod/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

func NewQuerier(keeper *Keeper) Querier {
	return Querier{Keeper: keeper}
}

func (k Querier) Validators(goCtx context.Context, req *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Validators not implemented")
	return &stakingtypes.QueryValidatorsResponse{}, nil
}

func (k Querier) Validator(goCtx context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Validator not implemented")
	return &stakingtypes.QueryValidatorResponse{}, nil
}

func (k Querier) ValidatorDelegations(goCtx context.Context, req *stakingtypes.QueryValidatorDelegationsRequest) (*stakingtypes.QueryValidatorDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("ValidatorDelegations not implemented")
	return &stakingtypes.QueryValidatorDelegationsResponse{}, nil
}

func (k Querier) ValidatorUnbondingDelegations(goCtx context.Context, req *stakingtypes.QueryValidatorUnbondingDelegationsRequest) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("ValidatorUnbondingDelegations not implemented")
	return &stakingtypes.QueryValidatorUnbondingDelegationsResponse{}, nil
}

func (k Querier) Delegation(goCtx context.Context, req *stakingtypes.QueryDelegationRequest) (*stakingtypes.QueryDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Delegation not implemented")
	return &stakingtypes.QueryDelegationResponse{}, nil
}

func (k Querier) UnbondingDelegation(goCtx context.Context, req *stakingtypes.QueryUnbondingDelegationRequest) (*stakingtypes.QueryUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("UnbondingDelegation not implemented")
	return &stakingtypes.QueryUnbondingDelegationResponse{}, nil
}

func (k Querier) DelegatorDelegations(goCtx context.Context, req *stakingtypes.QueryDelegatorDelegationsRequest) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorDelegations not implemented")
	return &stakingtypes.QueryDelegatorDelegationsResponse{}, nil
}

func (k Querier) DelegatorUnbondingDelegations(goCtx context.Context, req *stakingtypes.QueryDelegatorUnbondingDelegationsRequest) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorUnbondingDelegations not implemented")
	return &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{}, nil
}

func (k Querier) Redelegations(goCtx context.Context, req *stakingtypes.QueryRedelegationsRequest) (*stakingtypes.QueryRedelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Redelegations not implemented")
	return &stakingtypes.QueryRedelegationsResponse{}, nil
}

func (k Querier) DelegatorValidators(goCtx context.Context, req *stakingtypes.QueryDelegatorValidatorsRequest) (*stakingtypes.QueryDelegatorValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorValidators not implemented")
	return &stakingtypes.QueryDelegatorValidatorsResponse{}, nil
}

func (k Querier) DelegatorValidator(goCtx context.Context, req *stakingtypes.QueryDelegatorValidatorRequest) (*stakingtypes.QueryDelegatorValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorValidator not implemented")
	return &stakingtypes.QueryDelegatorValidatorResponse{}, nil
}

func (k Querier) HistoricalInfo(goCtx context.Context, req *stakingtypes.QueryHistoricalInfoRequest) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("HistoricalInfo not implemented")
	return &stakingtypes.QueryHistoricalInfoResponse{}, nil
}

func (k Querier) Pool(goCtx context.Context, req *stakingtypes.QueryPoolRequest) (*stakingtypes.QueryPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Pool not implemented")
	return &stakingtypes.QueryPoolResponse{}, nil
}

func (k Querier) Params(goCtx context.Context, req *stakingtypes.QueryParamsRequest) (*stakingtypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Params not implemented")
	return &stakingtypes.QueryParamsResponse{}, nil
}
