package lib

import (
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
)

// Module identification and protocol constants
const (
	MODULE_NAME = "kayrosp2p"
	PROTOCOL_ID = "kayrosp2p_1"
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

// KayrosConfig holds the configuration for Kayros API client
type KayrosConfig struct {
	ApiBaseUrl string `json:"api_base_url"` // e.g., "https://kayros.provable.dev"
}

// InstantiateMsg is the message passed during contract instantiation
type InstantiateMsg struct {
	KayrosApiUrl      string `json:"kayros_api_url"`       // e.g., "https://kayros.provable.dev"
	GenesisUUID       string `json:"genesis_uuid"`         // First UUID to start from on genesis
	ThresholdCommit   int    `json:"threshold_commit"`     // Percentage threshold for commit (e.g., 51) - determines when to rollback
	ThresholdFinalize int    `json:"threshold_finalize"`   // Percentage threshold for finalization (e.g., 75) - determines when block is finalized
}

// KayrosRecord represents a record in the Kayros database
type KayrosRecord struct {
	DataType    string `json:"data_type"`
	DataTypeHex string `json:"data_type_hex"`
	DataItemHex string `json:"data_item_hex"`
	UuidHex     string `json:"uuid_hex"`
	HashItemHex string `json:"hash_item_hex"` // Kayros unique ID
	PrevHashHex string `json:"prev_hash_hex,omitempty"` // Optional - may not be in all responses
	HashType    string `json:"hash_type"`
	Timestamp   string `json:"timestamp"`
}

// KayrosApiResponse is the common response structure from Kayros API
type KayrosApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// KayrosRecordResponse is the response for single record queries
type KayrosRecordResponse struct {
	KayrosApiResponse
	Data KayrosRecord `json:"data"`
}

// KayrosRecordsData is the data structure for multiple record queries
type KayrosRecordsData struct {
	Count   int            `json:"count"`
	Limit   int            `json:"limit"`
	Records []KayrosRecord `json:"records"`
}

// KayrosRecordsResponse is the response for multiple record queries
type KayrosRecordsResponse struct {
	KayrosApiResponse
	Data KayrosRecordsData `json:"data"`
}

// BlockCommit represents a commit from a validator for a specific block
type BlockCommit struct {
	BlockNumber int64  `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	Sender      string `json:"sender"`
	Timestamp   int64  `json:"timestamp"` // Timestamp from the sending node
	Signature   string `json:"signature"`
}

// BlockCommitMapping stores all commits for a specific block from various validators
type BlockCommitMapping struct {
	BlockNumber int64                  `json:"block_number"`
	Commits     map[string]BlockCommit `json:"commits"` // Map from sender to their commit (allows overwriting with newer timestamp)
	Finalized   bool                   `json:"finalized"`
}

// ConsensusConfig holds the consensus thresholds
type ConsensusConfig struct {
	ThresholdCommit   int    `json:"threshold_commit"`   // Percentage for commit threshold (e.g., 51)
	ThresholdFinalize int    `json:"threshold_finalize"` // Percentage for finalization threshold (e.g., 75)
	GenesisUUID       string `json:"genesis_uuid"`       // First UUID on genesis
}

// MissingTransactionsRequest is sent to peers to request missing transactions
type MissingTransactionsRequest struct {
	TxHashes    []string `json:"tx_hashes"`    // List of transaction hashes we need
	BlockNumber int64    `json:"block_number"` // The block number we're building
	Sender      string   `json:"sender"`       // The peer requesting
}

// MissingTransactionsResponse contains the requested transactions
type MissingTransactionsResponse struct {
	Transactions [][]byte `json:"transactions"` // Raw transaction bytes
	TxHashes     []string `json:"tx_hashes"`    // Corresponding hashes
	Sender       string   `json:"sender"`       // The peer responding
}

// KayrosRegistrationRequest is the request body for POST /api/grpc/single-hash
type KayrosRegistrationRequest struct {
	DataType string `json:"data_type"`
	DataItem string `json:"data_item"` // Transaction hash
}

// KayrosRegistrationResponse is the response from the registration API
type KayrosRegistrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UUID    string `json:"uuid,omitempty"`
}

// KayrosTxMetadata stores Kayros ordering information for a transaction
// This metadata is included in the block's Metainfo field
type KayrosTxMetadata struct {
	TxHash      string `json:"tx_hash"`       // Transaction hash (data_item)
	UUID        string `json:"uuid"`          // Kayros UUID for ordering
	HashItem    string `json:"hash_item"`     // Kayros unique ID (hash_item_hex)
	Timestamp   string `json:"timestamp"`     // Kayros timestamp
	PrevHash    string `json:"prev_hash"`     // Previous hash in the ordering chain
	BlockNumber int64  `json:"block_number"`  // Block this tx is included in
}

// KayrosBlockMetadata contains all Kayros metadata for transactions in a block
type KayrosBlockMetadata struct {
	Transactions []KayrosTxMetadata `json:"transactions"`
	FirstUUID    string             `json:"first_uuid"`  // First UUID in this block
	LastUUID     string             `json:"last_uuid"`   // Last UUID in this block
}
