package lib

import (
	sdkmath "cosmossdk.io/math"
	authlib "github.com/loredanacirstea/wasmx-auth/lib"
	banklib "github.com/loredanacirstea/wasmx-bank/lib"
	consensus "github.com/loredanacirstea/wasmx-consensus"
	distributionlib "github.com/loredanacirstea/wasmx-distribution/lib"
	wasmx "github.com/loredanacirstea/wasmx-env"
	govmod "github.com/loredanacirstea/wasmx-gov/gov"
	slashinglib "github.com/loredanacirstea/wasmx-slashing/lib"
	stakinglib "github.com/loredanacirstea/wasmx-staking/lib"
)

const MODULE_NAME = "multichain_registry"

const (
	DEFAULT_MIN_VALIDATORS_COUNT        = 3
	DEFAULT_EID_CHECK                   = false
	DEFAULT_ERC20_CODE_ID        uint64 = 28
	DEFAULT_DERC20_CODE_ID       uint64 = 29
)

const DEFAULT_INITIAL_BALANCE = "10000000000000000000"

const CROSS_CHAIN_TIMEOUT_MS = 120000 // 2 min

// Messages
type MsgInitialize struct {
	Params Params `json:"params"`
}

type Params struct {
	MinValidatorsCount  int32       `json:"min_validators_count"`
	EnableEidCheck      bool        `json:"enable_eid_check"`
	Erc20CodeId         uint64      `json:"erc20CodeId"`
	Derc20CodeId        uint64      `json:"derc20CodeId"`
	LevelInitialBalance sdkmath.Int `json:"level_initial_balance"`
}

type SubChainData struct {
	Data               consensus.InitSubChainDeterministicRequest     `json:"data"`
	GenTxs             [][]byte                                       `json:"genTxs"` // base64-encoded JSON strings
	WasmxContractState map[wasmx.Bech32String][]wasmx.ContractStorage `json:"wasmxContractState"`
	InitialBalance     sdkmath.Int                                    `json:"initial_balance"`
	Initialized        bool                                           `json:"initialized"`
	Level              int32                                          `json:"level"`
}

// Queries
type QueryGetSubChainsRequest struct{}

type QueryGetSubChainsByIdsRequest struct {
	Ids []string `json:"ids"`
}

type QuerySubChainConfigByIdsRequest struct {
	Ids []string `json:"ids"`
}

type QueryGetSubChainIdsRequest struct{}

type QueryGetSubChainRequest struct {
	ChainID string `json:"chainId"`
}

type QueryGetSubChainIdsByLevelRequest struct {
	Level int32 `json:"level"`
}

type QueryGetCurrentLevelRequest struct{}

type QueryGetCurrentLevelResponse struct {
	Level int32 `json:"level"`
}

type QueryGetSubChainIdsByValidatorRequest struct {
	ValidatorAddress wasmx.Bech32String `json:"validator_address"`
}

type QueryGetValidatorsByChainIdRequest struct {
	ChainID string `json:"chainId"`
}

type QueryValidatorAddressesByChainIdRequest struct {
	ChainID string `json:"chainId"`
}

type QueryConvertAddressByChainIdRequest struct {
	ChainID string             `json:"chainId"`
	Prefix  string             `json:"prefix"`
	Address wasmx.Bech32String `json:"address"`
	Type    string             `json:"type"`
}

// Register default subchain
type RegisterDefaultSubChainRequest struct {
	DenomUnit      string      `json:"denom_unit"`
	BaseDenomUnit  uint32      `json:"base_denom_unit"`
	ChainBaseName  string      `json:"chain_base_name"`
	LevelIndex     uint32      `json:"level_index"`
	InitialBalance sdkmath.Int `json:"initial_balance"`
	GenTxs         [][]byte    `json:"gen_txs"`
}

// Register subchain with full deterministic data
type RegisterSubChainRequest struct {
	Data           consensus.InitSubChainDeterministicRequest `json:"data"`
	GenTxs         [][]byte                                   `json:"genTxs"`
	InitialBalance sdkmath.Int                                `json:"initial_balance"`
}

type RemoveSubChainRequest struct {
	ChainID string `json:"chainId"`
}

type RegisterSubChainValidatorRequest struct {
	ChainID string `json:"chainId"`
	GenTx   []byte `json:"genTx"`
}

type InitSubChainRequest struct {
	ChainID string `json:"chainId"`
}

// CosmosmodGenesisState aggregates module genesis states
type CosmosmodGenesisState struct {
	Staking      stakinglib.GenesisState      `json:"staking"`
	Bank         banklib.GenesisState         `json:"bank"`
	Gov          govmod.GenesisState          `json:"gov"`
	Auth         authlib.GenesisState         `json:"auth"`
	Slashing     slashinglib.GenesisState     `json:"slashing"`
	Distribution distributionlib.GenesisState `json:"distribution"`
}

type ValidatorInfo struct {
	Validator      stakinglib.Validator `json:"validator"`
	OperatorPubkey *wasmx.PublicKey     `json:"operator_pubkey"`
	P2PAddress     string               `json:"p2p_address"`
}
