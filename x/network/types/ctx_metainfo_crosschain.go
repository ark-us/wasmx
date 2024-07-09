package types

import (
	"context"

	mctx "mythos/v1/context"
)

const MetaInfoCrossChainKey = "crosschain_internal_tx"
const MetaInfoCrossChainNextIndexKey = "crosschain_internal_tx_index"

func SetCrossChainCallMetaInfoNextIndex(ctx context.Context, index int32) error {
	execInfo, err := mctx.GetExecutionMetaInfo(ctx)
	if err != nil {
		return err
	}
	execInfo.TempData[MetaInfoCrossChainNextIndexKey] = index
	return nil
}

func GetCrossChainCallMetaInfoNextIndex(ctx context.Context) int32 {
	execInfo, err := mctx.GetExecutionMetaInfo(ctx)
	if err != nil {
		return 0
	}
	return execInfo.TempData[MetaInfoCrossChainNextIndexKey].(int32)
}

func GetCrossChainCallMetaInfo(ctx context.Context) (*AtomicTxCrossChainCallInfo, int32) {
	data := &AtomicTxCrossChainCallInfo{}
	execInfo, err := mctx.GetExecutionMetaInfo(ctx)
	if err != nil {
		return data, 0
	}
	datai, found := execInfo.Data[MetaInfoCrossChainKey]
	if found {
		data = datai.(*AtomicTxCrossChainCallInfo)
	}
	index := execInfo.TempData[MetaInfoCrossChainNextIndexKey].(int32)
	return data, index
}

// we only add deterministic requests
// we should consider if we need to also add
func AddCrossChainCallMetaInfo(ctx context.Context, req MsgExecuteCrossChainCallRequest, resp WrappedResponse) error {
	mcctx, err := GetMultiChainContext(ctx)
	if err != nil {
		return err
	}
	subtxIndex := int(mcctx.CurrentSubTxIndex)
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
