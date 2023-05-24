package vm

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ChainInfoJson struct {
	Denom       string      `json:"denom"`
	ChainId     CustomBytes `json:"chainId"`
	ChainIdFull string      `json:"chainIdFull"`
}

type BlockInfoJson struct {
	Height   CustomBytes `json:"height"`
	Time     CustomBytes `json:"time"`
	GasLimit CustomBytes `json:"gasLimit"`
	Hash     CustomBytes `json:"hash"`
	Proposer CustomBytes `json:"proposer"`
}

type TransactionInfoJson struct {
	Index    int32       `json:"index"`
	GasPrice CustomBytes `json:"gasPrice"`
}

type AccountInfoJson struct {
	Address  CustomBytes `json:"address"`
	Balance  CustomBytes `json:"balance"`
	CodeHash CustomBytes `json:"codeHash"`
	Bytecode CustomBytes `json:"bytecode"`
}

type CurrentCallInfoJson struct {
	Origin   CustomBytes `json:"origin"`
	Sender   CustomBytes `json:"sender"`
	Funds    CustomBytes `json:"funds"`
	GasLimit CustomBytes `json:"gasLimit"`
	CallData CustomBytes `json:"callData"`
}

type EnvJson struct {
	Chain       ChainInfoJson       `json:"chain"`
	Block       BlockInfoJson       `json:"block"`
	Transaction TransactionInfoJson `json:"transaction"`
	Contract    AccountInfoJson     `json:"contract"`
	CurrentCall CurrentCallInfoJson `json:"currentCall"`
}

type CallRequestJson struct {
	To       CustomBytes `json:"to"`
	From     CustomBytes `json:"from"`
	Value    CustomBytes `json:"value"`
	GasLimit CustomBytes `json:"gasLimit"`
	Calldata CustomBytes `json:"calldata"`
	Bytecode CustomBytes `json:"bytecode"`
	CodeHash CustomBytes `json:"codeHash"`
	IsQuery  CustomBytes `json:"isQuery"`
}

type CallResponseJson struct {
	Success int32       `json:"success"`
	Data    CustomBytes `json:"data"`
}

type CustomBytes struct {
	Value []byte
}

func NewCustomBytes(v []byte) CustomBytes {
	return CustomBytes{Value: v}
}

func (m CustomBytes) MarshalJSON() ([]byte, error) {
	strs := []string{}
	for _, v := range m.Value {
		strs = append(strs, strconv.FormatInt(int64(v), 10))
	}
	str := "[" + strings.Join(strs, ",") + "]"
	return []byte(str), nil
}

func (m *CustomBytes) UnmarshalJSON(data []byte) error {
	// Customize the JSON unmarshaling logic
	var value []int32
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	m.Value = m.fromInt32(value)
	return nil
}

func (m *CustomBytes) fromInt32(data []int32) []byte {
	intArray := make([]byte, len(data))
	for i, val := range data {
		intArray[i] = byte(val)
	}
	return intArray
}

type CallRequest struct {
	To       sdk.AccAddress
	From     sdk.AccAddress
	Value    *big.Int
	GasLimit *big.Int
	Calldata []byte
	Bytecode []byte
	CodeHash []byte
	IsQuery  bool
}

func (v CallRequestJson) Transform() CallRequest {
	return CallRequest{
		To:       sdk.AccAddress(v.To.Value),
		From:     sdk.AccAddress(v.From.Value),
		Value:    big.NewInt(0).SetBytes(v.Value.Value),
		GasLimit: big.NewInt(0).SetBytes(v.GasLimit.Value),
		Calldata: v.Calldata.Value,
		Bytecode: v.Bytecode.Value,
		CodeHash: v.CodeHash.Value,
		IsQuery:  big.NewInt(0).SetBytes(v.IsQuery.Value).Int64() == 1,
	}
}
