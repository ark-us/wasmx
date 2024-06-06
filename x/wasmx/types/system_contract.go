package types

import (
	bytes "bytes"
	"encoding/json"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
)

type SystemContracts = []SystemContract

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

var ADDR_SYS_PROXY = "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

func StarterPrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("SimplePrecompiles: cannot marshal init message")
	}
	return []SystemContract{
		// auth needs to be initialized first (account keeper)
		{
			Address:     ADDR_AUTH,
			Label:       AUTH_v001,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_AUTH,
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
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SHA2_256,
			Label:       "sha2-256",
			InitMessage: initMsg,
			Pinned:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_RIPMD160,
			Label:       "ripmd160",
			InitMessage: initMsg,
			Pinned:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_IDENTITY,
			Label:       "identity",
			InitMessage: initMsg,
			Pinned:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_MODEXP,
			Label:       "modexp",
			InitMessage: initMsg,
			Pinned:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ECADD,
			Label:       "ecadd",
			InitMessage: initMsg,
			Pinned:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ECMUL,
			Label:       "ecmul",
			InitMessage: initMsg,
			Pinned:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ECPAIRINGS,
			Label:       "ecpairings",
			InitMessage: initMsg,
			Pinned:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_BLAKE2F,
			Label:       "blake2f",
			InitMessage: initMsg,
			Pinned:      true,
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
			Pinned:      false,
			Role:        ROLE_INTERPRETER,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_INTERPRETER_PYTHON,
			Label:       INTERPRETER_PYTHON,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_INTERPRETER,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_INTERPRETER_JS,
			Label:       INTERPRETER_JS,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_INTERPRETER,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_INTERPRETER_FSM,
			Label:       INTERPRETER_FSM,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_INTERPRETER,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func BasePrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal init message")
	}

	storageInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"initialBlockIndex":1}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal storageInitMsg message")
	}
	return []SystemContract{
		{
			Address:     ADDR_STORAGE_CHAIN,
			Label:       STORAGE_CHAIN,
			InitMessage: storageInitMsg,
			Pinned:      false,
			Role:        ROLE_STORAGE,
			StorageType: ContractStorageType_MetaConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_ALIAS_ETH,
			Label:       "alias_eth",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_ALIAS,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_PROXY_INTERFACES,
			Label:       "proxy_interfaces",
			InitMessage: initMsg,
			Pinned:      false,
			Native:      true,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SYS_PROXY,
			Label:       "sys_proxy",
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
			Pinned:      false, //TODO
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SECP384R1_REGISTRY,
			Label:       "secp384r1_registry",
			InitMessage: initMsg,
			Pinned:      false, // TODO
			Role:        ROLE_EID_REGISTRY,
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
		panic("DefaultSystemContracts: cannot marshal bankInitMsg message")
	}

	// govInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"arbitrationDenom":"aarb","coefs":["0x100000","0x3","0x64","0x7d0","0x5dc","0xa","0x4","0x8","0x2710","0x5fb","0x3e8"],"defaultX":1425,"defaultY":1000}`)})
	// govInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"arbitrationDenom":"aarb","coefs":[],"defaultX":1425,"defaultY":1000}`)})
	govInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"arbitrationDenom":"aarb","coefs":[1048576, 3, 100, 2000, 1500, 10, 4, 8, 10000, 1531, 1000],"defaultX":1531,"defaultY":1000}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal govInitMsg message")
	}

	return []SystemContract{
		{
			Address:     ADDR_STAKING,
			Label:       STAKING_v001,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_STAKING,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_BANK,
			Label:       BANK_v001,
			InitMessage: bankInitMsg,
			Pinned:      false,
			Role:        ROLE_BANK,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		// we only need to create, not initialize the erc20 contract
		{
			Address:     "",
			Label:       ERC20_v001,
			InitMessage: initMsg,
			Pinned:      false,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		// we only need to create, not initialize the derc20 contract
		{
			Address:     "",
			Label:       DERC20_v001,
			InitMessage: initMsg,
			Pinned:      false,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_SLASHING,
			Label:       SLASHING_v001,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_SLASHING,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_DISTRIBUTION,
			Label:       DISTRIBUTION_v001,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_DISTRIBUTION,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_GOV,
			Label:       GOV_v001,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_GOVERNANCE,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_GOV_CONT,
			Label:       GOV_CONT_v001,
			InitMessage: govInitMsg,
			Pinned:      false,
			Role:        ROLE_GOVERNANCE,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func HookPrecompiles() SystemContracts {
	hooksbz, err := json.Marshal(DEFAULT_HOOKS)
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal hooks message")
	}
	hooksInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"hooks":%s}`, hooksbz))})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal hooksInitMsg message")
	}

	hooksnoncbz, err := json.Marshal(DEFAULT_HOOKS_NONC)
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal hooks nonc message")
	}
	hooksInitMsgNonC, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"hooks":%s}`, hooksnoncbz))})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal hooksInitMsgNonC message")
	}
	return []SystemContract{
		{
			Address:     ADDR_HOOKS,
			Label:       HOOKS_v001,
			InitMessage: hooksInitMsg,
			Pinned:      false,
			Role:        ROLE_HOOKS,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_HOOKS_NONC,
			Label:       HOOKS_v001,
			InitMessage: hooksInitMsgNonC,
			Pinned:      false,
			Role:        ROLE_HOOKS_NONC,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
	}
}

