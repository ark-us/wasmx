package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the interval is for?
func (k *Keeper) StartTimeout(goCtx context.Context, msg *types.MsgStartTimeoutRequest) (*types.MsgStartTimeoutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.goRoutineGroup.Go(func() error {
		err := k.startTimeoutInternalGoroutine(msg, ctx.ChainID())
		if err != nil {
			k.actionExecutor.GetLogger().Error(err.Error())
		}
		return nil
	})
	return &types.MsgStartTimeoutResponse{}, nil
}

func (k *Keeper) startTimeoutInternalGoroutine(
	msg *types.MsgStartTimeoutRequest,
	chainId string,
) error {
	select {
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start the delayed execution")
		return nil
	default:
		// continue
	}

	description := fmt.Sprintf("chain_id=%s, delay %dms, contract %s, args: %s ", chainId, msg.Delay, msg.Contract, string(msg.Args))

	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error, 1)
	defer close(intervalEnded)
	defer close(errCh)
	go func() {
		k.actionExecutor.GetLogger().Debug("eventual execution triggered", "description", description)
		err := k.startTimeoutInternal(description, msg, chainId)
		if err != nil {
			k.actionExecutor.GetLogger().Error("eventual execution failed", "err", err, "description", description)
			errCh <- err
		}
		k.actionExecutor.GetLogger().Debug("eventual execution ended", "description", description)
		intervalEnded <- true
	}()

	select {
	case err := <-errCh:
		k.actionExecutor.GetLogger().Error("eventual execution failed to start", "error", err.Error())
		return err
	case <-intervalEnded:
		return nil
	}
}

func (k *Keeper) startTimeoutInternal(
	description string,
	msg *types.MsgStartTimeoutRequest,
	chainId string,
) error {
	// sleep first and then load the context
	time.Sleep(time.Duration(msg.Delay) * time.Millisecond)

	select {
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start the delayed execution")
		return nil
	default:
		// continue
	}
	k.actionExecutor.GetLogger().Debug("eventual execution started", "description", description)

	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		msg := &types.MsgExecuteContract{
			Sender:   msg.Sender,
			Contract: msg.Contract,
			Msg:      msg.Args,
		}
		res, err := k.ExecuteEntryPoint(ctx, wasmxtypes.ENTRY_POINT_TIMED, msg)
		if err != nil {
			if err == types.ErrGoroutineClosed {
				k.actionExecutor.GetLogger().Error("Closing eventual thread", "description", description, err.Error())
				return res, nil
			}
			k.actionExecutor.GetLogger().Error("eventual execution failed", "error", err.Error())
			return nil, err
		}
		return res, nil
	}
	// disregard result
	bapp := k.actionExecutor.GetBaseApp()
	_, err := k.actionExecutor.Execute(k.goContextParent, bapp.LastBlockHeight(), cb)
	if err != nil {
		return err
	}
	return nil
}
