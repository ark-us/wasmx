package types

import (
	"context"

	mctx "wasmx/v1/context"
)

// these structures are used when building the meta info
// during the common block proposer's execution
const MetaInfoCrossChainKey = "crosschain_internal_tx"

func GetCrossChainCallMetaInfo(ctx context.Context, chainId string) *AtomicTxCrossChainCallInfo {
	data := &AtomicTxCrossChainCallInfo{}
	execInfo, err := mctx.GetExecutionMetaInfo(ctx)
	if err != nil {
		return nil
	}
	datai, found := execInfo.Data[MetaInfoCrossChainKey]
	if found {
		data = datai.(*AtomicTxCrossChainCallInfo)
	}
	return data
}

// we only add deterministic requests
// we should consider if we need to also add
func AddCrossChainCallMetaInfo(ctx context.Context, chainId string, req MsgExecuteCrossChainCallRequest, resp WrappedResponse) error {
	mcctx, err := GetMultiChainContext(ctx)
	if err != nil {
		return err
	}
	subtxIndex := mcctx.GetCurrentSubTxIndex(chainId)
	execInfo, err := mctx.GetExecutionMetaInfo(ctx)
	info := CrossChainCallInfo{Request: req, Response: resp}
	if err != nil {
		return err
	}
	data := &AtomicTxCrossChainCallInfo{}
	datai, found := execInfo.Data[MetaInfoCrossChainKey]
	if found {
		data = datai.(*AtomicTxCrossChainCallInfo)
	}
	for i := len(data.Subtx); i <= subtxIndex; i++ {
		data.Subtx = append(data.Subtx, SubTxCrossChainCallInfo{})
	}
	data.Subtx[subtxIndex].Calls = append(data.Subtx[subtxIndex].Calls, info)
	execInfo.Data[MetaInfoCrossChainKey] = data
	return nil
}
