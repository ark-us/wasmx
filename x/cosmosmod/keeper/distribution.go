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

// get outstanding rewards
func (k KeeperDistribution) GetValidatorOutstandingRewardsCoins(goCtx context.Context, val sdk.ValAddress) (sdk.DecCoins, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.GetValidatorOutstandingRewardsCoins not implemented")
	return nil, nil
}
