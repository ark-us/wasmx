package keeper

import (
	"context"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// QuerierStaking is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierStaking struct {
	Keeper *KeeperStaking
}

var _ stakingtypes.QueryServer = QuerierStaking{}

func NewQuerierStaking(keeper *KeeperStaking) QuerierStaking {
	return QuerierStaking{Keeper: keeper}
}

func (k QuerierStaking) Validators(goCtx context.Context, req *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	validators, err := k.Keeper.GetAllValidators(goCtx)
	if err != nil {
		return nil, err
	}
	if req.Status == "" {
		return &stakingtypes.QueryValidatorsResponse{Validators: validators}, nil
	}
	// TODO filtering in the contract
	filtered := make([]stakingtypes.Validator, 0)
	for _, valid := range validators {
		if valid.Status.String() == req.Status {
			filtered = append(filtered, valid)
		}
	}
	return &stakingtypes.QueryValidatorsResponse{Validators: filtered}, nil
}

func (k QuerierStaking) Validator(goCtx context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
	addr, err := k.Keeper.AddressCodec().StringToBytes(req.ValidatorAddr)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	v, err := k.Keeper.Validator(goCtx, addr)
	if err != nil {
		return nil, err
	}
	validator := v.(stakingtypes.Validator)
	return &stakingtypes.QueryValidatorResponse{Validator: validator}, nil
}

func (k QuerierStaking) ValidatorDelegations(goCtx context.Context, req *stakingtypes.QueryValidatorDelegationsRequest) (*stakingtypes.QueryValidatorDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.ValidatorDelegations not implemented")
	return &stakingtypes.QueryValidatorDelegationsResponse{}, nil
}

func (k QuerierStaking) ValidatorUnbondingDelegations(goCtx context.Context, req *stakingtypes.QueryValidatorUnbondingDelegationsRequest) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.ValidatorUnbondingDelegations not implemented")
	return &stakingtypes.QueryValidatorUnbondingDelegationsResponse{}, nil
}

func (k QuerierStaking) Delegation(goCtx context.Context, req *stakingtypes.QueryDelegationRequest) (*stakingtypes.QueryDelegationResponse, error) {
	delegator, err := k.Keeper.accBech32Codec.StringToAccAddressPrefixed(req.DelegatorAddr)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	addrVal, err := k.Keeper.valBech32Codec.StringToValAddressPrefixed(req.ValidatorAddr)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	delegation, err := k.Keeper.DelegationInternal(goCtx, delegator, addrVal)
	if err != nil {
		return nil, err
	}
	return delegation, nil
}

func (k QuerierStaking) UnbondingDelegation(goCtx context.Context, req *stakingtypes.QueryUnbondingDelegationRequest) (*stakingtypes.QueryUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.UnbondingDelegation not implemented")
	return &stakingtypes.QueryUnbondingDelegationResponse{}, nil
}

func (k QuerierStaking) DelegatorDelegations(goCtx context.Context, req *stakingtypes.QueryDelegatorDelegationsRequest) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.DelegatorDelegations not implemented")
	return &stakingtypes.QueryDelegatorDelegationsResponse{}, nil
}

func (k QuerierStaking) DelegatorUnbondingDelegations(goCtx context.Context, req *stakingtypes.QueryDelegatorUnbondingDelegationsRequest) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.DelegatorUnbondingDelegations not implemented")
	return &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{}, nil
}

func (k QuerierStaking) Redelegations(goCtx context.Context, req *stakingtypes.QueryRedelegationsRequest) (*stakingtypes.QueryRedelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.Redelegations not implemented")
	return &stakingtypes.QueryRedelegationsResponse{}, nil
}

func (k QuerierStaking) DelegatorValidators(goCtx context.Context, req *stakingtypes.QueryDelegatorValidatorsRequest) (*stakingtypes.QueryDelegatorValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.DelegatorValidators not implemented")
	return &stakingtypes.QueryDelegatorValidatorsResponse{}, nil
}

func (k QuerierStaking) DelegatorValidator(goCtx context.Context, req *stakingtypes.QueryDelegatorValidatorRequest) (*stakingtypes.QueryDelegatorValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.DelegatorValidator not implemented")
	return &stakingtypes.QueryDelegatorValidatorResponse{}, nil
}

func (k QuerierStaking) HistoricalInfo(goCtx context.Context, req *stakingtypes.QueryHistoricalInfoRequest) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierStaking.HistoricalInfo not implemented")
	return &stakingtypes.QueryHistoricalInfoResponse{}, nil
}

func (k QuerierStaking) Pool(goCtx context.Context, req *stakingtypes.QueryPoolRequest) (*stakingtypes.QueryPoolResponse, error) {
	return k.Keeper.Pool(goCtx, req)
}

func (k QuerierStaking) Params(goCtx context.Context, req *stakingtypes.QueryParamsRequest) (*stakingtypes.QueryParamsResponse, error) {
	params, err := k.Keeper.GetParams(goCtx)
	if err != nil {
		return nil, err
	}
	return &stakingtypes.QueryParamsResponse{Params: params}, nil
}
