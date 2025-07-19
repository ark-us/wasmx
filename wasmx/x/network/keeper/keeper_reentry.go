package keeper

import (
	"context"

	"github.com/loredanacirstea/wasmx/x/network/types"
)

// TODO these should not be used through cosmos message router
// rewrite this for clarity; these are created to use action executor,
// to avoid race conditions
// but if used through cosmos router, we get race conditions at getting
// the signers from the messages
func (k *Keeper) Reentry(goCtx context.Context, msg *types.MsgReentry) (*types.MsgReentryResponse, error) {
	res, err := k.reentryInternal(goCtx, msg.Sender, msg.Contract, msg.Msg, msg.EntryPoint)
	if err != nil {
		return nil, err
	}
	return &types.MsgReentryResponse{Data: res.Data}, nil
}
