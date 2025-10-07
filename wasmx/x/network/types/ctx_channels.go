package types

import (
	"context"
	"fmt"
	"slices"
	sync "sync"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
)

type ContextKey string

const MultiChainContextKey ContextKey = "multichain-context"

type MultiChainContext struct {
	// chainId => channel
	resultChannels    map[string]*chan MsgExecuteAtomicTxResponse
	mtxResultChannels sync.Mutex

	// chainId => channel
	internalCallChannels    map[string]*chan MsgExecuteCrossChainCallRequestIndexed
	mtxInternalCallChannels sync.Mutex

	// chainId => channel
	internalCallResultChannels    map[string]*chan MsgExecuteCrossChainCallResponseIndexed
	mtxInternalCallResultChannels sync.Mutex

	ChainIds            []string
	CurrentAtomicTxHash []byte
	// chain ids set when executing the atomic tx
	CurrentAtomicTxChainIds   []string
	currentSubTxIndex         map[string]int
	mtxCurrentSubTxIndex      sync.Mutex
	currentInternalCrossTx    map[string]int
	mtxCurrentInternalCrossTx sync.Mutex

	// TODO if we support multichain, we will advance in lock step, 1 crosschain call at a time, for each chain, so we will have a common subtx & crosstx indexes
	// a chain cannot go more than 1 crosschaincall in advance to other chains - a lock step advancement of the atomic tx across all participating chains
	// chains wait for the crosschaincall index to be incremented before advancing through preprocessed crosschain calls
	// CurrentSubTxIndexCommon:          0,
	// CurrentInternalCrossTxCommon:     0,
}

func (mcctx *MultiChainContext) GetCurrentInternalCrossTx(chainId string) int {
	mcctx.mtxCurrentInternalCrossTx.Lock()
	defer mcctx.mtxCurrentInternalCrossTx.Unlock()
	index, ok := mcctx.currentInternalCrossTx[chainId]
	if !ok {
		mcctx.currentInternalCrossTx[chainId] = 0
		return 0
	}
	return index
}

func (mcctx *MultiChainContext) SetCurrentInternalCrossTx(chainId string, index int) {
	mcctx.mtxCurrentInternalCrossTx.Lock()
	defer mcctx.mtxCurrentInternalCrossTx.Unlock()
	mcctx.currentInternalCrossTx[chainId] = index
}

func (mcctx *MultiChainContext) GetCurrentSubTxIndex(chainId string) int {
	mcctx.mtxCurrentSubTxIndex.Lock()
	defer mcctx.mtxCurrentSubTxIndex.Unlock()
	index, ok := mcctx.currentSubTxIndex[chainId]
	if !ok {
		mcctx.currentSubTxIndex[chainId] = 0
		return 0
	}
	return index
}

func (mcctx *MultiChainContext) SetCurrentSubTxIndex(chainId string, index int) {
	mcctx.mtxCurrentSubTxIndex.Lock()
	defer mcctx.mtxCurrentSubTxIndex.Unlock()
	mcctx.currentSubTxIndex[chainId] = index
}

func (mcctx *MultiChainContext) GetResultChannel(chainId string) (*chan MsgExecuteAtomicTxResponse, error) {
	mcctx.mtxResultChannels.Lock()
	defer mcctx.mtxResultChannels.Unlock()
	mcchannel, ok := mcctx.resultChannels[chainId]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s", chainId)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetResultChannel(chainId string, value *chan MsgExecuteAtomicTxResponse) error {
	mcctx.mtxResultChannels.Lock()
	defer mcctx.mtxResultChannels.Unlock()
	mcctx.resultChannels[chainId] = value
	if !slices.Contains(mcctx.ChainIds, chainId) {
		mcctx.ChainIds = append(mcctx.ChainIds, chainId)
	}
	return nil
}

func (mcctx *MultiChainContext) GetInternalCallChannel(chainId string) (*chan MsgExecuteCrossChainCallRequestIndexed, error) {
	mcctx.mtxInternalCallChannels.Lock()
	defer mcctx.mtxInternalCallChannels.Unlock()
	mcchannel, ok := mcctx.internalCallChannels[chainId]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s", chainId)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetInternalCallChannel(chainId string, value *chan MsgExecuteCrossChainCallRequestIndexed) error {
	mcctx.mtxInternalCallChannels.Lock()
	defer mcctx.mtxInternalCallChannels.Unlock()
	mcctx.internalCallChannels[chainId] = value
	return nil
}

func (mcctx *MultiChainContext) GetInternalCallResultChannel(chainId string) (*chan MsgExecuteCrossChainCallResponseIndexed, error) {
	mcctx.mtxInternalCallResultChannels.Lock()
	defer mcctx.mtxInternalCallResultChannels.Unlock()
	mcchannel, ok := mcctx.internalCallResultChannels[chainId]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s", chainId)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetInternalCallResultChannel(chainId string, value *chan MsgExecuteCrossChainCallResponseIndexed) error {
	mcctx.mtxInternalCallResultChannels.Lock()
	defer mcctx.mtxInternalCallResultChannels.Unlock()
	mcctx.internalCallResultChannels[chainId] = value
	return nil
}

func (mcctx *MultiChainContext) CloseChannels() error {
	mcctx.mtxResultChannels.Lock()
	defer mcctx.mtxResultChannels.Unlock()

	mcctx.mtxInternalCallChannels.Lock()
	defer mcctx.mtxInternalCallChannels.Unlock()

	mcctx.mtxInternalCallResultChannels.Lock()
	defer mcctx.mtxInternalCallResultChannels.Unlock()

	for _, channel := range mcctx.resultChannels {
		close(*channel)
	}
	for _, channel := range mcctx.internalCallChannels {
		close(*channel)
	}
	for _, channel := range mcctx.internalCallResultChannels {
		close(*channel)
	}
	mcctx.resultChannels = map[string]*chan MsgExecuteAtomicTxResponse{}
	mcctx.internalCallChannels = map[string]*chan MsgExecuteCrossChainCallRequestIndexed{}
	mcctx.internalCallResultChannels = map[string]*chan MsgExecuteCrossChainCallResponseIndexed{}
	return nil
}

func ContextWithMultiChainContext(g *errgroup.Group, ctx context.Context, logger log.Logger) context.Context {
	mcctx := &MultiChainContext{
		resultChannels:             make(map[string]*chan MsgExecuteAtomicTxResponse, 0),
		internalCallChannels:       make(map[string]*chan MsgExecuteCrossChainCallRequestIndexed, 0),
		internalCallResultChannels: make(map[string]*chan MsgExecuteCrossChainCallResponseIndexed, 0),
		CurrentAtomicTxHash:        make([]byte, 0),
		currentSubTxIndex:          make(map[string]int, 0),
		currentInternalCrossTx:     make(map[string]int, 0),
	}
	ctx = context.WithValue(ctx, MultiChainContextKey, mcctx)
	// close channels when parent context closes
	g.Go(func() error {
		announcedCtxClose := false
		for range ctx.Done() {
			if !announcedCtxClose {
				logger.Info("closing multichain channels")
			}
			announcedCtxClose = true
			// wait for any atomic tx to finish before closing the channels
			if len(mcctx.CurrentAtomicTxHash) == 0 {
				return mcctx.CloseChannels()
			}
		}
		return nil
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
