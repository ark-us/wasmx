package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/network/types"
)

func (k *Keeper) Reentry(goCtx context.Context, msg *types.MsgReentry) (*types.MsgReentryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	execmsg := &types.MsgExecuteContract{
		Sender:   msg.Sender,
		Contract: msg.Contract,
		Msg:      msg.Msg,
	}
	res, err := k.ExecuteEntryPoint(ctx, msg.EntryPoint, execmsg)
	if err != nil {
		return nil, err
	}
	return &types.MsgReentryResponse{Data: res.Data}, nil
}
