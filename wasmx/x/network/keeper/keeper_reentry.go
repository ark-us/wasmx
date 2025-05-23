package keeper

import (
	"context"

	"github.com/loredanacirstea/wasmx/x/network/types"
)

func (k *Keeper) Reentry(goCtx context.Context, msg *types.MsgReentry) (*types.MsgReentryResponse, error) {
	res, err := k.reentryInternal(goCtx, msg.Sender, msg.Contract, msg.Msg, msg.EntryPoint)
	if err != nil {
		return nil, err
	}
	return &types.MsgReentryResponse{Data: res.Data}, nil
}
