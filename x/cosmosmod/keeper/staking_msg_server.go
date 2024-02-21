package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type msgStakingServer struct {
	Keeper *KeeperStaking
}

// NewMsgStakingServerImpl returns an implementation of the MsgServer interface
func NewMsgStakingServerImpl(keeper *KeeperStaking) stakingtypes.MsgServer {
	return &msgStakingServer{
		Keeper: keeper,
	}
}

var _ stakingtypes.MsgServer = msgStakingServer{}

func (m msgStakingServer) CreateValidator(goCtx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"CreateValidator":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.ValidatorAddress,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.MsgCreateValidatorResponse{}, nil
}

func (m msgStakingServer) EditValidator(goCtx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("EditValidator not implemented")
	return &stakingtypes.MsgEditValidatorResponse{}, nil
}

func (m msgStakingServer) Delegate(goCtx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("Delegate not implemented")
	return &stakingtypes.MsgDelegateResponse{}, nil
}

func (m msgStakingServer) BeginRedelegate(goCtx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("BeginRedelegate not implemented")
	return &stakingtypes.MsgBeginRedelegateResponse{}, nil
}

func (m msgStakingServer) Undelegate(goCtx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("Undelegate not implemented")
	return &stakingtypes.MsgUndelegateResponse{}, nil
}

func (m msgStakingServer) CancelUnbondingDelegation(goCtx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("CancelUnbondingDelegation not implemented")
	return &stakingtypes.MsgCancelUnbondingDelegationResponse{}, nil
}

func (m msgStakingServer) UpdateParams(goCtx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("UpdateParams not implemented")
	return &stakingtypes.MsgUpdateParamsResponse{}, nil
}
