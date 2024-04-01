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
func (k KeeperDistribution) WithdrawValidatorCommission(ctx context.Context, valAddr sdk.ValAddress) (sdk.Coins, error) {
	return nil, nil
}

// withdraw rewards from a delegation
func (k KeeperDistribution) WithdrawDelegationRewards(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	return nil, nil
}

// delete all slash events
func (k KeeperDistribution) DeleteAllValidatorSlashEvents(ctx context.Context) {
}

// delete all historical rewards
func (k KeeperDistribution) DeleteAllValidatorHistoricalRewards(ctx context.Context) {
}

// get outstanding rewards
func (k KeeperDistribution) GetValidatorOutstandingRewardsCoins(ctx context.Context, val sdk.ValAddress) (sdk.DecCoins, error) {
	return nil, nil
}
