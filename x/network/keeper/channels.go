package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"mythos/v1/x/network/types"
)

type ContextKey string

const MultiChainContextKey ContextKey = "multichain-context"

type MultiChainContext struct {
	// chainId | txhash => channel
	ResultChannels map[string]*chan types.MsgExecuteAtomicTxResponse

	// chainId | txhash | call index => channel
	InternalCallChannels map[string]*chan types.MsgExecuteCrossChainTxRequest

	// chainId | txhash | call index => channel
	InternalCallResultChannels map[string]*chan types.MsgExecuteCrossChainTxResponse
}

func (mcctx *MultiChainContext) GetResultChannel(chainId string, txhash []byte) (*chan types.MsgExecuteAtomicTxResponse, error) {
	key := GetChannelIdTx(chainId, txhash)
	mcchannel, ok := mcctx.ResultChannels[key]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s; txhash: %s", chainId, hex.EncodeToString(txhash))
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetResultChannel(chainId string, txhash []byte, value *chan types.MsgExecuteAtomicTxResponse) error {
	key := GetChannelIdTx(chainId, txhash)
	mcctx.ResultChannels[key] = value
	return nil
}

func (mcctx *MultiChainContext) GetInternalCallChannel(chainId string, txhash []byte, index int32) (*chan types.MsgExecuteCrossChainTxRequest, error) {
	key := GetChannelIdInternalCall(chainId, txhash, index)
	mcchannel, ok := mcctx.InternalCallChannels[key]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s; txhash: %s; index: %d", chainId, hex.EncodeToString(txhash), index)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetInternalCallChannel(chainId string, txhash []byte, index int32, value *chan types.MsgExecuteCrossChainTxRequest) error {
	key := GetChannelIdInternalCall(chainId, txhash, index)
	mcctx.InternalCallChannels[key] = value
	return nil
}

func (mcctx *MultiChainContext) GetInternalCallResultChannel(chainId string, txhash []byte, index int32) (*chan types.MsgExecuteCrossChainTxResponse, error) {
	key := GetChannelIdInternalCall(chainId, txhash, index)
	mcchannel, ok := mcctx.InternalCallResultChannels[key]
	if !ok {
		return nil, fmt.Errorf("channel not found for chain_id: %s; txhash: %s; index: %d", chainId, hex.EncodeToString(txhash), index)
	}
	return mcchannel, nil
}

func (mcctx *MultiChainContext) SetInternalCallResultChannel(chainId string, txhash []byte, index int32, value *chan types.MsgExecuteCrossChainTxResponse) error {
	key := GetChannelIdInternalCall(chainId, txhash, index)
	mcctx.InternalCallResultChannels[key] = value
	return nil
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

func GetChannelIdTx(chainId string, txhash []byte) string {
	return fmt.Sprintf("%s_%s", chainId, hex.EncodeToString(txhash))
}

func GetChannelIdInternalCall(chainId string, txhash []byte, index int32) string {
	return fmt.Sprintf("%s_%s_%d", chainId, hex.EncodeToString(txhash), index)
}
