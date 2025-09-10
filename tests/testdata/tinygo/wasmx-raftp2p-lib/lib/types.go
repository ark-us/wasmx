package lib

import (
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
)

// Module identification and protocol constants
const (
	MODULE_NAME = "raftp2p"
	PROTOCOL_ID = "raftp2p_1"
)

// StateSyncRequest mirrors the AS request with start index
type StateSyncRequest struct {
	StartIndex int64 `json:"start_index"`
}

// StateSyncResponse mirrors the AS response with batch indexes and entries
type StateSyncResponse struct {
	StartBatchIndex int64                       `json:"start_batch_index"`
	LastBatchIndex  int64                       `json:"last_batch_index"`
	LastLogIndex    int64                       `json:"last_log_index"`
	TermID          int32                       `json:"termId"`
	Entries         []raftlib.LogEntryAggregate `json:"entries"`
}