func ConsensusPrecompiles() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("ConsensusPrecompiles: cannot marshal init message")
	}

	raftInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"validatorNodesInfo","value":"[]"},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"matchIndex","value":"[]"},{"key":"commitIndex","value":"0"},{"key":"currentTerm","value":"0"},{"key":"lastApplied","value":"0"},{"key":"blockTimeout","value":"heartbeatTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"prevLogIndex","value":"0"},{"key":"currentNodeId","value":"0"},{"key":"electionReset","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"electionTimeout","value":"0"},{"key":"maxElectionTime","value":"20000"},{"key":"minElectionTime","value":"10000"},{"key":"heartbeatTimeout","value":"5000"}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal raftInitMsg message")
	}

	tendermintInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":15000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"20000000"}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal tendermintInitMsg message")
	}

	tendermintP2PInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":"5000"},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"timeoutPropose","value":5000},{"key":"timeoutPrevote","value":5000},{"key":"timeoutPrecommit","value":5000}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal tendermintInitMsg message")
	}

	avaInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"sampleSize","value":"2"},{"key":"betaThreshold","value":2},{"key":"roundsCounter","value":"0"},{"key":"alphaThreshold","value":80}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal avaInitMsg message")
	}

	timeInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"params":{"chain_id":"time_666-1","interval_ms":100}}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal timeInitMsg message")
	}

	level0InitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"blockTimeout","value":"roundTimeout"},{"key":"max_tx_bytes","value":"65536"},{"key":"roundTimeout","value":4000},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"timeoutPrevote","value":3000},{"key":"timeoutPropose","value":3000},{"key":"timeoutPrecommit","value":3000}],"initialState":"uninitialized"}}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal level0InitMsg message")
	}

	mutichainLocalInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(`{"ids":[]}`)})
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal mutichainLocalInitMsg message")
	}

	return []SystemContract{
		{
			Address:     ADDR_CONSENSUS_RAFT_LIBRARY,
			Label:       "raft_library",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_LIBRARY,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_RAFTP2P_LIBRARY,
			Label:       "raftp2p_library",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_LIBRARY,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_TENDERMINT_LIBRARY,
			Label:       "tendermint_library",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_LIBRARY,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CONSENSUS_TENDERMINTP2P_LIBRARY,
			Label:       "tendermintp2p_library",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_LIBRARY,
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
			Label:       "ava_snowman_library",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_LIBRARY,
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
			Pinned:      false,
			Role:        ROLE_TIME,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_LEVEL0_LIBRARY,
			Label:       "level0_library",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_LIBRARY,
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
			Pinned:      false,
			Role:        ROLE_MULTICHAIN_REGISTRY_LOCAL,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
	}
}

