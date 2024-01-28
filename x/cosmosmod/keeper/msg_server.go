package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		Keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) CreateValidator(goCtx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := json.Marshal(msg)
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

func (m msgServer) EditValidator(goCtx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	return &stakingtypes.MsgEditValidatorResponse{}, nil
}

func (m msgServer) Delegate(goCtx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	return &stakingtypes.MsgDelegateResponse{}, nil
}

func (m msgServer) BeginRedelegate(goCtx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	return &stakingtypes.MsgBeginRedelegateResponse{}, nil
}

func (m msgServer) Undelegate(goCtx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	return &stakingtypes.MsgUndelegateResponse{}, nil
}

func (m msgServer) CancelUnbondingDelegation(goCtx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	return &stakingtypes.MsgCancelUnbondingDelegationResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	return &stakingtypes.MsgUpdateParamsResponse{}, nil
}
