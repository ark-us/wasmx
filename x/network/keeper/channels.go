package keeper

import (
	"context"
	"fmt"

	"mythos/v1/x/network/types"
)

type ContextKey string

const MultiChainContextKey ContextKey = "multichain-context"

// chainId => channel
type MultiChainContext struct {
	ResultChannels map[string]*chan types.MsgExecuteAtomicTxResponse
}

func ContextWithMultiChainContext(ctx context.Context) context.Context {
	procc := &MultiChainContext{ResultChannels: make(map[string]*chan types.MsgExecuteAtomicTxResponse, 0)}
	return context.WithValue(ctx, MultiChainContextKey, procc)
}

func GetMultiChainContext(ctx context.Context) (*MultiChainContext, error) {
	mcctx_ := ctx.Value(MultiChainContextKey)
	if mcctx_ == nil {
		return nil, fmt.Errorf("multichain context not set on context")
	}
	mcctx := (mcctx_).(*MultiChainContext)
	if mcctx == nil {
		return nil, fmt.Errorf("multichain context not set on context")
	}
	return mcctx, nil
}

func GetMultiChainChannel(ctx context.Context, chainId string) (*chan types.MsgExecuteAtomicTxResponse, error) {
	mcctx, err := GetMultiChainContext(ctx)
	if err != nil {
		return nil, err
	}
	mcchannel, ok := mcctx.ResultChannels[chainId]
	if !ok {
		return nil, fmt.Errorf("channel not found for %s", chainId)
	}
	return mcchannel, nil
}

func AddMultiChainContext(ctx context.Context, chainId string, value chan types.MsgExecuteAtomicTxResponse) error {
	mcctx, err := GetMultiChainContext(ctx)
	if err != nil {
		return err
	}
	mcctx.ResultChannels[chainId] = &value
	return nil
}

func SetMultiChainContext(ctx context.Context, chainId string, value *chan types.MsgExecuteAtomicTxResponse) error {
	mcctx, err := GetMultiChainContext(ctx)
	if err != nil {
		return err
	}
	mcctx.ResultChannels[chainId] = value
	return nil
}
