package crosschain

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	utils "github.com/loredanacirstea/wasmx-env-utils"
)

const loggerModule = "crosschain"

func ExecuteCrossChainTx(req MsgCrossChainCallRequest) (MsgCrossChainCallResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "ExecuteCrossChainTx", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(executeCrossChainTx_(utils.BytesToPackedPtr(bz)))
	var resp MsgCrossChainCallResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	return resp, nil
}

func ExecuteCrossChainQuery(req MsgCrossChainCallRequest) (MsgCrossChainCallResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "ExecuteCrossChainQuery", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(executeCrossChainQuery_(utils.BytesToPackedPtr(bz)))
	var resp MsgCrossChainCallResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	return resp, nil
}

func ExecuteCrossChainQueryNonDeterministic(req MsgCrossChainCallRequest) (MsgCrossChainCallResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "ExecuteCrossChainQueryNonDeterministic", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(executeCrossChainQueryNonDeterministic_(utils.BytesToPackedPtr(bz)))
	var resp MsgCrossChainCallResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	return resp, nil
}

func ExecuteCrossChainTxNonDeterministic(req MsgCrossChainCallRequest) (MsgCrossChainCallResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "ExecuteCrossChainTxNonDeterministic", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(executeCrossChainTxNonDeterministic_(utils.BytesToPackedPtr(bz)))
	var resp MsgCrossChainCallResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return MsgCrossChainCallResponse{}, err
	}
	return resp, nil
}

func IsAtomicTxInExecution(req MsgIsAtomicTxInExecutionRequest) (bool, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return false, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "IsAtomicTxInExecution", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(isAtomicTxInExecution_(utils.BytesToPackedPtr(bz)))
	var resp MsgIsAtomicTxInExecutionResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return false, err
	}
	return resp.IsInExecution, nil
}
