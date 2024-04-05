package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// TODO Hooks()
// AfterValidatorCreated
// AfterValidatorRemoved
// BeforeDelegationCreated
// BeforeDelegationSharesModified
// AfterDelegationModified ...

func (k KeeperDistribution) SetWithdrawAddress(ctx sdk.Context, msg *distributiontypes.MsgSetWithdrawAddress) (*distributiontypes.MsgSetWithdrawAddressResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "SetWithdrawAddress", msg)
	var cresp distributiontypes.MsgSetWithdrawAddressResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
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
	resp, err := k.ContractModuleCall(ctx, "WithdrawValidatorCommission", msg)
	var cresp distributiontypes.MsgWithdrawValidatorCommissionResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

// withdraw rewards from a delegation
func (k KeeperDistribution) WithdrawDelegationRewards(goCtx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.WithdrawDelegationRewards not implemented")
	resp, err := k.WithdrawDelegatorReward(goCtx, &distributiontypes.MsgWithdrawDelegatorReward{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Amount, nil
}

// withdraw rewards from a delegation
func (k KeeperDistribution) FundCommunityPool(goCtx context.Context, msg *distributiontypes.MsgFundCommunityPool) (*distributiontypes.MsgFundCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleCall(ctx, "FundCommunityPool", msg)
	var cresp distributiontypes.MsgFundCommunityPoolResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) UpdateParams(goCtx context.Context, msg *distributiontypes.MsgUpdateParams) (*distributiontypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleCall(ctx, "UpdateParams", msg)
	var cresp distributiontypes.MsgUpdateParamsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) CommunityPoolSpend(goCtx context.Context, msg *distributiontypes.MsgCommunityPoolSpend) (*distributiontypes.MsgCommunityPoolSpendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleCall(ctx, "CommunityPoolSpend", msg)
	var cresp distributiontypes.MsgCommunityPoolSpendResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DepositValidatorRewardsPool(goCtx context.Context, msg *distributiontypes.MsgDepositValidatorRewardsPool) (*distributiontypes.MsgDepositValidatorRewardsPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleCall(ctx, "DepositValidatorRewardsPool", msg)
	var cresp distributiontypes.MsgDepositValidatorRewardsPoolResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
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
	resp, err := k.ContractModuleCall(ctx, "WithdrawDelegatorReward", msg)
	var cresp distributiontypes.MsgWithdrawDelegatorRewardResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

// get outstanding rewards
func (k KeeperDistribution) GetValidatorOutstandingRewardsCoins(goCtx context.Context, val sdk.ValAddress) (sdk.DecCoins, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// k.Logger(ctx).Debug("KeeperDistribution.GetValidatorOutstandingRewardsCoins not implemented")
	// return nil, nil
	resp, err := k.ContractModuleCall(ctx, "WithdrawDelegatorReward", msg)
	var cresp distributiontypes.MsgWithdrawDelegatorRewardResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return cresp.Amount, nil
}

func (k KeeperDistribution) Params(goCtx context.Context) (*distributiontypes.Params, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperDistribution.Params not implemented")
	return nil, nil
}

func (k KeeperDistribution) ValidatorDistributionInfo(ctx sdk.Context, req *distributiontypes.QueryValidatorDistributionInfoRequest) (*distributiontypes.QueryValidatorDistributionInfoResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "ValidatorDistributionInfo", req)
	var cresp distributiontypes.QueryValidatorDistributionInfoResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ValidatorOutstandingRewards(ctx sdk.Context, req *distributiontypes.QueryValidatorOutstandingRewardsRequest) (*distributiontypes.QueryValidatorOutstandingRewardsResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "ValidatorOutstandingRewards", req)
	var cresp distributiontypes.QueryValidatorOutstandingRewardsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ValidatorCommission(ctx sdk.Context, req *distributiontypes.QueryValidatorCommissionRequest) (*distributiontypes.QueryValidatorCommissionResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "ValidatorCommission", req)
	var cresp distributiontypes.QueryValidatorCommissionResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ValidatorSlashes(ctx sdk.Context, req *distributiontypes.QueryValidatorSlashesRequest) (*distributiontypes.QueryValidatorSlashesResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "ValidatorSlashes", req)
	var cresp distributiontypes.QueryValidatorSlashesResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegationRewards(ctx sdk.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "DelegationRewards", req)
	var cresp distributiontypes.QueryDelegationRewardsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegationTotalRewards(ctx sdk.Context, req *distributiontypes.QueryDelegationTotalRewardsRequest) (*distributiontypes.QueryDelegationTotalRewardsResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "DelegationTotalRewards", req)
	var cresp distributiontypes.QueryDelegationTotalRewardsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegatorValidators(ctx sdk.Context, req *distributiontypes.QueryDelegatorValidatorsRequest) (*distributiontypes.QueryDelegatorValidatorsResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "DelegatorValidators", req)
	var cresp distributiontypes.QueryDelegatorValidatorsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegatorWithdrawAddress(ctx sdk.Context, req *distributiontypes.QueryDelegatorWithdrawAddressRequest) (*distributiontypes.QueryDelegatorWithdrawAddressResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "DelegatorWithdrawAddress", req)
	var cresp distributiontypes.QueryDelegatorWithdrawAddressResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) CommunityPool(ctx sdk.Context, req *distributiontypes.QueryCommunityPoolRequest) (*distributiontypes.QueryCommunityPoolResponse, error) {
	resp, err := k.ContractModuleCall(ctx, "CommunityPool", req)
	var cresp distributiontypes.QueryCommunityPoolResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ContractModuleCall(ctx sdk.Context, fname string, req interface{}) (*wasmxtypes.ContractResponse, error) {
	msgbz, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msgbz2 := []byte(fmt.Sprintf(`{"%s":%s}`, fname, string(msgbz)))
	res1, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_DISTRIBUTION,
		Contract: wasmxtypes.ROLE_DISTRIBUTION,
		Msg:      msgbz2,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res1.Data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
