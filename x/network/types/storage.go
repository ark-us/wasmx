package types

import (
	"encoding/base64"
	"strconv"

	cmtlibs "github.com/cometbft/cometbft/libs/bytes"
	cmttypes "github.com/cometbft/cometbft/types"
)

var LOG_INDEXER = "logs_"
var TX_INDEXER = "tx_"

type IndexedTransaction struct {
	Height int64  `json:"height"`
	Index  uint32 `json:"index"`
}

type LogEntry struct {
	Index    int64  `json:"index"`
	TermId   int32  `json:"termId"`
	LeaderId int32  `json:"leaderId"`
	Data     []byte `json:"data"`   // RequestProcessProposal
	Result   []byte `json:"result"` // ResponseFinalizeBlock
}

func GetTxKey(hash []byte) string {
	return TX_INDEXER + base64.StdEncoding.EncodeToString(hash)
}

func GetBlockKey(index int64) string {
	return LOG_INDEXER + strconv.Itoa(int(index))
}

type InitChainSetup struct {
	ChainID          string                    `json:"chain_id"`
	ConsensusParams  *cmttypes.ConsensusParams `json:"consensus_params"`
	Validators       []*cmttypes.Validator     `json:"validators"`
	AppHash          []byte                    `json:"app_hash"`
	LastResultsHash  []byte                    `json:"last_results_hash"`
	CurrentValidator cmtlibs.HexBytes          `json:"current_validator"`
	Version          Version                   `json:"version"`
}

type Consensus struct {
	Block uint64 `json:"block"`
	App   uint64 `json:"app"`
}

type Version struct {
	Consensus Consensus `json:"consensus"`
	Software  string    `json:"software"`
}
