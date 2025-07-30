package wasmx

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
)

type Bech32String string

type SimpleCallRequestRaw struct {
	To       Bech32String `json:"to"`
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
