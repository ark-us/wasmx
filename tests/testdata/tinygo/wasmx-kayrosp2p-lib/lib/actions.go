package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
)

// Chat room topics (parity with AS config)
const (
	CHAT_ROOM_PROTOCOL           = "kayros_p2p_protocol"
	CHAT_ROOM_MEMPOOL            = CHAT_ROOM_PROTOCOL
	STORAGE_KEY_KAYROS_CONFIG    = "kayros_config"
	STORAGE_KEY_LAST_UUID        = "kayros_last_uuid"
	STORAGE_KEY_CONSENSUS_CONFIG = "consensus_config"
	STORAGE_KEY_BLOCK_COMMITS    = "block_commits_"  // Prefix for block commit mappings
	STORAGE_KEY_PENDING_TXS      = "pending_txs"     // Transactions needed for current block
	STORAGE_KEY_KAYROS_TX_META   = "kayros_tx_meta_" // Prefix for Kayros tx metadata
	STORAGE_KEY_BLOCK_KAYROS     = "block_kayros_"   // Prefix for block Kayros metadata
	MAX_TXS_PER_BLOCK            = 100               // Default max transactions to fetch from Kayros per block

	CTX_KAYROS_BASE_URL    = "kayros_base_url"
	CTX_KAYROS_USER_KEY    = "kayros_user_key"
	CTX_DATA_TYPE_ID       = "data_type_id"
	CTX_THRESHOLD_COMMIT   = "threshold_commit"
	CTX_THRESHOLD_FINALIZE = "threshold_finalize"
	CTX_GENESIS_UUID       = "genesis_uuid"
)

// GetKayrosClient retrieves the Kayros client, initializing if needed
func GetKayrosClient() (*KayrosClient, error) {
	config := KayrosConfig{
		ApiBaseUrl: sGet(CTX_KAYROS_BASE_URL),
	}
	kayrosClient := NewKayrosClient(config)
	return kayrosClient, nil
}

// GetLastKayrosUUID retrieves the last processed Kayros UUID from state
func GetLastKayrosUUID() string {
	data := sGet(STORAGE_KEY_LAST_UUID)
	return string(data)
}

// SetLastKayrosUUID stores the last processed Kayros UUID in state
func SetLastKayrosUUID(uuid string) {
	sSet(STORAGE_KEY_LAST_UUID, uuid)
}

// GetConsensusConfig retrieves the consensus configuration from state
func GetConsensusConfig() (*ConsensusConfig, error) {
	thresholdCommit := 51   // default
	thresholdFinalize := 75 // default

	commitStr := sGet(CTX_THRESHOLD_COMMIT)
	if commitStr != "" {
		if val, err := fsm.ParseInt32(commitStr); err == nil {
			thresholdCommit = int(val)
		}
	}

	finalizeStr := sGet(CTX_THRESHOLD_FINALIZE)
	if finalizeStr != "" {
		if val, err := fsm.ParseInt32(finalizeStr); err == nil {
			thresholdFinalize = int(val)
		}
	}

	config := ConsensusConfig{
		ThresholdCommit:   thresholdCommit,
		ThresholdFinalize: thresholdFinalize,
		GenesisUUID:       sGet(CTX_GENESIS_UUID),
	}
	return &config, nil
}

