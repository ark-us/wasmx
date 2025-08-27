package consensus

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "wasmx_consensus"

// Protocol/version constants
const ABCISemVer = "2.0.0"
const ABCIVersion = ABCISemVer
const P2PProtocol uint64 = 8
const BlockProtocol uint64 = 11

// Type URLs
const TypeUrl_ExtensionOptionEthereumTx = "/mythos.wasmx.v1.ExtensionOptionEthereumTx"
const TypeUrl_ExtensionOptionAtomicMultiChainTx = "/mythos.network.v1.ExtensionOptionAtomicMultiChainTx"
const TypeUrl_ExtensionOptionMultiChainTx = "/mythos.network.v1.ExtensionOptionMultiChainTx"

type ResponseWrap struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type VersionConsensus struct {
	Block uint64 `json:"block"`
	App   uint64 `json:"app"`
}

type Version struct {
	Consensus VersionConsensus `json:"consensus"`
	Software  string           `json:"software"`
}

type PartSetHeader struct {
	Total uint32          `json:"total"`
	Hash  wasmx.HexString `json:"hash"`
}

type BlockID struct {
	Hash  wasmx.HexString `json:"hash"`
	Parts PartSetHeader   `json:"parts"`
}

type BlockIDProto struct {
	Hash          []byte        `json:"hash"`
	PartSetHeader PartSetHeader `json:"part_set_header"`
}

type Header struct {
	Version            VersionConsensus `json:"version"`
	ChainID            string           `json:"chain_id"`
	Height             int64            `json:"height"`
	Time               string           `json:"time"`
	LastBlockID        BlockID          `json:"last_block_id"`
	LastCommitHash     wasmx.HexString  `json:"last_commit_hash"`
	DataHash           wasmx.HexString  `json:"data_hash"`
	ValidatorsHash     wasmx.HexString  `json:"validators_hash"`
	NextValidatorsHash wasmx.HexString  `json:"next_validators_hash"`
	ConsensusHash      wasmx.HexString  `json:"consensus_hash"`
	AppHash            wasmx.HexString  `json:"app_hash"`
	LastResultsHash    wasmx.HexString  `json:"last_results_hash"`
	EvidenceHash       wasmx.HexString  `json:"evidence_hash"`
	ProposerAddress    wasmx.HexString  `json:"proposer_address"`
}

// BlockIDFlag corresponds to the vote flag
type BlockIDFlag int32

const (
	BlockIDFlagUnknown BlockIDFlag = 0
	BlockIDFlagAbsent  BlockIDFlag = 1
	BlockIDFlagCommit  BlockIDFlag = 2
	BlockIDFlagNil     BlockIDFlag = 3
)

type Validator struct {
	Address []byte `json:"address"`
	Power   int64  `json:"power"`
}

type ExtendedVoteInfo struct {
	Validator          Validator   `json:"validator"`
	VoteExtension      []byte      `json:"vote_extension"`
	ExtensionSignature []byte      `json:"extension_signature"`
	BlockIDFlag        BlockIDFlag `json:"block_id_flag"`
}

type ExtendedCommitInfo struct {
	Round int64              `json:"round"`
	Votes []ExtendedVoteInfo `json:"votes"`
}

type CommitSig struct {
	BlockIDFlag      BlockIDFlag     `json:"block_id_flag"`
	ValidatorAddress wasmx.HexString `json:"validator_address"`
	// Timestamp is optional; keep as pointer to allow null
	Timestamp *string `json:"timestamp,omitempty"`
	Signature []byte  `json:"signature"`
}

type BlockCommit struct {
	Height     int64       `json:"height"`
	Round      int64       `json:"round"`
	BlockID    BlockID     `json:"block_id"`
	Signatures []CommitSig `json:"signatures"`
}

type CanonicalVote struct {
	Type      int32        `json:"type"`
	Height    int64        `json:"height"`
	Round     int64        `json:"round"`
	BlockID   BlockIDProto `json:"block_id"`
	Timestamp string       `json:"timestamp"`
	ChainID   string       `json:"chain_id"`
}

