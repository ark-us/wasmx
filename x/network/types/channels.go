package types

import (
	"context"
	"fmt"
	"slices"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
)

type ContextKey string

const MultiChainContextKey ContextKey = "multichain-context"

type MultiChainContext struct {
	// chainId => channel
	ResultChannels map[string]*chan MsgExecuteAtomicTxResponse

	// chainId => channel
	InternalCallChannels map[string]*chan MsgExecuteCrossChainCallRequestIndexed

	// chainId => channel
	InternalCallResultChannels map[string]*chan MsgExecuteCrossChainCallResponseIndexed

	ChainIds               []string
	CurrentAtomicTxHash    []byte
	CurrentSubTxIndex      int32
	CurrentInternalCrossTx int32
}

func (mcctx *MultiChainContext) GetResultChannel(chainId string) (*chan MsgExecuteAtomicTxResponse, error) {
	mcchannel, ok := mcctx.ResultChannels[chainId]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s", chainId)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetResultChannel(chainId string, value *chan MsgExecuteAtomicTxResponse) error {
	mcctx.ResultChannels[chainId] = value
	if !slices.Contains(mcctx.ChainIds, chainId) {
		mcctx.ChainIds = append(mcctx.ChainIds, chainId)
	}
	return nil
}

func (mcctx *MultiChainContext) GetInternalCallChannel(chainId string) (*chan MsgExecuteCrossChainCallRequestIndexed, error) {
	mcchannel, ok := mcctx.InternalCallChannels[chainId]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s", chainId)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetInternalCallChannel(chainId string, value *chan MsgExecuteCrossChainCallRequestIndexed) error {
	mcctx.InternalCallChannels[chainId] = value
	return nil
}

func (mcctx *MultiChainContext) GetInternalCallResultChannel(chainId string) (*chan MsgExecuteCrossChainCallResponseIndexed, error) {
	mcchannel, ok := mcctx.InternalCallResultChannels[chainId]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s", chainId)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetInternalCallResultChannel(chainId string, value *chan MsgExecuteCrossChainCallResponseIndexed) error {
	mcctx.InternalCallResultChannels[chainId] = value
	return nil
}

func (mcctx *MultiChainContext) CloseChannels() error {
	for _, channel := range mcctx.ResultChannels {
		close(*channel)
	}
	for _, channel := range mcctx.InternalCallChannels {
		close(*channel)
	}
	for _, channel := range mcctx.InternalCallResultChannels {
		close(*channel)
	}
	return nil
}

func ContextWithMultiChainContext(g *errgroup.Group, ctx context.Context, logger log.Logger) context.Context {
	mcctx := &MultiChainContext{
		ResultChannels:             make(map[string]*chan MsgExecuteAtomicTxResponse, 0),
		InternalCallChannels:       make(map[string]*chan MsgExecuteCrossChainCallRequestIndexed, 0),
		InternalCallResultChannels: make(map[string]*chan MsgExecuteCrossChainCallResponseIndexed, 0),
		CurrentAtomicTxHash:        make([]byte, 0),
		CurrentSubTxIndex:          0,
		CurrentInternalCrossTx:     0,
	}
	ctx = context.WithValue(ctx, MultiChainContextKey, mcctx)
	// close channels when parent context closes
	g.Go(func() error {
		<-ctx.Done()
		logger.Info("closing multichain channels")
		return mcctx.CloseChannels()
	})
	return ctx
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
