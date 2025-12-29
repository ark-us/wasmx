package lib

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	consutils "github.com/loredanacirstea/wasmx-consensus-utils/lib"
	consensuswrap "github.com/loredanacirstea/wasmx-env-consensus/lib"
	typestnd "github.com/loredanacirstea/wasmx-env-consensus/lib"
	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
	stakinglib "github.com/loredanacirstea/wasmx-staking/lib"
)

const (
	METAINFO_KEY_KAYROS_RECORDS = "kayros_records"
)

// CommitBlock builds and commits a block using Kayros-ordered transactions
// This is called as an entry action when transitioning to the propose state
func CommitBlock(params []fsm.ActionParam, event fsm.EventObject) error {
	// Get pending transaction hashes (ordered by Kayros)
	pendingHashes, err := GetPendingTxHashes()
	if err != nil {
		return fmt.Errorf("failed to get pending tx hashes: %v", err)
	}

	LoggerInfo("committing block with Kayros-ordered transactions", []string{
		"txCount", fmt.Sprintf("%d", len(pendingHashes)),
	})

	// Get mempool to fetch actual transaction bytes
	mp, err := raftlib.GetMempool()
	if err != nil {
		return fmt.Errorf("failed to get mempool: %v", err)
	}

	// Collect transactions in Kayros order and track last UUID
	txs := make([][]byte, 0, len(pendingHashes))
	var lastUUID string
	for _, txHash := range pendingHashes {
		// Convert hex hash to base64 for mempool lookup (mempool uses base64 keys)
		txHashBase64, err := HexHashToBase64(txHash)
		if err != nil {
			LoggerError("failed to convert tx hash to base64", []string{"txHash", txHash, "error", err.Error()})
			continue
		}

		if mempoolTx, ok := mp.Map[txHashBase64]; ok {
			txs = append(txs, mempoolTx.Tx)
			// Get the Kayros metadata for this tx to track last UUID
			if txMeta, err := GetKayrosTxMetadata(txHash); err == nil && txMeta != nil {
				lastUUID = txMeta.UUID
			}
		} else {
			LoggerError("transaction not found in mempool", []string{"txHash", txHash, "txHashBase64", txHashBase64})
			// Skip missing transactions - they should have been fetched
		}
	}

	// Build and commit the block
	// In Kayros P2P, all validators build blocks independently, so disable optimistic execution
	err = buildKayrosBlockProposal(txs, pendingHashes, false, raftlib.MaxBlockSizeBytes)
	if err != nil {
		return err
	}

	// After successful block commit, update the last Kayros UUID
	if lastUUID != "" {
		SetLastKayrosUUID(lastUUID)
		LoggerDebug("updated last Kayros UUID after block commit", []string{"uuid", lastUUID})
	}

	// Clear pending transaction hashes for next block
	if err := SetPendingTxHashes([]string{}); err != nil {
		LoggerError("failed to clear pending tx hashes", []string{"error", err.Error()})
	}

	return nil
}

