package keeper

import (
	"context"

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
	return nil, nil
}

func (k QuerierDistribution) ValidatorDistributionInfo(goCtx context.Context, req *distributiontypes.QueryValidatorDistributionInfoRequest) (*distributiontypes.QueryValidatorDistributionInfoResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) ValidatorOutstandingRewards(goCtx context.Context, req *distributiontypes.QueryValidatorOutstandingRewardsRequest) (*distributiontypes.QueryValidatorOutstandingRewardsResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) ValidatorCommission(goCtx context.Context, req *distributiontypes.QueryValidatorCommissionRequest) (*distributiontypes.QueryValidatorCommissionResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) ValidatorSlashes(goCtx context.Context, req *distributiontypes.QueryValidatorSlashesRequest) (*distributiontypes.QueryValidatorSlashesResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) DelegationRewards(goCtx context.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) DelegationTotalRewards(goCtx context.Context, req *distributiontypes.QueryDelegationTotalRewardsRequest) (*distributiontypes.QueryDelegationTotalRewardsResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) DelegatorValidators(goCtx context.Context, req *distributiontypes.QueryDelegatorValidatorsRequest) (*distributiontypes.QueryDelegatorValidatorsResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) DelegatorWithdrawAddress(goCtx context.Context, req *distributiontypes.QueryDelegatorWithdrawAddressRequest) (*distributiontypes.QueryDelegatorWithdrawAddressResponse, error) {
	return nil, nil
}

func (k QuerierDistribution) CommunityPool(goCtx context.Context, req *distributiontypes.QueryCommunityPoolRequest) (*distributiontypes.QueryCommunityPoolResponse, error) {
	return nil, nil
}
