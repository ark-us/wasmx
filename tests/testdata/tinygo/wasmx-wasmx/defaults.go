package wasmxwasmx

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// Addresses
const (
	ADDR_ECRECOVER    = "0x0000000000000000000000000000000000000001"
	ADDR_ECRECOVERETH = "0x000000000000000000000000000000000000001f"
	ADDR_SHA2_256     = "0x0000000000000000000000000000000000000002"
	ADDR_RIPMD160     = "0x0000000000000000000000000000000000000003"
	ADDR_IDENTITY     = "0x0000000000000000000000000000000000000004"
	ADDR_MODEXP       = "0x0000000000000000000000000000000000000005"
	ADDR_ECADD        = "0x0000000000000000000000000000000000000006"
	ADDR_ECMUL        = "0x0000000000000000000000000000000000000007"
	ADDR_ECPAIRINGS   = "0x0000000000000000000000000000000000000008"
	ADDR_BLAKE2F      = "0x0000000000000000000000000000000000000009"

	ADDR_SECP384R1                       = "0x0000000000000000000000000000000000000020"
	ADDR_SECP384R1_REGISTRY              = "0x0000000000000000000000000000000000000021"
	ADDR_SECRET_SHARING                  = "0x0000000000000000000000000000000000000022"
	ADDR_INTERPRETER_EVM_SHANGHAI        = "0x0000000000000000000000000000000000000023"
	ADDR_ALIAS_ETH                       = "0x0000000000000000000000000000000000000024"
	ADDR_PROXY_INTERFACES                = "0x0000000000000000000000000000000000000025"
	ADDR_INTERPRETER_PYTHON              = "0x0000000000000000000000000000000000000026"
	ADDR_INTERPRETER_JS                  = "0x0000000000000000000000000000000000000027"
	ADDR_INTERPRETER_FSM                 = "0x0000000000000000000000000000000000000028"
	ADDR_STORAGE_CHAIN                   = "0x0000000000000000000000000000000000000029"
	ADDR_CONSENSUS_RAFT_LIBRARY          = "0x000000000000000000000000000000000000002a"
	ADDR_CONSENSUS_TENDERMINT_LIBRARY    = "0x000000000000000000000000000000000000002b"
	ADDR_CONSENSUS_RAFT                  = "0x000000000000000000000000000000000000002c"
	ADDR_CONSENSUS_TENDERMINT            = "0x000000000000000000000000000000000000002d"
	ADDR_CONSENSUS_AVA_SNOWMAN_LIBRARY   = "0x000000000000000000000000000000000000002e"
	ADDR_CONSENSUS_AVA_SNOWMAN           = "0x000000000000000000000000000000000000002f"
	ADDR_STAKING                         = "0x0000000000000000000000000000000000000030"
	ADDR_BANK                            = "0x0000000000000000000000000000000000000031"
	ADDR_HOOKS                           = "0x0000000000000000000000000000000000000034"
	ADDR_GOV                             = "0x0000000000000000000000000000000000000035"
	ADDR_GOV_CONT                        = "0x0000000000000000000000000000000000000038"
	ADDR_AUTH                            = "0x0000000000000000000000000000000000000039"
	ADDR_CONSENSUS_RAFTP2P_LIBRARY       = "0x0000000000000000000000000000000000000036"
	ADDR_CONSENSUS_RAFTP2P               = "0x0000000000000000000000000000000000000037"
	ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY = "0x0000000000000000000000000000000000000040"
	ADDR_CONSENSUS_TENDERMINTP2P         = "0x0000000000000000000000000000000000000041"
	ADDR_CHAT                            = "0x0000000000000000000000000000000000000042"
	ADDR_HOOKS_NONC                      = "0x0000000000000000000000000000000000000043"
	ADDR_CHAT_VERIFIER                   = "0x0000000000000000000000000000000000000044"
	ADDR_SLASHING                        = "0x0000000000000000000000000000000000000045"
	ADDR_DISTRIBUTION                    = "0x0000000000000000000000000000000000000046"
	ADDR_TIME                            = "0x0000000000000000000000000000000000000047"
	ADDR_LEVEL0                          = "0x0000000000000000000000000000000000000048"
	ADDR_LEVEL0_LIBRARY                  = "0x0000000000000000000000000000000000000049"
	ADDR_MULTICHAIN_REGISTRY             = "0x000000000000000000000000000000000000004a"
	ADDR_MULTICHAIN_REGISTRY_LOCAL       = "0x000000000000000000000000000000000000004b"
	ADDR_LOBBY                           = "0x000000000000000000000000000000000000004d"
	ADDR_LOBBY_LIBRARY                   = "0x000000000000000000000000000000000000004e"
	ADDR_METAREGISTRY                    = "0x000000000000000000000000000000000000004f"
	ADDR_INTERPRETER_TAY                 = "0x0000000000000000000000000000000000000050"

	ADDR_SYS_PROXY = "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
)

