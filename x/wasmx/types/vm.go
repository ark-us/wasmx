package types

import (
	bytes "bytes"
	"fmt"
	"math/big"
	"strings"

	address "cosmossdk.io/core/address"
	"cosmossdk.io/store/prefix"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	mcodec "mythos/v1/codec"
	cw8types "mythos/v1/x/wasmx/cw8/types"
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
	MaxWasmSize = 1500 * 1024 // extension point for chains to customize via compile flag.

	// MaxProposalWasmSize is the largest a gov proposal compiled contract code can be when storing code on chain
	MaxProposalWasmSize = 3 * 1024 * 1024 // extension point for chains to customize via compile flag.

	// 0x6000 must be minimum, to support Ethereum contracts
	MaxInterpretedCodeSize = 0xf000
)

var (
	ENTRY_POINT_INSTANTIATE = "instantiate"
	ENTRY_POINT_EXECUTE     = "execute"
	ENTRY_POINT_QUERY       = "query"
	ENTRY_POINT_REPLY       = "reply"
	ENTRY_POINT_TIMED       = "eventual"
	ENTRY_POINT_P2P_MSG     = "p2pmsg"
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
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	ReverseIterator(start, end []byte) Iterator
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
	ContractStore(ctx sdk.Context, storageType ContractStorageType, prefixStoreKey []byte) prefix.Store
	SubmitCosmosQuery(reqQuery *abci.RequestQuery) ([]byte, error)
	ExecuteCosmosMsgAny(any *cdctypes.Any) ([]sdk.Event, []byte, error)
	ExecuteCosmosMsgAnyBz(msgbz []byte) ([]sdk.Event, []byte, error)
	ExecuteCosmosMsg(msg sdk.Msg) ([]sdk.Event, []byte, error)
	DecodeCosmosTx(bz []byte) ([]byte, error)
	AnyToBz(anyMsg *cdctypes.Any) ([]byte, error)
	VerifyCosmosTx(bz []byte) (bool, error)
	WasmVMQueryHandler(caller mcodec.AccAddressPrefixed, request cw8types.QueryRequest) ([]byte, error)
	GetAccount(addr mcodec.AccAddressPrefixed) (mcodec.AccountI, error)
	GetCodeHash(contractAddress sdk.AccAddress) Checksum
	GetCode(contractAddress sdk.AccAddress) []byte
	GetBlockHash(blockNumber uint64) Checksum
	GetCodeInfo(addr sdk.AccAddress) CodeInfo
	GetContractInstance(contractAddress sdk.AccAddress) (ContractInfo, CodeInfo, []byte, error)
	Create(codeId uint64, creator mcodec.AccAddressPrefixed, initMsg []byte, label string, value *big.Int, funds sdk.Coins) (*mcodec.AccAddressPrefixed, error)
	Create2(codeId uint64, creator mcodec.AccAddressPrefixed, initMsg []byte, salt Checksum, label string, value *big.Int, funds sdk.Coins) (*mcodec.AccAddressPrefixed, error)
	Deploy(bytecode []byte, sender *mcodec.AccAddressPrefixed, provenance *mcodec.AccAddressPrefixed, initMsg []byte, value *big.Int, deps []string, metadata CodeMetadata, label string, salt []byte) (codeId uint64, checksum []byte, contractAddress mcodec.AccAddressPrefixed, err error)
	Execute(contractAddress mcodec.AccAddressPrefixed, sender mcodec.AccAddressPrefixed, execmsg []byte, value *big.Int, deps []string) (res []byte, err error)
	GetContractDependency(ctx sdk.Context, addr sdk.AccAddress) (ContractDependency, error)
	CanCallSystemContract(ctx sdk.Context, addr sdk.AccAddress) bool
	WithNewAddress(addr mcodec.AccAddressPrefixed) WasmxCosmosHandler
	GetAddressOrRole(ctx sdk.Context, addressOrRole string) (mcodec.AccAddressPrefixed, error)
	GetRoleByContractAddress(ctx sdk.Context, addr sdk.AccAddress) string
	JSONCodec() codec.JSONCodec
	GetAlias(addr mcodec.AccAddressPrefixed) (mcodec.AccAddressPrefixed, bool)
	Codec() codec.Codec
	AddressCodec() address.Codec
	ValidatorAddressCodec() address.Codec
	ConsensusAddressCodec() address.Codec
	AccBech32Codec() mcodec.AccBech32Codec
	TxConfig() client.TxConfig
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
var WASMX_CONS_VM_EXPORT = "wasmx_consensus_json_"
var SYS_VM_EXPORT = "sys_env_"
var CW_VM_EXPORT = "interface_version_"
var WASI_VM_EXPORT = "wasi_"

// simplest wasmx version 1 interface
var WASMX_ENV_1 = "wasmx_env_1"

// wasmx version 2 with env information
var WASMX_ENV_2 = "wasmx_env_2"

// only for core consensus
var WASMX_CONSENSUS_JSON_1 = "wasmx_consensus_json_1"

// non-deterministic system operations, only as queries
var SYS_ENV_1 = "sys_env_1"

// wasi
var WASI_SNAPSHOT_PREVIEW1 = "wasi_snapshot_preview1"
var WASI_UNSTABLE = "wasi_unstable"

// initial interface use in precompiles 1 -> 9
// TODO replace & remove
var EWASM_ENV_0 = "ewasm_interface_version_1"

// current ewasm interface
var EWASM_ENV_1 = "ewasm_env_1"

// current cosmwasm interface
var CW_ENV_8 = "interface_version_8"

var DEFAULT_SYS_DEP = EWASM_ENV_1

var SUPPORTED_HOST_INTERFACES = map[string]bool{
	WASMX_ENV_1:            true,
	WASMX_ENV_2:            true,
	EWASM_ENV_1:            true,
	CW_ENV_8:               true,
	WASMX_CONSENSUS_JSON_1: true,
}

var ROLE_EID_REGISTRY = "eid_registry"
var ROLE_STORAGE = "storage"
var ROLE_STAKING = "staking"
var ROLE_BANK = "bank"
var ROLE_HOOKS = "hooks"
var ROLE_HOOKS_NONC = "hooks_nonconsensus"
var ROLE_GOVERNANCE = "gov"
var ROLE_AUTH = "auth"
var ROLE_SLASHING = "slashing"
var ROLE_DISTRIBUTION = "distribution"
var ROLE_INTERPRETER = "interpreter"
var ROLE_PRECOMPILE = "precompile"
var ROLE_ALIAS = "alias"
var ROLE_CONSENSUS = "consensus"
var ROLE_INTERPRETER_PYTHON = "interpreter_python"
var ROLE_INTERPRETER_JS = "interpreter_javascript"
var ROLE_INTERPRETER_FSM = "interpreter_state_machine"
var ROLE_INTERPRETER_TAY = "interpreter_tay"

var ROLE_LIBRARY = "deplibrary"

var ROLE_CHAT = "chat"
var ROLE_TIME = "time"
var ROLE_LEVEL0 = "level0"
var ROLE_LEVELN = "leveln"

var ROLE_MULTICHAIN_REGISTRY = "multichain_registry"
var ROLE_MULTICHAIN_REGISTRY_LOCAL = "multichain_registry_local"
var ROLE_SECRET_SHARING = "secret_sharing"

var ROLE_LOBBY = "lobby"
var ROLE_METAREGISTRY = "metaregistry"

// interpreter_<code type>_<encoding>_<version>
// code type = "solidity" | "evm" | "python" | "pythonbz"
// encoding = ""

// hex -> stored as interpreted bytecode
// utf8 -> stored as a file
// wasm -> stored in the filesystem

// TODO "interpreter_evm_hex_shanghai" ?
var INTERPRETER_EVM_SHANGHAI = "interpreter_evm_shanghai"

// https://github.com/RustPython/RustPython version
var INTERPRETER_PYTHON = "interpreter_python_utf8_0.2.0"

var INTERPRETER_JS = "interpreter_javascript_utf8_0.1.0"

var INTERPRETER_FSM = "interpreter_state_machine_bz_0.1.0"

var INTERPRETER_TAY = "tay_interpreter_v0.0.1"

var STORAGE_CHAIN = "storage_chain"

var CONSENSUS_RAFT = "consensus_raft_0.0.1"
var CONSENSUS_RAFTP2P = "consensus_raftp2p_0.0.1"

var CONSENSUS_TENDERMINT = "consensus_tendermint_0.0.1"
var CONSENSUS_TENDERMINTP2P = "consensus_tendermintp2p_0.0.1"

var CONSENSUS_AVA_SNOWMAN = "consensus_ava_snowman_0.0.1"

var STAKING_v001 = "staking_0.0.1"

var BANK_v001 = "bank_0.0.1"

var ERC20_v001 = "erc20json"
var DERC20_v001 = "derc20json"

var HOOKS_v001 = "hooks_0.0.1"
var GOV_v001 = "gov_0.0.1"
var GOV_CONT_v001 = "gov_cont_0.0.1"
var AUTH_v001 = "auth_0.0.1"
var SLASHING_v001 = "slashing_0.0.1"
var DISTRIBUTION_v001 = "distribution_0.0.1"
var CHAT_v001 = "chat_0.0.1"
var CHAT_VERIFIER_v001 = "chat_verifier_0.0.1"
var TIME_v001 = "time_0.0.1"
var LEVEL0_v001 = "level0_0.0.1"
var LEVELN_v001 = "leveln_0.0.1"
var MULTICHAIN_REGISTRY_v001 = "multichain_registry_0.0.1"
var MULTICHAIN_REGISTRY_LOCAL_v001 = "multichain_registry_local_0.0.1"
var ERC20_ROLLUP_v001 = "erc20rollupjson_0.0.1"
var LOBBY_v001 = "lobby_json_0.0.1"
var METAREGISTRY_v001 = "metaregistry_json_0.0.1"

// var ALLOC_TYPE_AS = "alloc_assemblyscript_1"
// var ALLOC_DEFAULT = "alloc_default"
var MEMORY_EXPORT_MALLOC = "malloc"
var MEMORY_EXPORT_ALLOCATE = "allocate"
var MEMORY_EXPORT_AS = "__new"

var WASMX_MEMORY_DEFAULT = "memory_default_1"
var WASMX_MEMORY_ASSEMBLYSCRIPT = "memory_assemblyscript_1"
var WASMX_MEMORY_TAYLOR = "memory_taylor"

var TRUSTED_ADDRESS_LIMIT = big.NewInt(0).SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 128})

