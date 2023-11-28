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
	Calldata   types.RawBytes `json:"calldata"`
	Bytecode   types.RawBytes `json:"bytecode"`
	CodeHash   types.RawBytes `json:"codeHash"`
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
	To       sdk.AccAddress `json:"to"`
	Value    *big.Int       `json:"value"`
	GasLimit *big.Int       `json:"gasLimit"`
	Calldata types.RawBytes `json:"calldata"`
	IsQuery  bool           `json:"isQuery"`
}

type CallResponse struct {
	Success uint8          `json:"success"`
	Data    types.RawBytes `json:"data"`
}

type CreateAccountRequest struct {
	Bytecode types.RawBytes `json:"bytecode"`
	Balance  *big.Int       `json:"balance"`
}

type CreateAccountRequestRaw struct {
	Bytecode types.RawBytes `json:"bytecode"`
	Balance  types.RawBytes `json:"balance"`
}

type Create2AccountRequest struct {
	Bytecode types.RawBytes `json:"bytecode"`
	Balance  *big.Int       `json:"balance"`
	Salt     *big.Int       `json:"salt"`
}

type Create2AccountRequestRaw struct {
	Bytecode types.RawBytes `json:"bytecode"`
	Balance  types.RawBytes `json:"balance"`
	Salt     types.RawBytes `json:"salt"`
}

func (m CallRequest) MarshalJSON() ([]byte, error) {
	var to types.RawBytes = types.PaddLeftTo32(m.To.Bytes())
	var from types.RawBytes = types.PaddLeftTo32(m.From.Bytes())
	var value types.RawBytes = m.Value.FillBytes(make([]byte, 32))
	var gasLimit types.RawBytes = m.GasLimit.FillBytes(make([]byte, 32))
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

func (m *CreateAccountRequest) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var d CreateAccountRequestRaw
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	m.Balance = big.NewInt(0).SetBytes(d.Balance)
	m.Bytecode = d.Bytecode
	return nil
}

func (m *Create2AccountRequest) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var d Create2AccountRequestRaw
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
