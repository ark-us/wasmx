package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.SetWithdrawAddress(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m msgDistributionServer) WithdrawDelegatorReward(goCtx context.Context, msg *distributiontypes.MsgWithdrawDelegatorReward) (*distributiontypes.MsgWithdrawDelegatorRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.WithdrawDelegatorReward(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m msgDistributionServer) WithdrawValidatorCommission(goCtx context.Context, msg *distributiontypes.MsgWithdrawValidatorCommission) (*distributiontypes.MsgWithdrawValidatorCommissionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.WithdrawValidatorCommissionInternal(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m msgDistributionServer) FundCommunityPool(goCtx context.Context, msg *distributiontypes.MsgFundCommunityPool) (*distributiontypes.MsgFundCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.FundCommunityPool(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m msgDistributionServer) UpdateParams(goCtx context.Context, msg *distributiontypes.MsgUpdateParams) (*distributiontypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.UpdateParams(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m msgDistributionServer) CommunityPoolSpend(goCtx context.Context, msg *distributiontypes.MsgCommunityPoolSpend) (*distributiontypes.MsgCommunityPoolSpendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.CommunityPoolSpend(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m msgDistributionServer) DepositValidatorRewardsPool(goCtx context.Context, msg *distributiontypes.MsgDepositValidatorRewardsPool) (*distributiontypes.MsgDepositValidatorRewardsPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.DepositValidatorRewardsPool(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
