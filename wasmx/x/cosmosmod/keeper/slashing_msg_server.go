package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type msgSlashingServer struct {
	Keeper *KeeperSlashing
}

// NewMsgSlashingServerImpl returns an implementation of the MsgServer interface
func NewMsgSlashingServerImpl(keeper *KeeperSlashing) slashingtypes.MsgServer {
	return &msgSlashingServer{
		Keeper: keeper,
	}
}

var _ slashingtypes.MsgServer = msgSlashingServer{}

func (m msgSlashingServer) Unjail(goCtx context.Context, msg *slashingtypes.MsgUnjail) (*slashingtypes.MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := m.Keeper.Unjail(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m msgSlashingServer) UpdateParams(goCtx context.Context, msg *slashingtypes.MsgUpdateParams) (*slashingtypes.MsgUpdateParamsResponse, error) {
	if m.Keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf(
			"slashing.UpdateParams: expected gov account as only signer for proposal message; invalid authority; expected %s, got %s",
			m.Keeper.authority, msg.Authority)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.UpdateParams(ctx, msg.Params)
	if err != nil {
		return nil, err
	}
	return &slashingtypes.MsgUpdateParamsResponse{}, nil
}
