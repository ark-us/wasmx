package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"mythos/v1/x/cosmosmod/types"
)

// QuerierStaking is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierStaking struct {
	Keeper *KeeperStaking
}

var _ types.QueryStakingServer = QuerierStaking{}

func NewQuerierStaking(keeper *KeeperStaking) QuerierStaking {
	return QuerierStaking{Keeper: keeper}
}

func (k QuerierStaking) Validators(goCtx context.Context, req *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Validators not implemented")
	return &stakingtypes.QueryValidatorsResponse{}, nil
}

func (k QuerierStaking) Validator(goCtx context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking - Validator not implemented")
	// addr, err := sdk.AccAddressFromBech32(req.ValidatorAddr)
	// if err != nil {
	// 	return nil, sdkerr.Wrap(err, "sender")
	// }
	// validator, err := k.Keeper.Validator(goCtx, addr)
	// if err != nil {
	// 	return nil, err
	// }

	// return &stakingtypes.QueryValidatorResponse{Validator: validator}, nil
	return nil, nil
}

func (k QuerierStaking) ValidatorDelegations(goCtx context.Context, req *stakingtypes.QueryValidatorDelegationsRequest) (*stakingtypes.QueryValidatorDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("ValidatorDelegations not implemented")
	return &stakingtypes.QueryValidatorDelegationsResponse{}, nil
}

func (k QuerierStaking) ValidatorUnbondingDelegations(goCtx context.Context, req *stakingtypes.QueryValidatorUnbondingDelegationsRequest) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("ValidatorUnbondingDelegations not implemented")
	return &stakingtypes.QueryValidatorUnbondingDelegationsResponse{}, nil
}

func (k QuerierStaking) Delegation(goCtx context.Context, req *stakingtypes.QueryDelegationRequest) (*stakingtypes.QueryDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Delegation not implemented")
	return &stakingtypes.QueryDelegationResponse{}, nil
}

func (k QuerierStaking) UnbondingDelegation(goCtx context.Context, req *stakingtypes.QueryUnbondingDelegationRequest) (*stakingtypes.QueryUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("UnbondingDelegation not implemented")
	return &stakingtypes.QueryUnbondingDelegationResponse{}, nil
}

func (k QuerierStaking) DelegatorDelegations(goCtx context.Context, req *stakingtypes.QueryDelegatorDelegationsRequest) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorDelegations not implemented")
	return &stakingtypes.QueryDelegatorDelegationsResponse{}, nil
}

func (k QuerierStaking) DelegatorUnbondingDelegations(goCtx context.Context, req *stakingtypes.QueryDelegatorUnbondingDelegationsRequest) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorUnbondingDelegations not implemented")
	return &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{}, nil
}

func (k QuerierStaking) Redelegations(goCtx context.Context, req *stakingtypes.QueryRedelegationsRequest) (*stakingtypes.QueryRedelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Redelegations not implemented")
	return &stakingtypes.QueryRedelegationsResponse{}, nil
}

func (k QuerierStaking) DelegatorValidators(goCtx context.Context, req *stakingtypes.QueryDelegatorValidatorsRequest) (*stakingtypes.QueryDelegatorValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorValidators not implemented")
	return &stakingtypes.QueryDelegatorValidatorsResponse{}, nil
}

func (k QuerierStaking) DelegatorValidator(goCtx context.Context, req *stakingtypes.QueryDelegatorValidatorRequest) (*stakingtypes.QueryDelegatorValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("DelegatorValidator not implemented")
	return &stakingtypes.QueryDelegatorValidatorResponse{}, nil
}

func (k QuerierStaking) HistoricalInfo(goCtx context.Context, req *stakingtypes.QueryHistoricalInfoRequest) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("HistoricalInfo not implemented")
	return &stakingtypes.QueryHistoricalInfoResponse{}, nil
}

func (k QuerierStaking) Pool(goCtx context.Context, req *stakingtypes.QueryPoolRequest) (*stakingtypes.QueryPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Pool not implemented")
	return &stakingtypes.QueryPoolResponse{}, nil
}

func (k QuerierStaking) Params(goCtx context.Context, req *stakingtypes.QueryParamsRequest) (*stakingtypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Params not implemented")
	return &stakingtypes.QueryParamsResponse{}, nil
}
