package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"wasmx/x/websrv/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) RegisterRoute(goCtx context.Context, msg *types.MsgRegisterRoute) (*types.MsgRegisterRouteResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRegisterRoute,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		sdk.NewAttribute(types.AttributeKeyRoute, msg.Path),
		sdk.NewAttribute(types.AttributeKeyContract, contractAddress.String()),
	))

	m.Keeper.RegisterRoute(ctx, msg.Path, contractAddress)
	return &types.MsgRegisterRouteResponse{}, nil
}