type VoteTendermint struct {
	Type             int32        `json:"type"`
	Height           int64        `json:"height"`
	Round            int64        `json:"round"`
	BlockID          BlockIDProto `json:"block_id"`
	Timestamp        string       `json:"timestamp"`
	ValidatorAddress []byte       `json:"validator_address"`
	ValidatorIndex   int32        `json:"validator_index"`
}

type Misbehavior struct{}

type EvidenceData struct {
	Evidence []Evidence `json:"evidence"`
}

type Evidence struct{}

type RequestPrepareProposal struct {
	MaxTxBytes         int64              `json:"max_tx_bytes"`
	Txs                [][]byte           `json:"txs"`
	LocalLastCommit    ExtendedCommitInfo `json:"local_last_commit"`
	Misbehavior        []Misbehavior      `json:"misbehavior"`
	Height             int64              `json:"height"`
	Time               string             `json:"time"`
	NextValidatorsHash []byte             `json:"next_validators_hash"`
	ProposerAddress    wasmx.HexString    `json:"proposer_address"`
}

type ResponsePrepareProposal struct {
	Txs [][]byte `json:"txs"`
}

type VoteInfo struct {
	Validator   Validator   `json:"validator"`
	BlockIDFlag BlockIDFlag `json:"block_id_flag"`
}

type CommitInfo struct {
	Round int64      `json:"round"`
	Votes []VoteInfo `json:"votes"`
}

type RequestProcessProposal struct {
	Txs                [][]byte        `json:"txs"`
	ProposedLastCommit CommitInfo      `json:"proposed_last_commit"`
	Misbehavior        []Misbehavior   `json:"misbehavior"`
	Hash               []byte          `json:"hash"`
	Height             int64           `json:"height"`
	Time               string          `json:"time"`
	NextValidatorsHash []byte          `json:"next_validators_hash"`
	ProposerAddress    wasmx.HexString `json:"proposer_address"`
}

type ProposalStatus int32

const (
	ProposalStatus_UNKNOWN ProposalStatus = 0
	ProposalStatus_ACCEPT  ProposalStatus = 1
	ProposalStatus_REJECT  ProposalStatus = 2
)

type ResponseProcessProposal struct {
	Status ProposalStatus `json:"status"`
}

type ResponseOptimisticExecution struct {
	Metainfo map[string][]byte `json:"metainfo"`
}

type RequestFinalizeBlock struct {
	Txs                [][]byte        `json:"txs"`
	DecidedLastCommit  CommitInfo      `json:"decided_last_commit"`
	Misbehavior        []Misbehavior   `json:"misbehavior"`
	Hash               []byte          `json:"hash"`
	Height             int64           `json:"height"`
	Time               string          `json:"time"`
	NextValidatorsHash []byte          `json:"next_validators_hash"`
	ProposerAddress    wasmx.HexString `json:"proposer_address"`
}

type WrapRequestFinalizeBlock struct {
	Request  RequestFinalizeBlock `json:"request"`
	Metainfo map[string][]byte    `json:"metainfo"`
}

type RequestProcessProposalWithMetaInfo struct {
	Request             RequestProcessProposal `json:"request"`
	OptimisticExecution bool                   `json:"optimistic_execution"`
	Metainfo            map[string][]byte      `json:"metainfo"`
}

// ExecTxResultExternal mirrors ExecTxResult but stringifies gas fields
type ExecTxResultExternal struct {
	Code      uint32        `json:"code"`
	Data      []byte        `json:"data"`
	Log       string        `json:"log"`
	Info      string        `json:"info"`
	GasWanted string        `json:"gas_wanted"`
	GasUsed   string        `json:"gas_used"`
	Events    []wasmx.Event `json:"events"`
	Codespace string        `json:"codespace"`
}

