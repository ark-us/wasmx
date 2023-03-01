package types

import (
	"encoding/json"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
)

// Coin is a string representation of the sdk.Coin type (more portable than sdk.Int)
type Coin struct {
	Denom  string `json:"denom"`  // type, eg. "ATOM"
	Amount string `json:"amount"` // string encoing of decimal value, eg. "12.3456"
}

func NewCoin(amount uint64, denom string) Coin {
	return Coin{
		Denom:  denom,
		Amount: strconv.FormatUint(amount, 10),
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
	HasIBCEntryPoints bool
	// Deprecated, use RequiredCapabilities. For now both fields contain the same value.
	RequiredFeatures     string
	RequiredCapabilities string
}

//---------- Env ---------

// Env defines the state of the blockchain environment this contract is
// running in. This must contain only trusted data - nothing from the Tx itself
// that has not been verfied (like Signer).
//
// Env are json encoded to a byte slice before passing to the wasm contract.
type Env struct {
	Block       BlockInfo        `json:"block"`
	Transaction *TransactionInfo `json:"transaction"`
	Contract    EnvContractInfo  `json:"contract"`
	Chain       ChainInfo        `json:"chain"`
}

type ChainInfo struct {
	Denom   string  `json:"denom"`
	ChainId big.Int `json:"chain_id"`
}

type EnvContractInfo struct {
	Address sdk.AccAddress
}

type BlockInfo struct {
	// block height this transaction is executed
	Height uint64 `json:"height"`
	// time in nanoseconds since unix epoch. Uses string to ensure JavaScript compatibility.
	Time     uint64         `json:"time,string"`
	ChainID  string         `json:"chain_id"`
	GasLimit uint64         `json:"gas_limit"`
	Hash     string         `json:"hash"`
	Proposer sdk.AccAddress `json:"proposer"`
}

type TransactionInfo struct {
	// Position of this transaction in the block.
	// The first transaction has index 0
	//
	// Along with BlockInfo.Height, this allows you to get a unique
	// transaction identifier for the chain for future queries
	Index    uint32 `json:"index"`
	GasPrice string `json:"gas_price"`
}

type MessageInfo struct {
	// Bech32 encoded sdk.AccAddress from which the calls originated
	Origin sdk.AccAddress `json:"origin"`
	// Bech32 encoded sdk.AccAddress executing the contract
	Sender sdk.AccAddress `json:"sender"`
	// Amount of funds send to the contract along with this message
	Funds        *big.Int     `json:"funds"`
	CallCacheMap CallCacheMap `json:"call_cache_map"`
	ReadOnly     bool         `json:"readonly"`
	IsQuery      bool         `json:"is_query"`
}

type CallCache struct {
	Index   uint32 `json:"index"`
	Success uint32 `json:"success"`
	Data    string `json:"data"`
}

type CallCacheMap = []CallCache

type ContractDependency struct {
	Address  sdk.AccAddress
	Store    types.KVStore
	FilePath string
}
