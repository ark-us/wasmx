package types

import (
	bytes "bytes"
	"encoding/json"
	"math/big"

	mcodec "github.com/loredanacirstea/wasmx/codec"
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
	Address    mcodec.AccAddressPrefixed `json:"address"`
	CodeHash   []byte                    `json:"codeHash"`
	CodeId     uint64                    `json:"codeId"`
	SystemDeps []string                  `json:"deps"`
	// instantiate -> this is the constructor + runtime + constructor args
	// execute -> this is the runtime bytecode
	Bytecode []byte `json:"bytecode"`
}

type BlockInfo struct {
	// block height this transaction is executed
	Height uint64 `json:"height"`
	// time in nanoseconds since unix epoch. Uses string to ensure JavaScript compatibility.
	Timestamp uint64                    `json:"timestamp"`
	GasLimit  uint64                    `json:"gasLimit"`
	Hash      []byte                    `json:"hash"`
	Proposer  mcodec.AccAddressPrefixed `json:"proposer"`
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
	Origin mcodec.AccAddressPrefixed `json:"origin"`
	// Bech32 encoded sdk.AccAddress executing the contract
	Sender mcodec.AccAddressPrefixed `json:"sender"`
	// Amount of funds send to the contract along with this message
	Funds    *big.Int `json:"funds"`
	GasLimit *big.Int `json:"gasLimit"`
	CallData []byte   `json:"callData"`
}

type ContractDependency struct {
	Address       mcodec.AccAddressPrefixed
	Role          string
	Label         string
	StoreKey      []byte
	CodeFilePath  string
	AotFilePath   string
	SystemDeps    []SystemDep
	Bytecode      []byte
	CodeHash      []byte
	CodeId        uint64
	SystemDepsRaw []string
	StorageType   ContractStorageType
	Pinned        bool
	MeteringOff   bool
}

func (v ContractDependency) Clone() *ContractDependency {
	deps := make([]SystemDep, len(v.SystemDeps))
	for i, dep := range v.SystemDeps {
		deps[i] = dep.Clone()
	}
	return &ContractDependency{
		Address:       mcodec.NewAccAddressPrefixed(cloneBytes(v.Address.Bytes()), v.Address.Prefix()),
		Role:          v.Role,
		Label:         v.Label,
		StoreKey:      cloneBytes(v.StoreKey),
		CodeFilePath:  v.CodeFilePath,
		AotFilePath:   v.AotFilePath,
		SystemDeps:    deps,
		Bytecode:      cloneBytes(v.Bytecode),
		CodeHash:      cloneBytes(v.CodeHash),
		CodeId:        v.CodeId,
		SystemDepsRaw: cloneStrings(v.SystemDepsRaw),
		StorageType:   v.StorageType,
		Pinned:        v.Pinned,
	}
}

func (v *Env) Clone() *Env {
	return &Env{
		Chain:       v.Chain.Clone(),
		Block:       v.Block.Clone(),
		Transaction: v.Transaction.Clone(),
		Contract:    v.Contract.Clone(),
		CurrentCall: v.CurrentCall.Clone(),
	}
}

func (v MessageInfo) Clone() MessageInfo {
	return MessageInfo{
		Origin:   v.Origin,
		Sender:   v.Sender,
		Funds:    cloneBigInt(v.Funds),
		GasLimit: cloneBigInt(v.GasLimit),
		CallData: cloneBytes(v.CallData),
	}
}

func (v EnvContractInfo) Clone() EnvContractInfo {
	return EnvContractInfo{
		Address:    v.Address,
		CodeHash:   cloneBytes(v.CodeHash),
		CodeId:     v.CodeId,
		SystemDeps: cloneStrings(v.SystemDeps),
		Bytecode:   cloneBytes(v.Bytecode),
	}
}

func (v *TransactionInfo) Clone() *TransactionInfo {
	return &TransactionInfo{
		Index:    v.Index,
		GasPrice: cloneBigInt(v.GasPrice),
	}
}