// buildKayrosBlockProposal creates a block proposal with Kayros metadata
func buildKayrosBlockProposal(txs [][]byte, txHashes []string, optimisticExecution bool, maxDataBytes int64) error {
	if txs == nil {
		txs = make([][]byte, 0)
	}

	// Get last log index for height
	last, err := raftlib.GetLastLogIndex()
	if err != nil {
		return err
	}
	height := last + 1
	LoggerDebug("start Kayros block proposal", []string{"height", raftlib.Int64ToString(height)})

	// Gather state and validators
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	validators, err := raftlib.GetAllValidators()
	if err != nil {
		return err
	}
	activeInfos, err := consutils.GetActiveValidatorInfo(validators)
	if err != nil {
		return err
	}
	validatorInfos := consutils.SortTendermintValidators(activeInfos)
	validatorSet := typestnd.TendermintValidators{Validators: validatorInfos}

	// Use the first validator as proposer for deterministic block hash across all validators
	var proposerAddress wasmx.HexString
	var proposerBech32 wasmx.Bech32String
	if len(validatorInfos) > 0 {
		proposerAddress = validatorInfos[0].HexAddress
		proposerBech32 = validatorInfos[0].OperatorAddress
	} else {
		proposerAddress = wasmx.HexString("")
		proposerBech32 = wasmx.Bech32String("")
	}

	lastBlockCommit := getLastBlockCommitInternal(st)

	// Get previous block validators for the last block commit signatures
	previousBlock, err := raftlib.GetLogEntryAggregate(height - 1)
	var previousValidatorSet typestnd.TendermintValidators
	if previousBlock != nil {
		err = json.Unmarshal(previousBlock.Data.ValidatorInfo, &previousValidatorSet)
	} else {
		previousValidatorSet = validatorSet
	}

	signatures := consutils.FilterAndSortCommitSignatures(lastBlockCommit.Signatures, previousValidatorSet.Validators)
	if len(signatures) != len(previousValidatorSet.Validators) && height > (raftlib.LOG_START+1) {
		raftlib.Revert(fmt.Sprintf(`last block validator set length mismatch with signature list: expected %d, got %d`, len(signatures), len(previousValidatorSet.Validators)))
	}

	lastCommit := typestnd.CommitInfo{Round: 0, Votes: []typestnd.VoteInfo{}}
	localLastCommit := typestnd.ExtendedCommitInfo{Round: 0, Votes: []typestnd.ExtendedVoteInfo{}}
	for i := 0; i < len(signatures); i++ {
		commitSig := signatures[i]
		val := previousValidatorSet.Validators[i]

		vaddress := wasmx.AddrCanonicalize(string(val.OperatorAddress))
		validator := typestnd.Validator{Address: vaddress, Power: val.VotingPower}
		voteInfo := typestnd.VoteInfo{Validator: validator, BlockIDFlag: commitSig.BlockIDFlag}
		lastCommit.Votes = append(lastCommit.Votes, voteInfo)

		extendedVoteInfo := typestnd.ExtendedVoteInfo{
			Validator:          validator,
			VoteExtension:      []byte{},
			ExtensionSignature: []byte{},
			BlockIDFlag:        commitSig.BlockIDFlag,
		}
		localLastCommit.Votes = append(localLastCommit.Votes, extendedVoteInfo)
	}

	nextValsHash, err := consensuswrap.ValidatorsHash(validatorInfos)
	if err != nil {
		return err
	}
	misbehavior := []typestnd.Misbehavior{}
	timeNow := time.Now()
	timeISO := timeNow.UTC().Format(time.RFC3339Nano)

	t, _ := time.Parse(time.RFC3339Nano, st.LastTime)
	if t.UnixMilli() >= timeNow.UnixMilli() {
		raftlib.Revert(fmt.Sprintf(`last block time %s higher than current time %s, revert and skip round`, st.LastTime, timeISO))
	}

	prepareReq := typestnd.RequestPrepareProposal{
		MaxTxBytes:         maxDataBytes,
		Txs:                txs,
		LocalLastCommit:    localLastCommit,
		Misbehavior:        misbehavior,
		Height:             height,
		Time:               timeISO,
		NextValidatorsHash: nextValsHash,
		ProposerAddress:    proposerAddress, // First validator for deterministic hash
	}
	prepareResp, err := consensuswrap.PrepareProposal(prepareReq)
	if err != nil {
		return err
	}

	sortedBlockCommits := lastBlockCommit
	if height > (raftlib.LOG_START + 1) {
		sortedBlockCommits, err = consutils.GetSortedBlockCommits(lastBlockCommit, previousValidatorSet.Validators)
		sortedBlockCommits = consutils.CleanAbsentCommits(sortedBlockCommits)
	}
	evidence := typestnd.Evidence{}
	consHash := []byte{}
	if params, err := raftlib.GetConsensusParams(height); err == nil && params != nil {
		if h, err2 := consutils.GetConsensusParamsHash(*params); err2 == nil {
			consHash = h
		}
	}

	lastCommitHashHex := wasmx.HexString(strings.ToUpper(hex.EncodeToString(consutils.GetCommitHash(sortedBlockCommits))))

	// In Kayros consensus, all validators build blocks independently
	// ProposerAddress must be deterministic to ensure same block hash across all validators
	header := typestnd.Header{
		Version:            typestnd.VersionConsensus{Block: typestnd.BlockProtocol, App: st.Version.Consensus.App},
		ChainID:            st.ChainID,
		Height:             height,
		Time:               prepareReq.Time,
		LastBlockID:        st.LastBlockID,
		LastCommitHash:     lastCommitHashHex,
		DataHash:           wasmx.HexString(strings.ToUpper(hex.EncodeToString(consutils.GetTxsHash(prepareResp.Txs)))),
		ValidatorsHash:     wasmx.HexString(strings.ToUpper(hex.EncodeToString(nextValsHash))),
		NextValidatorsHash: wasmx.HexString(strings.ToUpper(hex.EncodeToString(nextValsHash))),
		ConsensusHash:      wasmx.HexString(strings.ToUpper(hex.EncodeToString(consHash))),
		AppHash:            wasmx.HexString(strings.ToUpper(hex.EncodeToString(st.AppHash))),
		LastResultsHash:    wasmx.HexString(strings.ToUpper(hex.EncodeToString(st.LastResultsHash))),
		EvidenceHash:       wasmx.HexString(strings.ToUpper(hex.EncodeToString(consutils.GetEvidenceHash(evidence)))),
		ProposerAddress:    proposerAddress, // First validator for deterministic hash
	}
	hhash, err := consensuswrap.HeaderHash(header)
	if err != nil {
		return err
	}
	LoggerInfo("Kayros block proposal", []string{
		"height", raftlib.Int64ToString(height),
		"hash", base64.StdEncoding.EncodeToString(hhash),
		"txCount", fmt.Sprintf("%d", len(txs)),
	})

	processReq := typestnd.RequestProcessProposal{
		Txs:                prepareResp.Txs,
		ProposedLastCommit: lastCommit,
		Misbehavior:        prepareReq.Misbehavior,
		Hash:               hhash,
		Height:             prepareReq.Height,
		Time:               prepareReq.Time,
		NextValidatorsHash: prepareReq.NextValidatorsHash,
		ProposerAddress:    proposerAddress, // First validator for deterministic hash
	}
	processResp, err := consensuswrap.ProcessProposal(processReq)
	if err != nil {
		return err
	}
	if processResp.Status == typestnd.ProposalStatus_REJECT {
		LoggerError("Kayros block rejected", []string{"height", raftlib.Int64ToString(processReq.Height)})
		return nil
	}

	// Build metainfo with Kayros records
	metainfo := map[string][]byte{}
	if optimisticExecution {
		if oe, err := consensuswrap.OptimisticExecution(processReq, processResp); err == nil {
			metainfo = oe.Metainfo
		}
	}

	// Add Kayros metadata to metainfo
	kayrosBlockMeta, err := BuildBlockKayrosMetadata(height)
	if err != nil {
		LoggerError("failed to build Kayros block metadata", []string{"error", err.Error()})
	} else if kayrosBlockMeta != nil && len(kayrosBlockMeta.Transactions) > 0 {
		// kayrosMetaBytes, err := json.Marshal(kayrosBlockMeta)
		// if err != nil {
		// 	LoggerError("failed to marshal Kayros metadata", []string{"error", err.Error()})
		// } else {
		// 	metainfo[METAINFO_KEY_KAYROS_RECORDS] = kayrosMetaBytes
		// 	LoggerInfo("added Kayros metadata to block", []string{
		// 		"txCount", fmt.Sprintf("%d", len(kayrosBlockMeta.Transactions)),
		// 		"firstUUID", kayrosBlockMeta.FirstUUID,
		// 		"lastUUID", kayrosBlockMeta.LastUUID,
		// 	})
		// }
	}

	st, _ = raftlib.GetCurrentState()
	st.NextHash = hhash
	raftlib.SetCurrentState(st)

	err = appendLogWithKayrosMeta(processReq, header, sortedBlockCommits, optimisticExecution, metainfo, validatorSet, proposerBech32)
	if err != nil {
		return err
	}

	// In Kayros P2P, finalize the block immediately after appending to log
	// (no voting phase - each validator finalizes independently)
	entryobj, err := raftlib.GetLogEntryAggregate(height)
	if err != nil {
		return fmt.Errorf("failed to get log entry for finalization: %v", err)
	}
	if entryobj != nil {
		_, err := raftlib.StartBlockFinalizationInternal(entryobj, false)
		if err != nil {
			return fmt.Errorf("failed to finalize block: %v", err)
		}
		// Update last applied to mark this block as finalized
		if err := raftlib.SetLastApplied(height); err != nil {
			return fmt.Errorf("failed to set last applied: %v", err)
		}
	}

	return nil
}

