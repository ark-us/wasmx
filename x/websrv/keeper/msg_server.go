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

func (m msgServer) RegisterOAuthClient(goCtx context.Context, msg *types.MsgRegisterOAuthClient) (*types.MsgRegisterOAuthClientResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr, err := sdk.AccAddressFromBech32(msg.ClientAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.ClientAddress),
		sdk.NewAttribute(sdk.AttributeKeyAction, "register_outh2_client"),
	))

	codeId, checksum, err := m.Keeper.Create(ctx, senderAddr, msg.WasmByteCode)
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreCodeResponse{
		CodeId:   codeId,
		Checksum: checksum,
	}, nil
}