// GetBlockCommitMapping retrieves the commit mapping for a specific block
func GetBlockCommitMapping(blockNumber int64) (*BlockCommitMapping, error) {
	key := fmt.Sprintf("%s%d", STORAGE_KEY_BLOCK_COMMITS, blockNumber)
	data := wasmx.StorageLoad([]byte(key))
	if len(data) == 0 {
		// Return empty mapping if not found
		return &BlockCommitMapping{
			BlockNumber: blockNumber,
			Commits:     make(map[string]BlockCommit),
			Finalized:   false,
		}, nil
	}
	var mapping BlockCommitMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

// SetBlockCommitMapping stores the commit mapping for a specific block
func SetBlockCommitMapping(mapping *BlockCommitMapping) error {
	key := fmt.Sprintf("%s%d", STORAGE_KEY_BLOCK_COMMITS, mapping.BlockNumber)
	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	wasmx.StorageStore([]byte(key), data)
	return nil
}

// GetPendingTxHashes retrieves the list of transaction hashes needed for the current block
func GetPendingTxHashes() ([]string, error) {
	data := wasmx.StorageLoad([]byte(STORAGE_KEY_PENDING_TXS))
	if len(data) == 0 {
		return []string{}, nil
	}
	var hashes []string
	if err := json.Unmarshal(data, &hashes); err != nil {
		return nil, err
	}
	return hashes, nil
}

// SetPendingTxHashes stores the list of transaction hashes needed for the current block
func SetPendingTxHashes(hashes []string) error {
	data, err := json.Marshal(hashes)
	if err != nil {
		return err
	}
	wasmx.StorageStore([]byte(STORAGE_KEY_PENDING_TXS), data)
	return nil
}

// GetKayrosTxMetadata retrieves Kayros metadata for a specific transaction
func GetKayrosTxMetadata(txHash string) (*KayrosTxMetadata, error) {
	key := fmt.Sprintf("%s%s", STORAGE_KEY_KAYROS_TX_META, txHash)
	data := wasmx.StorageLoad([]byte(key))
	if len(data) == 0 {
		return nil, nil
	}
	var meta KayrosTxMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// SetKayrosTxMetadata stores Kayros metadata for a specific transaction
func SetKayrosTxMetadata(meta *KayrosTxMetadata) error {
	key := fmt.Sprintf("%s%s", STORAGE_KEY_KAYROS_TX_META, meta.TxHash)
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	wasmx.StorageStore([]byte(key), data)
	return nil
}

// GetBlockKayrosMetadata retrieves Kayros metadata for all transactions in a block
func GetBlockKayrosMetadata(blockNumber int64) (*KayrosBlockMetadata, error) {
	key := fmt.Sprintf("%s%d", STORAGE_KEY_BLOCK_KAYROS, blockNumber)
	data := wasmx.StorageLoad([]byte(key))
	if len(data) == 0 {
		return nil, nil
	}
	var meta KayrosBlockMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// SetBlockKayrosMetadata stores Kayros metadata for all transactions in a block
func SetBlockKayrosMetadata(blockNumber int64, meta *KayrosBlockMetadata) error {
	key := fmt.Sprintf("%s%d", STORAGE_KEY_BLOCK_KAYROS, blockNumber)
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	wasmx.StorageStore([]byte(key), data)
	return nil
}

// BuildBlockKayrosMetadata collects Kayros metadata for all transactions in the pending list
func BuildBlockKayrosMetadata(blockNumber int64) (*KayrosBlockMetadata, error) {
	pendingHashes, err := GetPendingTxHashes()
	if err != nil {
		return nil, err
	}

	blockMeta := &KayrosBlockMetadata{
		Transactions: make([]KayrosTxMetadata, 0, len(pendingHashes)),
	}

	for _, txHash := range pendingHashes {
		txMeta, err := GetKayrosTxMetadata(txHash)
		if err != nil {
			LoggerError("failed to get Kayros metadata for tx", []string{"txHash", txHash, "error", err.Error()})
			continue
		}
		if txMeta != nil {
			txMeta.BlockNumber = blockNumber
			blockMeta.Transactions = append(blockMeta.Transactions, *txMeta)

			// Track first and last UUIDs
			if blockMeta.FirstUUID == "" {
				blockMeta.FirstUUID = txMeta.UUID
			}
			blockMeta.LastUUID = txMeta.UUID
		}
	}

	// Store the block metadata
	if err := SetBlockKayrosMetadata(blockNumber, blockMeta); err != nil {
		return nil, err
	}

	return blockMeta, nil
}

// RegisterWithKayros registers a transaction with Kayros via POST /api/grpc/single-hash
// This action is called when a new transaction is received
func RegisterWithKayros(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract transaction from event params
	txB64 := ""
	for _, p := range event.Params {
		if p.Key == "transaction" {
			txB64 = p.Value
			break
		}
	}
	if txB64 == "" {
		return fmt.Errorf("no transaction found in event")
	}

	// Decode transaction
	txBytes, err := base64.StdEncoding.DecodeString(txB64)
	if err != nil {
		return err
	}

	// Compute transaction hash (hex encoded for Kayros)
	txHashBytes := wasmx.Sha256(txBytes)
	txHash := fmt.Sprintf("%x", txHashBytes)

	// Get Kayros client
	client, err := GetKayrosClient()
	if err != nil {
		LoggerError("failed to get Kayros client", []string{"error", err.Error()})
		// Don't fail the transaction if Kayros is unavailable
		return nil
	}

	// Register transaction with Kayros via POST /api/grpc/single-hash
	resp, err := client.RegisterTransaction(txHash)
	if err != nil {
		LoggerError("failed to register transaction with Kayros", []string{"txHash", txHash, "error", err.Error()})
		// Don't fail the transaction if Kayros registration fails
		return nil
	}

	if !resp.Success {
		LoggerDebug("Kayros registration returned failure", []string{"txHash", txHash, "message", resp.Message})
		return nil
	}

	LoggerInfo("transaction registered with Kayros", []string{
		"txHash", txHash,
		"uuid", resp.UUID,
	})

	return nil
}

// GetKayrosTxs fetches transactions from Kayros for block production
// This action is called periodically by validators to produce blocks
// It first checks the mempool for transactions, then requests missing ones from peers
func GetKayrosTxs(params []fsm.ActionParam, event fsm.EventObject) error {
	client, err := GetKayrosClient()
	if err != nil {
		return err
	}

	// Get the last processed Kayros UUID
	lastUUID := GetLastKayrosUUID()

	// If no last UUID, use the genesis UUID from config
	if lastUUID == "" {
		config, err := GetConsensusConfig()
		if err != nil {
			LoggerError("failed to get consensus config for genesis UUID", []string{"error", err.Error()})
			return err
		}
		lastUUID = config.GenesisUUID
		LoggerInfo("using genesis UUID", []string{"uuid", lastUUID})
	}

	// Fetch records from Kayros starting from the last UUID
	limit := MAX_TXS_PER_BLOCK
	records, err := client.GetRecordsFromPrev(lastUUID, limit)
	if err != nil {
		LoggerError("failed to fetch records from Kayros", []string{"error", err.Error(), "lastUUID", lastUUID})
		return err
	}

	LoggerInfo("fetched records from Kayros", []string{
		"count", fmt.Sprintf("%d", len(records)),
		"lastUUID", lastUUID,
	})

	if len(records) == 0 {
		// No new transactions to process - clear pending
		SetPendingTxHashes([]string{})
		return nil
	}

	// Get mempool to check which transactions we already have
	mp, err := raftlib.GetMempool()
	if err != nil {
		return err
	}

	// Collect transaction hashes we need and ones we're missing
	missingTxHashes := []string{}
	allTxHashes := []string{}

	for _, record := range records {
		txHash := record.DataItemHex
		allTxHashes = append(allTxHashes, txHash)

		// Store Kayros metadata for this transaction
		txMeta := &KayrosTxMetadata{
			TxHash:    txHash,
			UUID:      record.UuidHex,
			HashItem:  record.HashItemHex,
			Timestamp: record.Timestamp,
			PrevHash:  record.PrevHashHex,
		}
		if err := SetKayrosTxMetadata(txMeta); err != nil {
			LoggerError("failed to store Kayros tx metadata", []string{"txHash", txHash, "error", err.Error()})
		}

		// Convert hex hash to base64 for mempool check (mempool uses base64 format)
		txHashBase64, err := HexHashToBase64(txHash)
		if err != nil {
			LoggerError("failed to decode tx hash from Kayros", []string{"txHash", txHash, "error", err.Error()})
			continue
		}

		// Check if we have this transaction in our mempool
		if mp.HasSeen(txHashBase64) {
			LoggerDebug("transaction in mempool", []string{"txHash", txHash, "kayrosId", record.HashItemHex})
			continue
		}

		// Transaction not in mempool - add to missing list
		missingTxHashes = append(missingTxHashes, txHash)
		LoggerDebug("transaction not in mempool, will request from peers", []string{
			"txHash", txHash,
			"kayrosId", record.HashItemHex,
			"uuid", record.UuidHex,
			"timestamp", record.Timestamp,
		})
	}

	// Store the list of all transactions needed for this block
	if err := SetPendingTxHashes(allTxHashes); err != nil {
		LoggerError("failed to store pending tx hashes", []string{"error", err.Error()})
	}

	// If we have missing transactions, request them from peers
	if len(missingTxHashes) > 0 {
		LoggerInfo("requesting missing transactions from peers", []string{"count", fmt.Sprintf("%d", len(missingTxHashes))})
		if err := RequestMissingTransactions(missingTxHashes); err != nil {
			LoggerError("failed to request missing transactions", []string{"error", err.Error()})
			// Don't return error - we'll retry on next round
		}
	}

	// Note: LastKayrosUUID is updated in CommitBlock after successful block commit
	// This ensures we don't skip transactions if block commit fails
	return nil
}

// ReceiveCommit handles receiving a commit from another validator
// In Kayros consensus, validators send block hashes to verify consensus
// Until a block is finalized, we may receive multiple commits from the same node
// so we use timestamps to overwrite previous commits if a newer one comes in
func ReceiveCommit(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract commit data from event params
	commitB64 := ""
	sender := ""
	for _, p := range event.Params {
		if p.Key == "commit" {
			commitB64 = p.Value
		} else if p.Key == "sender" {
			sender = p.Value
		}
	}

	if commitB64 == "" {
		return fmt.Errorf("no commit found in event")
	}

	// Decode commit data
	commitData, err := base64.StdEncoding.DecodeString(commitB64)
	if err != nil {
		return err
	}

	// Parse the commit
	var commit BlockCommit
	if err := json.Unmarshal(commitData, &commit); err != nil {
		LoggerError("failed to parse commit", []string{"error", err.Error()})
		return err
	}

	// Set sender from event if not in commit data
	if commit.Sender == "" {
		commit.Sender = sender
	}

	LoggerDebug("received commit", []string{
		"sender", commit.Sender,
		"blockNumber", fmt.Sprintf("%d", commit.BlockNumber),
		"blockHash", commit.BlockHash,
		"timestamp", fmt.Sprintf("%d", commit.Timestamp),
	})

	// Get the commit mapping for this block
	mapping, err := GetBlockCommitMapping(commit.BlockNumber)
	if err != nil {
		LoggerError("failed to get block commit mapping", []string{"error", err.Error()})
		return err
	}

	// If block is already finalized, ignore new commits
	if mapping.Finalized {
		LoggerDebug("block already finalized, ignoring commit", []string{"blockNumber", fmt.Sprintf("%d", commit.BlockNumber)})
		return nil
	}

	// Check if we already have a commit from this sender
	existingCommit, exists := mapping.Commits[commit.Sender]
	if exists {
		// Only overwrite if new commit has a newer timestamp
		if commit.Timestamp <= existingCommit.Timestamp {
			LoggerDebug("ignoring older commit from same sender", []string{
				"sender", commit.Sender,
				"existingTimestamp", fmt.Sprintf("%d", existingCommit.Timestamp),
				"newTimestamp", fmt.Sprintf("%d", commit.Timestamp),
			})
			return nil
		}
		LoggerDebug("overwriting commit with newer timestamp", []string{
			"sender", commit.Sender,
			"oldTimestamp", fmt.Sprintf("%d", existingCommit.Timestamp),
			"newTimestamp", fmt.Sprintf("%d", commit.Timestamp),
		})
	}

	// Store the commit
	mapping.Commits[commit.Sender] = commit

	// Save the updated mapping
	if err := SetBlockCommitMapping(mapping); err != nil {
		LoggerError("failed to save block commit mapping", []string{"error", err.Error()})
		return err
	}

	// Check consensus thresholds
	if err := CheckConsensusThresholds(mapping); err != nil {
		LoggerError("failed to check consensus thresholds", []string{"error", err.Error()})
		return err
	}

	return nil
}

// CheckConsensusThresholds checks if we have reached consensus thresholds for a block
func CheckConsensusThresholds(mapping *BlockCommitMapping) error {
	config, err := GetConsensusConfig()
	if err != nil {
		return err
	}

	// Get total validator count
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	totalValidators := len(nodes)
	if totalValidators == 0 {
		return nil
	}

	// Count commits by block hash
	hashCounts := make(map[string]int)
	for _, commit := range mapping.Commits {
		hashCounts[commit.BlockHash]++
	}

	// Find the hash with the most votes
	maxCount := 0
	maxHash := ""
	for hash, count := range hashCounts {
		if count > maxCount {
			maxCount = count
			maxHash = hash
		}
	}

	// Calculate percentages
	commitPercentage := (len(mapping.Commits) * 100) / totalValidators
	hashPercentage := (maxCount * 100) / totalValidators

	LoggerDebug("consensus status", []string{
		"blockNumber", fmt.Sprintf("%d", mapping.BlockNumber),
		"totalCommits", fmt.Sprintf("%d", len(mapping.Commits)),
		"totalValidators", fmt.Sprintf("%d", totalValidators),
		"commitPercentage", fmt.Sprintf("%d", commitPercentage),
		"maxHashCount", fmt.Sprintf("%d", maxCount),
		"maxHash", maxHash,
		"hashPercentage", fmt.Sprintf("%d", hashPercentage),
	})

	// Check finalization threshold (e.g., 75% agree on the same hash)
	if hashPercentage >= config.ThresholdFinalize {
		LoggerInfo("block finalized", []string{
			"blockNumber", fmt.Sprintf("%d", mapping.BlockNumber),
			"blockHash", maxHash,
			"percentage", fmt.Sprintf("%d", hashPercentage),
		})
		mapping.Finalized = true
		if err := SetBlockCommitMapping(mapping); err != nil {
			return err
		}
		return nil
	}

	// Check commit threshold for potential rollback (e.g., 51% have committed but disagree)
	if commitPercentage >= config.ThresholdCommit && len(hashCounts) > 1 {
		// We have enough commits but they disagree - check if we need to rollback
		LoggerInfo("commit threshold reached but hashes disagree, checking rollback", []string{
			"blockNumber", fmt.Sprintf("%d", mapping.BlockNumber),
			"uniqueHashes", fmt.Sprintf("%d", len(hashCounts)),
			"commitPercentage", fmt.Sprintf("%d", commitPercentage),
		})

		// Get our block hash to compare
		st, err := raftlib.GetCurrentState()
		if err != nil {
			return err
		}
		ourHash := base64.StdEncoding.EncodeToString(st.NextHash)

		// If our hash doesn't match the majority hash, we need to rollback
		if ourHash != maxHash {
			LoggerInfo("our hash doesn't match majority, initiating rollback", []string{
				"blockNumber", fmt.Sprintf("%d", mapping.BlockNumber),
				"ourHash", ourHash,
				"majorityHash", maxHash,
			})

			// Rollback to the block before the disagreement
			// Using raft.Rollback which handles multiple blocks
			rollbackHeight := mapping.BlockNumber - 1
			err := raftlib.Rollback([]fsm.ActionParam{
				{Key: "height", Value: fmt.Sprintf("%d", rollbackHeight)},
			}, fsm.EventObject{})
			if err != nil {
				LoggerError("rollback failed", []string{"error", err.Error()})
				return err
			}

			// Clear pending txs and reset Kayros UUID to re-fetch
			SetPendingTxHashes([]string{})

			// Reset LastKayrosUUID to the UUID from the last good block
			// This will cause GetKayrosTxs to re-fetch from that point
			lastGoodBlockMeta, err := GetBlockKayrosMetadata(rollbackHeight)
			if err == nil && lastGoodBlockMeta != nil && lastGoodBlockMeta.LastUUID != "" {
				SetLastKayrosUUID(lastGoodBlockMeta.LastUUID)
				LoggerInfo("reset Kayros UUID after rollback", []string{"uuid", lastGoodBlockMeta.LastUUID})
			}

			LoggerInfo("rollback completed", []string{"toHeight", fmt.Sprintf("%d", rollbackHeight)})
		}
	}

	return nil
}

// VerifyCommitLight performs light verification of a commit
func VerifyCommitLight(params []fsm.ActionParam, event fsm.EventObject) error {
	// TODO: Implement light client verification of commits
	// This could verify signatures and basic structure without full state verification
	return nil
}

// RequestMissingTransactions sends a P2P request to peers for missing transactions
func RequestMissingTransactions(txHashes []string) error {
	if len(txHashes) == 0 {
		return nil
	}

	// Get current block number
	lastBlockIndex, err := raftlib.GetLastBlockIndex()
	if err != nil {
		return err
	}
	nextBlockNumber := lastBlockIndex + 1

	// Get our node info
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	ourId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	if int(ourId) >= len(nodes) {
		return fmt.Errorf("our node ID out of range")
	}
	ourAddress := string(nodes[ourId].Address)

	// Create request
	request := MissingTransactionsRequest{
		TxHashes:    txHashes,
		BlockNumber: nextBlockNumber,
		Sender:      ourAddress,
	}

	reqBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Build P2P message payload
	payload := struct {
		Run struct {
			Event struct {
				Type   string `json:"type"`
				Params []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveMissingTransactionsRequest"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		{Key: "request", Value: base64.StdEncoding.EncodeToString(reqBytes)},
		{Key: "sender", Value: ourAddress},
	}
	msg, _ := json.Marshal(&payload)

	// Send to all peers
	peers := make([]string, 0, len(nodes)-1)
	for i := range nodes {
		if int32(i) != ourId && raftlib.IsNodeActive(nodes[i]) {
			peers = append(peers, nodes[i].Node.IP)
		}
	}

	if len(peers) == 0 {
		LoggerDebug("no peers available to request missing transactions", nil)
		return nil
	}

	contract := wasmx.GetAddress()
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{
		Contract:   contract,
		Sender:     contract,
		Msg:        msg,
		ProtocolId: PROTOCOL_ID,
		Peers:      peers,
	})

	if err != nil {
		LoggerError("failed to send missing transactions request", []string{"error", err.Error()})
		return err
	}

	LoggerInfo("sent missing transactions request", []string{
		"txCount", fmt.Sprintf("%d", len(txHashes)),
		"peerCount", fmt.Sprintf("%d", len(peers)),
	})

	return nil
}

// ReceiveMissingTransactionsRequest handles incoming requests for missing transactions
// This is called when another node asks us for transactions it doesn't have
func ReceiveMissingTransactionsRequest(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract request from event params
	requestB64 := ""
	sender := ""
	for _, p := range event.Params {
		if p.Key == "request" {
			requestB64 = p.Value
		} else if p.Key == "sender" {
			sender = p.Value
		}
	}

	if requestB64 == "" {
		return fmt.Errorf("no request found in event")
	}

	// Decode request
	requestData, err := base64.StdEncoding.DecodeString(requestB64)
	if err != nil {
		return err
	}

	var request MissingTransactionsRequest
	if err := json.Unmarshal(requestData, &request); err != nil {
		LoggerError("failed to parse missing transactions request", []string{"error", err.Error()})
		return err
	}

	LoggerDebug("received missing transactions request", []string{
		"sender", sender,
		"txCount", fmt.Sprintf("%d", len(request.TxHashes)),
		"blockNumber", fmt.Sprintf("%d", request.BlockNumber),
	})

	// Get mempool to find transactions
	mp, err := raftlib.GetMempool()
	if err != nil {
		return err
	}

	// Collect transactions we have
	transactions := make([][]byte, 0)
	foundHashes := make([]string, 0)

	for _, txHash := range request.TxHashes {
		// Check if transaction exists in mempool map
		if mempoolTx, ok := mp.Map[txHash]; ok {
			transactions = append(transactions, mempoolTx.Tx)
			foundHashes = append(foundHashes, txHash)
		}
	}

	if len(transactions) == 0 {
		LoggerDebug("no requested transactions found in mempool", []string{"sender", sender})
		return nil
	}

	// Get our node info
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	ourId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	ourAddress := string(nodes[ourId].Address)

	// Create response
	response := MissingTransactionsResponse{
		Transactions: transactions,
		TxHashes:     foundHashes,
		Sender:       ourAddress,
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	// Build P2P message payload
	payload := struct {
		Run struct {
			Event struct {
				Type   string `json:"type"`
				Params []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveMissingTransactions"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		{Key: "response", Value: base64.StdEncoding.EncodeToString(respBytes)},
		{Key: "sender", Value: ourAddress},
	}
	msg, _ := json.Marshal(&payload)

	// Send response to the requesting peer
	contract := wasmx.GetAddress()
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{
		Contract:   contract,
		Sender:     contract,
		Msg:        msg,
		ProtocolId: PROTOCOL_ID,
		Peers:      []string{sender},
	})

	if err != nil {
		LoggerError("failed to send missing transactions response", []string{"error", err.Error()})
		return err
	}

	LoggerInfo("sent missing transactions response", []string{
		"txCount", fmt.Sprintf("%d", len(transactions)),
		"to", sender,
	})

	return nil
}

// ReceiveMissingTransactions handles incoming missing transactions from peers
// This is called when a peer responds to our request for missing transactions
func ReceiveMissingTransactions(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract response from event params
	responseB64 := ""
	sender := ""
	for _, p := range event.Params {
		if p.Key == "response" {
			responseB64 = p.Value
		} else if p.Key == "sender" {
			sender = p.Value
		}
	}

	if responseB64 == "" {
		return fmt.Errorf("no response found in event")
	}

	// Decode response
	responseData, err := base64.StdEncoding.DecodeString(responseB64)
	if err != nil {
		return err
	}

	var response MissingTransactionsResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		LoggerError("failed to parse missing transactions response", []string{"error", err.Error()})
		return err
	}

	LoggerDebug("received missing transactions", []string{
		"sender", sender,
		"txCount", fmt.Sprintf("%d", len(response.Transactions)),
	})

	// Add transactions to mempool
	mp, err := raftlib.GetMempool()
	if err != nil {
		return err
	}

	addedCount := 0
	chainId := wasmx.GetChainId()
	for i, tx := range response.Transactions {
		txHash := response.TxHashes[i]
		if !mp.HasSeen(txHash) {
			// Add to mempool with default gas (0) and our chain as leader
			mp.Add(txHash, tx, 0, chainId)
			addedCount++
		}
	}

	// Save the updated mempool
	if addedCount > 0 {
		if err := raftlib.SetMempool(mp); err != nil {
			LoggerError("failed to save mempool", []string{"error", err.Error()})
		}
	}

	LoggerInfo("added missing transactions to mempool", []string{
		"addedCount", fmt.Sprintf("%d", addedCount),
		"fromPeer", sender,
	})

	return nil
}

// IfAllTransactions is a guard that checks if we have all transactions needed for the current block
func IfAllTransactions(params []fsm.ActionParam, event fsm.EventObject) bool {
	// Get pending transaction hashes
	pendingHashes, err := GetPendingTxHashes()
	if err != nil {
		LoggerError("failed to get pending tx hashes", []string{"error", err.Error()})
		return false
	}

	// If no pending transactions, we're ready (empty block)
	if len(pendingHashes) == 0 {
		LoggerDebug("no pending transactions, ready for block", nil)
		return true
	}

	// Get mempool to check which transactions we have
	mp, err := raftlib.GetMempool()
	if err != nil {
		LoggerError("failed to get mempool", []string{"error", err.Error()})
		return false
	}

	// Check if all pending transactions are in mempool
	missingCount := 0
	for _, txHash := range pendingHashes {
		// Convert hex hash to base64 for mempool check (mempool uses base64 format)
		txHashBase64, err := HexHashToBase64(txHash)
		if err != nil {
			LoggerError("failed to convert tx hash to base64", []string{"txHash", txHash, "error", err.Error()})
			missingCount++
			continue
		}

		if !mp.HasSeen(txHashBase64) {
			missingCount++
		}
	}

	if missingCount > 0 {
		LoggerDebug("still missing transactions", []string{
			"missingCount", fmt.Sprintf("%d", missingCount),
			"totalPending", fmt.Sprintf("%d", len(pendingHashes)),
		})
		return false
	}

	LoggerDebug("all transactions available", []string{
		"count", fmt.Sprintf("%d", len(pendingHashes)),
	})
	return true
}