// Additional dynamic addresses
const (
	ADDR_LEVEL0_ONDEMAND         = "0x0000000000000000000000000000000000000051"
	ADDR_LEVEL0_ONDEMAND_LIBRARY = "0x0000000000000000000000000000000000000052"
	ADDR_ROLES                   = "0x0000000000000000000000000000000000000060"
	ADDR_STORAGE_CONTRACTS       = "0x0000000000000000000000000000000000000061"
	ADDR_DTYPE                   = "0x0000000000000000000000000000000000000062"
	ADDR_EMAIL_HANDLER           = "0x0000000000000000000000000000000000000063"
)

// Labels
const (
	ECRECOVER                      = "ecrecover"
	ECRECOVERETH                   = "ecrecovereth"
	SHA2_256                       = "sha2-256"
	RIPMD160                       = "ripmd160"
	IDENTITY                       = "identity"
	MODEXP                         = "modexp"
	ECADD                          = "ecadd"
	ECMUL                          = "ecmul"
	ECPAIRINGS                     = "ecpairings"
	BLAKE2F                        = "blake2f"
	ALIAS_ETH                      = "alias_eth"
	PROXY_INTERFACES               = "proxy_interfaces"
	SYS_PROXY                      = "sys_proxy"
	SECP384R1                      = "secp384r1"
	SECP384R1_REGISTRY             = "secp384r1_registry"
	SECRET_SHARING                 = "secret_sharing"
	RAFT_LIBRARY                   = "raft_library"
	RAFTP2P_LIBRARY                = "raftp2p_library"
	TENDERMINT_LIBRARY             = "tendermint_library"
	TENDERMINTP2P_LIBRARY          = "tendermintp2p_library"
	AVA_SNOWMAN_LIBRARY            = "ava_snowman_library"
	LEVEL0_LIBRARY                 = "level0_library"
	LEVEL0_ONDEMAND_LIBRARY        = "level0_ondemand_library"
	LOBBY_LIBRARY                  = "lobby_library"
	INTERPRETER_EVM_SHANGHAI       = "interpreter_evm_shanghai"
	INTERPRETER_PYTHON             = "interpreter_python_utf8_0.2.0"
	INTERPRETER_JS                 = "interpreter_javascript_utf8_0.1.0"
	INTERPRETER_FSM                = "interpreter_state_machine_bz_0.1.0"
	INTERPRETER_TAY                = "tay_interpreter_v0.0.1"
	STORAGE_CHAIN                  = "storage_chain"
	CONSENSUS_RAFT                 = "consensus_raft_0.0.1"
	CONSENSUS_RAFTP2P              = "consensus_raftp2p_0.0.1"
	CONSENSUS_TENDERMINT           = "consensus_tendermint_0.0.1"
	CONSENSUS_TENDERMINTP2P        = "consensus_tendermintp2p_0.0.1"
	CONSENSUS_AVA_SNOWMAN          = "consensus_ava_snowman_0.0.1"
	STAKING_v001                   = "staking_0.0.1"
	BANK_v001                      = "bank_0.0.1"
	ERC20_v001                     = "erc20json"
	DERC20_v001                    = "derc20json"
	HOOKS_v001                     = "hooks_0.0.1"
	GOV_v001                       = "gov_0.0.1"
	GOV_CONT_v001                  = "gov_cont_0.0.1"
	AUTH_v001                      = "auth_0.0.1"
	ROLES_v001                     = "roles_0.0.1"
	SLASHING_v001                  = "slashing_0.0.1"
	DISTRIBUTION_v001              = "distribution_0.0.1"
	CHAT_v001                      = "chat_0.0.1"
	CHAT_VERIFIER_v001             = "chat_verifier_0.0.1"
	TIME_v001                      = "time_0.0.1"
	LEVEL0_v001                    = "level0_0.0.1"
	LEVEL0_ONDEMAND_v001           = "level0_ondemand_0.0.1"
	MULTICHAIN_REGISTRY_v001       = "multichain_registry_0.0.1"
	MULTICHAIN_REGISTRY_LOCAL_v001 = "multichain_registry_local_0.0.1"
	LOBBY_v001                     = "lobby_json_0.0.1"
	METAREGISTRY_v001              = "metaregistry_json_0.0.1"
	STORAGE_CONTRACTS_v001         = "storage_contracts_0.0.1"
)

// Memory types
const (
	WASMX_MEMORY_DEFAULT        = "memory_default_1"
	WASMX_MEMORY_ASSEMBLYSCRIPT = "memory_assemblyscript_1"
	WASMX_MEMORY_TAYLOR         = "memory_taylor"
	WASMX_MEMORY_RUSTi64        = "memory_rust_i64_1"
)

var EMPTY_INIT_MSG = []byte(`{"data":""}`)

const EMPTY_ROLE = ""

