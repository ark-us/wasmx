package wasmx

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
)

type SimpleCallRequestRaw struct {
	To       string       `json:"to"`
	Value    *sdkmath.Int `json:"value"`
	GasLimit *big.Int     `json:"gasLimit"`
	Calldata []byte       `json:"calldata"`
	IsQuery  bool         `json:"isQuery"`
}

type CallResult struct {
	Success int    `json:"success"`
	Data    []byte `json:"data"`
}

type LoggerLog struct {
	Msg   string   `json:"msg"`
	Parts []string `json:"parts"`
}
