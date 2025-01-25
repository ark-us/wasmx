package types

import (
	bytes "bytes"
	"encoding/json"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/common"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

type SystemContracts = []SystemContract

const FEE_COLLECTOR = "fee_collector"

var ADDR_ECRECOVER = "0x0000000000000000000000000000000000000001"
var ADDR_SHA2_256 = "0x0000000000000000000000000000000000000002"
var ADDR_RIPMD160 = "0x0000000000000000000000000000000000000003"
var ADDR_IDENTITY = "0x0000000000000000000000000000000000000004"
var ADDR_MODEXP = "0x0000000000000000000000000000000000000005"
var ADDR_ECADD = "0x0000000000000000000000000000000000000006"
var ADDR_ECMUL = "0x0000000000000000000000000000000000000007"
var ADDR_ECPAIRINGS = "0x0000000000000000000000000000000000000008"
var ADDR_BLAKE2F = "0x0000000000000000000000000000000000000009"

var ADDR_ECRECOVERETH = "0x000000000000000000000000000000000000001f"

var ADDR_SECP384R1 = "0x0000000000000000000000000000000000000020"
var ADDR_SECP384R1_REGISTRY = "0x0000000000000000000000000000000000000021"
var ADDR_SECRET_SHARING = "0x0000000000000000000000000000000000000022"
var ADDR_INTERPRETER_EVM_SHANGHAI = "0x0000000000000000000000000000000000000023"
var ADDR_ALIAS_ETH = "0x0000000000000000000000000000000000000024"
var ADDR_PROXY_INTERFACES = "0x0000000000000000000000000000000000000025"
var ADDR_INTERPRETER_PYTHON = "0x0000000000000000000000000000000000000026"
var ADDR_INTERPRETER_JS = "0x0000000000000000000000000000000000000027"
var ADDR_INTERPRETER_FSM = "0x0000000000000000000000000000000000000028"
var ADDR_STORAGE_CHAIN = "0x0000000000000000000000000000000000000029"
var ADDR_CONSENSUS_RAFT_LIBRARY = "0x000000000000000000000000000000000000002a"
var ADDR_CONSENSUS_TENDERMINT_LIBRARY = "0x000000000000000000000000000000000000002b"
var ADDR_CONSENSUS_RAFT = "0x000000000000000000000000000000000000002c"
var ADDR_CONSENSUS_TENDERMINT = "0x000000000000000000000000000000000000002d"
var ADDR_CONSENSUS_AVA_SNOWMAN_LIBRARY = "0x000000000000000000000000000000000000002e"
var ADDR_CONSENSUS_AVA_SNOWMAN = "0x000000000000000000000000000000000000002f"
var ADDR_STAKING = "0x0000000000000000000000000000000000000030"
var ADDR_BANK = "0x0000000000000000000000000000000000000031"
var ADDR_HOOKS = "0x0000000000000000000000000000000000000034"
var ADDR_GOV = "0x0000000000000000000000000000000000000035"
var ADDR_GOV_CONT = "0x0000000000000000000000000000000000000038"
var ADDR_AUTH = "0x0000000000000000000000000000000000000039"
var ADDR_CONSENSUS_RAFTP2P_LIBRARY = "0x0000000000000000000000000000000000000036"
var ADDR_CONSENSUS_RAFTP2P = "0x0000000000000000000000000000000000000037"
var ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY = "0x0000000000000000000000000000000000000040"
var ADDR_CONSENSUS_TENDERMINTP2P = "0x0000000000000000000000000000000000000041"
var ADDR_CHAT = "0x0000000000000000000000000000000000000042"
var ADDR_HOOKS_NONC = "0x0000000000000000000000000000000000000043"
var ADDR_CHAT_VERIFIER = "0x0000000000000000000000000000000000000044"
var ADDR_SLASHING = "0x0000000000000000000000000000000000000045"
var ADDR_DISTRIBUTION = "0x0000000000000000000000000000000000000046"
var ADDR_TIME = "0x0000000000000000000000000000000000000047"
var ADDR_LEVEL0 = "0x0000000000000000000000000000000000000048"
var ADDR_LEVEL0_LIBRARY = "0x0000000000000000000000000000000000000049"
var ADDR_MULTICHAIN_REGISTRY = "0x000000000000000000000000000000000000004a"
var ADDR_MULTICHAIN_REGISTRY_LOCAL = "0x000000000000000000000000000000000000004b"
var ADDR_LOBBY = "0x000000000000000000000000000000000000004d"
var ADDR_LOBBY_LIBRARY = "0x000000000000000000000000000000000000004e"
var ADDR_METAREGISTRY = "0x000000000000000000000000000000000000004f"
var ADDR_INTERPRETER_TAY = "0x0000000000000000000000000000000000000050"

var ADDR_LEVEL0_ONDEMAND = "0x0000000000000000000000000000000000000051"
var ADDR_LEVEL0_ONDEMAND_LIBRARY = "0x0000000000000000000000000000000000000052"

var ADDR_ROLES = "0x0000000000000000000000000000000000000060"
var ADDR_STORAGE_CONTRACTS = "0x0000000000000000000000000000000000000061"

var ADDR_SYS_PROXY = "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

func StarterPrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("SimplePrecompiles: cannot marshal init message")
	}
	return []SystemContract{
		// contract storage needs to be initialized first, roles second, auth third
		{
			Address:     ADDR_STORAGE_CONTRACTS,
			Label:       STORAGE_CONTRACTS_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_STORAGE_CONTRACTS, Label: ROLE_STORAGE_CONTRACTS, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ROLES,
			Label:       ROLES_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_ROLES, Label: ADDR_ROLES, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_AUTH,
			Label:       AUTH_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_AUTH, Label: ADDR_AUTH, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func SimplePrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("SimplePrecompiles: cannot marshal init message")
	}
	return []SystemContract{
		{
			Address:     ADDR_ECRECOVER,
			Label:       "ecrecover",
			InitMessage: initMsg,
			Pinned:      false,
			Native:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		// Ethereum ecrecover
		{
			Address:     ADDR_ECRECOVERETH,
			Label:       "ecrecovereth",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SHA2_256,
			Label:       "sha2-256",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_RIPMD160,
			Label:       "ripmd160",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_IDENTITY,
			Label:       "identity",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_MODEXP,
			Label:       "modexp",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ECADD,
			Label:       "ecadd",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ECMUL,
			Label:       "ecmul",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ECPAIRINGS,
			Label:       "ecpairings",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_BLAKE2F,
			Label:       "blake2f",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func InterpreterPrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("InterpreterPrecompiles: cannot marshal init message")
	}
	return []SystemContract{
		{
			Address:     ADDR_INTERPRETER_EVM_SHANGHAI,
			Label:       INTERPRETER_EVM_SHANGHAI,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_INTERPRETER, Label: INTERPRETER_EVM_SHANGHAI, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_INTERPRETER_PYTHON,
			Label:       INTERPRETER_PYTHON,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_INTERPRETER, Label: INTERPRETER_PYTHON},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_INTERPRETER_JS,
			Label:       INTERPRETER_JS,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_INTERPRETER, Label: INTERPRETER_JS},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_INTERPRETER_FSM,
			Label:       INTERPRETER_FSM,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_INTERPRETER, Label: INTERPRETER_FSM},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_INTERPRETER_TAY,
			Label:       INTERPRETER_TAY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_INTERPRETER, Label: INTERPRETER_TAY},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{WASMX_MEMORY_TAYLOR},
		},
	}
}

func BasePrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("BasePrecompiles: cannot marshal init message")
	}

	storageInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"initialBlockIndex":1}`)})
	if err != nil {
		panic("BasePrecompiles: cannot marshal storageInitMsg message")
	}
	return []SystemContract{
		{
			Address:     ADDR_STORAGE_CHAIN,
			Label:       STORAGE_CHAIN,
			InitMessage: storageInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_STORAGE, Label: STORAGE_CHAIN, Primary: true},
			StorageType: ContractStorageType_MetaConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ALIAS_ETH,
			Label:       ALIAS_ETH,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        &SystemContractRole{Role: ROLE_ALIAS, Label: ALIAS_ETH, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_PROXY_INTERFACES,
			Label:       PROXY_INTERFACES,
			InitMessage: initMsg,
			Pinned:      false,
			Native:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SYS_PROXY,
			Label:       SYS_PROXY,
			InitMessage: initMsg,
			Pinned:      false,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func EIDPrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("EIDPrecompiles: cannot marshal init message")
	}
	return []SystemContract{
		{
			Address:     ADDR_SECP384R1,
			Label:       "secp384r1",
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SECP384R1_REGISTRY,
			Label:       SECP384r1_REGISTRY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_EID_REGISTRY, Label: SECP384r1_REGISTRY, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SECRET_SHARING,
			Label:       "secret_sharing",
			InitMessage: initMsg,
			Pinned:      false,
			Native:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func CosmosPrecompiles(feeCollectorBech32 string, mintBech32 string) SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("CosmosPrecompiles: cannot marshal init message")
	}
	// TODO remove/replace minter
	// note we use ROLE_BANK for redirected messages through cosmosmod
	bankInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"authorities":["%s","%s","%s","%s","%s"]}`, ROLE_STAKING, ROLE_GOVERNANCE, ROLE_BANK, feeCollectorBech32, mintBech32))})
	if err != nil {
		panic("CosmosPrecompiles: cannot marshal bankInitMsg message")
	}

	// govInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"arbitrationDenom":"aarb","coefs":["0x100000","0x3","0x64","0x7d0","0x5dc","0xa","0x4","0x8","0x2710","0x5fb","0x3e8"],"defaultX":1425,"defaultY":1000}`)})
	// govInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"arbitrationDenom":"aarb","coefs":[],"defaultX":1425,"defaultY":1000}`)})
	govInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"arbitrationDenom":"aarb","coefs":[1048576, 3, 100, 2000, 1500, 10, 4, 8, 10000, 1531, 1000],"defaultX":1531,"defaultY":1000}`)})
	if err != nil {
		panic("CosmosPrecompiles: cannot marshal govInitMsg message")
	}

	return []SystemContract{
		{
			Address:     ADDR_STAKING,
			Label:       STAKING_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_STAKING, Label: STAKING_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_BANK,
			Label:       BANK_v001,
			InitMessage: bankInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_BANK, Label: BANK_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		// we only need to create, not initialize the erc20 contract
		{
			Address:     "",
			Label:       ERC20_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		// we only need to create, not initialize the derc20 contract
		{
			Address:     "",
			Label:       DERC20_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SLASHING,
			Label:       SLASHING_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_SLASHING, Label: SLASHING_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_DISTRIBUTION,
			Label:       DISTRIBUTION_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_DISTRIBUTION, Label: DISTRIBUTION_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_GOV,
			Label:       GOV_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_GOVERNANCE, Label: GOV_v001},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_GOV_CONT,
			Label:       GOV_CONT_v001,
			InitMessage: govInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_GOVERNANCE, Label: GOV_CONT_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func HookPrecompiles() SystemContracts {
	hooksbz, err := json.Marshal(DEFAULT_HOOKS)
	if err != nil {
		panic("HookPrecompiles: cannot marshal hooks message")
	}
	hooksInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"hooks":%s}`, hooksbz))})
	if err != nil {
		panic("HookPrecompiles: cannot marshal hooksInitMsg message")
	}

	hooksnoncbz, err := json.Marshal(DEFAULT_HOOKS_NONC)
	if err != nil {
		panic("HookPrecompiles: cannot marshal hooks nonc message")
	}
	hooksInitMsgNonC, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"hooks":%s}`, hooksnoncbz))})
	if err != nil {
		panic("HookPrecompiles: cannot marshal hooksInitMsgNonC message")
	}
	return []SystemContract{
		{
			Address:     ADDR_HOOKS,
			Label:       HOOKS_v001,
			InitMessage: hooksInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_HOOKS, Label: ROLE_HOOKS + "_" + HOOKS_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_HOOKS_NONC,
			Label:       HOOKS_v001,
			InitMessage: hooksInitMsgNonC,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_HOOKS_NONC, Label: ROLE_HOOKS_NONC + "_" + HOOKS_v001, Primary: true},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
	}
}

func ConsensusPrecompiles(minValidatorCount int32, enableEIDCheck bool, currentLevel int32, initialPortValues string, erc20CodeId int32, derc20CodeId int32) SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal init message")
	}

	raftInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"validatorNodesInfo","value":"[]"},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"matchIndex","value":"[]"},{"key":"commitIndex","value":"0"},{"key":"currentTerm","value":"0"},{"key":"lastApplied","value":"0"},{"key":"blockTimeout","value":"heartbeatTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"prevLogIndex","value":"0"},{"key":"currentNodeId","value":"0"},{"key":"electionReset","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"electionTimeout","value":"0"},{"key":"maxElectionTime","value":"20000"},{"key":"minElectionTime","value":"10000"},{"key":"heartbeatTimeout","value":"5000"}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal raftInitMsg message")
	}

	tendermintInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":15000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"20000000"}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal tendermintInitMsg message")
	}

	tendermintP2PInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":"2000"},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"60000000"},{"key":"timeoutPropose","value":15000},{"key":"timeoutPrevote","value":15000},{"key":"timeoutPrecommit","value":20000}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal tendermintInitMsg message")
	}

	avaInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"sampleSize","value":"2"},{"key":"betaThreshold","value":2},{"key":"roundsCounter","value":"0"},{"key":"alphaThreshold","value":80}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal avaInitMsg message")
	}

	timeInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"params":{"chain_id":"time_666-1","interval_ms":100}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal timeInitMsg message")
	}

	level0InitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":3000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"60000000"},{"key":"timeoutPropose","value":20000},{"key":"timeoutPrecommit","value":20000}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal level0InitMsg message")
	}

	lobbyInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"instantiate":{"context":[{"key":"heartbeatTimeout","value":5000},{"key":"newchainTimeout","value":20000},{"key":"current_level","value":0},{"key":"min_validators_count","value":%d},{"key":"enable_eid_check","value":%t},{"key":"erc20CodeId","value":%d},{"key":"derc20CodeId","value":%d},{"key":"level_initial_balance","value":10000000000000000000}],"initialState":"uninitialized"}}`, minValidatorCount, enableEIDCheck, erc20CodeId, derc20CodeId))})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal lobbyInitMsg message")
	}

	mutichainLocalInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"ids":[],"initialPorts":%s}`, initialPortValues))})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal mutichainLocalInitMsg message")
	}

	metaregistryInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"params":{"current_level":0}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal metaregistryInitMsg message")
	}

	level0OnDemandInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":2000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"60000000"},{"key":"timeoutPropose","value":20000},{"key":"timeoutPrecommit","value":20000},{"key":"batchTimeout","value":1000}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal level0OnDemandInitMsg message")
	}

	return []SystemContract{
		{
			Address:     ADDR_CONSENSUS_RAFT_LIBRARY,
			Label:       CONSENSUS_RAFT_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: CONSENSUS_RAFT_LIBRARY},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_RAFTP2P_LIBRARY,
			Label:       CONSENSUS_RAFTP2P_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: CONSENSUS_RAFTP2P_LIBRARY},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_TENDERMINT_LIBRARY,
			Label:       CONSENSUS_TENDERMINT_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: CONSENSUS_TENDERMINT_LIBRARY},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY,
			Label:       CONSENSUS_TENDERMINTP2P_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: CONSENSUS_TENDERMINTP2P_LIBRARY, Primary: true},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_RAFT,
			Label:       CONSENSUS_RAFT,
			InitMessage: raftInitMsg,
			Pinned:      false,
			// Role:        ROLE_CONSENSUS,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_RAFT_LIBRARY, ROLE_LIBRARY)},
		},
		{
			Address:     ADDR_CONSENSUS_RAFTP2P,
			Label:       CONSENSUS_RAFTP2P,
			InitMessage: raftInitMsg,
			Pinned:      false,
			// Role:        ROLE_CONSENSUS,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_RAFTP2P_LIBRARY, ROLE_LIBRARY)},
		},
		{
			Address:     ADDR_CONSENSUS_TENDERMINT,
			Label:       CONSENSUS_TENDERMINT,
			InitMessage: tendermintInitMsg,
			Pinned:      false,
			// Role:        ROLE_CONSENSUS,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_TENDERMINT_LIBRARY, ROLE_LIBRARY)},
		},
		{
			Address:     ADDR_CONSENSUS_TENDERMINTP2P,
			Label:       CONSENSUS_TENDERMINTP2P,
			InitMessage: tendermintP2PInitMsg,
			Pinned:      false,
			// Role:        ROLE_CONSENSUS,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY, ROLE_LIBRARY)},
		},
		{
			Address:     ADDR_CONSENSUS_AVA_SNOWMAN_LIBRARY,
			Label:       CONSENSUS_AVA_SNOWMAN_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: CONSENSUS_AVA_SNOWMAN_LIBRARY},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_AVA_SNOWMAN,
			Label:       CONSENSUS_AVA_SNOWMAN,
			InitMessage: avaInitMsg,
			Pinned:      false,
			// Role:        ROLE_CONSENSUS,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_CONSENSUS_AVA_SNOWMAN_LIBRARY, ROLE_LIBRARY)},
		},
		{
			Address:     ADDR_TIME,
			Label:       TIME_v001,
			InitMessage: timeInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_TIME, Label: TIME_v001, Primary: true},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_LEVEL0_LIBRARY,
			Label:       CONSENSUS_LEVEL_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: CONSENSUS_LEVEL_LIBRARY},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_LEVEL0,
			Label:       LEVEL0_v001,
			InitMessage: level0InitMsg,
			Pinned:      false,
			// Role:        ROLE_CONSENSUS,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_LEVEL0_LIBRARY, ROLE_LIBRARY)},
		},
		{
			Address:     ADDR_MULTICHAIN_REGISTRY_LOCAL,
			Label:       MULTICHAIN_REGISTRY_LOCAL_v001,
			InitMessage: mutichainLocalInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_MULTICHAIN_REGISTRY_LOCAL, Label: MULTICHAIN_REGISTRY_LOCAL_v001, Primary: true},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_LOBBY_LIBRARY,
			Label:       LOBBY_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: LOBBY_LIBRARY},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_LOBBY,
			Label:       LOBBY_v001,
			InitMessage: lobbyInitMsg,
			Pinned:      false,
			Role:        &SystemContractRole{Role: ROLE_LOBBY, Label: LOBBY_v001, Primary: true},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_LOBBY_LIBRARY, ROLE_LIBRARY)},
		},
		{
			Address:     ADDR_METAREGISTRY,
			Label:       METAREGISTRY_v001,
			InitMessage: metaregistryInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_METAREGISTRY, Label: METAREGISTRY_v001, Primary: true},
			StorageType: ContractStorageType_MetaConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_LEVEL0_ONDEMAND_LIBRARY,
			Label:       LEVEL0_ONDEMAND_LIBRARY,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_LIBRARY, Label: LEVEL0_ONDEMAND_LIBRARY},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_LEVEL0_ONDEMAND,
			Label:       LEVEL0_ONDEMAND_v001,
			InitMessage: level0OnDemandInitMsg,
			Pinned:      false,
			// Role:        ROLE_CONSENSUS,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{INTERPRETER_FSM, BuildDep(ADDR_LEVEL0_ONDEMAND_LIBRARY, ROLE_LIBRARY)},
		},
	}
}

func MultiChainPrecompiles(minValidatorCount int32, enableEIDCheck bool, erc20CodeId int32, derc20CodeId int32) SystemContracts {
	mutichainInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"params":{"min_validators_count":%d,"enable_eid_check":%t,"erc20CodeId":%d,"derc20CodeId":%d,"level_initial_balance":"10000000000000000000"}}`, minValidatorCount, enableEIDCheck, erc20CodeId, derc20CodeId))})
	if err != nil {
		panic("MultiChainPrecompiles: cannot marshal mutichainInitMsg message")
	}

	return []SystemContract{
		{
			Address:     ADDR_MULTICHAIN_REGISTRY,
			Label:       MULTICHAIN_REGISTRY_v001,
			InitMessage: mutichainInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_MULTICHAIN_REGISTRY, Label: MULTICHAIN_REGISTRY_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func ChatPrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("ChatPrecompiles: cannot marshal init message")
	}
	return []SystemContract{
		{
			Address:     ADDR_CHAT,
			Label:       CHAT_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_CHAT, Label: CHAT_v001, Primary: true},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CHAT_VERIFIER,
			Label:       CHAT_VERIFIER_v001,
			InitMessage: initMsg,
			Pinned:      true,
			MeteringOff: true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func DefaultSystemContracts(accBech32Codec mcodec.AccBech32Codec, feeCollectorBech32 string, mintBech32 string, minValidatorCount int32, enableEIDCheck bool, initialPortValues string) SystemContracts {

	precompiles := StarterPrecompiles()
	precompiles = append(precompiles, SimplePrecompiles()...)
	precompiles = append(precompiles, InterpreterPrecompiles()...)
	precompiles = append(precompiles, BasePrecompiles()...)
	precompiles = append(precompiles, EIDPrecompiles()...)
	precompiles = append(precompiles, HookPrecompiles()...)
	precompiles = append(precompiles, CosmosPrecompiles(feeCollectorBech32, mintBech32)...)

	erc20CodeId := int32(0)
	derc20CodeId := int32(0)
	for i, p := range precompiles {
		if p.Label == ERC20_v001 {
			erc20CodeId = int32(i + 1)
		}
	}
	for i, p := range precompiles {
		if p.Label == DERC20_v001 {
			derc20CodeId = int32(i + 1)
		}
	}
	if erc20CodeId == int32(0) || derc20CodeId == int32(0) {
		panic(fmt.Sprintf("erc20 or derc20 contracts not found: erc20CodeId %d, derc20CodeId %d", erc20CodeId, derc20CodeId))
	}

	consensusPrecompiles := ConsensusPrecompiles(minValidatorCount, enableEIDCheck, 0, initialPortValues, erc20CodeId, derc20CodeId)
	for i, val := range consensusPrecompiles {
		if val.Label == CONSENSUS_TENDERMINTP2P {
			consensusPrecompiles[i].Role = &SystemContractRole{Role: ROLE_CONSENSUS, Label: CONSENSUS_TENDERMINTP2P, Primary: true}
		}
	}
	precompiles = append(precompiles, consensusPrecompiles...)
	precompiles = append(precompiles, MultiChainPrecompiles(minValidatorCount, enableEIDCheck, erc20CodeId, derc20CodeId)...)
	precompiles = append(precompiles, ChatPrecompiles()...)

	precompiles, err := FillRoles(precompiles, accBech32Codec, feeCollectorBech32)
	if err != nil {
		panic(err)
	}
	return precompiles
}

func DefaultTimeChainContracts(accBech32Codec mcodec.AccBech32Codec, feeCollectorBech32 string, mintBech32 string, minValidatorCount int32, enableEIDCheck bool, initialPortValues string) SystemContracts {
	// DEFAULT_HOOKS_NONC
	hooksNonC := []Hook{
		{
			Name:          HOOK_START_NODE,
			SourceModules: []string{ROLE_HOOKS_NONC},
			TargetModules: []string{ROLE_CONSENSUS, ROLE_MULTICHAIN_REGISTRY_LOCAL, ROLE_TIME, ROLE_LOBBY},
		},
		{
			Name:          HOOK_SETUP_NODE,
			SourceModules: []string{ROLE_HOOKS_NONC},
			TargetModules: []string{ROLE_CONSENSUS, ROLE_LOBBY},
		},
		{
			Name:          HOOK_NEW_SUBCHAIN,
			SourceModules: []string{ROLE_HOOKS_NONC, ROLE_LOBBY, ROLE_CONSENSUS},
			TargetModules: []string{ROLE_METAREGISTRY, ROLE_MULTICHAIN_REGISTRY_LOCAL},
		},
	}
	hooksbz, err := json.Marshal(DEFAULT_HOOKS)
	if err != nil {
		panic("DefaultTimeChainContracts: cannot marshal hooks message")
	}
	hooksInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"hooks":%s}`, hooksbz))})
	if err != nil {
		panic("DefaultTimeChainContracts: cannot marshal hooksInitMsg message")
	}

	hooksnoncbz, err := json.Marshal(hooksNonC)
	if err != nil {
		panic("DefaultTimeChainContracts: cannot marshal hooks nonc message")
	}
	hooksInitMsgNonC, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"hooks":%s}`, hooksnoncbz))})
	if err != nil {
		panic("DefaultTimeChainContracts: cannot marshal hooksInitMsgNonC message")
	}
	hooksPrecompiles := []SystemContract{
		{
			Address:     ADDR_HOOKS,
			Label:       HOOKS_v001,
			InitMessage: hooksInitMsg,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_HOOKS, Label: ROLE_HOOKS + "_" + HOOKS_v001, Primary: true},
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_HOOKS_NONC,
			Label:       HOOKS_v001,
			InitMessage: hooksInitMsgNonC,
			Pinned:      true,
			MeteringOff: true,
			Role:        &SystemContractRole{Role: ROLE_HOOKS_NONC, Label: ROLE_HOOKS_NONC + "_" + HOOKS_v001, Primary: true},
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
	}

	precompiles := StarterPrecompiles()
	precompiles = append(precompiles, SimplePrecompiles()...)
	precompiles = append(precompiles, InterpreterPrecompiles()...)
	precompiles = append(precompiles, BasePrecompiles()...)
	precompiles = append(precompiles, EIDPrecompiles()...)
	precompiles = append(precompiles, hooksPrecompiles...)
	precompiles = append(precompiles, CosmosPrecompiles(feeCollectorBech32, mintBech32)...)

	erc20CodeId := int32(0)
	derc20CodeId := int32(0)
	for i, p := range precompiles {
		if p.Label == ERC20_v001 {
			erc20CodeId = int32(i + 1)
		}
	}
	for i, p := range precompiles {
		if p.Label == DERC20_v001 {
			derc20CodeId = int32(i + 1)
		}
	}
	if erc20CodeId == int32(0) || derc20CodeId == int32(0) {
		panic(fmt.Sprintf("erc20 or derc20 contracts not found: erc20CodeId %d, derc20CodeId %d", erc20CodeId, derc20CodeId))
	}

	consensusPrecompiles := ConsensusPrecompiles(minValidatorCount, enableEIDCheck, 0, initialPortValues, erc20CodeId, derc20CodeId)
	for i, val := range consensusPrecompiles {
		if val.Label == LEVEL0_ONDEMAND_v001 {
			consensusPrecompiles[i].Role = &SystemContractRole{Role: ROLE_CONSENSUS, Label: LEVEL0_ONDEMAND_v001, Primary: true}
		}
	}
	precompiles = append(precompiles, consensusPrecompiles...)
	precompiles = append(precompiles, MultiChainPrecompiles(minValidatorCount, enableEIDCheck, erc20CodeId, derc20CodeId)...)
	precompiles = append(precompiles, ChatPrecompiles()...)

	precompiles, err = FillRoles(precompiles, accBech32Codec, feeCollectorBech32)
	if err != nil {
		panic(err)
	}
	return precompiles
}