func WasmxExecMsg(data string) []byte {
	inner := base64.StdEncoding.EncodeToString([]byte(data))
	outer := fmt.Sprintf(`{"data":"%s"}`, inner)
	return []byte(outer)
}

func BuildDep(addr, deptype string) string { return fmt.Sprintf("%s:%s", addr, deptype) }

// Helpers for init messages
var (
	storageInitMsg = WasmxExecMsg(`{"initialBlockIndex":1}`)
)

func govInitMsg(bondBaseDenom string) []byte {
	return WasmxExecMsg(fmt.Sprintf(`{"bond_base_denom":"%s"}`, bondBaseDenom))
}

var (
	govContInitMsg        = WasmxExecMsg(`{"arbitrationDenom":"aarb","coefs":[1048576, 3, 100, 2000, 1500, 10, 4, 8, 10000, 1531, 1000],"defaultX":1531,"defaultY":1000}`)
	raftInitMsg           = WasmxExecMsg(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"validatorNodesInfo","value":"[]"},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"matchIndex","value":"[]"},{"key":"commitIndex","value":"0"},{"key":"currentTerm","value":"0"},{"key":"lastApplied","value":"0"},{"key":"blockTimeout","value":"heartbeatTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"prevLogIndex","value":"0"},{"key":"currentNodeId","value":"0"},{"key":"electionReset","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"electionTimeout","value":"0"},{"key":"maxElectionTime","value":"20000"},{"key":"minElectionTime","value":"10000"},{"key":"heartbeatTimeout","value":"5000"}],"initialState":"uninitialized"}}`)
	tendermintInitMsg     = WasmxExecMsg(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":15000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"20000000"}],"initialState":"uninitialized"}}`)
	tendermintP2PInitMsg  = WasmxExecMsg(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":"5000"},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"60000000"},{"key":"timeoutPropose","value":20000},{"key":"timeoutPrevote","value":20000},{"key":"timeoutPrecommit","value":20000}],"initialState":"uninitialized"}}`)
	avaInitMsg            = WasmxExecMsg(`{"instantiate":{"context":[{"key":"sampleSize","value":"2"},{"key":"betaThreshold","value":2},{"key":"roundsCounter","value":"0"},{"key":"alphaThreshold","value":80}],"initialState":"uninitialized"}}`)
	timeInitMsg           = WasmxExecMsg(`{"params":{"chain_id":"time_666-1","interval_ms":100}}`)
	level0InitMsg         = WasmxExecMsg(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":5000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"60000000"},{"key":"timeoutPropose","value":20000},{"key":"timeoutPrecommit","value":20000}],"initialState":"uninitialized"}}`)
	level0OnDemandInitMsg = WasmxExecMsg(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":5000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"60000000"},{"key":"timeoutPropose","value":20000},{"key":"timeoutPrecommit","value":20000},{"key":"batchTimeout","value":1000}],"initialState":"uninitialized"}}`)
)

func mutichainLocalInitMsg(initialPorts string) []byte {
	return WasmxExecMsg(fmt.Sprintf(`{"ids":[],"initialPorts":%s}`, initialPorts))
}

func lobbyInitMsg(minValidatorsCount int32, enableEID bool, currentLevel int32, erc20CodeId int32, derc20CodeId int32) []byte {
	return WasmxExecMsg(fmt.Sprintf(`{"instantiate":{"context":[{"key":"heartbeatTimeout","value":5000},{"key":"newchainTimeout","value":20000},{"key":"current_level","value":%d},{"key":"min_validators_count","value":%d},{"key":"enable_eid_check","value":%t},{"key":"erc20CodeId","value":%d},{"key":"derc20CodeId","value":%d},{"key":"level_initial_balance","value":10000000000000000000},{"key":"newchainRequestTimeout","value":1000}],"initialState":"uninitialized"}}`, currentLevel, minValidatorsCount, enableEID, erc20CodeId, derc20CodeId))
}

func metaregistryInitMsg(currentLevel int32) []byte {
	return WasmxExecMsg(fmt.Sprintf(`{"params":{"current_level":%d}}`, currentLevel))
}

func bankInitMsg(feeCollectorBech32 string, mintBech32 string) []byte {
	return WasmxExecMsg(fmt.Sprintf(`{"authorities":["%s","%s","%s","%s","%s"]}`,
		wasmx.ROLE_STAKING, wasmx.ROLE_GOVERNANCE, wasmx.ROLE_BANK, feeCollectorBech32, mintBech32,
	))
}

func mutichainInitMsg(minValidatorCount int32, enableEIDCheck bool, erc20CodeId int32, derc20CodeId int32) []byte {
	return WasmxExecMsg(fmt.Sprintf(`{"params":{"min_validators_count":%d,"enable_eid_check":%t,"erc20CodeId":%d,"derc20CodeId":%d,"level_initial_balance":"10000000000000000000"}}`,
		minValidatorCount, enableEIDCheck, erc20CodeId, derc20CodeId,
	))
}

// sc_* system contracts
var (
	sc_auth  = wasmx.SystemContract{Address: ADDR_AUTH, Label: AUTH_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_AUTH, Label: AUTH_v001, Primary: true}}
	sc_roles = wasmx.SystemContract{Address: ADDR_ROLES, Label: ROLES_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_ROLES, Label: ROLES_v001, Primary: true}}

	sc_ecrecover    = wasmx.SystemContract{Address: ADDR_ECRECOVER, Label: ECRECOVER, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: false, MeteringOff: false, Native: true}
	sc_ecrecovereth = wasmx.SystemContract{Address: ADDR_ECRECOVERETH, Label: ECRECOVERETH, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_sha2_256     = wasmx.SystemContract{Address: ADDR_SHA2_256, Label: SHA2_256, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_ripmd160     = wasmx.SystemContract{Address: ADDR_RIPMD160, Label: RIPMD160, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_identity     = wasmx.SystemContract{Address: ADDR_IDENTITY, Label: IDENTITY, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_modexp       = wasmx.SystemContract{Address: ADDR_MODEXP, Label: MODEXP, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_ecadd        = wasmx.SystemContract{Address: ADDR_ECADD, Label: ECADD, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_ecmul        = wasmx.SystemContract{Address: ADDR_ECMUL, Label: ECMUL, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_ecpairings   = wasmx.SystemContract{Address: ADDR_ECPAIRINGS, Label: ECPAIRINGS, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_blake2f      = wasmx.SystemContract{Address: ADDR_BLAKE2F, Label: BLAKE2F, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}

	sc_secp384r1          = wasmx.SystemContract{Address: ADDR_SECP384R1, Label: SECP384R1, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_secp384r1_registry = wasmx.SystemContract{Address: ADDR_SECP384R1_REGISTRY, Label: SECP384R1_REGISTRY, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_EID_REGISTRY, Label: SECP384R1_REGISTRY, Primary: true}}
	sc_secret_sharing     = wasmx.SystemContract{Address: ADDR_SECRET_SHARING, Label: SECRET_SHARING, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: false, MeteringOff: false, Native: true, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_SECRET_SHARING, Label: SECRET_SHARING, Primary: true}}
	sc_interpreter_evm    = wasmx.SystemContract{Address: ADDR_INTERPRETER_EVM_SHANGHAI, Label: INTERPRETER_EVM_SHANGHAI, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_INTERPRETER, Label: INTERPRETER_EVM_SHANGHAI, Primary: true}}
	sc_interpreter_py     = wasmx.SystemContract{Address: ADDR_INTERPRETER_PYTHON, Label: INTERPRETER_PYTHON, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_INTERPRETER, Label: INTERPRETER_PYTHON, Primary: false}}
	sc_interpreter_js     = wasmx.SystemContract{Address: ADDR_INTERPRETER_JS, Label: INTERPRETER_JS, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_INTERPRETER, Label: INTERPRETER_JS, Primary: false}}
	sc_interpreter_fsm    = wasmx.SystemContract{Address: ADDR_INTERPRETER_FSM, Label: INTERPRETER_FSM, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_INTERPRETER, Label: INTERPRETER_FSM, Primary: false}}
	sc_interpreter_tay    = wasmx.SystemContract{Address: ADDR_INTERPRETER_TAY, Label: INTERPRETER_TAY, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_INTERPRETER, Label: INTERPRETER_TAY, Primary: false}, Deps: []string{WASMX_MEMORY_TAYLOR}}

	sc_aliaseth         = wasmx.SystemContract{Address: ADDR_ALIAS_ETH, Label: ALIAS_ETH, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_ALIAS, Label: ALIAS_ETH, Primary: true}}
	sc_proxy_interfaces = wasmx.SystemContract{Address: ADDR_PROXY_INTERFACES, Label: PROXY_INTERFACES, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: false, MeteringOff: false, Native: true}

	sc_storage_codes = wasmx.SystemContract{Address: ADDR_STORAGE_CONTRACTS, Label: STORAGE_CONTRACTS_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_STORAGE_CONTRACTS, Label: STORAGE_CONTRACTS_v001, Primary: true}}
	sc_storage_chain = wasmx.SystemContract{Address: ADDR_STORAGE_CHAIN, Label: STORAGE_CHAIN, StorageType: wasmx.StorageMetaConsensus, InitMessage: storageInitMsg, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_STORAGE, Label: STORAGE_CHAIN, Primary: true}}

	sc_raft_library          = wasmx.SystemContract{Address: ADDR_CONSENSUS_RAFT_LIBRARY, Label: RAFT_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LIBRARY, Label: RAFT_LIBRARY, Primary: false}}
	sc_raftp2p_library       = wasmx.SystemContract{Address: ADDR_CONSENSUS_RAFTP2P_LIBRARY, Label: RAFTP2P_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LIBRARY, Label: RAFTP2P_LIBRARY, Primary: false}}
	sc_tendermint_library    = wasmx.SystemContract{Address: ADDR_CONSENSUS_TENDERMINT_LIBRARY, Label: TENDERMINT_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LIBRARY, Label: TENDERMINT_LIBRARY, Primary: false}}
	sc_tendermintp2p_library = wasmx.SystemContract{Address: ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY, Label: TENDERMINTP2P_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LIBRARY, Label: TENDERMINTP2P_LIBRARY, Primary: false}}

	sc_raft          = wasmx.SystemContract{Address: ADDR_CONSENSUS_RAFT, Label: CONSENSUS_RAFT, StorageType: wasmx.StorageSingleConsensus, InitMessage: raftInitMsg, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: CONSENSUS_RAFT, Primary: false}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_RAFT_LIBRARY, wasmx.ROLE_LIBRARY)}}
	sc_raftp2p       = wasmx.SystemContract{Address: ADDR_CONSENSUS_RAFTP2P, Label: CONSENSUS_RAFTP2P, StorageType: wasmx.StorageSingleConsensus, InitMessage: raftInitMsg, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: CONSENSUS_RAFTP2P, Primary: false}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_RAFTP2P_LIBRARY, wasmx.ROLE_LIBRARY)}}
	sc_tendermint    = wasmx.SystemContract{Address: ADDR_CONSENSUS_TENDERMINT, Label: CONSENSUS_TENDERMINT, StorageType: wasmx.StorageSingleConsensus, InitMessage: tendermintInitMsg, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: CONSENSUS_TENDERMINT, Primary: false}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_TENDERMINT_LIBRARY, wasmx.ROLE_LIBRARY)}}
	sc_tendermintp2p = wasmx.SystemContract{Address: ADDR_CONSENSUS_TENDERMINTP2P, Label: CONSENSUS_TENDERMINTP2P, StorageType: wasmx.StorageSingleConsensus, InitMessage: tendermintP2PInitMsg, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: CONSENSUS_TENDERMINTP2P, Primary: false}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY, wasmx.ROLE_LIBRARY)}}

	sc_ava_snowman_library = wasmx.SystemContract{Address: ADDR_CONSENSUS_AVA_SNOWMAN_LIBRARY, Label: AVA_SNOWMAN_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: AVA_SNOWMAN_LIBRARY, Primary: false}}
	sc_ava_snowman         = wasmx.SystemContract{Address: ADDR_CONSENSUS_AVA_SNOWMAN, Label: CONSENSUS_TENDERMINTP2P, StorageType: wasmx.StorageSingleConsensus, InitMessage: tendermintP2PInitMsg, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: CONSENSUS_TENDERMINTP2P, Primary: false}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY, wasmx.ROLE_LIBRARY)}}

	sc_staking = wasmx.SystemContract{Address: ADDR_STAKING, Label: STAKING_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_STAKING, Label: STAKING_v001, Primary: true}}
	// sc_bank is a function due to dynamic init message

	// create-only ERC20/DERC20
	sc_erc20  = wasmx.SystemContract{Address: "", Label: ERC20_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_derc20 = wasmx.SystemContract{Address: "", Label: DERC20_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}

	sc_slashing     = wasmx.SystemContract{Address: ADDR_SLASHING, Label: SLASHING_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_SLASHING, Label: SLASHING_v001, Primary: true}}
	sc_distribution = wasmx.SystemContract{Address: ADDR_DISTRIBUTION, Label: DISTRIBUTION_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_DISTRIBUTION, Label: DISTRIBUTION_v001, Primary: true}}
	// sc_gov is a function due to param
	sc_gov_cont = wasmx.SystemContract{Address: ADDR_GOV_CONT, Label: GOV_CONT_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: govContInitMsg, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_GOVERNANCE, Label: GOV_CONT_v001, Primary: false}}

	sc_chat          = wasmx.SystemContract{Address: ADDR_CHAT, Label: CHAT_v001, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CHAT, Label: CHAT_v001, Primary: true}}
	sc_chat_verifier = wasmx.SystemContract{Address: ADDR_CHAT_VERIFIER, Label: CHAT_VERIFIER_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false}
	sc_time          = wasmx.SystemContract{Address: ADDR_TIME, Label: TIME_v001, StorageType: wasmx.StorageSingleConsensus, InitMessage: timeInitMsg, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_TIME, Label: TIME_v001, Primary: true}}

	sc_level0_library = wasmx.SystemContract{Address: ADDR_LEVEL0_LIBRARY, Label: LEVEL0_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LIBRARY, Label: LEVEL0_LIBRARY, Primary: false}}
	sc_level0         = wasmx.SystemContract{Address: ADDR_LEVEL0, Label: LEVEL0_v001, StorageType: wasmx.StorageSingleConsensus, InitMessage: level0InitMsg, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: LEVEL0_v001, Primary: false}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_LEVEL0_LIBRARY, wasmx.ROLE_LIBRARY)}}

	sc_level0_ondemand_library = wasmx.SystemContract{Address: ADDR_LEVEL0_ONDEMAND_LIBRARY, Label: LEVEL0_ONDEMAND_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LIBRARY, Label: LEVEL0_ONDEMAND_LIBRARY, Primary: false}}
	sc_level0_ondemand         = wasmx.SystemContract{Address: ADDR_LEVEL0_ONDEMAND, Label: LEVEL0_ONDEMAND_v001, StorageType: wasmx.StorageSingleConsensus, InitMessage: level0OnDemandInitMsg, Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_CONSENSUS, Label: LEVEL0_ONDEMAND_v001, Primary: false}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_LEVEL0_ONDEMAND_LIBRARY, wasmx.ROLE_LIBRARY)}}

	sc_lobby_library = wasmx.SystemContract{Address: ADDR_LOBBY_LIBRARY, Label: LOBBY_LIBRARY, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LIBRARY, Label: LOBBY_LIBRARY, Primary: false}}
	// sc_lobby is a function due to params

	sc_sys_proxy  = wasmx.SystemContract{Address: ADDR_SYS_PROXY, Label: SYS_PROXY, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: false, MeteringOff: false, Native: false}
	sc_hooks      = wasmx.SystemContract{Address: ADDR_HOOKS, Label: HOOKS_v001, StorageType: wasmx.StorageCoreConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_HOOKS, Label: wasmx.ROLE_HOOKS + "_" + HOOKS_v001, Primary: true}}
	sc_hooks_nonc = wasmx.SystemContract{Address: ADDR_HOOKS_NONC, Label: HOOKS_v001, StorageType: wasmx.StorageSingleConsensus, InitMessage: EMPTY_INIT_MSG, Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_HOOKS_NONC, Label: wasmx.ROLE_HOOKS_NONC + "_" + HOOKS_v001, Primary: true}}
)

