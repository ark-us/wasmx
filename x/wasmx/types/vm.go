package types

import (
	bytes "bytes"
	"math/big"
	"strings"

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
	MaxInterpretedCodeSize = 0xf000
)

var (
	ENTRY_POINT_INSTANTIATE = "instantiate"
	ENTRY_POINT_EXECUTE     = "execute"
	ENTRY_POINT_QUERY       = "query"
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
	Deploy(bytecode []byte, sender sdk.AccAddress, provenance sdk.AccAddress, initMsg []byte, value *big.Int, deps []string, metadata CodeMetadata, label string, salt []byte) (codeId uint64, checksum []byte, contractAddress sdk.AccAddress, err error)
	GetContractDependency(ctx sdk.Context, addr sdk.AccAddress) (ContractDependency, error)
	CanCallSystemContract(ctx sdk.Context, addr sdk.AccAddress) bool
}

// LibWasmEdgeVersion returns the version of the loaded wasmedge library
// at runtime. This can be used for debugging to verify the loaded version
// matches the expected version.
//
// When cgo is disabled at build time, this returns an error at runtime.
func LibWasmEdgeVersion() string {
	return wasmedge.GetVersion()
}

var EWASM_VM_EXPORT = "ewasm_env_"
var WASMX_VM_EXPORT = "wasmx_env_"
var SYS_VM_EXPORT = "sys_env_"
var CW_VM_EXPORT = "interface_version_"

// simplest wasmx version 1 interface
var WASMX_ENV_1 = "wasmx_env_1"

// wasmx version 2 with env information
var WASMX_ENV_2 = "wasmx_env_2"

// non-deterministic system operations, only as queries
var SYS_ENV_1 = "sys_env_1"

// initial interface use in precompiles 1 -> 9
// TODO replace & remove
var EWASM_ENV_0 = "ewasm_interface_version_1"

// current ewasm interface
var EWASM_ENV_1 = "ewasm_env_1"

// current cosmwasm interface
var CW_ENV_8 = "interface_version_8"

var SUPPORTED_HOST_INTERFACES = map[string]bool{
	WASMX_ENV_1: true,
	WASMX_ENV_2: true,
	EWASM_ENV_1: true,
	CW_ENV_8:    true,
}

var ROLE_INTERPRETER = "interpreter"
var ROLE_PRECOMPILE = "precompile"
var ROLE_ALIAS = "alias"

var INTERPRETER_EVM_SHANGHAI = "interpreter_evm_shanghai"

var TRUSTED_ADDRESS_LIMIT = big.NewInt(0).SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 128})

// var SUPPORTED_INTERPRETERS = map[string]bool{
// 	INTERPRETER_EVM_SHANGHAI: true,
// }

type SystemDep struct {
	Role     string
	Label    string
	FilePath string
	Deps     []SystemDep
}

func GetMaxCodeSize(sdeps []string) int {
	for _, dep := range sdeps {
		// _, found := SUPPORTED_INTERPRETERS[dep]
		isInterpreter := strings.Contains(dep, "interpreter")
		if isInterpreter {
			return MaxInterpretedCodeSize
		}
	}
	return MaxWasmSize
}

// func IsWasmDeps(sdeps []string) bool {
// 	for _, dep := range sdeps {
// 		_, found := SUPPORTED_INTERPRETERS[dep]
// 		if found {
// 			return false
// 		}
// 	}
// 	return true
// }
