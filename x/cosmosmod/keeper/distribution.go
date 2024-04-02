package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// TODO Hooks()
// AfterValidatorCreated
// AfterValidatorRemoved
// BeforeDelegationCreated
// BeforeDelegationSharesModified
// AfterDelegationModified ...

func (k KeeperDistribution) SetWithdrawAddress(ctx sdk.Context, msg *distributiontypes.MsgSetWithdrawAddress) (*distributiontypes.MsgSetWithdrawAddressResponse, error) {
	// TODO
	k.Logger(ctx).Debug("KeeperDistribution.SetWithdrawAddress not implemented")
	return nil, nil
}

// withdraw validator commission
func (k KeeperDistribution) WithdrawValidatorCommission(goCtx context.Context, valAddr sdk.ValAddress) (sdk.Coins, error) {
	resp, err := k.WithdrawValidatorCommissionInternal(goCtx, &distributiontypes.MsgWithdrawValidatorCommission{ValidatorAddress: valAddr.String()})
	if err != nil {
		return nil, err
	}
	return resp.Amount, nil
}

func (k KeeperDistribution) WithdrawValidatorCommissionInternal(goCtx context.Context, msg *distributiontypes.MsgWithdrawValidatorCommission) (*distributiontypes.MsgWithdrawValidatorCommissionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.WithdrawValidatorCommission not implemented")
	return nil, nil
}

// withdraw rewards from a delegation
func (k KeeperDistribution) WithdrawDelegationRewards(goCtx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.WithdrawDelegationRewards not implemented")
	return nil, nil
}

// withdraw rewards from a delegation
func (k KeeperDistribution) FundCommunityPool(goCtx context.Context, msg *distributiontypes.MsgFundCommunityPool) (*distributiontypes.MsgFundCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.FundCommunityPool not implemented")
	return nil, nil
}

func (k KeeperDistribution) UpdateParams(goCtx context.Context, msg *distributiontypes.MsgUpdateParams) (*distributiontypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.UpdateParams not implemented")
	return nil, nil
}

func (k KeeperDistribution) CommunityPoolSpend(goCtx context.Context, msg *distributiontypes.MsgCommunityPoolSpend) (*distributiontypes.MsgCommunityPoolSpendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.UpdateParams not implemented")
	return nil, nil
}

func (k KeeperDistribution) DepositValidatorRewardsPool(goCtx context.Context, msg *distributiontypes.MsgDepositValidatorRewardsPool) (*distributiontypes.MsgDepositValidatorRewardsPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.UpdateParams not implemented")
	return nil, nil
}

// delete all slash events
func (k KeeperDistribution) DeleteAllValidatorSlashEvents(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.DeleteAllValidatorSlashEvents not implemented")
}

// delete all historical rewards
func (k KeeperDistribution) DeleteAllValidatorHistoricalRewards(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.DeleteAllValidatorHistoricalRewards not implemented")
}

func (k KeeperDistribution) WithdrawDelegatorReward(goCtx context.Context, msg *distributiontypes.MsgWithdrawDelegatorReward) (*distributiontypes.MsgWithdrawDelegatorRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.WithdrawDelegatorReward not implemented")
	return &distributiontypes.MsgWithdrawDelegatorRewardResponse{}, nil
}

// get outstanding rewards
func (k KeeperDistribution) GetValidatorOutstandingRewardsCoins(goCtx context.Context, val sdk.ValAddress) (sdk.DecCoins, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.GetValidatorOutstandingRewardsCoins not implemented")
	return nil, nil
}

func (k KeeperDistribution) Params(goCtx context.Context) (*distributiontypes.Params, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.Params not implemented")
	return nil, nil
}