// appendLogWithKayrosMeta creates a BlockEntry and LogEntryAggregate with Kayros metadata
func appendLogWithKayrosMeta(processReq typestnd.RequestProcessProposal, header typestnd.Header, blockCommit typestnd.BlockCommit, optimisticExecution bool, meta map[string][]byte, validatorSet typestnd.TendermintValidators, proposerBech32 wasmx.Bech32String) error {
	if meta == nil {
		meta = make(map[string][]byte)
	}

	// Create RequestProcessProposalWithMetaInfo
	wrap := typestnd.RequestProcessProposalWithMetaInfo{
		Request:             processReq,
		OptimisticExecution: optimisticExecution,
		Metainfo:            meta,
	}

	blockData, err := json.Marshal(&wrap)
	if err != nil {
		return fmt.Errorf("failed to marshal block data: %v", err)
	}

	blockHeader, err := json.Marshal(&header)
	if err != nil {
		return fmt.Errorf("failed to marshal block header: %v", err)
	}

	commit, err := json.Marshal(&blockCommit)
	if err != nil {
		return fmt.Errorf("failed to marshal block commit: %v", err)
	}

	leaderId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return fmt.Errorf("failed to get current node ID: %v", err)
	}

	contractAddress := wasmx.GetAddressBz()

	validatorSetBytes, err := json.Marshal(&validatorSet)
	if err != nil {
		return fmt.Errorf("failed to marshal validator set: %v", err)
	}

	// Create BlockEntry with deterministic proposer address to ensure same block hash
	blockEntry := blocks.BlockEntry{
		Index:           processReq.Height,
		ReaderContract:  contractAddress,
		WriterContract:  contractAddress,
		Data:            blockData,
		Header:          blockHeader,
		ProposerAddress: proposerBech32, // First validator for deterministic hash
		LastCommit:      commit,
		Evidence:        []byte(`{"evidence":[]}`),
		Result:          []byte{},
		ValidatorInfo:   validatorSetBytes,
	}

	// Create LogEntryAggregate
	entry := raftlib.LogEntryAggregate{
		Index:    processReq.Height,
		TermID:   0,
		LeaderID: leaderId,
		Data:     blockEntry,
	}

	return raftlib.AppendLogEntry(entry)
}

