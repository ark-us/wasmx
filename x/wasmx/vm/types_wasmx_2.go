package vm

import (
	"encoding/json"
	"strconv"
	"strings"
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

type ContractInfoJson struct {
	Address  CustomBytes `json:"address"`
	Bytecode CustomBytes `json:"bytecode"`
	Balance  CustomBytes `json:"balance"`
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
	Contract    ContractInfoJson    `json:"contract"`
	CurrentCall CurrentCallInfoJson `json:"currentCall"`
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