func (p SystemContract) Validate() error {
	if err := validateString(p.Label); err != nil {
		return err
	}
	if err := validateString(p.Address); err != nil {
		return err
	}
	if err := p.InitMessage.ValidateBasic(); err != nil {
		return err
	}
	if p.InitMessage == nil {
		return fmt.Errorf("initialization message cannot be nil")
	}
	if p.Address != "" {
		return ValidateNonZeroAddress(p.Address)
	}
	return nil
}

func validateString(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBytes(i interface{}) error {
	_, ok := i.([]byte)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

const (
	// AddressLengthCW is the expected length of a Wasmx and CosmWasm address
	AddressLengthWasmx = 32
	// AddressLengthCW is the expected length of an Ethereum address
	AddressLengthEth = 20
)

// TODO have addresses be 32bytes

// IsZeroAddress returns true if the address corresponds to an empty ethereum hex address.
func IsZeroAddress(address string) bool {
	return bytes.Equal(common.HexToAddress(address).Bytes(), common.Address{}.Bytes())
}

// ValidateAddress returns an error if the provided string is either not a hex formatted string address
func ValidateAddress(address string) error {
	if !IsHexAddress(address) {
		return sdkerr.Wrapf(
			sdkerrors.ErrInvalidAddress, "address '%s' is not a valid ethereum hex address",
			address,
		)
	}
	return nil
}

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// WasmX or Ethereum address or not.
func IsHexAddress(s string) bool {
	if has0xPrefix(s) {
		s = s[2:]
	}
	return isHex(s) && (len(s) == 2*AddressLengthWasmx || len(s) == 2*AddressLengthEth)
}

// has0xPrefix validates str begins with '0x' or '0X'.
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

// ValidateNonZeroAddress returns an error if the provided string is not a hex
// formatted string address or is equal to zero
func ValidateNonZeroAddress(address string) error {
	if IsZeroAddress(address) {
		return sdkerr.Wrapf(
			sdkerrors.ErrInvalidAddress, "address '%s' must not be zero",
			address,
		)
	}
	return ValidateAddress(address)
}

func FillRoles(precompiles []SystemContract, accBech32Codec mcodec.AccBech32Codec, feeCollectorBech32 string) ([]SystemContract, error) {
	// Map to store roles and their corresponding RoleJSON objects
	roleMap := make(map[string]*RoleJSON)

	// First pass: Collect roles with Primary: true
	for _, precompile := range precompiles {
		if precompile.Role != nil {
			role := precompile.Role.Role

			// Initialize a RoleJSON entry if it doesn't exist
			if _, exists := roleMap[role]; !exists {
				roleMap[role] = &RoleJSON{
					Role:        role,
					StorageType: int32(precompile.StorageType),
					Primary:     0,
					Multiple:    false,
					Labels:      []string{},
					Addresses:   []string{},
				}
			}

			if precompile.Role.Primary {
				prefixedAddr := accBech32Codec.BytesToAccAddressPrefixed(AccAddressFromHex(precompile.Address))
				roleMap[role].Primary = int32(len(roleMap[role].Addresses))
				roleMap[role].Labels = append(roleMap[role].Labels, precompile.Role.Label)
				roleMap[role].Addresses = append(roleMap[role].Addresses, prefixedAddr.String())
			}
		}
	}

	// Second pass: Add other contracts for roles with multiple = true
	for _, precompile := range precompiles {
		if precompile.Role != nil {
			role := precompile.Role.Role
			if entry, exists := roleMap[role]; exists {
				// Check if the contract has already been added
				prefixedAddr := accBech32Codec.BytesToAccAddressPrefixed(AccAddressFromHex(precompile.Address))
				if !contains(entry.Addresses, prefixedAddr.String()) {
					entry.Multiple = true
					entry.Labels = append(entry.Labels, precompile.Role.Label)
					entry.Addresses = append(entry.Addresses, prefixedAddr.String())
				}
			}
		}
	}

	// add denom role
	if _, exists := roleMap[ROLE_DENOM]; !exists {
		roleMap[ROLE_DENOM] = &RoleJSON{
			Role:        ROLE_DENOM,
			StorageType: int32(ContractStorageType_CoreConsensus),
			Primary:     0,
			Multiple:    true,
			Labels:      []string{},
			Addresses:   []string{},
		}
	}

	if _, exists := roleMap[FEE_COLLECTOR]; !exists {
		feeCollector, _ := accBech32Codec.BytesToString(authtypes.NewModuleAddress(FEE_COLLECTOR))
		roleMap[FEE_COLLECTOR] = &RoleJSON{
			Role:        FEE_COLLECTOR,
			StorageType: int32(ContractStorageType_CoreConsensus),
			Primary:     0,
			Multiple:    false,
			Labels:      []string{FEE_COLLECTOR},
			Addresses:   []string{feeCollector},
		}
	}

	// Prepare the RolesGenesis message
	roles := make([]RoleJSON, 0, len(roleMap))
	for _, roleJSON := range roleMap {
		roles = append(roles, *roleJSON)
	}

	// IndividualMigration - a list of roles that handle their own role migration for action type Replace
	msgInit := RolesGenesis{Roles: roles, IndividualMigration: []string{ROLE_CONSENSUS}}
	msgInitBz, err := json.Marshal(msgInit)
	if err != nil {
		return nil, err
	}

	msgbz, err := json.Marshal(&WasmxExecutionMessage{Data: msgInitBz})
	if err != nil {
		return nil, err
	}

	// Update the InitMessage for the ROLE_ROLES contract
	for i, precompile := range precompiles {
		if precompile.Role != nil && precompile.Role.Role == ROLE_ROLES {
			precompiles[i].InitMessage = msgbz
		}
	}

	return precompiles, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
