package lib

// Constants mirroring AssemblyScript config.ts
// Raft and blockchain sizing
const (
	LOG_START          = int64(1)
	STATE_SYNC_BATCH   = 200
	ERROR_INVALID_TX   = "transaction is invalid"
	MaxBlockSizeBytes  = 104857600 // 100MB
	BlockPartSizeBytes = 65536     // 64kB
	MaxBlockPartsCount = (MaxBlockSizeBytes / BlockPartSizeBytes) + 1
)

// Node update types
const (
	NODE_UPDATE_REMOVE = 0
	NODE_UPDATE_ADD    = 1
	NODE_UPDATE_UPDATE = 2
)

// Context/storage keys
const (
	// Context values
	NODE_IPS             = "validatorNodesInfo"
	CURRENT_NODE_ID      = "currentNodeId"
	HEARTBEAT_TIMEOUT    = "heartbeatTimeout"
	ELECTION_TIMEOUT_KEY = "electionTimeout"

	// Persistent/volatile state
	TERM_ID       = "currentTerm"
	VOTED_FOR_KEY = "votedFor"

	// Volatile state on all servers
	LAST_APPLIED = "lastApplied"
	COMMIT_INDEX = "commitIndex"

	// Volatile state on leaders
	MATCH_INDEX_ARRAY = "matchIndex"
	NEXT_INDEX_ARRAY  = "nextIndex"
	VOTE_INDEX_ARRAY  = "voteIndex"

	// Cosmos-specific
	MEMPOOL_KEY  = "mempool"
	MAX_TX_BYTES = "max_tx_bytes"

	// Blockchain
	STATE_KEY = "state"
)
