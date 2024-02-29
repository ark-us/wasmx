package keeper

import (
	"context"
	"mythos/v1/x/network/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) P2PReceiveMessage(goCtx context.Context, msg *types.MsgP2PReceiveMessageRequest) (*types.MsgP2PReceiveMessageResponse, error) {
	k.goRoutineGroup.Go(func() error {
		err := k.p2pReceiveMessageInternalGoroutine(msg)
		if err != nil {
			k.actionExecutor.GetLogger().Error(err.Error())
		}
		return nil
	})
	return &types.MsgP2PReceiveMessageResponse{}, nil
}

func (k *Keeper) p2pReceiveMessageInternalGoroutine(
	msg *types.MsgP2PReceiveMessageRequest,
) error {
	select {
	case <-k.goContextParent.Done():
		k.actionExecutor.GetLogger().Info("parent context was closed, we do not start p2p message receival execution")
		return nil
	default:
		// continue
	}

	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error, 1)
	defer close(intervalEnded)
	defer close(errCh)
	go func() {
		k.actionExecutor.GetLogger().Debug("p2p message receival started")
		err := k.p2pReceiveMessageInternal(msg)
		if err != nil {
			k.actionExecutor.GetLogger().Error("p2p message receival failed", "err", err)
			errCh <- err
		}
		k.actionExecutor.GetLogger().Debug("p2p message receival ended")
		intervalEnded <- true
	}()

	select {
	case err := <-errCh:
		k.actionExecutor.GetLogger().Error("p2p message receival failed to start", "error", err.Error())
		return err
	case <-intervalEnded:
		return nil
	}
}

func (k *Keeper) p2pReceiveMessageInternal(msg *types.MsgP2PReceiveMessageRequest) error {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		msg := &types.MsgExecuteContract{
			Sender:   msg.Sender,
			Contract: msg.Contract,
			Msg:      msg.Data,
		}
		// res, err := k.ExecuteEntryPoint(ctx, wasmxtypes.ENTRY_POINT_P2P_MSG, msg)
		res, err := k.ExecuteContract(ctx, msg)
		if err != nil {
			if err == types.ErrGoroutineClosed {
				k.actionExecutor.GetLogger().Error("closing p2p message receival thread", err.Error())
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