func sc_bank(feeCollectorBech32, mintBech32 string) wasmx.SystemContract {
	return wasmx.SystemContract{
		Address: ADDR_BANK, Label: BANK_v001, StorageType: wasmx.StorageCoreConsensus,
		InitMessage: bankInitMsg(feeCollectorBech32, mintBech32), Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_BANK, Label: BANK_v001, Primary: true},
	}
}

func sc_gov(bondBaseDenom string) wasmx.SystemContract {
	return wasmx.SystemContract{
		Address: ADDR_GOV, Label: GOV_v001, StorageType: wasmx.StorageCoreConsensus,
		InitMessage: govInitMsg(bondBaseDenom), Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_GOVERNANCE, Label: GOV_v001, Primary: true},
	}
}

func sc_multichain_registry(minValidatorCount int32, enableEIDCheck bool, erc20CodeId int32, derc20CodeId int32) wasmx.SystemContract {
	return wasmx.SystemContract{
		Address: ADDR_MULTICHAIN_REGISTRY, Label: MULTICHAIN_REGISTRY_v001, StorageType: wasmx.StorageCoreConsensus,
		InitMessage: mutichainInitMsg(minValidatorCount, enableEIDCheck, erc20CodeId, derc20CodeId), Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_MULTICHAIN_REGISTRY, Label: MULTICHAIN_REGISTRY_v001, Primary: true},
	}
}

