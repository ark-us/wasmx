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
    StartIndex  int64  `json:"start_index"`
    PeerAddress string `json:"peer_address"`
}

// StateSyncResponse mirrors the AS response with batch indexes and entries
type StateSyncResponse struct {
    StartBatchIndex int64                       `json:"start_batch_index"`
    LastBatchIndex  int64                       `json:"last_batch_index"`
    LastLogIndex    int64                       `json:"last_log_index"`
    TrustedLogIndex int64                       `json:"trusted_log_index"`
    TrustedLogHash  []byte                      `json:"trusted_log_hash"`
    TermID          int32                       `json:"termId"`
    PeerAddress     string                      `json:"peer_address"`
    Entries         []raftlib.LogEntryAggregate `json:"entries"`
}
