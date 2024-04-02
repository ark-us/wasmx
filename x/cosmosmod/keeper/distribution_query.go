package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// QuerierDistribution is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierDistribution struct {
	Keeper *KeeperDistribution
}

var _ distributiontypes.QueryServer = QuerierDistribution{}

func NewQuerierDistribution(keeper *KeeperDistribution) QuerierDistribution {
	return QuerierDistribution{Keeper: keeper}
}

func (k QuerierDistribution) Params(goCtx context.Context, req *distributiontypes.QueryParamsRequest) (*distributiontypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := k.Keeper.Params(ctx)
	if err != nil {
		return nil, err
	}
	return &distributiontypes.QueryParamsResponse{Params: *params}, nil
}

func (k QuerierDistribution) ValidatorDistributionInfo(goCtx context.Context, req *distributiontypes.QueryValidatorDistributionInfoRequest) (*distributiontypes.QueryValidatorDistributionInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.ValidatorDistributionInfo not implemented")
	return nil, nil
}

func (k QuerierDistribution) ValidatorOutstandingRewards(goCtx context.Context, req *distributiontypes.QueryValidatorOutstandingRewardsRequest) (*distributiontypes.QueryValidatorOutstandingRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.ValidatorOutstandingRewards not implemented")
	return nil, nil
}

func (k QuerierDistribution) ValidatorCommission(goCtx context.Context, req *distributiontypes.QueryValidatorCommissionRequest) (*distributiontypes.QueryValidatorCommissionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.ValidatorCommission not implemented")
	return nil, nil
}

func (k QuerierDistribution) ValidatorSlashes(goCtx context.Context, req *distributiontypes.QueryValidatorSlashesRequest) (*distributiontypes.QueryValidatorSlashesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.ValidatorSlashes not implemented")
	return nil, nil
}

func (k QuerierDistribution) DelegationRewards(goCtx context.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.DelegationRewards not implemented")
	return nil, nil
}

func (k QuerierDistribution) DelegationTotalRewards(goCtx context.Context, req *distributiontypes.QueryDelegationTotalRewardsRequest) (*distributiontypes.QueryDelegationTotalRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.DelegationTotalRewards not implemented")
	return nil, nil
}

func (k QuerierDistribution) DelegatorValidators(goCtx context.Context, req *distributiontypes.QueryDelegatorValidatorsRequest) (*distributiontypes.QueryDelegatorValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.DelegatorValidators not implemented")
	return nil, nil
}

func (k QuerierDistribution) DelegatorWithdrawAddress(goCtx context.Context, req *distributiontypes.QueryDelegatorWithdrawAddressRequest) (*distributiontypes.QueryDelegatorWithdrawAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.DelegatorWithdrawAddress not implemented")
	return nil, nil
}

func (k QuerierDistribution) CommunityPool(goCtx context.Context, req *distributiontypes.QueryCommunityPoolRequest) (*distributiontypes.QueryCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Debug("KeeperDistribution.CommunityPool not implemented")
	return nil, nil
}
