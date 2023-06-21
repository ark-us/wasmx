package types

import (
	bytes "bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Coin is a string representation of the sdk.Coin type (more portable than sdk.Int)
type Coin struct {
	Denom  string   `json:"denom"`  // type, eg. "ATOM"
	Amount *big.Int `json:"amount"` // string encoing of decimal value, eg. "12.3456"
}

func NewCoin(amount uint64, denom string) Coin {
	return Coin{
		Denom:  denom,
		Amount: big.NewInt(int64(amount)),
	}
}

// Coins handles properly serializing empty amounts
type Coins []Coin

// MarshalJSON ensures that we get [] for empty arrays
func (c Coins) MarshalJSON() ([]byte, error) {
	if len(c) == 0 {
		return []byte("[]"), nil
	}
	var d []Coin = c
	return json.Marshal(d)
}

// UnmarshalJSON ensures that we get [] for empty arrays
func (c *Coins) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var d []Coin
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	*c = d
	return nil
}

type OutOfGasError struct{}

var _ error = OutOfGasError{}

func (o OutOfGasError) Error() string {
	return "Out of gas"
}

// Contains static analysis info of the contract (the Wasm code to be precise).
// This type is returned by VM.AnalyzeCode().
type AnalysisReport struct {
	Dependencies []string
}

//---------- Env ---------

type RawBytes []byte

// Env defines the state of the blockchain environment this contract is
// running in. This must contain only trusted data - nothing from the Tx itself
// that has not been verfied (like Signer).
//
// Env are json encoded to a byte slice before passing to the wasm contract.
type Env struct {
	Chain       ChainInfo        `json:"chain"`
	Block       BlockInfo        `json:"block"`
	Transaction *TransactionInfo `json:"transaction"`
	Contract    EnvContractInfo  `json:"contract"`
	CurrentCall MessageInfo      `json:"currentCall"`
}

type ChainInfo struct {
	Denom       string   `json:"denom"`
	ChainId     *big.Int `json:"chainId"`
	ChainIdFull string   `json:"chainIdFull"`
}

type EnvContractInfo struct {
	Address  sdk.AccAddress `json:"address"`
	CodeHash RawBytes       `json:"codeHash"`
	// instantiate -> this is the constructor + runtime + constructor args
	// execute -> this is the runtime bytecode
	Bytecode RawBytes `json:"bytecode"`
}

type BlockInfo struct {
	// block height this transaction is executed
	Height uint64 `json:"height"`
	// time in nanoseconds since unix epoch. Uses string to ensure JavaScript compatibility.
	Timestamp uint64         `json:"timestamp"`
	GasLimit  uint64         `json:"gasLimit"`
	Hash      RawBytes       `json:"hash"`
	Proposer  sdk.AccAddress `json:"proposer"`
}

type TransactionInfo struct {
	// Position of this transaction in the block.
	// The first transaction has index 0
	//
	// Along with BlockInfo.Height, this allows you to get a unique
	// transaction identifier for the chain for future queries
	Index    uint32   `json:"index"`
	GasPrice *big.Int `json:"gasPrice"`
}

type MessageInfo struct {
	// Bech32 encoded sdk.AccAddress from which the calls originated
	Origin sdk.AccAddress `json:"origin"`
	// Bech32 encoded sdk.AccAddress executing the contract
	Sender sdk.AccAddress `json:"sender"`
	// Amount of funds send to the contract along with this message
	Funds    *big.Int `json:"funds"`
	GasLimit *big.Int `json:"gasLimit"`
	CallData RawBytes `json:"callData"`
}

type ContractDependency struct {
	Address    sdk.AccAddress
	Role       string
	Label      string
	StoreKey   []byte
	FilePath   string
	SystemDeps []SystemDep
	Bytecode   []byte
	CodeHash   []byte
}

func (u RawBytes) MarshalJSON() ([]byte, error) {
	var result string
	if u == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", u)), ",")
	}
	return []byte(result), nil
}

func (m ChainInfo) MarshalJSON() ([]byte, error) {
	var chainId RawBytes = m.ChainId.FillBytes(make([]byte, 32))
	return json.Marshal(map[string]interface{}{
		"denom":       m.Denom,
		"chainId":     chainId,
		"chainIdFull": m.ChainIdFull,
	})
}

func (m BlockInfo) MarshalJSON() ([]byte, error) {
	var height RawBytes = big.NewInt(int64(m.Height)).FillBytes(make([]byte, 32))
	var timestamp RawBytes = big.NewInt(int64(m.Timestamp)).FillBytes(make([]byte, 32))
	var gasLimit RawBytes = big.NewInt(int64(m.GasLimit)).FillBytes(make([]byte, 32))
	var proposer RawBytes = PaddLeftTo32(m.Proposer.Bytes())
	return json.Marshal(map[string]interface{}{
		"height":    height,
		"timestamp": timestamp,
		"gasLimit":  gasLimit,
		"hash":      m.Hash,
		"proposer":  proposer,
	})
}

func (m TransactionInfo) MarshalJSON() ([]byte, error) {
	var gasPrice RawBytes = m.GasPrice.FillBytes(make([]byte, 32))
	return json.Marshal(map[string]interface{}{
		"index":    m.Index,
		"gasPrice": gasPrice,
	})
}

func (m EnvContractInfo) MarshalJSON() ([]byte, error) {
	var address RawBytes = PaddLeftTo32(m.Address.Bytes())
	return json.Marshal(map[string]interface{}{
		"address":  address,
		"codeHash": m.CodeHash,
		"bytecode": m.Bytecode,
	})
}

func (m MessageInfo) MarshalJSON() ([]byte, error) {
	var origin RawBytes = PaddLeftTo32(m.Origin.Bytes())
	var sender RawBytes = PaddLeftTo32(m.Sender.Bytes())
	var funds RawBytes = m.Funds.FillBytes(make([]byte, 32))
	var gasLimit RawBytes = m.GasLimit.FillBytes(make([]byte, 32))
	return json.Marshal(map[string]interface{}{
		"origin":   origin,
		"sender":   sender,
		"funds":    funds,
		"gasLimit": gasLimit,
		"callData": m.CallData,
	})
}

func PaddLeftTo32(data []byte) []byte {
	length := len(data)
	if length >= 32 {
		return data
	}
	data = append(bytes.Repeat([]byte{0}, 32-length), data...)
	return data
}
