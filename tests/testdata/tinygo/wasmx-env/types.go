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

type CallResponse struct {
	Success int    `json:"success"`
	Data    string `json:"data"`
}

type CallResult struct {
	Success int    `json:"success"`
	Data    []byte `json:"data"`
}

type LoggerLog struct {
	Msg   string   `json:"msg"`
	Parts []string `json:"parts"`
}

// Common SDK types mirrored from AS

type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"` // base64 string
	Index bool   `json:"index"`
}

type Event struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

type StorageRangeReq struct {
	StartKey string `json:"start_key"`
	EndKey   string `json:"end_key"`
	Reverse  bool   `json:"reverse"`
}

type StorageDeleteRange struct {
	StartKey string `json:"start_key"`
	EndKey   string `json:"end_key"`
}

type StorageRange struct {
	StartKey string `json:"start_key"` // base64
	EndKey   string `json:"end_key"`   // base64
	Reverse  bool   `json:"reverse"`
}

type StoragePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type StoragePairs struct {
	Values []StoragePair `json:"values"`
}

type Account struct {
	Address       Bech32String `json:"address"`
	PubKey        string       `json:"pubKey"`
	AccountNumber int64        `json:"accountNumber"`
	Sequence      int64        `json:"sequence"`
}

type WasmxLog struct {
	Data   []byte
	Topics [][32]byte
}

// create accounts
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
type CreateAccountRequest struct {
	CodeID uint64 `json:"code_id"`
	Msg    string `json:"msg"`
	Funds  []Coin `json:"funds"`
	Label  string `json:"label"`
}
type CreateAccountResponse struct {
	Address Bech32String `json:"address"`
}

type Create2AccountRequest struct {
	CodeID uint64 `json:"code_id"`
	Msg    string `json:"msg"`
	Salt   string `json:"salt"`
	Funds  []Coin `json:"funds"`
	Label  string `json:"label"`
}
type Create2AccountResponse struct {
	Address Bech32String `json:"address"`
}

type BlockInfo struct {
	// block height this transaction is executed
	Height uint64 `json:"height"`
	// time in nanoseconds since unix epoch.
	Timestamp uint64       `json:"timestamp"`
	GasLimit  uint64       `json:"gasLimit"`
	Hash      []byte       `json:"hash"`
	Proposer  Bech32String `json:"proposer"`
}

type WasmxExecutionMessage struct {
	Data []byte `json:"data"`
}