// SendCommit broadcasts a commit message to other validators
func SendCommit(params []fsm.ActionParam, event fsm.EventObject) error {
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}

	lastIndex, err := raftlib.GetLastLogIndex()
	if err != nil {
		return err
	}

	// Get current node info
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

	// Create commit message
	commit := BlockCommit{
		BlockNumber: lastIndex,
		BlockHash:   base64.StdEncoding.EncodeToString(st.NextHash),
		Sender:      ourAddress,
		Timestamp:   time.Now().UnixMilli(),
	}

	// Sign the commit
	commitBytes, err := json.Marshal(commit)
	if err != nil {
		return err
	}
	signature, err := raftlib.SignMessage(string(commitBytes))
	if err != nil {
		LoggerError("failed to sign commit", []string{"error", err.Error()})
	}
	commit.Signature = signature

	commitBytes, _ = json.Marshal(commit)

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
	payload.Run.Event.Type = "receiveCommit"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		{Key: "commit", Value: base64.StdEncoding.EncodeToString(commitBytes)},
		{Key: "sender", Value: ourAddress},
	}
	msg, _ := json.Marshal(&payload)

	// Broadcast to chat room instead of sending to individual peers
	contract := wasmx.GetAddress()
	chainId := wasmx.GetChainId()
	protocolId := fmt.Sprintf("%s_%s", PROTOCOL_ID, chainId)
	topic := fmt.Sprintf("%s_%s", CHAT_ROOM_PROTOCOL, chainId)

	_, err = p2p.SendMessageToChatRoom(p2p.SendMessageToChatRoomRequest{
		Contract:   contract,
		Sender:     contract,
		Msg:        msg,
		ProtocolId: protocolId,
		Topic:      topic,
	})

	if err != nil {
		LoggerError("failed to send commit to chat room", []string{"error", err.Error(), "topic", topic})
		return err
	}

	LoggerInfo("sent commit to chat room", []string{
		"blockNumber", fmt.Sprintf("%d", commit.BlockNumber),
		"blockHash", commit.BlockHash,
		"topic", topic,
	})

	return nil
}

// Helper functions copied/adapted from raft

// getLastBlockCommitInternal gets the last block commit from state
func getLastBlockCommitInternal(st raftlib.CurrentState) typestnd.BlockCommit {
	return typestnd.BlockCommit{
		Height:     st.NextHeight - 1,
		Round:      0,
		BlockID:    st.LastBlockID,
		Signatures: st.LastBlockSigs,
	}
}

// getValidatorByHexAddrInternal queries the staking module for a validator
func getValidatorByHexAddrInternal(addr wasmx.HexString) (stakinglib.Validator, error) {
	payload := map[string]any{"ValidatorByHexAddr": map[string]any{"validator_addr": string(addr)}}
	bz, err := json.Marshal(&payload)
	if err != nil {
		return stakinglib.Validator{}, err
	}
	resp, err := callStakingInternal(string(bz), true)
	if err != nil {
		return stakinglib.Validator{}, err
	}
	if resp.Success > 0 {
		return stakinglib.Validator{}, fmt.Errorf("validator not found: %s", addr)
	}
	if resp.Data == "" {
		return stakinglib.Validator{}, fmt.Errorf("validator not found: %s", addr)
	}
	var result struct {
		Validator stakinglib.Validator `json:"validator"`
	}
	if err := json.Unmarshal([]byte(resp.Data), &result); err != nil {
		return stakinglib.Validator{}, err
	}
	return result.Validator, nil
}

// callStakingInternal calls the staking contract
func callStakingInternal(calldata string, isQuery bool) (wasmx.CallResponse, error) {
	addr := wasmx.GetAddressByRole("staking")
	ok, data := wasmx.CallSimple(addr, []byte(calldata), isQuery, MODULE_NAME)
	resp := wasmx.CallResponse{Success: 0, Data: string(data)}
	if !ok {
		resp.Success = 1
	}
	return resp, nil
}
