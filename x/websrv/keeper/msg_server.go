package keeper

import (
	"context"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/websrv/types"
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

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerr.Wrap(err, "owner")
	}

	if m.Keeper.GetOauthClientRegistrationOnlyEId(ctx) {
		if !m.Keeper.isEIdActive(ctx, owner) {
			return nil, sdkerr.Wrap(err, "action requires an active eID")
		}
	}

	clientId := m.Keeper.autoIncrementClientId(ctx, types.KeyLastClientId)
	info := types.OauthClientInfo{
		ClientId: clientId,
		Owner:    msg.Owner,
		Domain:   msg.Domain,
		Public:   true,
	}
	m.Keeper.SetClientIdToInfo(ctx, clientId, info)
	m.Keeper.SetNewClientId(ctx, owner, clientId)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRegisterOauthClient,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Owner),
		sdk.NewAttribute(types.AttributeKeyOauthClientId, fmt.Sprint(clientId)),
	))

	return &types.MsgRegisterOAuthClientResponse{
		ClientId: clientId,
	}, nil
}

func (m msgServer) EditOAuthClient(goCtx context.Context, msg *types.MsgEditOAuthClient) (*types.MsgEditOAuthClientResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerr.Wrap(err, "owner")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeEditOauthClient,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Owner),
		sdk.NewAttribute(types.AttributeKeyOauthClientId, fmt.Sprint(msg.ClientId)),
	))

	info, err := m.Keeper.GetClientIdToInfo(ctx, msg.ClientId)
	if err != nil {
		return nil, sdkerr.Wrap(err, "invalid client id")
	}
	if info.Owner != msg.Owner {
		return nil, sdkerr.Wrap(err, "unauthorized")
	}

	info.Domain = msg.Domain
	err = m.Keeper.SetClientIdToInfo(ctx, msg.ClientId, *info)
	if err != nil {
		return nil, err
	}

	return &types.MsgEditOAuthClientResponse{}, nil
}

func (m msgServer) DeregisterOAuthClient(goCtx context.Context, msg *types.MsgDeregisterOAuthClient) (*types.MsgDeregisterOAuthClientResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerr.Wrap(err, "owner")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDeregisterOauthClient,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Owner),
		sdk.NewAttribute(types.AttributeKeyOauthClientId, fmt.Sprint(msg.ClientId)),
	))

	m.Keeper.DeleteClientIdFromOwner(ctx, owner, msg.ClientId)
	m.Keeper.DeleteClientIdToInfo(ctx, msg.ClientId)

	return &types.MsgDeregisterOAuthClientResponse{}, nil
}
