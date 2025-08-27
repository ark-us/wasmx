package crosschain

import (
	wasmx "github.com/loredanacirstea/wasmx-env"
)

// Cross-chain types moved from wasmx-env

type MsgCrossChainCallRequest struct {
	From         string       `json:"from"`
	To           string       `json:"to"`
	Msg          []byte       `json:"msg"`
	Funds        []wasmx.Coin `json:"funds"`
	Dependencies []string     `json:"dependencies"`
	FromChainId  string       `json:"from_chain_id"`
	ToChainId    string       `json:"to_chain_id"`
	IsQuery      bool         `json:"is_query"`
	TimeoutMs    uint64       `json:"timeout_ms"`
}

type MsgCrossChainCallResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type MsgIsAtomicTxInExecutionRequest struct {
	SubChainId string `json:"sub_chain_id"`
	TxHash     []byte `json:"tx_hash"`
}

type MsgIsAtomicTxInExecutionResponse struct {
	IsInExecution bool `json:"is_in_execution"`
}
