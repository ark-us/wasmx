package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mctx "wasmx/v1/context"
	"wasmx/v1/x/network/types"
	wasmxtypes "wasmx/v1/x/wasmx/types"
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

// TODO make sure this is not be called from outside, only from wasmx
func (k *Keeper) CancelTimeout(goCtx context.Context, msg *types.MsgCancelTimeoutRequest) (*types.MsgCancelTimeoutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &types.MsgCancelTimeoutResponse{}

	timeoutKey := TimeoutKey(ctx.ChainID(), msg.Sender, msg.Id)
	cancelfn, err := mctx.GetTimeoutGoroutine(k.goContextParent, timeoutKey)
	if err != nil || cancelfn == nil {
		return resp, err
	}
	cancelfn()
	err = mctx.RemoveTimeoutGoroutine(k.goContextParent, timeoutKey)
	if err != nil {
		k.actionExecutor.GetLogger().Error("error removing goroutine", "error", err.Error())
	}
	return &types.MsgCancelTimeoutResponse{}, nil
}

func (k *Keeper) startTimeoutInternalGoroutine(
	msg *types.MsgStartTimeoutRequest,
	chainId string,
) error {
	description := fmt.Sprintf("chain_id=%s, delay %dms, contract %s, args: %s ", chainId, msg.Delay, msg.Contract, string(msg.Args))

	select {
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start the delayed execution", "description", description)
		return nil
	default:
		// continue
	}

	timeoutKey := TimeoutKey(chainId, msg.Sender, msg.Id)
	goctx, cancel := context.WithCancel(k.goContextParent)

	mctx.SetTimeoutGoroutine(k.goContextParent, timeoutKey, cancel)

	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error, 1)
	defer close(intervalEnded)
	defer close(errCh)
	go func() {
		k.actionExecutor.GetLogger().Debug("eventual execution triggered", "description", description)
		err := k.startTimeoutInternal(goctx, description, msg, chainId)
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
	goctx context.Context,
	description string,
	msg *types.MsgStartTimeoutRequest,
	chainId string,
) error {
	duration := time.Duration(msg.Delay) * time.Millisecond

	// either sleep action finishes first or the goroutine context is canceled or the parent context is finished (node is stopping)
	select {
	case <-goctx.Done():
		k.actionExecutor.GetLogger().Debug("delayed action was canceled, we do not start the timeout", "description", description)
		return nil
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start the delayed execution", "description", description)
		return nil
	case <-time.After(duration):
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
	_, err := k.actionExecutor.Execute(goctx, bapp.LastBlockHeight(), cb)
	if err != nil {
		return err
	}
	return nil
}

func TimeoutKey(chainId string, sender string, id string) string {
	return fmt.Sprintf("%s_%s_%s", chainId, sender, id)
}
