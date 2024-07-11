package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	mcfg "mythos/v1/config"
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
	resp, err := k.ContractModuleExecution(ctx, "SetWithdrawAddress", msg)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.MsgSetWithdrawAddressResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

// withdraw validator commission
func (k KeeperDistribution) WithdrawValidatorCommission(goCtx context.Context, valAddr sdk.ValAddress) (sdk.Coins, error) {
	valAddrStr, err := k.ValidatorAddressCodec().BytesToString(valAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "validator: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}

	resp, err := k.WithdrawValidatorCommissionInternal(goCtx, &distributiontypes.MsgWithdrawValidatorCommission{ValidatorAddress: valAddrStr})
	if err != nil {
		return nil, err
	}
	return resp.Amount, nil
}

func (k KeeperDistribution) WithdrawValidatorCommissionInternal(goCtx context.Context, msg *distributiontypes.MsgWithdrawValidatorCommission) (*distributiontypes.MsgWithdrawValidatorCommissionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleExecution(ctx, "WithdrawValidatorCommission", msg)
	if err != nil {
		return nil, err
	}
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

	valAddrStr, err := k.ValidatorAddressCodec().BytesToString(valAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "validator: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}
	delAddrStr, err := k.AddressCodec().BytesToString(delAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "delegator: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}

	resp, err := k.WithdrawDelegatorReward(goCtx, &distributiontypes.MsgWithdrawDelegatorReward{
		DelegatorAddress: delAddrStr,
		ValidatorAddress: valAddrStr,
	})
	if err != nil {
		return nil, err
	}
	return resp.Amount, nil
}

// withdraw rewards from a delegation
func (k KeeperDistribution) FundCommunityPool(goCtx context.Context, msg *distributiontypes.MsgFundCommunityPool) (*distributiontypes.MsgFundCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleExecution(ctx, "FundCommunityPool", msg)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.MsgFundCommunityPoolResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) UpdateParams(goCtx context.Context, msg *distributiontypes.MsgUpdateParams) (*distributiontypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleExecution(ctx, "UpdateParams", msg)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.MsgUpdateParamsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) CommunityPoolSpend(goCtx context.Context, msg *distributiontypes.MsgCommunityPoolSpend) (*distributiontypes.MsgCommunityPoolSpendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleExecution(ctx, "CommunityPoolSpend", msg)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.MsgCommunityPoolSpendResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DepositValidatorRewardsPool(goCtx context.Context, msg *distributiontypes.MsgDepositValidatorRewardsPool) (*distributiontypes.MsgDepositValidatorRewardsPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleExecution(ctx, "DepositValidatorRewardsPool", msg)
	if err != nil {
		return nil, err
	}
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
	resp, err := k.ContractModuleExecution(ctx, "WithdrawDelegatorReward", msg)
	if err != nil {
		return nil, err
	}
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

	valAddrStr, err := k.ValidatorAddressCodec().BytesToString(val)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "validator: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}

	rewards, err := k.ValidatorOutstandingRewards(ctx, &distributiontypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: valAddrStr})
	if err != nil {
		return nil, err
	}
	return rewards.Rewards.Rewards, nil
}

func (k KeeperDistribution) Params(goCtx context.Context) (*distributiontypes.Params, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleQuery(ctx, "Params", &distributiontypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryParamsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp.Params, nil
}

func (k KeeperDistribution) ValidatorDistributionInfo(ctx sdk.Context, req *distributiontypes.QueryValidatorDistributionInfoRequest) (*distributiontypes.QueryValidatorDistributionInfoResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "ValidatorDistributionInfo", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryValidatorDistributionInfoResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ValidatorOutstandingRewards(ctx sdk.Context, req *distributiontypes.QueryValidatorOutstandingRewardsRequest) (*distributiontypes.QueryValidatorOutstandingRewardsResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "ValidatorOutstandingRewards", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryValidatorOutstandingRewardsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ValidatorCommission(ctx sdk.Context, req *distributiontypes.QueryValidatorCommissionRequest) (*distributiontypes.QueryValidatorCommissionResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "ValidatorCommission", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryValidatorCommissionResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ValidatorSlashes(ctx sdk.Context, req *distributiontypes.QueryValidatorSlashesRequest) (*distributiontypes.QueryValidatorSlashesResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "ValidatorSlashes", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryValidatorSlashesResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegationRewards(ctx sdk.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "DelegationRewards", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryDelegationRewardsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegationTotalRewards(ctx sdk.Context, req *distributiontypes.QueryDelegationTotalRewardsRequest) (*distributiontypes.QueryDelegationTotalRewardsResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "DelegationTotalRewards", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryDelegationTotalRewardsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegatorValidators(ctx sdk.Context, req *distributiontypes.QueryDelegatorValidatorsRequest) (*distributiontypes.QueryDelegatorValidatorsResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "DelegatorValidators", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryDelegatorValidatorsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) DelegatorWithdrawAddress(ctx sdk.Context, req *distributiontypes.QueryDelegatorWithdrawAddressRequest) (*distributiontypes.QueryDelegatorWithdrawAddressResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "DelegatorWithdrawAddress", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryDelegatorWithdrawAddressResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) CommunityPool(ctx sdk.Context, req *distributiontypes.QueryCommunityPoolRequest) (*distributiontypes.QueryCommunityPoolResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "CommunityPool", req)
	if err != nil {
		return nil, err
	}
	var cresp distributiontypes.QueryCommunityPoolResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

func (k KeeperDistribution) ContractModuleQuery(ctx sdk.Context, fname string, req interface{}) (*wasmxtypes.ContractResponse, error) {
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

func (k KeeperDistribution) ContractModuleExecution(ctx sdk.Context, fname string, req interface{}) (*wasmxtypes.ContractResponse, error) {
	msgbz, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msgbz2 := []byte(fmt.Sprintf(`{"%s":%s}`, fname, string(msgbz)))
	res1, err := k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
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
