package keeper

import (
	"context"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
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
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"EditValidator":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.ValidatorAddress,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.MsgEditValidatorResponse{}, nil
}

func (m msgStakingServer) Delegate(goCtx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"Delegate":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.ValidatorAddress,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.MsgDelegateResponse{}, nil
}

func (m msgStakingServer) BeginRedelegate(goCtx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"BeginRedelegate":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.DelegatorAddress,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.MsgBeginRedelegateResponse{}, nil
}

func (m msgStakingServer) Undelegate(goCtx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"Undelegate":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.DelegatorAddress,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.MsgUndelegateResponse{}, nil
}

func (m msgStakingServer) CancelUnbondingDelegation(goCtx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"CancelUnbondingDelegation":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.DelegatorAddress,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.MsgCancelUnbondingDelegationResponse{}, nil
}

func (m msgStakingServer) UpdateParams(goCtx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority := m.Keeper.GetAuthority()
	if authority != msg.Authority {
		return nil, sdkerr.Wrapf(errortypes.ErrUnauthorized, "invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"UpdateParams":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   authority,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.MsgUpdateParamsResponse{}, nil
}
