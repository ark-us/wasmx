package lib

import (
	sdkmath "cosmossdk.io/math"
	consensus "github.com/loredanacirstea/wasmx-env-consensus/lib"
	registrylib "github.com/loredanacirstea/wasmx-multichain-registry/lib"
)

const MODULE_NAME = "lobby"

// NetworkNode represents a network node (placeholder - needs to be defined based on actual p2p types)
type NetworkNode struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	PeerID  string `json:"peer_id"`
}

// NodeInfo represents node information (placeholder - needs to be defined based on actual p2p types)
type NodeInfo struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

// Params represents lobby module parameters
type Params struct {
	CurrentLevel        int32       `json:"current_level"`
	MinValidatorsCount  int32       `json:"min_validators_count"`
	EnableEidCheck      bool        `json:"enable_eid_check"`
	Erc20CodeId         uint64      `json:"erc20CodeId"`
	Derc20CodeId        uint64      `json:"derc20CodeId"`
	LevelInitialBalance sdkmath.Int `json:"level_initial_balance"`
}

type MsgLastChainId struct {
	ID consensus.ChainId `json:"id"`
}

type MsgLastNodeId struct {
	ID string `json:"id"`
}

type PotentialValidator struct {
	Node               NetworkNode `json:"node"`
	AddressBytes       string      `json:"addressBytes"`       // base64 string
	ConsensusPublicKey string      `json:"consensusPublicKey"` // base64 string
}

type PotentialValidatorWithSignature struct {
	Validator PotentialValidator `json:"validator"`
	Signature string             `json:"signature"`
}

type MsgNewChainRequest struct {
	Level     int32              `json:"level"`
	Validator PotentialValidator `json:"validator"`
}

type MsgNewChainAccepted struct {
	Level      int32                `json:"level"`
	ChainId    consensus.ChainId    `json:"chainId"`
	Validators []PotentialValidator `json:"validators"`
}

type MsgNewChainResponse struct {
	Msg        MsgNewChainAccepted `json:"msg"`
	Signatures []string            `json:"signatures"` // base64 strings
}

type MsgNewChainGenesisData struct {
	Data       registrylib.SubChainData `json:"data"`
	Validators []PotentialValidator     `json:"validators"`
	Signatures []string                 `json:"signatures"` // signature on data.data, base64 strings
}

type CurrentChainSetup struct {
	Data consensus.InitChainSetup `json:"data"`
	Node NodeInfo                 `json:"node"`
}

type ChainConfigData struct {
	ChainId     string                `json:"chain_id"`
	ChainConfig consensus.ChainConfig `json:"chain_config"`
}

// Calldata structure
type CallData struct {
	GetParams *MsgGetParams `json:"GetParams"`

	InitGenesis *MsgInitGenesis `json:"InitGenesis"`
}

type MsgInitGenesis struct{}

type MsgGetParams struct{}