func (v BlockInfo) Clone() BlockInfo {
	return BlockInfo{
		Height:    v.Height,
		Timestamp: v.Timestamp,
		GasLimit:  v.GasLimit,
		Hash:      cloneBytes(v.Hash),
		Proposer:  v.Proposer,
	}
}

func (v ChainInfo) Clone() ChainInfo {
	return ChainInfo{
		Denom:       v.Denom,
		ChainId:     cloneBigInt(v.ChainId),
		ChainIdFull: v.ChainIdFull,
	}
}

// func (u RawBytes) MarshalJSON() ([]byte, error) {
// 	var result string
// 	if u == nil {
// 		result = "null"
// 	} else {
// 		result = strings.Join(strings.Fields(fmt.Sprintf("%d", u)), ",")
// 	}
// 	return []byte(result), nil
// }

// func (m ChainInfo) MarshalJSON() ([]byte, error) {
// 	var chainId []byte = m.ChainId.FillBytes(make([]byte, 32))
// 	return json.Marshal(map[string]interface{}{
// 		"denom":       m.Denom,
// 		"chainId":     chainId,
// 		"chainIdFull": m.ChainIdFull,
// 	})
// }

// func (m *ChainInfo) UnmarshalJSON(data []byte) error {
// 	var value map[string]interface{}

// 	err := json.Unmarshal(data, &value)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("-ChainId-", value["chainId"])
// 	m.ChainId = big.NewInt(0).SetBytes(value["chainId"].([]byte))
// 	m.ChainIdFull = value["chainIdFull"].(string)
// 	m.Denom = value["denom"].(string)
// 	return nil
// }

// func (m BlockInfo) MarshalJSON() ([]byte, error) {
// 	var height []byte = big.NewInt(int64(m.Height)).FillBytes(make([]byte, 32))
// 	var timestamp []byte = big.NewInt(int64(m.Timestamp)).FillBytes(make([]byte, 32))
// 	var gasLimit []byte = big.NewInt(int64(m.GasLimit)).FillBytes(make([]byte, 32))
// 	var proposer []byte = PaddLeftTo32(m.Proposer.Bytes())
// 	return json.Marshal(map[string]interface{}{
// 		"height":    height,
// 		"timestamp": timestamp,
// 		"gasLimit":  gasLimit,
// 		"hash":      m.Hash,
// 		"proposer":  proposer,
// 	})
// }

// func (m TransactionInfo) MarshalJSON() ([]byte, error) {
// 	var gasPrice []byte = m.GasPrice.FillBytes(make([]byte, 32))
// 	return json.Marshal(map[string]interface{}{
// 		"index":    m.Index,
// 		"gasPrice": gasPrice,
// 	})
// }

// func (m EnvContractInfo) MarshalJSON() ([]byte, error) {
// 	var address []byte = PaddLeftTo32(m.Address.Bytes())
// 	return json.Marshal(map[string]interface{}{
// 		"address":  address,
// 		"codeHash": m.CodeHash,
// 		// "bytecode": m.Bytecode,
// 		"codeId": m.CodeId,
// 		"deps":   m.SystemDeps,
// 	})
// }

// func (m MessageInfo) MarshalJSON() ([]byte, error) {
// 	var origin []byte = PaddLeftTo32(m.Origin.Bytes())
// 	var sender []byte = PaddLeftTo32(m.Sender.Bytes())
// 	var funds []byte = m.Funds.FillBytes(make([]byte, 32))
// 	var gasLimit []byte = m.GasLimit.FillBytes(make([]byte, 32))
// 	return json.Marshal(map[string]interface{}{
// 		"origin":   origin,
// 		"sender":   sender,
// 		"funds":    funds,
// 		"gasLimit": gasLimit,
// 		"callData": m.CallData,
// 	})
// }

func PaddLeftTo32(data []byte) []byte {
	length := len(data)
	if length >= 32 {
		return data
	}
	data = append(bytes.Repeat([]byte{0}, 32-length), data...)
	return data
}

func cloneBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func cloneStrings(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func cloneBigInt(src *big.Int) *big.Int {
	return new(big.Int).Set(src)
}
