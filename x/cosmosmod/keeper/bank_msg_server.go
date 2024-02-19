package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/x/cosmosmod/types"
)

type msgBankServer struct {
	Keeper *KeeperBank
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgBankServerImpl(keeper *KeeperBank) types.MsgBankServer {
	return &msgBankServer{
		Keeper: keeper,
	}
}

var _ types.MsgBankServer = msgBankServer{}

func (m msgBankServer) Send(goCtx context.Context, msg *banktypes.MsgSend) (*banktypes.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("Send not implemented")
	return &banktypes.MsgSendResponse{}, nil
}

func (m msgBankServer) MultiSend(goCtx context.Context, msg *banktypes.MsgMultiSend) (*banktypes.MsgMultiSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("MultiSend not implemented")
	return &banktypes.MsgMultiSendResponse{}, nil
}

func (m msgBankServer) UpdateParams(goCtx context.Context, msg *banktypes.MsgUpdateParams) (*banktypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("UpdateParams not implemented")
	return &banktypes.MsgUpdateParamsResponse{}, nil
}

func (m msgBankServer) SetSendEnabled(goCtx context.Context, msg *banktypes.MsgSetSendEnabled) (*banktypes.MsgSetSendEnabledResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("SetSendEnabled not implemented")
	return &banktypes.MsgSetSendEnabledResponse{}, nil
}
