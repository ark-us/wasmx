package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/network/types"
)

func (k *Keeper) Reentry(goCtx context.Context, msg *types.MsgReentry) (*types.MsgReentryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.goRoutineGroup.Go(func() error {
		err := k.reentryInternalGoroutine(msg, ctx.ChainID())
		if err != nil {
			k.actionExecutor.GetLogger().Error(err.Error())
		}
		return nil
	})
	return &types.MsgReentryResponse{}, nil
}

func (k *Keeper) reentryInternalGoroutine(
	msg *types.MsgReentry,
	chainId string,
) error {
	select {
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start the reentry execution", "entry_point", msg.EntryPoint)
		return nil
	default:
		// continue
	}

	goctx, cancel := context.WithCancel(k.goContextParent)

	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error, 1)
	defer close(intervalEnded)
	defer close(errCh)
	go func() {
		k.actionExecutor.GetLogger().Debug("contract reentry triggered", "entry_point", msg.EntryPoint)
		err := k.reentryInternal(goctx, msg, chainId)
		if err != nil {
			k.actionExecutor.GetLogger().Error("reentry execution failed", "err", err, "entry_point", msg.EntryPoint)
			errCh <- err
		}
		k.actionExecutor.GetLogger().Debug("reentry execution ended", "entry_point", msg.EntryPoint)
		intervalEnded <- true
	}()

	select {
	case err := <-errCh:
		k.actionExecutor.GetLogger().Error("reentry execution failed", "error", err.Error())
		// cancel context
		cancel()
		return err
	case <-intervalEnded:
		cancel()
		return nil
	}
}

func (k *Keeper) reentryInternal(
	goctx context.Context,
	msg *types.MsgReentry,
	chainId string,
) error {
	k.actionExecutor.GetLogger().Debug("reentry started", "entry_point", msg.EntryPoint)

	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		execmsg := &types.MsgExecuteContract{
			Sender:   msg.Sender,
			Contract: msg.Contract,
			Msg:      msg.Msg,
		}
		res, err := k.ExecuteEntryPoint(ctx, msg.EntryPoint, execmsg)
		if err != nil {
			if err == types.ErrGoroutineClosed {
				k.actionExecutor.GetLogger().Error("Closing reentry thread", "entry_point", msg.EntryPoint, err.Error())
				return res, nil
			}
			k.actionExecutor.GetLogger().Error("reentry execution failed", "error", err.Error())
			return nil, err
		}
		return res, nil
	}
	// disregard result
	bapp := k.actionExecutor.GetBaseApp()
	_, err := k.actionExecutor.Execute(goctx, bapp.LastBlockHeight(), cb)
	if err != nil {
		return err
	}
	return nil
}
