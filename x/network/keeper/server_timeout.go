package keeper

import (
	"context"
	"fmt"
	"mythos/v1/x/network/types"
	"time"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the interval is for?
func (k *Keeper) StartTimeout(goCtx context.Context, msg *types.MsgStartTimeoutRequest) (*types.MsgStartTimeoutResponse, error) {
	k.goRoutineGroup.Go(func() error {
		err := k.startTimeoutInternalGoroutine(msg)
		if err != nil {
			k.actionExecutor.GetLogger().Error(err.Error())
		}
		return nil
	})
	return &types.MsgStartTimeoutResponse{}, nil
}

func (k *Keeper) startTimeoutInternalGoroutine(
	msg *types.MsgStartTimeoutRequest,
) error {
	select {
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start the delayed execution")
		return nil
	default:
		// continue
	}

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return sdkerr.Wrap(err, "ExecuteEth could not parse sender address")
	}
	description := fmt.Sprintf("timed action: delay %dms, args: %s ", msg.Delay, string(msg.Args))
	timeDelay := msg.Delay
	msgbz := msg.Args

	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error, 1)
	defer close(intervalEnded)
	defer close(errCh)
	go func() {
		k.actionExecutor.GetLogger().Info("eventual execution starting", "description", description)
		err := k.startTimeoutInternal(description, timeDelay, msgbz, contractAddress)
		if err != nil {
			k.actionExecutor.GetLogger().Error("eventual execution failed", "err", err)
			errCh <- err
		}
		k.actionExecutor.GetLogger().Info("eventual execution ended", "description", description)
		intervalEnded <- true
	}()

	select {
	case err := <-errCh:
		k.actionExecutor.GetLogger().Error("eventual execution failed to start", "error", err.Error())
		return err
	case <-intervalEnded:
		k.actionExecutor.GetLogger().Info("intervalEnded", "description", description)
		return nil
	}
}

func (k *Keeper) startTimeoutInternal(
	description string,
	timeDelay int64,
	msgbz []byte,
	contractAddress sdk.AccAddress,
) error {
	// sleep first and then load the context
	time.Sleep(time.Duration(timeDelay) * time.Millisecond)

	select {
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start the delayed execution")
		return nil
	default:
		// continue
	}

	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		msg := &types.MsgExecuteContract{
			Sender:   contractAddress.String(),
			Contract: contractAddress.String(),
			Msg:      msgbz,
		}
		res, err := k.ExecuteEventual(ctx, msg)
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
	_, err := k.actionExecutor.Execute(k.goContextParent, k.actionExecutor.GetApp().LastBlockHeight(), cb)

	return err
}
