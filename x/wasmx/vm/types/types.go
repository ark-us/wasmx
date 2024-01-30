package types

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

type ContextI interface {
	GetCosmosHandler() types.WasmxCosmosHandler
}

// Internal call request
type CallRequest struct {
	To         sdk.AccAddress `json:"to"`
	From       sdk.AccAddress `json:"from"`
	Value      *big.Int       `json:"value"`
	GasLimit   *big.Int       `json:"gasLimit"`
	Calldata   []byte         `json:"calldata"`
	Bytecode   []byte         `json:"bytecode"`
	CodeHash   []byte         `json:"codeHash"`
	FilePath   string         `json:"filePath"`
	CodeId     uint64         `json:"codeId"`
	SystemDeps []string       `json:"systemDeps"`
	IsQuery    bool           `json:"isQuery"`
}

type CallRequestRaw struct {
	To       types.RawBytes `json:"to"`
	From     types.RawBytes `json:"from"`
	Value    types.RawBytes `json:"value"`
	GasLimit types.RawBytes `json:"gasLimit"`
	Calldata types.RawBytes `json:"calldata"`
	Bytecode types.RawBytes `json:"bytecode"`
	CodeHash types.RawBytes `json:"codeHash"`
	IsQuery  bool           `json:"isQuery"`
}

type SimpleCallRequestRaw struct {
	To       string   `json:"to"`
	Value    *big.Int `json:"value"`
	GasLimit *big.Int `json:"gasLimit"`
	Calldata []byte   `json:"calldata"`
	IsQuery  bool     `json:"isQuery"`
}

type CallResponse struct {
	Success uint8  `json:"success"`
	Data    []byte `json:"data"`
}

type CreateAccountInterpretedRequest struct {
	Bytecode []byte   `json:"bytecode"`
	Balance  *big.Int `json:"balance"`
}

type CreateAccountInterpretedRequestRaw struct {
	Bytecode types.RawBytes `json:"bytecode"`
	Balance  types.RawBytes `json:"balance"`
}

type Create2AccountInterpretedRequest struct {
	Bytecode []byte   `json:"bytecode"`
	Balance  *big.Int `json:"balance"`
	Salt     *big.Int `json:"salt"`
}

type Create2AccountInterpretedRequestRaw struct {
	Bytecode types.RawBytes `json:"bytecode"`
	Balance  types.RawBytes `json:"balance"`
	Salt     types.RawBytes `json:"salt"`
}

type InstantiateAccountRequest struct {
	CodeId uint64    `json:"code_id"`
	Msg    []byte    `json:"msg"`
	Funds  sdk.Coins `json:"funds"`
	Label  string    `json:"label"`
}

type InstantiateAccountResponse struct {
	Address sdk.AccAddress `json:"address"`
}

type Instantiate2AccountRequest struct {
	CodeId uint64    `json:"code_id"`
	Msg    []byte    `json:"msg"`
	Funds  sdk.Coins `json:"funds"`
	Label  string    `json:"label"`
	Salt   []byte    `json:"salt"`
}

type Instantiate2AccountResponse struct {
	Address sdk.AccAddress `json:"address"`
}

func (m CallRequest) MarshalJSON() ([]byte, error) {
	var to []byte = types.PaddLeftTo32(m.To.Bytes())
	var from []byte = types.PaddLeftTo32(m.From.Bytes())
	var value []byte = m.Value.FillBytes(make([]byte, 32))
	var gasLimit []byte = m.GasLimit.FillBytes(make([]byte, 32))
	return json.Marshal(map[string]interface{}{
		"to":       to,
		"from":     from,
		"value":    value,
		"gasLimit": gasLimit,
		"calldata": m.Calldata,
		"bytecode": m.Bytecode,
		"codeHash": m.CodeHash,
		"isQuery":  m.IsQuery,
	})
}

func (m *CallRequest) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var d CallRequestRaw
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	m.To = sdk.AccAddress(CleanupAddress(d.To))
	m.From = sdk.AccAddress(CleanupAddress(d.From))
	m.Value = big.NewInt(0).SetBytes(d.Value)
	m.GasLimit = big.NewInt(0).SetBytes(d.GasLimit)
	m.Calldata = d.Calldata
	m.Bytecode = d.Bytecode
	m.CodeHash = d.CodeHash
	m.IsQuery = d.IsQuery
	return nil
}

func (m *CreateAccountInterpretedRequest) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var d CreateAccountInterpretedRequestRaw
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	m.Balance = big.NewInt(0).SetBytes(d.Balance)
	m.Bytecode = d.Bytecode
	return nil
}

func (m *Create2AccountInterpretedRequest) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var d Create2AccountInterpretedRequestRaw
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	m.Balance = big.NewInt(0).SetBytes(d.Balance)
	m.Bytecode = d.Bytecode
	m.Salt = big.NewInt(0).SetBytes(d.Salt)
	return nil
}

func CleanupAddress(addr []byte) []byte {
	if IsEvmAddress(types.BytesToAddressCW(addr)) {
		return addr[12:]
	}
	return addr
}

func IsEvmAddress(addr types.AddressCW) bool {
	return hex.EncodeToString(addr.Bytes()[:12]) == hex.EncodeToString(make([]byte, 12))
}
