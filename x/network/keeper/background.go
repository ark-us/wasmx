package keeper

import (
	"context"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "mythos/v1/codec"
	"mythos/v1/x/network/types"
)

// internal method
func (k *Keeper) StartBackgroundProcess(goCtx context.Context, msg *types.MsgStartBackgroundProcessRequest) (*types.MsgStartBackgroundProcessResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.goRoutineGroup.Go(func() error {
		err := k.startBackgroundProcessInternalGoroutine(ctx, msg)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		}
		return nil
	})
	return &types.MsgStartBackgroundProcessResponse{}, nil
}

func (k *Keeper) startBackgroundProcessInternalGoroutine(
	ctx sdk.Context,
	msg *types.MsgStartBackgroundProcessRequest,
) error {
	senderAddr, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Sender)
	if err != nil {
		return sdkerr.Wrap(err, "sender")
	}
	contractAddr, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Contract)
	if err != nil {
		return sdkerr.Wrap(err, "contract")
	}

	description := fmt.Sprintf("background process: chain_id %s, contract %s, args: %s ", ctx.ChainID(), msg.Contract, string(msg.Args))

	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error, 1)
	defer close(intervalEnded)
	defer close(errCh)
	go func() {
		k.Logger(ctx).Info("background process started", "description", description)
		err := k.startBackgroundProcessInternal(description, contractAddr, senderAddr, msg.Args)
		if err != nil {
			k.Logger(ctx).Error("background process failed", "description", description, "err", err)
			errCh <- err
		}
		k.Logger(ctx).Info("background process ended", "description", description)
		intervalEnded <- true
	}()

	select {
	case <-k.goContextParent.Done():
		k.Logger(ctx).Info("stopping background process: parent context closing", "description", description)
		return nil
	case err := <-errCh:
		k.Logger(ctx).Error("background process failed to start", "description", description, "error", err.Error())
		return err
	case <-intervalEnded:
		return nil
	}
}

func (k *Keeper) startBackgroundProcessInternal(
	description string,
	contractAddr mcodec.AccAddressPrefixed,
	senderAddr mcodec.AccAddressPrefixed,
	msgbz []byte,
) error {
	goCtx := k.goContextParent
	if goCtx == nil {
		return fmt.Errorf("goContextParent not set for background processes")
	}
	// we cannot use the ActionExecutor, because it will block all other executions (lock)
	// we need to start this process in its own goroutine
	mythosapp := k.actionExecutor.GetApp()
	height := mythosapp.GetBaseApp().LastBlockHeight()
	sdkCtx, commitCacheCtx, ctxcachems, err := CreateQueryContext(mythosapp.GetBaseApp(), k.actionExecutor.logger, height, false)
	if err != nil {
		return err
	}

	goCtx = context.WithValue(goCtx, sdk.SdkContextKey, sdkCtx)
	ctx_ := sdk.UnwrapSDKContext(goCtx)
	_, err = k.wasmxKeeper.Execute(ctx_, contractAddr, senderAddr, msgbz, nil, nil, true)
	// we only commit if callback was successful
	err = commitCtx(mythosapp, sdkCtx, commitCacheCtx, ctxcachems)
	if err != nil {
		return err
	}
	return nil
}
