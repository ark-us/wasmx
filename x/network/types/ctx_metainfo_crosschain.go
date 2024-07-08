package types

import (
	"context"

	mctx "mythos/v1/context"
)

const MetaInfoCrossChainKey = "crosschain_internal_tx"

type WrappedResponse struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

type CrossChainCallInfo struct {
	Request MsgExecuteCrossChainCallRequest
	Result  WrappedResponse
}

type MetaInfoCrossChain struct {
	Requests [][]CrossChainCallInfo
}

func SetCrossChainCallMetaInfoEmpty(ctx context.Context) error {
	data := &MetaInfoCrossChain{}
	execInfo, err := mctx.GetExecutionMetaInfo(ctx)
	if err != nil {
		return err
	}
	execInfo.Data[MetaInfoCrossChainKey] = data
	return nil
}

func GetCrossChainCallMetaInfo(ctx context.Context) *MetaInfoCrossChain {
	data := &MetaInfoCrossChain{}
	execInfo, err := mctx.GetExecutionMetaInfo(ctx)
	if err != nil {
		return data
	}
	datai, found := execInfo.Data[MetaInfoCrossChainKey]
	if found {
		data = datai.(*MetaInfoCrossChain)
	}
	return data
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
	info := CrossChainCallInfo{Request: req, Result: resp}
	if err != nil {
		return err
	}
	data := &MetaInfoCrossChain{}
	datai, found := execInfo.Data[MetaInfoCrossChainKey]
	if found {
		data = datai.(*MetaInfoCrossChain)
	}
	for i := len(data.Requests); i <= subtxIndex; i++ {
		data.Requests = append(data.Requests, []CrossChainCallInfo{})
	}
	data.Requests[subtxIndex] = append(data.Requests[subtxIndex], info)
	execInfo.Data[MetaInfoCrossChainKey] = data
	return nil
}