func sc_multichain_registry_local(initialPorts string) wasmx.SystemContract {
	return wasmx.SystemContract{
		Address: ADDR_MULTICHAIN_REGISTRY_LOCAL, Label: MULTICHAIN_REGISTRY_LOCAL_v001, StorageType: wasmx.StorageSingleConsensus,
		InitMessage: mutichainLocalInitMsg(initialPorts), Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_MULTICHAIN_REGISTRY_LOCAL, Label: MULTICHAIN_REGISTRY_LOCAL_v001, Primary: true},
	}
}

func sc_lobby(minValidatorCount int32, enableEIDCheck bool, currentLevel int32, erc20CodeId int32, derc20CodeId int32) wasmx.SystemContract {
	return wasmx.SystemContract{
		Address: ADDR_LOBBY, Label: LOBBY_v001, StorageType: wasmx.StorageSingleConsensus,
		InitMessage: lobbyInitMsg(minValidatorCount, enableEIDCheck, currentLevel, erc20CodeId, derc20CodeId), Pinned: false, MeteringOff: false, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_LOBBY, Label: LOBBY_v001, Primary: true}, Deps: []string{INTERPRETER_FSM, BuildDep(ADDR_LOBBY_LIBRARY, wasmx.ROLE_LIBRARY)},
	}
}

