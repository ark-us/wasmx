package types

import (
	"encoding/base64"
	"fmt"
	"strconv"

	cmtlibs "github.com/cometbft/cometbft/libs/bytes"
	cmttypes "github.com/cometbft/cometbft/types"
)

var BLOCK_LAST_INDEX = "block_last_index"
var BLOCK_INDEX_KEY = "block_"
var BLOCK_HASH_KEY = "block_by_hash_"
var TX_INDEXER = "tx_"
var PARAMS_KEY = "consensus_params"
var DATA_INDEXER = "data_"
var INTERVAL_ID_KEY = "intervalIdKey"
var BLOCK_DELAY = "blockTimeout"
var VALIDATORS_KEY = "validators"
var STATE_KEY = "state"
var FSM_CONTEXT_KEY = "context_"

type IndexedTransaction struct {
	Height int64  `json:"height"`
	Index  uint32 `json:"index"`
}

type LogEntry struct {
	Index    int64  `json:"index"`
	TermId   int32  `json:"termId"`
	LeaderId int32  `json:"leaderId"`
	Data     []byte `json:"data"` // empty for finalized blocks
}

type BlockEntry struct {
	Index          int64  `json:"index"`
	ReaderContract []byte `json:"readerContract"`
	WriterContract []byte `json:"writerContract"`
	Data           []byte `json:"data"`        // RequestProcessProposal
	Header         []byte `json:"header"`      // Block Header
	LastCommit     []byte `json:"last_commit"` // BlockCommit
	Evidence       []byte `json:"evidence"`    // EvidenceData
	Result         []byte `json:"result"`      // ResponseFinalizeBlock
}

func GetTopicDataKey(topic string) string {
	return DATA_INDEXER + topic
}

func GetTxKey(hash []byte) string {
	return TX_INDEXER + base64.StdEncoding.EncodeToString(hash)
}

func GetBlockKey(index int64) string {
	return BLOCK_INDEX_KEY + strconv.Itoa(int(index))
}

func GetLastBlockHeightKey() string {
	return BLOCK_LAST_INDEX
}

func GetLastIntervalIdKey() string {
	return INTERVAL_ID_KEY
}

func GetLastIntervalIdByStateKey(state string, delay string) string {
	return fmt.Sprintf(`%s_%s_%s`, INTERVAL_ID_KEY, state, delay)
}

type InitChainSetup struct {
	ChainID          string                    `json:"chain_id"`
	ConsensusParams  *cmttypes.ConsensusParams `json:"consensus_params"`
	AppHash          []byte                    `json:"app_hash"`
	LastResultsHash  []byte                    `json:"last_results_hash"`
	Version          Version                   `json:"version"`
	ValidatorAddress cmtlibs.HexBytes          `json:"validator_address"`
	ValidatorPrivKey []byte                    `json:"validator_privkey"`
	ValidatorPubKey  []byte                    `json:"validator_pubkey"`
	Peers            []string                  `json:"peers"`
}

type Consensus struct {
	Block uint64 `json:"block"`
	App   uint64 `json:"app"`
}

type Version struct {
	Consensus Consensus `json:"consensus"`
	Software  string    `json:"software"`
}