type ExecTxResult struct {
	Code      uint32        `json:"code"`
	Data      []byte        `json:"data"`
	Log       string        `json:"log"`
	Info      string        `json:"info"`
	GasWanted int64         `json:"gas_wanted"`
	GasUsed   int64         `json:"gas_used"`
	Events    []wasmx.Event `json:"events"`
	Codespace string        `json:"codespace"`
}

func (e ExecTxResult) MarshalJSON() ([]byte, error) {
	ext := ExecTxResultExternal{
		Code:      e.Code,
		Data:      e.Data,
		Log:       e.Log,
		Info:      e.Info,
		GasWanted: fmt.Sprintf("%d", e.GasWanted),
		GasUsed:   fmt.Sprintf("%d", e.GasUsed),
		Events:    e.Events,
		Codespace: e.Codespace,
	}
	return json.Marshal(ext)
}

func (e *ExecTxResult) UnmarshalJSON(b []byte) error {
	var ext ExecTxResultExternal
	if err := json.Unmarshal(b, &ext); err != nil {
		return err
	}
	var gw, gu int64
	if ext.GasWanted != "" {
		if _, err := fmt.Sscan(ext.GasWanted, &gw); err != nil {
			return fmt.Errorf("parse gas_wanted: %w", err)
		}
	}
	if ext.GasUsed != "" {
		if _, err := fmt.Sscan(ext.GasUsed, &gu); err != nil {
			return fmt.Errorf("parse gas_used: %w", err)
		}
	}
	e.Code = ext.Code
	e.Data = ext.Data
	e.Log = ext.Log
	e.Info = ext.Info
	e.GasWanted = gw
	e.GasUsed = gu
	e.Events = ext.Events
	e.Codespace = ext.Codespace
	return nil
}

type ExtensionOptionMultiChainTx struct {
	ChainID string `json:"chain_id"`
	Index   int32  `json:"index"`
	TxCount int32  `json:"tx_count"`
}

type ExtensionOptionAtomicMultiChainTx struct {
	LeaderChainID string   `json:"leader_chain_id"`
	ChainIDs      []string `json:"chain_ids"`
}

