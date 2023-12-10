package types

import (
	"encoding/base64"
	"strconv"
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