// dynamic metaregistry builder with currentLevel param
func sc_metaregistry(currentLevel int32) wasmx.SystemContract {
	return wasmx.SystemContract{Address: ADDR_METAREGISTRY, Label: METAREGISTRY_v001, StorageType: wasmx.StorageMetaConsensus, InitMessage: metaregistryInitMsg(currentLevel), Pinned: true, MeteringOff: true, Native: false, Role: &wasmx.SystemContractRole{Role: wasmx.ROLE_METAREGISTRY, Label: METAREGISTRY_v001, Primary: true}}
}

// Utility: hex (optionally 0x-prefixed) -> bech32 with prefix
func toBech32(prefix, hexStr string) (string, error) {
	s := hexStr
	if len(s) >= 2 && (s[0:2] == "0x" || s[0:2] == "0X") {
		s = s[2:]
	}
	bz, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	return wasmx.AddrHumanizeMC(bz, prefix), nil
}

// fillRoles mirrors AssemblyScript logic and updates the roles contract init message
func fillRoles(precompiles []wasmx.SystemContract, bech32PrefixAccAddr string, feeCollectorBech32 string) ([]wasmx.SystemContract, error) {
	// Map to store roles and their corresponding RoleJSON objects
	roleMap := map[string]*wasmx.Role{}
	// ensure same order
	order := []string{}

	// First pass: Collect roles with Primary: true
	for _, precompile := range precompiles {
		if precompile.Role != nil {
			role := precompile.Role.Role

			// Initialize a RoleJSON entry if it doesn't exist
			if _, exists := roleMap[role]; !exists {
				order = append(order, role)
				st, ok := wasmx.ContractStorageTypeByString[precompile.StorageType]
				if !ok {
					return nil, fmt.Errorf("storage type does not exist: %s", precompile.StorageType)
				}
				roleMap[role] = &wasmx.Role{
					Role:        role,
					StorageType: st,
					Primary:     0,
					Multiple:    false,
					Labels:      []string{},
					Addresses:   []string{},
				}
			}

			if precompile.Role.Primary {
				prefixedAddr, err := toBech32(bech32PrefixAccAddr, precompile.Address)
				if err != nil {
					return nil, fmt.Errorf("fillRoles failed: %s", err.Error())
				}
				roleMap[role].Primary = int32(len(roleMap[role].Addresses))
				roleMap[role].Labels = append(roleMap[role].Labels, precompile.Role.Label)
				roleMap[role].Addresses = append(roleMap[role].Addresses, prefixedAddr)
			}
		}
	}

	// Second pass: Add other contracts for roles with multiple = true
	for _, precompile := range precompiles {
		if precompile.Role != nil {
			role := precompile.Role.Role
			if entry, exists := roleMap[role]; exists {
				// Check if the contract has already been added
				prefixedAddr, err := toBech32(bech32PrefixAccAddr, precompile.Address)
				if err != nil {
					return nil, fmt.Errorf("fillRoles failed: %s", err.Error())
				}
				if !slices.Contains(entry.Addresses, prefixedAddr) {
					entry.Multiple = true
					entry.Labels = append(entry.Labels, precompile.Role.Label)
					entry.Addresses = append(entry.Addresses, prefixedAddr)
				}
			}
		}
	}

	// add denom role
	if _, exists := roleMap[wasmx.ROLE_DENOM]; !exists {
		order = append(order, wasmx.ROLE_DENOM)
		roleMap[wasmx.ROLE_DENOM] = &wasmx.Role{
			Role:        wasmx.ROLE_DENOM,
			StorageType: wasmx.CoreConsensus,
			Primary:     0,
			Multiple:    true,
			Labels:      []string{},
			Addresses:   []string{},
		}
	}

	if _, exists := roleMap[wasmx.ROLE_FEE_COLLECTOR]; !exists {
		order = append(order, wasmx.ROLE_FEE_COLLECTOR)
		feeCollector, err := toBech32(bech32PrefixAccAddr, wasmx.ROLE_FEE_COLLECTOR)
		if err != nil {
			return nil, fmt.Errorf("fillRoles failed: %s", err.Error())
		}
		roleMap[wasmx.ROLE_FEE_COLLECTOR] = &wasmx.Role{
			Role:        wasmx.ROLE_FEE_COLLECTOR,
			StorageType: wasmx.CoreConsensus,
			Primary:     0,
			Multiple:    false,
			Labels:      []string{wasmx.ROLE_FEE_COLLECTOR},
			Addresses:   []string{feeCollector},
		}
	}

	// Prepare the RolesGenesis message
	if len(order) != len(roleMap) {
		return nil, fmt.Errorf("role map length mismatch")
	}
	roles := make([]wasmx.Role, 0, len(order))
	for _, roleName := range order {
		roles = append(roles, *roleMap[roleName])
	}

	// IndividualMigration - a list of roles that handle their own role migration for action type Replace
	msgInit := wasmx.RolesGenesis{Roles: roles, IndividualMigration: []string{wasmx.ROLE_CONSENSUS}}
	msgInitBz, err := json.Marshal(msgInit)
	if err != nil {
		return nil, err
	}

	msgbz, err := json.Marshal(&wasmx.WasmxExecutionMessage{Data: msgInitBz})
	if err != nil {
		return nil, err
	}

	// Update the InitMessage for the ROLE_ROLES contract
	for i, precompile := range precompiles {
		if precompile.Role != nil && precompile.Role.Role == wasmx.ROLE_ROLES {
			precompiles[i].InitMessage = msgbz
		}
	}

	return precompiles, nil
}

