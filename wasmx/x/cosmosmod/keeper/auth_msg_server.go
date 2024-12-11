package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type msgAuthServer struct {
	Keeper *KeeperAuth
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgAuthServerImpl(keeper *KeeperAuth) authtypes.MsgServer {
	return &msgAuthServer{
		Keeper: keeper,
	}
}

var _ authtypes.MsgServer = msgAuthServer{}

func (m msgAuthServer) UpdateParams(goCtx context.Context, msg *authtypes.MsgUpdateParams) (*authtypes.MsgUpdateParamsResponse, error) {
	if m.Keeper.authority != msg.Authority {
		return nil, fmt.Errorf(
			"expected gov account as only signer for proposal message; invalid authority; expected %s, got %s",
			m.Keeper.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("Auth.UpdateParams not implemented")
	// if err := m.Keeper.SetParams(ctx, msg.Params); err != nil {
	// 	return nil, err
	// }

	return &authtypes.MsgUpdateParamsResponse{}, nil
}
