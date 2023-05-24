package types

import (
	bytes "bytes"
	"math/big"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

// DefaultMaxQueryStackSize maximum size of the stack of contract instances doing queries
const DefaultMaxQueryStackSize uint32 = 10

var PINNED_FOLDER = "pinned"

var EMPTY_BYTES32 = bytes.Repeat([]byte{0}, 32)

// MaxSaltSize is the longest salt that can be used when instantiating a contract
const MaxSaltSize = 64

var (
	// MaxLabelSize is the longest label that can be used when instantiating a contract
	MaxLabelSize = 128 // extension point for chains to customize via compile flag.

	// MaxWasmSize is the largest a compiled contract code can be when storing code on chain
	MaxWasmSize = 800 * 1024 // extension point for chains to customize via compile flag.

	// MaxProposalWasmSize is the largest a gov proposal compiled contract code can be when storing code on chain
	MaxProposalWasmSize = 3 * 1024 * 1024 // extension point for chains to customize via compile flag.

	// 0x6000 must be minimum, to support Ethereum contracts
	MaxInterpretedCodeSize = 0x6000
)

// Checksum represents a hash of the Wasm bytecode that serves as an ID. Must be generated from this library.
type Checksum []byte

// WasmCode is an alias for raw bytes of the wasm compiled code
type WasmCode []byte

// KVStore is a reference to some sub-kvstore that is valid for one instance of a code
type KVStore interface {
	Get(key []byte) []byte
	Set(key, value []byte)
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	Iterator(start, end []byte) dbm.Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	ReverseIterator(start, end []byte) dbm.Iterator
}

// Querier lets us make read-only queries on other modules
type Querier interface {
	Query(request QueryRequest, gasLimit uint64) ([]byte, error)
	GasConsumed() uint64
}

type QueryRequest struct{}

// GasMeter is a read-only version of the sdk gas meter
type Gas = uint64
type GasMeter interface {
	GasConsumed() Gas
	ConsumeGas(gas uint64, descriptor string)
}

type WasmxCosmosHandler interface {
	ContractStore(ctx sdk.Context, prefixStoreKey []byte) prefix.Store
	SubmitCosmosQuery(reqQuery abci.RequestQuery) ([]byte, error)
	ExecuteCosmosMsg(any *cdctypes.Any) ([]byte, error)
	GetBalance(addr sdk.AccAddress) *big.Int
	SendCoin(addr sdk.AccAddress, value *big.Int) error
	GetCodeHash(contractAddress sdk.AccAddress) Checksum
	GetBlockHash(blockNumber uint64) Checksum
	GetCodeInfo(addr sdk.AccAddress) CodeInfo
	Create(codeId uint64, creator sdk.AccAddress, initMsg []byte, label string, value *big.Int) (sdk.AccAddress, error)
	Create2(codeId uint64, creator sdk.AccAddress, initMsg []byte, salt Checksum, label string, value *big.Int) (sdk.AccAddress, error)
	GetContractDependency(ctx sdk.Context, addr sdk.AccAddress) (ContractDependency, error)
}

// LibWasmEdgeVersion returns the version of the loaded wasmedge library
// at runtime. This can be used for debugging to verify the loaded version
// matches the expected version.
//
// When cgo is disabled at build time, this returns an error at runtime.
func LibWasmEdgeVersion() string {
	return wasmedge.GetVersion()
}

// simplest wasmx version 1 interface
var WASMX_WASMX_1 = "wasmx_wasmx_1"

// wasmx version 2 with env information
var WASMX_WASMX_2 = "wasmx_wasmx_2"

// current ewasm interface
var EWASM_ENV_1 = "ewasm_env_1"

var SUPPORTED_HOST_INTERFACES = map[string]bool{
	WASMX_WASMX_1: true,
	WASMX_WASMX_2: true,
	EWASM_ENV_1:   true,
}

var INTERPRETER_EWASM_1 = "ewasm_ewasm_1" // outdated
var INTERPRETER_EVM_SHANGHAI = "interpreter_evm_shanghai"

var SUPPORTED_INTERPRETERS = map[string]bool{
	INTERPRETER_EVM_SHANGHAI: true,
}

func GetMaxCodeSize(sdeps []string) int {
	for _, dep := range sdeps {
		_, found := SUPPORTED_INTERPRETERS[dep]
		if found {
			return MaxInterpretedCodeSize
		}
	}
	return MaxWasmSize
}

func IsWasmDeps(sdeps []string) bool {
	for _, dep := range sdeps {
		_, found := SUPPORTED_INTERPRETERS[dep]
		if found {
			return false
		}
	}
	return true
}