// GetDefaultSystemContracts builds the list mirroring AssemblyScript
func GetDefaultSystemContracts(feeCollectorBech32 string, mintBech32 string, minValidatorCount int32, enableEIDCheck bool, currentLevel int32, initialPorts string, bech32PrefixAccAddr string, bondBaseDenom string) ([]wasmx.SystemContract, error) {
	precompiles := []wasmx.SystemContract{
		// StarterPrecompiles: storage -> roles -> auth
		sc_storage_codes,
		sc_roles,
		sc_auth,

		// SimplePrecompiles
		sc_ecrecover,
		sc_ecrecovereth,
		sc_sha2_256,
		sc_ripmd160,
		sc_identity,
		sc_modexp,
		sc_ecadd,
		sc_ecmul,
		sc_ecpairings,
		sc_blake2f,

		// InterpreterPrecompiles
		sc_interpreter_evm,
		sc_interpreter_py,
		sc_interpreter_js,
		sc_interpreter_fsm,
		sc_interpreter_tay,

		// BasePrecompiles
		sc_storage_chain,
		sc_aliaseth,
		sc_proxy_interfaces,
		sc_sys_proxy,

		// EIDPrecompiles
		sc_secp384r1,
		sc_secp384r1_registry,
		sc_secret_sharing,

		// HookPrecompiles
		sc_hooks,
		sc_hooks_nonc,

		// CosmosPrecompiles
		sc_staking,
		sc_bank(feeCollectorBech32, mintBech32),
		sc_erc20,
		sc_derc20,
	}

	// CodeID starts at 1; capture ERC20/DERC20 code IDs by position
	erc20CodeId := int32(len(precompiles) - 1)
	derc20CodeId := int32(len(precompiles))

	precompiles = append(precompiles,
		// Cosmos continued
		sc_slashing,
		sc_distribution,
		sc_gov(bondBaseDenom),
		sc_gov_cont,

		// Consensus
		sc_raft_library,
		sc_raftp2p_library,
		sc_tendermint_library,
		sc_tendermintp2p_library,
		sc_raft,
		sc_raftp2p,
		sc_tendermint,
		sc_tendermintp2p,
		sc_ava_snowman_library,
		sc_ava_snowman,
		sc_time,
		sc_level0_library,
		sc_level0,
		sc_multichain_registry_local(initialPorts),
		sc_lobby_library,
		sc_lobby(minValidatorCount, enableEIDCheck, currentLevel, erc20CodeId, derc20CodeId),
		sc_metaregistry(currentLevel),
		sc_level0_ondemand_library,
		sc_level0_ondemand,

		// MultiChain
		sc_multichain_registry(minValidatorCount, enableEIDCheck, erc20CodeId, derc20CodeId),

		// Chat
		sc_chat,
		sc_chat_verifier,
	)

	// fill roles and update roles init message
	return fillRoles(precompiles, bech32PrefixAccAddr, feeCollectorBech32)
}

