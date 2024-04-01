package keeper

import (
	"context"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type msgDistributionServer struct {
	Keeper *KeeperDistribution
}

// NewMsgDistributionServerImpl returns an implementation of the MsgServer interface
func NewMsgDistributionServerImpl(keeper *KeeperDistribution) distributiontypes.MsgServer {
	return &msgDistributionServer{
		Keeper: keeper,
	}
}

var _ distributiontypes.MsgServer = msgDistributionServer{}

func (m msgDistributionServer) SetWithdrawAddress(goCtx context.Context, msg *distributiontypes.MsgSetWithdrawAddress) (*distributiontypes.MsgSetWithdrawAddressResponse, error) {
	return nil, nil
}

func (m msgDistributionServer) WithdrawDelegatorReward(goCtx context.Context, msg *distributiontypes.MsgWithdrawDelegatorReward) (*distributiontypes.MsgWithdrawDelegatorRewardResponse, error) {
	return nil, nil
}

func (m msgDistributionServer) WithdrawValidatorCommission(goCtx context.Context, msg *distributiontypes.MsgWithdrawValidatorCommission) (*distributiontypes.MsgWithdrawValidatorCommissionResponse, error) {
	return nil, nil
}

func (m msgDistributionServer) FundCommunityPool(goCtx context.Context, msg *distributiontypes.MsgFundCommunityPool) (*distributiontypes.MsgFundCommunityPoolResponse, error) {
	return nil, nil
}

func (m msgDistributionServer) UpdateParams(goCtx context.Context, msg *distributiontypes.MsgUpdateParams) (*distributiontypes.MsgUpdateParamsResponse, error) {
	return nil, nil
}

func (m msgDistributionServer) CommunityPoolSpend(goCtx context.Context, msg *distributiontypes.MsgCommunityPoolSpend) (*distributiontypes.MsgCommunityPoolSpendResponse, error) {
	return nil, nil
}

func (m msgDistributionServer) DepositValidatorRewardsPool(goCtx context.Context, msg *distributiontypes.MsgDepositValidatorRewardsPool) (*distributiontypes.MsgDepositValidatorRewardsPoolResponse, error) {
	return nil, nil
}