func MultiChainPrecompiles(minValidatorCount int32, enableEIDCheck bool) SystemContracts {
	mutichainInitMsg, err := json.Marshal(WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"params":{"min_validators_count":%d,"enable_eid_check":%t,"erc20CodeId":27,"derc20CodeId":28,"level_initial_balance":"10000000000000000000"}}`, minValidatorCount, enableEIDCheck))})
	if err != nil {
		panic("MultiChainPrecompiles: cannot marshal mutichainInitMsg message")
	}

	return []SystemContract{
		{
			Address:     ADDR_MULTICHAIN_REGISTRY,
			Label:       MULTICHAIN_REGISTRY_v001,
			InitMessage: mutichainInitMsg,
			Pinned:      false,
			Role:        ROLE_MULTICHAIN_REGISTRY,
			StorageType: ContractStorageType_SingleConsensus,
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
			Pinned:      false,
			Role:        ROLE_CHAT,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_CHAT_VERIFIER,
			Label:       CHAT_VERIFIER_v001,
			InitMessage: initMsg,
			Pinned:      false,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
	}
}

func DefaultSystemContracts(feeCollectorBech32 string, mintBech32 string, minValidatorCount int32, enableEIDCheck bool) SystemContracts {
	consensusPrecompiles := ConsensusPrecompiles()
	for i, val := range consensusPrecompiles {
		if val.Label == CONSENSUS_TENDERMINTP2P {
			consensusPrecompiles[i].Role = ROLE_CONSENSUS
		}
	}

	precompiles := StarterPrecompiles()
	precompiles = append(precompiles, SimplePrecompiles()...)
	precompiles = append(precompiles, InterpreterPrecompiles()...)
	precompiles = append(precompiles, BasePrecompiles()...)
	precompiles = append(precompiles, EIDPrecompiles()...)
	precompiles = append(precompiles, HookPrecompiles()...)
	precompiles = append(precompiles, CosmosPrecompiles(feeCollectorBech32, mintBech32)...)
	precompiles = append(precompiles, consensusPrecompiles...)
	precompiles = append(precompiles, MultiChainPrecompiles(minValidatorCount, enableEIDCheck)...)
	precompiles = append(precompiles, ChatPrecompiles()...)
	return precompiles
}

func DefaultTimeChainContracts(feeCollectorBech32 string, mintBech32 string, minValidatorCount int32, enableEIDCheck bool) SystemContracts {
	hooksNonC := []Hook{
		{
			Name:          HOOK_START_NODE,
			SourceModule:  ROLE_HOOKS_NONC,
			TargetModules: []string{ROLE_CONSENSUS, ROLE_TIME},
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
			Pinned:      false,
			Role:        ROLE_HOOKS,
			StorageType: ContractStorageType_CoreConsensus,
			Deps:        []string{},
		},
		{
			Address:     ADDR_HOOKS_NONC,
			Label:       HOOKS_v001,
			InitMessage: hooksInitMsgNonC,
			Pinned:      false,
			Role:        ROLE_HOOKS_NONC,
			StorageType: ContractStorageType_SingleConsensus,
			Deps:        []string{},
		},
	}
	consensusPrecompiles := ConsensusPrecompiles()
	for i, val := range consensusPrecompiles {
		if val.Label == LEVEL0_v001 {
			consensusPrecompiles[i].Role = ROLE_CONSENSUS
		}
	}

	precompiles := StarterPrecompiles()
	precompiles = append(precompiles, SimplePrecompiles()...)
	precompiles = append(precompiles, InterpreterPrecompiles()...)
	precompiles = append(precompiles, BasePrecompiles()...)
	precompiles = append(precompiles, EIDPrecompiles()...)
	precompiles = append(precompiles, hooksPrecompiles...)
	precompiles = append(precompiles, CosmosPrecompiles(feeCollectorBech32, mintBech32)...)
	precompiles = append(precompiles, consensusPrecompiles...)
	precompiles = append(precompiles, MultiChainPrecompiles(minValidatorCount, enableEIDCheck)...)
	precompiles = append(precompiles, ChatPrecompiles()...)
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
	return ValidateNonZeroAddress(p.Address)
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