// GetDefaultGenesis mirrors AS getDefaultGenesis
func GetDefaultGenesis(bootstrapAccountBech32 string, feeCollectorBech32 string, mintBech32 string, minValidatorCount int32, enableEIDCheck bool, currentLevel int32, initialPorts string, bech32PrefixAccAddr string, bondBaseDenom string) (GenesisState, error) {
	systemContracts, err := GetDefaultSystemContracts(feeCollectorBech32, mintBech32, minValidatorCount, enableEIDCheck, currentLevel, initialPorts, bech32PrefixAccAddr, bondBaseDenom)
	if err != nil {
		return GenesisState{}, err
	}
	return GenesisState{
		Params:                  Params{},
		BootstrapAccountAddress: wasmx.Bech32String(bootstrapAccountBech32),
		SystemContracts:         systemContracts,
		Codes:                   []Code{},
		Contracts:               []Contract{},
		Sequences:               []Sequence{},
		CompiledFolderPath:      "",
	}, nil
}

// hooks init messages populated via DEFAULT_HOOKS
func init() {
	// Embed hooks messages into hooks contracts
	hooksBz, _ := json.Marshal(wasmx.DEFAULT_HOOKS)
	hooksNoncBz, _ := json.Marshal(wasmx.DEFAULT_HOOKS_NONC)
	sc_hooks.InitMessage = WasmxExecMsg(fmt.Sprintf(`{"hooks":%s}`, string(hooksBz)))
	sc_hooks_nonc.InitMessage = WasmxExecMsg(fmt.Sprintf(`{"hooks":%s}`, string(hooksNoncBz)))
}
