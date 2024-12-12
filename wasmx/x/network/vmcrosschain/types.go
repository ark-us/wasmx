package vmcrosschain

import (
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const HOST_WASMX_ENV_CROSSCHAIN_VER1 = "wasmx_crosschain_1"

const HOST_WASMX_ENV_CROSSCHAIN_EXPORT = "wasmx_crosschain_"

const HOST_WASMX_ENV_CROSSCHAIN = "crosschain"

type Context struct {
	*vmtypes.Context
}

type MsgIsAtomicTxInExecutionRequest struct {
	SubChainId string `json:"sub_chain_id"`
	TxHash     []byte `json:"tx_hash"`
}

type MsgIsAtomicTxInExecutionResponse struct {
	IsInExecution bool `json:"is_in_execution"`
}
