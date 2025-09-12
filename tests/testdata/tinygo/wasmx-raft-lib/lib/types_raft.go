package lib

import (
	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "raft"

// Raft log types
type LogEntry struct {
	Index    int64  `json:"index"`
	TermID   int32  `json:"termId"`
	LeaderID int32  `json:"leaderId"`
	Data     []byte `json:"data"`
}

type LogEntryAggregate struct {
	Index    int64
	TermID   int32
	LeaderID int32
	Data     blocks.BlockEntry
}

// Transaction types
type Transaction struct {
	From  string       `json:"from"`
	To    string       `json:"to"`
	Funds []wasmx.Coin `json:"funds"`
	Data  []byte       `json:"data"`
	Gas   int64        `json:"gas"`
	Price int64        `json:"price"`
}

type TransactionResponse struct {
	TermID   int32 `json:"termId"`
	LeaderID int32 `json:"leaderId"`
	Index    int64 `json:"index"`
}

// Append entries request/response
type AppendEntry struct {
	TermID       int32               `json:"termId"`
	LeaderID     int32               `json:"leaderId"`
	PrevLogIndex int64               `json:"prevLogIndex"`
	PrevLogTerm  int32               `json:"prevLogTerm"`
	Entries      []LogEntryAggregate `json:"entries"`
	LeaderCommit int64               `json:"leaderCommit"`
	NodeIPs      []p2p.NodeInfo      `json:"nodeIps"`
}

type AppendEntryResponse struct {
	TermID    int32 `json:"termId"`
	Success   bool  `json:"success"`
	LastIndex int64 `json:"lastIndex"`
}

// Voting
type VoteRequest struct {
	TermID       int32 `json:"termId"`
	CandidateID  int32 `json:"candidateId"`
	LastLogIndex int64 `json:"lastLogIndex"`
	LastLogTerm  int32 `json:"lastLogTerm"`
}

type VoteResponse struct {
	TermID      int32 `json:"termId"`
	VoteGranted bool  `json:"voteGranted"`
}

// Node update
type NodeUpdate struct {
	Node  p2p.NodeInfo `json:"node"`
	Index int32        `json:"index"`
	Type  int32        `json:"type"`
}

type UpdateNodeResponse struct {
	Nodes          []p2p.NodeInfo `json:"nodes"`
	SyncNodeID     int32          `json:"sync_node_id"`
	LastEntryIndex int64          `json:"last_entry_index"`
}