var FILE_EXTENSIONS = map[string]string{
	ROLE_INTERPRETER_PYTHON: "py",
	INTERPRETER_PYTHON:      "py",
	ROLE_INTERPRETER_JS:     "js",
	INTERPRETER_JS:          "js",
}

// var SUPPORTED_INTERPRETERS = map[string]bool{
// 	INTERPRETER_EVM_SHANGHAI: true,
// 	INTERPRETER_PYTHON:       true,
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

func HasUtf8Dep(deps []string) bool {
	for _, dep := range deps {
		if strings.Contains(dep, "utf8") {
			return true
		}
	}
	return false
}

func HasUtf8SystemDep(sysDeps []SystemDep) bool {
	for _, dep := range sysDeps {
		if strings.Contains(dep.Label, "utf8") {
			return true
		}
	}
	return false
}

func HasInterpreterDep(deps []string) bool {
	for _, dep := range deps {
		if strings.Contains(dep, "interpreter_") {
			return true
		}
	}
	return false
}

func HasInterpreterSystemDep(sysDeps []SystemDep) bool {
	for _, dep := range sysDeps {
		if strings.Contains(dep.Label, "interpreter_") {
			return true
		}
	}
	return false
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

func BuildDep(addr string, deptype string) string {
	return fmt.Sprintf("%s:%s", addr, deptype)
}

func ParseDep(dep string) (string, string) {
	parts := strings.Split(dep, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return dep, ""
}