func ExtensionOptionAtomicMultiChainTxFromAnyWrap(value wasmx.AnyWrap) (*ExtensionOptionAtomicMultiChainTx, error) {
	bz, err := base64.StdEncoding.DecodeString(value.Value)
	if err != nil {
		return nil, err
	}
	var out ExtensionOptionAtomicMultiChainTx
	if err := json.Unmarshal(bz, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ExtensionOptionEthereumTx struct{}

type ValidatorUpdate struct {
	PubKey *wasmx.PublicKey `json:"pub_key"`
	Power  int64            `json:"power"`
}

type ValidatorInfo struct {
	Address          wasmx.HexString `json:"address"`
	PubKey           []byte          `json:"pub_key"`
	VotingPower      int64           `json:"voting_power"`
	ProposerPriority int64           `json:"proposer_priority"`
}

type TendermintValidator struct {
	OperatorAddress  wasmx.Bech32String `json:"operator_address"`
	HexAddress       wasmx.HexString    `json:"hex_address"`
	PubKey           *wasmx.PublicKey   `json:"pub_key"`
	VotingPower      int64              `json:"voting_power"`
	ProposerPriority int64              `json:"proposer_priority"`
}

type TendermintValidators struct {
	Validators []TendermintValidator `json:"validators"`
}

type BlockParams struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
}

type EvidenceParams struct {
	MaxAgeNumBlocks int64 `json:"max_age_num_blocks"`
	MaxAgeDuration  int64 `json:"max_age_duration"`
	MaxBytes        int64 `json:"max_bytes"`
}

type ValidatorParams struct {
	PubKeyTypes []string `json:"pub_key_types"`
}

type VersionParams struct {
	App uint64 `json:"app"`
}

type ABCIParams struct {
	VoteExtensionsEnableHeight int64 `json:"vote_extensions_enable_height"`
}

type ConsensusParams struct {
	Block     BlockParams     `json:"block"`
	Evidence  EvidenceParams  `json:"evidence"`
	Validator ValidatorParams `json:"validator"`
	Version   VersionParams   `json:"version"`
	ABCI      ABCIParams      `json:"abci"`
}

type ResponseFinalizeBlock struct {
	Events                []wasmx.Event     `json:"events"`
	TxResults             []ExecTxResult    `json:"tx_results"`
	ValidatorUpdates      []ValidatorUpdate `json:"validator_updates"`
	ConsensusParamUpdates *ConsensusParams  `json:"consensus_param_updates"`
	AppHash               []byte            `json:"app_hash"`
}

type ResponseFinalizeBlockWrap struct {
	Error string                 `json:"error"`
	Data  *ResponseFinalizeBlock `json:"data"`
}

type ResponseBeginBlock struct {
	Events []wasmx.Event `json:"events"`
}

type ResponseBeginBlockWrap struct {
	Error string              `json:"error"`
	Data  *ResponseBeginBlock `json:"data"`
}

type ResponseCommit struct {
	RetainHeight int64 `json:"retainHeight"`
}

type CheckTxType int32

const (
	CheckTxTypeNew     CheckTxType = 0
	CheckTxTypeRecheck CheckTxType = 1
)

type RequestCheckTx struct {
	Tx   []byte      `json:"tx"`
	Type CheckTxType `json:"type"`
}

type CodeType uint32

const (
	CodeTypeOk              CodeType = 0
	CodeTypeEncodingError   CodeType = 1
	CodeTypeInvalidTxFormat CodeType = 2
	CodeTypeUnauthorized    CodeType = 3
	CodeTypeUnused          CodeType = 4
	CodeTypeExecuted        CodeType = 5
)

type ResponseCheckTx struct {
	Code      uint32        `json:"code"`
	Data      []byte        `json:"data"`
	Log       string        `json:"log"`
	Info      string        `json:"info"`
	GasWanted int64         `json:"gas_wanted"`
	GasUsed   int64         `json:"gas_used"`
	Events    []wasmx.Event `json:"events"`
	Codespace string        `json:"codespace"`
}

type Transaction struct {
	Gas int32 `json:"gas"`
}

type TxResult struct {
	Height int64        `json:"height"`
	Index  uint32       `json:"index"`
	Tx     []byte       `json:"tx"`
	Result ExecTxResult `json:"result"`
}

type RequestApplySnapshotChunk struct{}
type ResponseApplySnapshotChunk struct{}
type RequestLoadSnapshotChunk struct{}
type ResponseLoadSnapshotChunk struct{}
type RequestOfferSnapshot struct{}
type ResponseOfferSnapshot struct{}
type RequestListSnapshots struct{}
type ResponseListSnapshots struct{}

type ValidatorSet struct {
	Validators []TendermintValidator `json:"validators"`
	Proposer   TendermintValidator   `json:"proposer"`
}

type State struct {
	Version                          Version         `json:"Version"`
	ChainID                          string          `json:"ChainID"`
	InitialHeight                    int64           `json:"InitialHeight"`
	LastBlockHeight                  int64           `json:"LastBlockHeight"`
	LastBlockID                      BlockID         `json:"LastBlockID"`
	LastBlockTime                    string          `json:"LastBlockTime"`
	NextValidators                   []byte          `json:"NextValidators"`
	Validators                       []byte          `json:"Validators"`
	LastValidators                   []byte          `json:"LastValidators"`
	LastHeightValidatorsChanged      int64           `json:"LastHeightValidatorsChanged"`
	ConsensusParams                  ConsensusParams `json:"ConsensusParams"`
	LastHeightConsensusParamsChanged int64           `json:"LastHeightConsensusParamsChanged"`
	LastResultsHash                  []byte          `json:"LastResultsHash"`
	AppHash                          []byte          `json:"AppHash"`
}
