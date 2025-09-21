package lib

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	consutils "github.com/loredanacirstea/wasmx-consensus-utils/lib"
	consensuswrap "github.com/loredanacirstea/wasmx-env-consensus/lib"
	typestnd "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	multichain "github.com/loredanacirstea/wasmx-env-multichain/lib"
	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
)

// Minimal actions to align with AS entrypoint. Many are placeholders for now.

// Guard: are we voted leader?
func IsVotedLeader(_ []fsm.ActionParam, _ fsm.EventObject) (bool, error) {
	votes, err := GetVoteIndexArray()
	if err != nil {
		return false, err
	}
	count := int32(0)
	for _, v := range votes {
		count += v
	}
	valid, err := GetAllValidators()
	if err != nil {
		return false, err
	}
	activeValid, err := consutils.GetActiveValidatorInfo(valid)
	if err != nil {
		return false, err
	}
	ncount := len(activeValid)
	majority := GetMajority(ncount)
	LoggerDebug("check if is voted Leader", []string{"yes", Int32ToString(count), "total_votes", Int32ToString(int32(ncount)), "majority", Int64ToString(majority)})
	return int64(count) >= majority, nil
}

// WrapGuard returns "1"/"0" bytes like AS
func WrapGuard(value bool) []byte {
	if value {
		return []byte("1")
	}
	return []byte("0")
}

// Placeholders for now; functions return errors to surface failures.
func SetupNode(_ []fsm.ActionParam, event fsm.EventObject) error {
	// Mirror AS: read base64-encoded InitChainSetup payload from event.params["data"],
	// set indices, parse peers, set node IPs, init chain state, initialize index arrays.
	dataB64 := ""
	for _, p := range event.Params {
		if p.Key == "data" {
			dataB64 = p.Value
			break
		}
	}
	if dataB64 == "" {
		return errors.New("no initChainSetup found")
	}
	raw, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return err
	}
	// Decode into InitChainSetup
	var init InitChainSetup
	if err := json.Unmarshal(raw, &init); err != nil {
		return err
	}

	// Reset commit/applied
	if err := SetLastApplied(LOG_START); err != nil {
		return err
	}

	// Current node id
	if err := SetCurrentNodeId(init.NodeIndex); err != nil {
		return err
	}

	// Parse peers as address@host:port -> NodeInfo
	peers := make([]p2p.NodeInfo, len(init.Peers))
	for i, peer := range init.Peers {
		parts := strings.Split(peer, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid node format; found: %s", peer)
		}
		peers[i] = p2p.NodeInfo{Address: wasmx.Bech32String(parts[0]), Node: p2p.NetworkNode{IP: parts[1]}, OutOfSync: false}
	}
	if err := SetNodeIPs(peers); err != nil {
		return err
	}

	// Initialize chain state and store consensus params for next height
	if err := InitChain(init); err != nil {
		return err
	}

	// Initialize Next/Match index arrays
	return InitializeIndexArrays(len(peers))
}

// this is for Nodes, not validators
// TODO signature
func ReceiveCommit(_ []fsm.ActionParam, event fsm.EventObject) error {
	// Extract entry and signature
	entryB64 := ""
	// sig := ""
	for _, p := range event.Params {
		if p.Key == "entry" {
			entryB64 = p.Value
		}
		// if p.Key == "signature" {
		// 	sig = p.Value
		// }
	}
	if entryB64 == "" {
		return errors.New("update node: empty entry")
	}
	// if sig == "" {
	// 	return errors.New("update node: empty signature")
	// }
	entryBytes, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	LoggerDebugExtended("received new entry", []string{"AppendEntry", string(entryBytes)})

	var appendEntry LogEntryAggregate
	if err := json.Unmarshal(entryBytes, &appendEntry); err != nil {
		return err
	}
	// // verify signature from leader
	// ok, err := verifyMessage(appendEntry.LeaderID, sig, string(entryBytes))
	// if err != nil {
	// 	return err
	// }
	// if !ok {
	// 	LoggerError("signature verification failed for ReceiveCommit", []string{"leaderId", Int32ToString(appendEntry.LeaderID), "termId", Int32ToString(appendEntry.TermID)})
	// 	return nil
	// }

	LoggerInfo("received new entry", []string{
		"leaderId", Int32ToString(appendEntry.LeaderID),
		"termId", Int32ToString(appendEntry.TermID),
		"height", Int32ToString(int32(appendEntry.Index)),
	})

	// Get current state to check if we should process this commit
	lastIndex, err := GetLastBlockIndex()
	if err != nil {
		return err
	}
	// Skip if we already have this block or newer
	if lastIndex >= appendEntry.Index {
		LoggerDebug("commit already processed", []string{"lastIndex", Int64ToString(lastIndex), "entryIndex", Int64ToString(appendEntry.Index)})
		return nil
	}

	// Store the block - make sure to overwrite any existing block because this is a trusted commit
	st, err := GetCurrentState()
	if err != nil {
		return err
	}
	if appendEntry.Index == st.NextHeight {
		// Process as current block
		if err := ProcessAppendEntry(appendEntry); err != nil {
			return err
		}
	} else {
		// Store out of order block
		SetLogEntryAggregate(appendEntry)
		SetLastLogIndex(appendEntry.Index)
	}

	// Try to finalize blocks from last finalized index to this commit height (AS parity)
	lastFinalizedIndex, err := GetLastBlockIndex()
	if err != nil {
		return err
	}
	for i := lastFinalizedIndex + 1; i <= appendEntry.Index; i++ {
		// Check if we have the block at this height
		entry, err := GetLogEntryAggregate(i)
		if err != nil {
			break
		}
		if entry == nil || entry.Index == 0 {
			// Empty block, stop finalization
			LoggerDebug("empty block found, stopping finalization", []string{"height", Int64ToString(entry.Index)})
			break
		}
		// StartBlockFinalizationFollower
		// Finalize this block
		LoggerDebug("finalizing block from commit", []string{"height", Int64ToString(entry.Index)})
		bz, _ := json.Marshal(entry.Data)
		LoggerDebugExtended("start block finalization", []string{"height", Int64ToString(entry.Index), "leaderId", Int32ToString(entry.LeaderID), "termId", Int32ToString(entry.TermID), "data", string(bz)})
		if _, err := StartBlockFinalizationInternal(entry, false); err != nil {
			LoggerError("failed to finalize block from commit", []string{"height", Int64ToString(i), "error", err.Error()})
			return err
		}
	}

	return nil
}

func ProcessAppendEntries(_ []fsm.ActionParam, event fsm.EventObject) error {
	// here we receive new entries/logs/blocks
	// we need to run ProcessProposal on each block
	// and then FinalizeBlock & Commit
	// TODO we also look at termId, as we might need to rollback changes in case of a network split

	// Extract entry and signature
	entryB64 := ""
	sig := ""
	for _, p := range event.Params {
		if p.Key == "entry" {
			entryB64 = p.Value
		}
		if p.Key == "signature" {
			sig = p.Value
		}
	}
	if entryB64 == "" {
		return errors.New("update node: empty entry")
	}
	if sig == "" {
		return errors.New("update node: empty signature")
	}
	entryBytes, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	LoggerDebugExtended("received new entries", []string{"AppendEntry", string(entryBytes)})

	var appendEntry AppendEntry
	if err := json.Unmarshal(entryBytes, &appendEntry); err != nil {
		return err
	}
	// verify signature from leader
	ok, err := verifyMessage(appendEntry.LeaderID, sig, string(entryBytes))
	if err != nil {
		return err
	}
	if !ok {
		LoggerError("signature verification failed for AppendEntry", []string{"leaderId", Int32ToString(appendEntry.LeaderID), "termId", Int32ToString(appendEntry.TermID)})
		return nil
	}

	lastCommitIndex, err := GetLastBlockIndex()
	if err != nil {
		return err
	}
	lastLogIndex, err := GetLastLogIndex()
	if err != nil {
		return err
	}

	LoggerInfo("received new entries", []string{
		"leaderId", Int32ToString(appendEntry.LeaderID),
		"termId", Int32ToString(appendEntry.TermID),
		"leaderCommit", Int64ToString(appendEntry.LeaderCommit),
		"prevLogIndex", Int64ToString(appendEntry.PrevLogIndex),
		"prevLogTerm", Int32ToString(appendEntry.PrevLogTerm),
		"count", Int32ToString(int32(len(appendEntry.Entries))),
		"nodeIps", Int32ToString(int32(len(appendEntry.NodeIPs))),
		"our_last_commit_index", Int64ToString(lastCommitIndex),
		"our_last_log_index", Int64ToString(lastLogIndex),
	})

	// return if this is an outdated batch; batch is ordered ASC on block height
	if len(appendEntry.Entries) > 0 {
		if lastCommitIndex >= appendEntry.Entries[len(appendEntry.Entries)-1].Index {
			return nil
		}
	}

	// update our nodeips
	ips, err := GetNodeIPs()
	if err != nil {
		return err
	}
	nodeId, err := GetCurrentNodeId()
	if err != nil {
		return err
	}
	if int(nodeId) >= len(ips) {
		return errors.New("current node id out of range")
	}
	nodeIp := ips[nodeId]
	// check that our nodeId is still in line
	newId := nodeId
	for i, nodeInfo := range appendEntry.NodeIPs {
		if nodeInfo.Address == nodeIp.Address {
			newId = int32(i)
			break
		}
	}
	if newId != nodeId {
		LoggerDebug("our node ID has changed", []string{"old", Int32ToString(nodeId), "new", Int32ToString(newId)})
		if err := SetCurrentNodeId(newId); err != nil {
			return err
		}
	}
	if err := SetNodeIPs(appendEntry.NodeIPs); err != nil {
		return err
	}
	if err := SetTermId(appendEntry.TermID); err != nil {
		return err
	}

	// we commit as close to the transition end as possible
	// we make sure to commit the last block before running ProcessProposal on the new block
	// TODO
	// entry.leaderId ?
	minVal := lastLogIndex
	if appendEntry.LeaderCommit < lastLogIndex {
		minVal = appendEntry.LeaderCommit
	}
	maxCommitIndex := minVal
	for i := lastCommitIndex + 1; i <= maxCommitIndex; i++ {
		if _, err := StartBlockFinalizationFollower(i); err != nil {
			return err
		}
		if err := SetLastApplied(i); err != nil {
			return err
		}
	}

	lastCommitIndex, _ = GetLastBlockIndex()

	// now we check the new block
	for _, e := range appendEntry.Entries {
		if lastCommitIndex >= e.Index {
			continue
		}
		if err := ProcessAppendEntry(e); err != nil {
			return err
		}
		lastCommitIndex = e.Index
	}
	LoggerDebug("new entries processing finished", []string{
		"leaderId", Int32ToString(appendEntry.LeaderID),
		"leaderCommit", Int64ToString(appendEntry.LeaderCommit),
		"prevLogIndex", Int64ToString(appendEntry.PrevLogIndex),
		"count", Int32ToString(int32(len(appendEntry.Entries))),
	})
	return nil
}

// processAppendEntry decodes the proposal from entry and processes it before appending
func ProcessAppendEntry(entry LogEntryAggregate) error {
	// decode wrap from entry data
	var processReqWithMeta typestnd.RequestProcessProposalWithMetaInfo
	if err := json.Unmarshal(entry.Data.Data, &processReqWithMeta); err != nil {
		return err
	}
	processReq := processReqWithMeta.Request
	resp, err := consensuswrap.ProcessProposal(processReq)
	if err != nil {
		return err
	}
	if resp.Status == typestnd.ProposalStatus_REJECT {
		// TODO - what to do here? returning just discards the block and does not return a response to the leader
		// but this node will not sync with the leader anymore
		LoggerError("new block rejected", []string{"height", Int64ToString(processReq.Height), "node type", "Follower"})
		return nil
	}
	return AppendLogEntry(entry)
}

func SendHeartbeatResponse(params []fsm.ActionParam, event fsm.EventObject) error {
	entry, err := ExtractAppendEntry(params, event)
	if err != nil {
		return err
	}
	resp, err := PrepareHeartbeatResponse(entry)
	if err != nil {
		return err
	}
	bz, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	wasmx.SetFinishData(bz)
	return nil
}

func SendAppendEntries(_ []fsm.ActionParam, _ fsm.EventObject) error {
	leaderId, err := GetCurrentNodeId()
	if err != nil {
		return err
	}
	ips, err := GetNodeIPs()
	if err != nil {
		return err
	}
	LoggerDebug("diseminate entries...", []string{"nodeId", Int32ToString(leaderId)})
	for i, ip := range ips {
		if int32(i) == leaderId || !IsNodeActive(ip) {
			continue
		}
		// determine next index for this node
		nextArr, err := GetNextIndexArray()
		if err != nil {
			return err
		}
		var nextIndex int64 = LOG_START
		if i < len(nextArr) {
			nextIndex = nextArr[i]
		} else {
			nextIndex = LOG_START
		}
		lastIndex, err := GetLastLogIndex()
		if err != nil {
			return err
		}
		// batch limit similar to AS
		lastIndexToSend := lastIndex
		if (lastIndex - nextIndex) > int64(STATE_SYNC_BATCH) {
			lastIndexToSend = nextIndex + int64(STATE_SYNC_BATCH)
		}
		data, err := prepareAppendEntry(ips, nextIndex, lastIndexToSend)
		if err != nil {
			return err
		}
		msg, err := PrepareAppendEntryMessage(int32(i), nextIndex, lastIndex, lastIndexToSend, ip, data)
		if err != nil {
			return err
		}
		// send
		contract := wasmx.GetAddress()
		resp, err := SendGrpcJSONBase64(ip.Node.IP, contract, msg)
		if err != nil {
			continue
		}
		if resp.Error != "" || resp.Data == "" {
			continue
		}
		var vresp AppendEntryResponse
		if err := json.Unmarshal([]byte(resp.Data), &vresp); err != nil {
			continue
		}
		if vresp.Success {
			nextArr, err := GetNextIndexArray()
			if err != nil {
				return err
			}
			// update next index with number of entries we sent
			sent := int64(len(data.Entries))
			if i < len(nextArr) {
				nextArr[i] = nextIndex + sent
			} else {
				// grow
				needed := i - len(nextArr) + 1
				for n := 0; n < needed; n++ {
					nextArr = append(nextArr, LOG_START)
				}
				nextArr[i] = nextIndex + sent
			}
			if err := SetNextIndexArray(nextArr); err != nil {
				return err
			}
		}
	}
	return nil
}

func SendNewTransactionResponse(_ []fsm.ActionParam, _ fsm.EventObject) error {
	e, err := GetLogEntryObjIndexLast()
	if err != nil {
		return err
	}
	resp := TransactionResponse{TermID: e.TermID, LeaderID: e.LeaderID, Index: e.Index}
	bz, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	wasmx.SetFinishData(bz)
	return nil
}

func AddToMempool(_ []fsm.ActionParam, event fsm.EventObject) error {
	txB64 := ""
	for _, p := range event.Params {
		if p.Key == "transaction" {
			txB64 = p.Value
			break
		}
	}
	if txB64 == "" {
		return errors.New("no transaction found")
	}
	// Decode transaction payload (base64 => []byte)
	txBytes, err := base64.StdEncoding.DecodeString(txB64)
	if err != nil {
		return err
	}
	LoggerDebug("new transaction received", []string{"transaction", txB64})
	return AddTransactionToMempool(txBytes)
}

func AddTransactionToMempool(txBytes []byte) error {
	// AS parity with TinyGo semantics: keep []byte and surface errors.
	// 1) compute hash and mark seen
	txhash := base64.StdEncoding.EncodeToString(wasmx.Sha256(txBytes))
	LoggerDebug("new transaction received", []string{"transaction", base64.StdEncoding.EncodeToString(txBytes), "hash", txhash})
	mp, err := GetMempool()
	if err != nil {
		return err
	}
	if mp.HasSeen(txhash) {
		LoggerDebug("transaction already processed", []string{"hash", txhash})
		return nil
	}
	mp.Seen(txhash)
	SetMempool(mp)

	// 2) decode tx, compute gas and atomic info BEFORE CheckTx (AS runs CheckTx last)
	txDecoded, err := DecodeTx(txBytes)
	if err != nil {
		return errors.New(ERROR_INVALID_TX)
	}
	// Determine gas from fee if present, default to 1_000_000
	var txGas uint64 = 1000000
	if txDecoded.AuthInfo != nil && txDecoded.AuthInfo.Fee != nil && txDecoded.AuthInfo.Fee.GasLimit.GT(sdkmath.NewInt(0)) {
		txGas = txDecoded.AuthInfo.Fee.GasLimit.Uint64()
	}
	// Enforce consensus max gas if configured
	if cparams, err := GetConsensusParams(0); err == nil && cparams != nil {
		if cparams.Block.MaxGas > -1 && uint64(cparams.Block.MaxGas) < txGas {
			return fmt.Errorf("out of gas: %d; max %d", txGas, cparams.Block.MaxGas)
		}
	}

	leaderChain := ""
	atomicChains := []string{}
	if len(txDecoded.Body.ExtensionOptions) > 0 {
		for _, any := range txDecoded.Body.ExtensionOptions {
			if any.TypeURL == typestnd.TypeUrl_ExtensionOptionAtomicMultiChainTx {
				ext, err := typestnd.ExtensionOptionAtomicMultiChainTxFromAnyWrap(any)
				if err != nil {
					return err
				}
				ourchain := wasmx.GetChainId()
				// verify leader correctness
				computed := multichain.GetLeaderChain(ext.ChainIDs)
				if ext.LeaderChainID != computed {
					return fmt.Errorf("atomic transaction wrong leader: expected %s, got %s", computed, ext.LeaderChainID)
				}

				// this tx is not for our chain -> skip adding (AS returns hash; we just skip without error)
				found := false
				for _, cid := range ext.ChainIDs {
					if cid == ourchain {
						found = true
						break
					}
				}
				if !found {
					return nil
				}

				// don't propose atomic transactions if we do not have all subchains
				subchains, err := multichain.GetSubChainIds()
				if err == nil {
					weCanInclude := true
					atomicChains = ext.ChainIDs
					for _, cid := range ext.ChainIDs {
						present := false
						for _, sc := range subchains {
							if sc == cid {
								present = true
								break
							}
						}
						if !present {
							weCanInclude = false
							break
						}
					}
					if !weCanInclude {
						LoggerInfo("atomic transaction not added to mempool, node cannot be proposer", []string{"txhash", txhash, "subchains", strings.Join(ext.ChainIDs, ",")})
						return nil
					}
				}

				leaderChain = ext.LeaderChainID
				break
			}
		}
	}

	// 3) CheckTx last; if invalid, surface error
	req := typestnd.RequestCheckTx{Tx: txBytes, Type: typestnd.CheckTxTypeNew}
	resp, err := consensuswrap.CheckTx(req)
	if err != nil {
		return err
	}
	if resp.Code != uint32(typestnd.CodeTypeOk) {
		return fmt.Errorf("%s; code %d; %s", ERROR_INVALID_TX, resp.Code, resp.Log)
	}

	// 4) add to mempool
	mp.Add(txhash, txBytes, txGas, leaderChain)
	if err := SetMempool(mp); err != nil {
		return err
	}
	if leaderChain != "" {
		LoggerInfo("new transaction added to mempool", []string{"txhash", txhash, "atomic_crosschain_tx_leader", leaderChain, "subchains", strings.Join(atomicChains, ",")})
	} else {
		LoggerInfo("new transaction added to mempool", []string{"txhash", txhash})
	}
	return nil
}

func CommitBlocks(_ []fsm.ActionParam, _ fsm.EventObject) error {
	changed, err := checkCommits()
	if err != nil {
		return err
	}
	if changed {
		// propagate commit to others
		return SendAppendEntries(nil, fsm.EventObject{})
	}
	return nil
}

// CheckCommits exposes the commit check logic without dissemination (AS parity helper)
func CheckCommits(_ []fsm.ActionParam, _ fsm.EventObject) (bool, error) {
	return checkCommits()
}

func SetRandomElectionTimeout(params []fsm.ActionParam, event fsm.EventObject) error {
	var minS, maxS string
	if len(params) == 0 {
		params = event.Params
	}
	for _, p := range params {
		if p.Key == "min" {
			minS = p.Value
		}
		if p.Key == "max" {
			maxS = p.Value
		}
	}
	if minS == "" {
		return errors.New("no min found")
	}
	if maxS == "" {
		return errors.New("no max found")
	}
	min, err := strconv.ParseInt(minS, 10, 64)
	if err != nil {
		return err
	}
	max, err := strconv.ParseInt(maxS, 10, 64)
	if err != nil {
		return err
	}
	v, err := GetRandomInRange(min, max)
	if err != nil {
		return err
	}
	SetElectionTimeout(v)
	return nil
}

func InitializeNextIndex(_ []fsm.ActionParam, _ fsm.EventObject) error {
	last, err := GetLastLogIndex()
	if err != nil {
		return err
	}
	next := last + 1
	arr, err := GetNextIndexArray()
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		// init from node count
		n, err := GetNodeCount()
		if err != nil {
			return err
		}
		arr = make([]int64, n)
	}
	for i := range arr {
		arr[i] = next
	}
	return SetNextIndexArray(arr)
}

func InitializeMatchIndex(_ []fsm.ActionParam, _ fsm.EventObject) error {
	arr, err := GetMatchIndexArray()
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		n, err := GetNodeCount()
		if err != nil {
			return err
		}
		arr = make([]int64, n)
	}
	for i := range arr {
		arr[i] = 0
	}
	return SetMatchIndexArray(arr)
}

func AddNodeNextIndexInternal(arr []int64, index int64) []int64 {
	if int(index) < len(arr) {
		return arr
	}
	for i := len(arr); i <= int(index); i++ {
		arr = append(arr, 0)
	}
	return arr
}

func AddNodeNextIndex(index int64) error {
	arr, err := GetNextIndexArray()
	if err != nil {
		return err
	}
	arr = AddNodeNextIndexInternal(arr, index)
	SetNextIndexArray(arr)
	return nil
}

func AddNodeMatchIndexInternal(arr []int64, index int64) []int64 {
	if int(index) < len(arr) {
		return arr
	}
	for i := len(arr); i <= int(index); i++ {
		arr = append(arr, 0)
	}
	return arr
}

func AddNodeMatchIndex(index int64) error {
	arr, err := GetMatchIndexArray()
	if err != nil {
		return err
	}
	arr = AddNodeMatchIndexInternal(arr, index)
	SetMatchIndexArray(arr)
	return nil
}

// PrepareAppendEntry exports the internal helper for reuse (AS parity)
func PrepareAppendEntry(nodeIps []p2p.NodeInfo, nextIndex int64, lastIndex int64) (AppendEntry, error) {
	return prepareAppendEntry(nodeIps, nextIndex, lastIndex)
}

func IncrementCurrentTerm(_ []fsm.ActionParam, _ fsm.EventObject) error {
	t, err := GetTermId()
	if err != nil {
		return err
	}
	return SetTermId(t + 1)
}

func Vote(_ []fsm.ActionParam, event fsm.EventObject) error {
	entryB64 := ""
	sig := ""
	for _, p := range event.Params {
		if p.Key == "entry" {
			entryB64 = p.Value
		}
		if p.Key == "signature" {
			sig = p.Value
		}
	}
	if entryB64 == "" {
		return errors.New("vote: empty entry")
	}
	if sig == "" {
		return errors.New("vote: empty signature")
	}
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	var req VoteRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}
	// verify signature against candidate's pubkey
	ok, err := verifyMessage(req.CandidateID, sig, string(data))
	if err != nil {
		return err
	}
	if !ok {
		LoggerError("signature verification failed for VoteRequest", []string{"candidateId", Int32ToString(req.CandidateID), "termId", Int32ToString(req.TermID)})
		return nil
	}
	resp, err := voteInternal(req.TermID, req.CandidateID, req.LastLogIndex, req.LastLogTerm)
	if err != nil {
		return err
	}
	bz, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	wasmx.SetFinishData(bz)
	return nil
}

func SelfVote(_ []fsm.ActionParam, _ fsm.EventObject) error {
	myId, err := GetCurrentNodeId()
	if err != nil {
		return err
	}
	if err := SetVotedFor(myId); err != nil {
		return err
	}
	ips, err := GetNodeIPs()
	if err != nil {
		return err
	}
	voteArray := make([]int32, len(ips))
	for i := range voteArray {
		voteArray[i] = 0
	}
	if int(myId) < len(voteArray) {
		voteArray[myId] = 1
	}
	return SetVoteIndexArray(voteArray)
}

func ForwardTxsToLeader(_ []fsm.ActionParam, _ fsm.EventObject) error {
	mp, err := GetMempool()
	if err != nil {
		return err
	}
	if len(mp.Map) == 0 {
		return nil
	}
	// get leader from last log
	last, err := GetLogEntryObjIndexLast()
	if err != nil {
		return err
	}
	if last.Index == 0 {
		return nil
	}
	nodeId := last.LeaderID
	ips, err := GetNodeIPs()
	if err != nil {
		return err
	}
	if int(nodeId) >= len(ips) {
		return nil
	}
	target := ips[nodeId]
	contract := wasmx.GetAddress()
	limit := mp.Count()
	if limit > 5 {
		limit = 5
	}
	LoggerDebug("forwarding txs to leader", []string{"nodeId", Int32ToString(nodeId), "nodeIp", target.Node.IP, "count", Int32ToString(int32(limit))})
	txhs := mp.Order[0:limit]
	for i, txhash := range txhs {
		tx, ok := mp.Map[txhash]
		if !ok {
			continue
		}
		msg := fmt.Sprintf(`{"run":{"event":{"type":"newTransaction","params":[{"key":"transaction","value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(tx.Tx))
		resp, err := SendGrpcJSONBase64(target.Node.IP, contract, msg)
		LoggerDebug("forwarding tx to leader", []string{"nodeId", Int32ToString(nodeId), "nodeIp", target.Node.IP, "batch", Int32ToString(int32(i)), "error", resp.Error})
		if err == nil && resp.Error == "" {
			mp.Remove(txhash)
			continue
		}
		if err == nil && resp.Error != "" && resp.Error == ERROR_INVALID_TX {
			mp.Remove(txhash)
			LoggerDebug("forwarded invalid transaction", []string{"tx", txhash})
		} else {
			break
		}
	}
	return SetMempool(mp)
}
func UpdateNodeAndReturn(params []fsm.ActionParam, event fsm.EventObject) error {
	entry, err := extractUpdateNodeEntryAndVerify(params, event)
	if err != nil {
		return err
	}
	resp, err := updateNodeEntry(entry)
	if err != nil {
		return err
	}
	bz, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	wasmx.SetFinishData(bz)
	return nil
}
func RegisteredCheck(_ []fsm.ActionParam, _ fsm.EventObject) error {
	ips, err := GetNodeIPs()
	if err != nil {
		return err
	}
	needed, err := registeredCheckNeeded(ips)
	if err != nil {
		return err
	}
	if !needed {
		return nil
	}
	nodeId, err := GetCurrentNodeId()
	if err != nil {
		return err
	}
	msgstr, err := registeredCheckMessage(ips, nodeId)
	if err != nil {
		return err
	}
	LoggerInfo("register request", []string{"req", msgstr})
	contract := wasmx.GetAddress()
	for i, node := range ips {
		if int32(i) == nodeId || node.Node.IP == "" {
			continue
		}
		LoggerInfo("register request", []string{"IP", node.Node.IP, "address", string(node.Address)})
		resp, err := SendGrpcJSONBase64(node.Node.IP, contract, msgstr)
		if err != nil || resp.Error != "" || resp.Data == "" {
			continue
		}
		var update UpdateNodeResponse
		if err := json.Unmarshal([]byte(resp.Data), &update); err != nil {
			continue
		}
		ourId := -1
		for j, n := range update.Nodes {
			if n.Address == ips[nodeId].Address {
				ourId = j
				break
			}
		}
		if ourId != -1 {
			if err := SetCurrentNodeId(int32(ourId)); err != nil {
				return err
			}
			if err := SetNodeIPs(update.Nodes); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func SendVoteRequests(_ []fsm.ActionParam, _ fsm.EventObject) error {
	termId, err := GetTermId()
	if err != nil {
		return err
	}
	candidateId, err := GetCurrentNodeId()
	if err != nil {
		return err
	}
	lastLogIndex, err := GetLastLogIndex()
	if err != nil {
		return err
	}
	lastEntry, err := GetLogEntryObjIndexLast()
	if err != nil {
		return err
	}
	lastLogTerm := lastEntry.TermID
	request := VoteRequest{TermID: termId, CandidateID: candidateId, LastLogIndex: lastLogIndex, LastLogTerm: lastLogTerm}
	ips, err := GetNodeIPs()
	if err != nil {
		return err
	}
	if len(ips) > 1 {
		bz, _ := json.Marshal(&ips)
		LoggerInfo("sending vote requests...", []string{"candidateId", Int32ToString(candidateId), "termId", Int32ToString(termId), "lastLogIndex", Int64ToString(lastLogIndex), "lastLogTerm", Int32ToString(lastLogTerm), "ips", string(bz)})
	}
	for i, ip := range ips {
		if int32(i) == candidateId || !IsNodeActive(ip) {
			continue
		}
		if err := sendVoteRequest(int32(i), ip, request, termId); err != nil {
			return err
		}
	}
	return nil
}

func ProposeBlock(_ []fsm.ActionParam, _ fsm.EventObject) error {
	// Only propose if last block is committed
	height, err := GetLastLogIndex()
	if err != nil {
		return err
	}
	lastCommit, err := GetLastBlockIndex()
	if err != nil {
		return err
	}
	if lastCommit < height {
		LoggerInfo("cannot propose new block, last block not committed", []string{"height", Int64ToString(height), "lastCommitIndex", Int64ToString(lastCommit)})
		return nil
	}
	// gather batch
	mp, err := GetMempool()
	if err != nil {
		return err
	}
	// Load consensus params to enforce gas/bytes limits
	var maxGas int64 = -1
	var maxBytes int64 = 0
	if cparams, err2 := GetConsensusParams(0); err2 == nil && cparams != nil {
		maxGas = cparams.Block.MaxGas
		maxBytes = cparams.Block.MaxBytes
	}
	if maxBytes <= 0 { // AS fallback to MaxBlockSizeBytes
		v, err3 := GetMaxTxBytes()
		if err3 == nil && v > 0 {
			maxBytes = v
		}
		if maxBytes <= 0 {
			maxBytes = MaxBlockSizeBytes
		}
	}
	batch, err := mp.Batch(maxGas, maxBytes, wasmx.GetChainId())
	if err != nil {
		return err
	}
	LoggerDebug("batch transactions", []string{"count", Int32ToString(int32(len(batch.Txs)))})
	// optimistic execution if atomic tx leader
	optimistic := batch.IsAtomicTx && batch.IsLeader
	if err := buildBlockProposal(batch.Txs, optimistic, batch.CummulatedGas, maxBytes); err != nil {
		return err
	}
	return SetMempool(mp)
}

// buildBlockProposal matches the AS utility: prepare + process + optional optimistic exec, then append log
func buildBlockProposal(txs [][]byte, optimisticExecution bool, _cummulatedGas int64, maxDataBytes int64) error {
	if txs == nil {
		txs = make([][]byte, 0)
	}
	// PrepareProposal
	last, err := GetLastLogIndex()
	if err != nil {
		return err
	}
	height := last + 1
	LoggerDebug("start block proposal", []string{"height", Int64ToString(height)})

	// Gather state and validators
	st, err := GetCurrentState()
	if err != nil {
		return err
	}
	validators, err := GetAllValidators()
	if err != nil {
		return err
	}
	activeInfos, err := consutils.GetActiveValidatorInfo(validators)
	if err != nil {
		return err
	}
	validatorInfos := consutils.SortTendermintValidators(activeInfos)
	validatorSet := typestnd.TendermintValidators{Validators: validatorInfos}

	lastBlockCommit := getLastBlockCommit(st)

	// we get the previous block validators for the last block commit signatures
	previousBlock, err := GetLogEntryAggregate(height - 1)
	var previousValidatorSet typestnd.TendermintValidators
	if previousBlock != nil {
		err = json.Unmarshal(previousBlock.Data.ValidatorInfo, &previousValidatorSet)
	} else {
		previousValidatorSet = validatorSet
	}

	signatures := consutils.FilterAndSortCommitSignatures(lastBlockCommit.Signatures, previousValidatorSet.Validators)
	if len(signatures) != len(previousValidatorSet.Validators) && height > (LOG_START+1) {
		Revert(fmt.Sprintf(`last block validator set length mismatch with signature list: expected %d, got %d`, len(signatures), len(previousValidatorSet.Validators)))
	}

	lastCommit := typestnd.CommitInfo{Round: 0, Votes: []typestnd.VoteInfo{}}
	localLastCommit := typestnd.ExtendedCommitInfo{Round: 0, Votes: []typestnd.ExtendedVoteInfo{}}
	for i := 0; i < len(signatures); i++ {
		commitSig := signatures[i]
		val := previousValidatorSet.Validators[i]

		// TODO VoteInfo should be hex
		// but then we need a mapping hex => pubkey or hex => operator_address
		// hex format -> bytes -> base64
		// commitSig.validator_address

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
	timeISO := time.Now().UTC().Format(time.RFC3339)
	prepareReq := typestnd.RequestPrepareProposal{
		MaxTxBytes:         maxDataBytes,
		Txs:                txs,
		LocalLastCommit:    localLastCommit,
		Misbehavior:        misbehavior,
		Height:             height,
		Time:               timeISO,
		NextValidatorsHash: nextValsHash,
		ProposerAddress:    st.ValidatorAddress,
	}
	prepareResp, err := consensuswrap.PrepareProposal(prepareReq)
	if err != nil {
		return err
	}

	sortedBlockCommits := lastBlockCommit
	// for height = 2, we have no signatures
	if height > (LOG_START + 1) {
		// sort active validators by power & address
		sortedBlockCommits, err = consutils.GetSortedBlockCommits(lastBlockCommit, previousValidatorSet.Validators)
		sortedBlockCommits = consutils.CleanAbsentCommits(sortedBlockCommits)
	}
	evidence := typestnd.Evidence{}
	consHash := []byte{}
	if params, err := GetConsensusParams(height); err == nil && params != nil {
		if h, err2 := consutils.GetConsensusParamsHash(*params); err2 == nil {
			consHash = h
		}
	}

	// Compute LastCommitHash from the commit signatures we will include
	lastCommitHashHex := wasmx.HexString(strings.ToUpper(hex.EncodeToString(consutils.GetCommitHash(sortedBlockCommits))))

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
		ProposerAddress:    st.ValidatorAddress,
	}
	hhash, err := consensuswrap.HeaderHash(header)
	if err != nil {
		return err
	}
	LoggerInfo("start block proposal", []string{"height", Int64ToString(height), "hash", base64.StdEncoding.EncodeToString(hhash), "optimistic_execution", fmt.Sprintf("%v", optimisticExecution)})

	processReq := typestnd.RequestProcessProposal{
		Txs:                prepareResp.Txs,
		ProposedLastCommit: lastCommit,
		Misbehavior:        prepareReq.Misbehavior,
		Hash:               hhash,
		Height:             prepareReq.Height,
		Time:               prepareReq.Time,
		NextValidatorsHash: prepareReq.NextValidatorsHash,
		ProposerAddress:    prepareReq.ProposerAddress,
	}
	processResp, err := consensuswrap.ProcessProposal(processReq)
	if err != nil {
		return err
	}
	if processResp.Status == typestnd.ProposalStatus_REJECT {
		LoggerError("new block rejected", []string{"height", Int64ToString(processReq.Height), "node type", "Leader"})
		return nil
	}
	metainfo := map[string][]byte{}
	if optimisticExecution {
		if oe, err := doOptimisticExecution(processReq, processResp); err == nil {
			metainfo = oe.Metainfo
		}
	}

	st, _ = GetCurrentState()
	st.NextHash = hhash
	SetCurrentState(st)

	return appendLogInternalVerified(processReq, header, sortedBlockCommits, optimisticExecution, metainfo, validatorSet)
}
func Setup(params []fsm.ActionParam, event fsm.EventObject) error {
	LoggerInfo("setting up new raft consensus contract", nil)
	oldContract := ""
	if len(params) > 0 {
		oldContract = params[0].Value
	} else if len(event.Params) > 0 {
		oldContract = event.Params[0].Value
	}
	if oldContract == "" {
		return errors.New("previous contract address not provided")
	}
	// helper to call getContextValue
	getCtx := func(key string) (string, error) {
		payload := struct {
			GetContextValue struct {
				Key string `json:"key"`
			} `json:"getContextValue"`
		}{}
		payload.GetContextValue.Key = key
		bz, _ := json.Marshal(&payload)
		addr := wasmx.Bech32String(oldContract)
		resp, err := callContract(addr, string(bz), true, MODULE_NAME)
		if err != nil {
			return "", err
		}
		if resp.Success > 0 {
			return "", fmt.Errorf("cannot get %s from previous contract", key)
		}
		// Data may be base64 or raw JSON; try base64 first
		if dbz, err := base64.StdEncoding.DecodeString(resp.Data); err == nil {
			return string(dbz), nil
		}
		return resp.Data, nil
	}
	// NODE_IPS
	data, err := getCtx(NODE_IPS)
	if err != nil {
		return err
	}
	LoggerInfo("setting up nodeIPs", []string{"ips", data})
	var nodeIps []p2p.NodeInfo
	if err := json.Unmarshal([]byte(data), &nodeIps); err != nil {
		return err
	}
	if err := SetNodeIPs(nodeIps); err != nil {
		return err
	}
	// state
	data, err = getCtx(STATE_KEY)
	if err != nil {
		return err
	}
	LoggerInfo("setting up state", []string{"data", data})
	var state CurrentState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return err
	}
	if err := SetCurrentState(state); err != nil {
		return err
	}
	// mempool
	data, err = getCtx(MEMPOOL_KEY)
	if err != nil {
		return err
	}
	LoggerInfo("setting up mempool", []string{"data", data})
	var mempool Mempool
	if err := json.Unmarshal([]byte(data), &mempool); err != nil {
		return err
	}
	if err := SetMempool(mempool); err != nil {
		return err
	}
	// currentNodeId
	data, err = getCtx(CURRENT_NODE_ID)
	if err != nil {
		return err
	}
	LoggerInfo("setting up currentNodeId", []string{"data", data})
	if v, err := ParseI32(data); err == nil {
		if err := SetCurrentNodeId(v); err != nil {
			return err
		}
	}
	// currentTerm
	data, err = getCtx(TERM_ID)
	if err != nil {
		return err
	}
	LoggerInfo("setting up currentTerm", []string{"data", data})
	if v, err := ParseI32(data); err == nil {
		if err := SetTermId(v); err != nil {
			return err
		}
	}
	// last block index from storage contract
	lastIndex, err := GetLastBlockIndex()
	if err != nil {
		return err
	}
	LoggerInfo("setting up last log index", []string{"index", Int64ToString(lastIndex)})
	if err := SetLastLogIndex(lastIndex); err != nil {
		return err
	}
	// after we set last log index
	if err := InitializeIndexArrays(len(nodeIps)); err != nil {
		return err
	}
	return nil
}

// Helpers
func IsNodeActive(node p2p.NodeInfo) bool {
	return !node.OutOfSync && (node.Node.IP != "" || (node.Node.Host != "" && node.Node.Port != ""))
}

func sendVoteRequest(nodeId int32, node p2p.NodeInfo, request VoteRequest, termId int32) error {
	datastrBz, err := json.Marshal(&request)
	if err != nil {
		return err
	}
	datastr := string(datastrBz)
	signature, err := SignMessage(datastr)
	if err != nil {
		return err
	}
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(datastr))
	msgstr := fmt.Sprintf(`{"run":{"event":{"type":"receiveVoteRequest","params":[{"key":"entry","value":"%s"},{"key":"signature","value":"%s"}]}}}`, dataBase64, signature)
	contract := wasmx.GetAddress()
	LoggerDebug("sending vote request", []string{"nodeId", Int32ToString(nodeId), "nodeIp", node.Node.IP, "termId", Int32ToString(termId), "data", datastr})
	resp, err := wasmxcore.GrpcRequest(node.Node.IP, contract, base64.StdEncoding.EncodeToString([]byte(msgstr)))
	if err != nil {
		return err
	}
	LoggerDebug("vote request response", []string{"nodeId", Int32ToString(nodeId), "nodeIp", node.Node.IP, "termId", Int32ToString(termId), "data", resp.Data, "error", resp.Error})
	if resp.Error != "" || resp.Data == "" {
		return nil
	}
	var vresp VoteResponse
	if err := json.Unmarshal([]byte(resp.Data), &vresp); err != nil {
		return err
	}
	if vresp.VoteGranted {
		voteArray, err := GetVoteIndexArray()
		if err != nil {
			return err
		}
		if int(nodeId) >= 0 && int(nodeId) < len(voteArray) {
			voteArray[nodeId] = 1
		}
		return SetVoteIndexArray(voteArray)
	}
	if vresp.TermID > termId {
		return SetTermId(vresp.TermID)
	}
	return nil
}

// ExtractAppendEntry similar to AS
func ExtractAppendEntry(_ []fsm.ActionParam, event fsm.EventObject) (AppendEntry, error) {
	entryBase64 := ""
	for _, p := range event.Params {
		if p.Key == "entry" {
			entryBase64 = p.Value
			break
		}
	}
	if entryBase64 == "" {
		return AppendEntry{}, errors.New("update node: empty entry")
	}
	data, err := base64.StdEncoding.DecodeString(entryBase64)
	if err != nil {
		return AppendEntry{}, err
	}
	var entry AppendEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return AppendEntry{}, err
	}
	return entry, nil
}

func PrepareHeartbeatResponse(entry AppendEntry) (AppendEntryResponse, error) {
	termId, err := GetTermId()
	if err != nil {
		return AppendEntryResponse{}, err
	}
	lastLogIndex, err := GetLastLogIndex()
	if err != nil {
		return AppendEntryResponse{}, err
	}
	successful := true
	for _, e := range entry.Entries {
		if e.Index > lastLogIndex {
			successful = false
			break
		}
	}
	resp := AppendEntryResponse{TermID: termId, Success: successful, LastIndex: lastLogIndex}
	LoggerDebug("send heartbeat response", []string{"termId", Int32ToString(termId), "success", "true", "lastLogIndex", Int64ToString(lastLogIndex)})
	return resp, nil
}

// Get latest log entry
func GetLogEntryObjIndexLast() (LogEntry, error) {
	idx, err := GetLastLogIndex()
	if err != nil {
		return LogEntry{}, err
	}
	return GetLogEntryObj(idx)
}

// GetLogEntryAggregate returns a LogEntryAggregate with Data populated from
// uncommitted entry bytes if present, or from finalized storage block JSON.
func GetLogEntryAggregate(index int64) (*LogEntryAggregate, error) {
	e, err := GetLogEntryObj(index)
	if err != nil {
		return nil, err
	}
	if e.Index == 0 {
		return nil, nil
	}
	var data []byte
	if len(e.Data) > 0 {
		data = e.Data
	} else {
		s, err := getFinalBlock(index)
		if err != nil {
			return nil, err
		}
		if s == "" {
			return nil, nil
		}
		data = []byte(s)
	}
	var blockData blocks.BlockEntry
	err = json.Unmarshal(data, &blockData)
	if err != nil {
		return nil, err
	}
	agg := LogEntryAggregate{Index: e.Index, TermID: e.TermID, LeaderID: e.LeaderID, Data: blockData}
	return &agg, nil
}

// voteInternal simplified version with error surfacing
func voteInternal(termId int32, candidateId int32, lastLogIndex int64, lastLogTerm int32) (VoteResponse, error) {
	myTerm, err := GetTermId()
	if err != nil {
		return VoteResponse{}, err
	}
	nodes, err := GetNodeIPs()
	if err != nil {
		return VoteResponse{}, err
	}
	if int(candidateId) >= len(nodes) || !IsNodeActive(nodes[candidateId]) {
		return VoteResponse{TermID: myTerm, VoteGranted: false}, nil
	}
	if termId < myTerm {
		return VoteResponse{TermID: myTerm, VoteGranted: false}, nil
	}
	if termId > myTerm {
		// Match AS behavior: on higher term, update term and grant vote
		if err := SetTermId(termId); err != nil {
			return VoteResponse{}, err
		}
		if err := SetVotedFor(candidateId); err != nil {
			return VoteResponse{}, err
		}
		return VoteResponse{TermID: termId, VoteGranted: true}, nil
	}
	// Same term: if already voted, do not grant
	hv, err := HasVotedFor()
	if err != nil {
		return VoteResponse{}, err
	}
	if hv {
		return VoteResponse{TermID: termId, VoteGranted: false}, nil
	}
	// Candidate’s log at least as up-to-date
	myLast, err := GetLastLogIndex()
	if err != nil {
		return VoteResponse{}, err
	}
	if lastLogIndex < myLast {
		return VoteResponse{TermID: termId, VoteGranted: false}, nil
	}
	if err := SetVotedFor(candidateId); err != nil {
		return VoteResponse{}, err
	}
	return VoteResponse{TermID: termId, VoteGranted: true}, nil
}

// registeredCheckNeeded logic from AS
func registeredCheckNeeded(ips []p2p.NodeInfo) (bool, error) {
	lastIndex, err := GetLastLogIndex()
	if err != nil {
		return false, err
	}
	if lastIndex > LOG_START {
		return false, nil
	}
	if len(ips) == 1 {
		return false, nil
	}
	term, err := GetTermId()
	if err != nil {
		return false, err
	}
	if term < 2 {
		return false, nil
	}
	return true, nil
}

func registeredCheckMessage(ips []p2p.NodeInfo, nodeId int32) (string, error) {
	if int(nodeId) >= len(ips) {
		return "", errors.New("invalid node id")
	}
	LoggerInfo("trying to register node IP with Leader", nil)
	update := NodeUpdate{Node: ips[nodeId], Index: nodeId, Type: NODE_UPDATE_ADD}
	dataBz, err := json.Marshal(&update)
	if err != nil {
		return "", err
	}
	signature, err := SignMessage(string(dataBz))
	if err != nil {
		return "", err
	}
	dataBase64 := base64.StdEncoding.EncodeToString(dataBz)
	msg := fmt.Sprintf(`{"run":{"event":{"type":"updateNode","params":[{"key":"entry","value":"%s"},{"key":"signature","value":"%s"}]}}}`, dataBase64, signature)
	LoggerInfo("register request", []string{"req", msg})
	return msg, nil
}

func extractUpdateNodeEntryAndVerify(_ []fsm.ActionParam, event fsm.EventObject) (NodeUpdate, error) {
	entryB64 := ""
	sig := ""
	for _, p := range event.Params {
		if p.Key == "entry" {
			entryB64 = p.Value
		}
		if p.Key == "signature" {
			sig = p.Value
		}
	}
	if entryB64 == "" {
		return NodeUpdate{}, errors.New("updateNodeAndReturn: empty entry")
	}
	if sig == "" {
		return NodeUpdate{}, errors.New("updateNodeAndReturn: empty signature")
	}
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return NodeUpdate{}, err
	}
	var entry NodeUpdate
	if err := json.Unmarshal(data, &entry); err != nil {
		return NodeUpdate{}, err
	}
	LoggerDebug("updateNodeAndReturn", []string{"entry", string(data), "signature", sig})

	// Verify signature from the node that's trying to register/update
	// For new nodes joining, we need to verify their signature using their consensus public key
	// This is different from regular RAFT operations where we use nodeIndex for verification

	// First try to verify using the node's address directly
	ok, err := VerifyMessageByAddr(entry.Node.Address, sig, string(data))
	if err != nil {
		LoggerError("signature verification error", []string{"address", string(entry.Node.Address), "error", err.Error()})
		return NodeUpdate{}, fmt.Errorf("signature verification failed: address %s: %v", entry.Node.Address, err)
	}
	if !ok {
		LoggerError("signature verification failed", []string{"address", string(entry.Node.Address)})
		return NodeUpdate{}, fmt.Errorf("signature verification failed: address %s", entry.Node.Address)
	}

	LoggerDebug("updateNodeAndReturn signature verified successfully", []string{"entry", string(data), "signature", sig, "address", string(entry.Node.Address)})
	return entry, nil
}

func updateNodeEntry(entry NodeUpdate) (UpdateNodeResponse, error) {
	ips, err := GetNodeIPs()
	if err != nil {
		return UpdateNodeResponse{}, err
	}
	if entry.Type == NODE_UPDATE_ADD {
		if entry.Node.Node.IP == "" && entry.Node.Node.Host == "" {
			return UpdateNodeResponse{}, errors.New("validator info missing from node update")
		}
		ndx := -1
		for i, n := range ips {
			if n.Address == entry.Node.Address {
				ndx = i
				break
			}
		}
		if ndx > -1 {
			ips[ndx].Node = entry.Node.Node
		} else {
			ips = append(ips, entry.Node)
			AddNodeNextIndex(int64(len(ips) - 1))
			AddNodeMatchIndex(int64(len(ips) - 1))
		}
	} else if entry.Type == NODE_UPDATE_REMOVE {
		if int(entry.Index) >= 0 && int(entry.Index) < len(ips) {
			ips[entry.Index] = removeNode(ips, int(entry.Index))[entry.Index]
		}
	}
	if err := SetNodeIPs(ips); err != nil {
		return UpdateNodeResponse{}, err
	}
	nodeId, err := GetCurrentNodeId()
	if err != nil {
		return UpdateNodeResponse{}, err
	}
	last, err := GetLastLogIndex()
	if err != nil {
		return UpdateNodeResponse{}, err
	}
	return UpdateNodeResponse{Nodes: ips, SyncNodeID: nodeId, LastEntryIndex: last}, nil
}

func removeNode(nodes []p2p.NodeInfo, index int) []p2p.NodeInfo {
	nodes[index].Node.IP = ""
	nodes[index].Node.Host = ""
	nodes[index].Node.Port = ""
	return nodes
}

// checkCommits minimal implementation
func checkCommits() (bool, error) {
	lastCommit, err := GetLastBlockIndex()
	if err != nil {
		return false, err
	}
	lastLog, err := GetLastLogIndex()
	if err != nil {
		return false, err
	}
	nextCommit := lastCommit + 1
	if lastLog < nextCommit {
		return false, nil
	}
	nextArr, err := GetNextIndexArray()
	if err != nil {
		return false, err
	}
	count := 1 // leader
	for _, v := range nextArr {
		if v > nextCommit {
			count++
		}
	}
	valid, err := GetAllValidators()
	if err != nil {
		return false, err
	}
	activeValid, err := consutils.GetActiveValidatorInfo(valid)
	if err != nil {
		return false, err
	}
	ncount := len(activeValid)
	committing := int64(count) >= GetMajority(ncount)
	LoggerDebug("committing diseminated block", []string{"height", Int64ToString(nextCommit)})
	if committing {
		changed, err2 := StartBlockFinalizationLeader(nextCommit)
		if err2 != nil {
			return false, err2
		}
		if err := SetLastApplied(nextCommit); err != nil {
			return false, err
		}
		return changed, nil
	}
	return false, nil
}

func StartBlockFinalizationInternal(entryobj *LogEntryAggregate, retry bool) (bool, error) {
	// entry data may be wrap JSON or BlockEntry JSON
	var processReqWithMeta typestnd.RequestProcessProposalWithMetaInfo
	if err := json.Unmarshal(entryobj.Data.Data, &processReqWithMeta); err != nil {
		return false, err
	}
	processReq := processReqWithMeta.Request
	// some blocks are stored out of order, so we run the block verification again
	if err := verifyBlockProposal(entryobj.Data, processReq); err != nil {
		LoggerError("new block rejected", []string{"height", Int64ToString(processReq.Height), "error", err.Error()})
		return false, nil
	}
	finReq := typestnd.RequestFinalizeBlock{
		Txs:                processReq.Txs,
		DecidedLastCommit:  processReq.ProposedLastCommit,
		Misbehavior:        processReq.Misbehavior,
		Hash:               processReq.Hash,
		Height:             processReq.Height,
		Time:               processReq.Time,
		NextValidatorsHash: processReq.NextValidatorsHash,
		ProposerAddress:    processReq.ProposerAddress,
	}
	w := typestnd.WrapRequestFinalizeBlock{Request: finReq, Metainfo: processReqWithMeta.Metainfo}
	// if we have done optimisting execution, BeginBlock was already ran
	selfNode, err := getSelfNodeInfo()
	if err != nil {
		return false, fmt.Errorf("failed to get self node info: %v", err)
	}

	oeran := processReqWithMeta.OptimisticExecution && (entryobj.Data.ProposerAddress == selfNode.Address)
	if !oeran {
		resbegin, err := consensuswrap.BeginBlock(finReq)
		if err != nil {
			return false, err
		}
		if resbegin.Error != "" && !retry {
			// ERR invalid height: X; expected: X+1
			mismatch := fmt.Sprintf("expected: %d", finReq.Height+1)
			LoggerInfo("begin block error", []string{"error", resbegin.Error})
			if strings.Contains(resbegin.Error, "invalid height") && strings.Contains(resbegin.Error, mismatch) {
				if err := consensuswrap.RollbackToVersion(finReq.Height - 1); err != nil {
					return false, fmt.Errorf("consensus break: %s; %v", resbegin.Error, err)
				}
				if !retry {
					return StartBlockFinalizationInternal(entryobj, true)
				}
				return false, fmt.Errorf(resbegin.Error)
			}
			return false, fmt.Errorf(resbegin.Error)
		}
	}
	resfin, err := consensuswrap.FinalizeBlock(w)
	if err != nil {
		return false, err
	}
	if resfin.Error != "" && !retry {
		// ERR invalid height: X; expected: X+1
		mismatch := fmt.Sprintf("expected: %d", finReq.Height+1)
		if strings.Contains(resfin.Error, "invalid height") && strings.Contains(resfin.Error, mismatch) {
			LoggerInfo("trying to rollback", []string{"height", Int64ToString(finReq.Height - 1)})
			if err := consensuswrap.RollbackToVersion(finReq.Height - 1); err != nil {
				return false, fmt.Errorf("consensus break: %s; %v", resfin.Error, err)
			}
			// repeat FinalizeBlock
			return StartBlockFinalizationInternal(entryobj, true)
		} else {
			return false, fmt.Errorf(resfin.Error)
		}
	}
	if resfin.Error != "" {
		return false, fmt.Errorf(resfin.Error)
	}
	finalizeResp := resfin.Data
	if finalizeResp == nil {
		return false, errors.New("FinalizeBlock response is null")
	}

	// AS: Store finalize response as base64 in entryobj.data.result (line 1181-1184)
	resultBz, err := json.Marshal(finalizeResp)
	if err != nil {
		return false, err
	}
	// This modifies the entry data which will be stored
	entryobj.Data.Result = resultBz

	// AS: Parse commit from entryobj.data.last_commit (line 1186-1187)
	if len(entryobj.Data.LastCommit) == 0 {
		return false, errors.New("missing last_commit in entry data")
	}
	commitBz, err := base64.StdEncoding.DecodeString(string(entryobj.Data.LastCommit))
	if err != nil {
		// try as raw JSON
		commitBz = entryobj.Data.LastCommit
	}
	var commit typestnd.BlockCommit
	if err := json.Unmarshal(commitBz, &commit); err != nil {
		return false, fmt.Errorf("failed to parse commit: %v", err)
	}

	lastCommitHash := consutils.GetCommitHash(commit)
	lastResultsHash := consutils.GetResultsHash(finalizeResp.TxResults)

	// AS: Update current state (lines 1193-1200)
	LoggerDebug("updating current state...", nil)
	st, err := GetCurrentState()
	if err != nil {
		return false, err
	}
	st.AppHash = finalizeResp.AppHash
	st.LastBlockID = GetBlockID(processReq.Hash)
	st.LastCommitHash = lastCommitHash
	st.LastResultsHash = lastResultsHash
	st.LastRound = int64(entryobj.TermID)
	st.LastTime = finReq.Time
	st.NextHeight = finReq.Height + 1
	sigs, err := getCommitSigsFromPrecommitArray(st, finReq.Height, processReq.Hash, int64(entryobj.TermID))
	if err != nil {
		return false, err
	}
	st.LastBlockSigs = sigs
	// Reset transient consensus fields (AS parity)
	st.ValidValue = 0
	st.ValidRound = 0
	st.LockedValue = 0
	st.LockedRound = 0

	if err := SetCurrentState(st); err != nil {
		return false, err
	}

	// AS: Update consensus params (lines 1202-1204)
	LoggerDebug("updating consensus parameters...", nil)
	if err := updateConsensusParams(processReq.Height, finalizeResp.ConsensusParamUpdates); err != nil {
		return false, err
	}

	// AS: Update validator info (lines 1207-1208)
	LoggerDebug("updating validator info...", []string{"validator_updates_count", fmt.Sprint(len(finalizeResp.ValidatorUpdates))})
	if err := updateValidators(finalizeResp.ValidatorUpdates); err != nil {
		return false, err
	}
	LoggerDebug("validator info updated successfully", []string{})

	// AS: Save final block and remove tx from mempool (lines 1213-1227)
	LoggerDebug("processing transaction hashes", []string{"tx_count", fmt.Sprint(len(processReq.Txs))})
	mp, err := GetMempool()
	if err != nil {
		return false, err
	}
	txhashes := make([]string, len(processReq.Txs))
	txHashBytes := make([][]byte, len(processReq.Txs))
	for i, tx := range processReq.Txs {
		hashbz := wasmx.Sha256(tx)
		hash := base64.StdEncoding.EncodeToString(hashbz)
		txhashes[i] = hash
		txHashBytes[i] = hashbz
		mp.Remove(hash)
	}
	if err := SetMempool(mp); err != nil {
		return false, err
	}

	blockData, err := json.Marshal(&entryobj.Data)
	if err != nil {
		return false, err
	}
	indexedTopics := extractIndexedTopics(*finalizeResp, txHashBytes)
	if err := setFinalizedBlock(string(blockData), base64.StdEncoding.EncodeToString(processReq.Hash), txHashBytes, indexedTopics); err != nil {
		return false, err
	}

	// AS: Remove temporary block data (line 1230)
	if err := RemoveLogEntry(entryobj.Index); err != nil {
		return false, err
	}

	// AS: EndBlock will execute passed governance proposals (lines 1233-1236)
	respend, err := consensuswrap.EndBlock(string(blockData))
	if err != nil {
		return false, err
	}
	if respend.Error != "" {
		return false, fmt.Errorf("%s", respend.Error)
	}

	// AS: Parse events (lines 1240-1243)
	aggregated := consensuswrap.AggregateEvents(respend.Data.TxResults, respend.Data.Events)
	info := consensuswrap.DefaultFinalizeResponseEventsParse(finalizeResp.TxResults, aggregated)

	// AS: Update AppHash after EndBlock (lines 1246-1249)
	st, _ = GetCurrentState()
	st.AppHash = respend.Data.AppHash
	_ = SetCurrentState(st)

	// AS: Handle created validators (lines 1251-1275)
	if len(info.CreatedValidators) > 0 {
		for _, cv := range info.CreatedValidators {
			if int(cv.TxIndex) >= len(processReq.Txs) {
				continue
			}
			decodedTx, _ := DecodeTx(processReq.Txs[cv.TxIndex])
			LoggerInfo("new validator", []string{"height", Int64ToString(entryobj.Index), "address", string(cv.OperatorAddress), "p2p_address", decodedTx.Body.Memo})
			resp := parseNodeAddress(decodedTx.Body.Memo)
			if resp.Error == "" && resp.NodeInfo != nil {
				nodeInfo := resp.NodeInfo
				if cv.OperatorAddress != nodeInfo.Address {
					LoggerError("validator operator address mismatch, using operator_address", []string{"operator_address", string(cv.OperatorAddress), "memo", decodedTx.Body.Memo, "memo_nodeinfo_address", string(nodeInfo.Address)})
					nodeInfo.Address = cv.OperatorAddress
				}
				// AS: Add new node info to our validator info list
				if _, err := updateNodeEntry(NodeUpdate{Node: *nodeInfo, Index: 0, Type: NODE_UPDATE_ADD}); err != nil {
					LoggerError("failed to update node entry", []string{"error", err.Error()})
				}
				// AS: Move node info to validator info if it exists
				bz, _ := json.Marshal(nodeInfo)
				_ = callHookContract(wasmx.HOOK_CREATE_VALIDATOR, string(bz))
			} else {
				LoggerError("validator node invalid address format", []string{"memo", decodedTx.Body.Memo})
			}
		}
	}

	// AS: init subchains (lines ~1280-1287)
	if len(info.InitChainRequests) > 0 {
		for _, raw := range info.InitChainRequests {
			var det multichain.InitSubChainDeterministicRequest
			if err := json.Unmarshal(raw, &det); err != nil {
				LoggerError("init subchain parse error", []string{"error", err.Error()})
				continue
			}
			// Build init message using current validator identity
			cur, err := GetCurrentState()
			if err != nil {
				LoggerError("init subchain state error", []string{"error", err.Error()})
				continue
			}
			msg := multichain.InitSubChainMsg{
				InitChainRequest: det.InitChainRequest,
				ChainConfig:      det.ChainConfig,
				ValidatorAddress: string(cur.ValidatorAddress),
				ValidatorPrivkey: cur.ValidatorPrivkey,
				ValidatorPubkey:  cur.ValidatorPubkey,
				Peers:            det.Peers,
				CurrentNodeID:    0,
				InitialPorts:     multichain.DefaultNodePorts(),
			}
			if _, err := multichain.InitSubChain(msg); err != nil {
				LoggerError("init subchain failed", []string{"error", err.Error()})
			}
		}
	}

	// Before Commit: consensus contract upgrade/setup (AS lines 1288-1362)
	newContractSetup := false
	if info.ConsensusContract != "" {
		// setup next consensus contract with previous address
		myaddr := string(wasmx.GetAddress())
		LoggerInfo("setting up next consensus contract", []string{"new contract", info.ConsensusContract, "previous contract", myaddr})

		calldSetup := fmt.Sprintf(`{"run":{"event":{"type":"setup","params":[{"key":"address","value":"%s"}]}}}`, myaddr)
		if resp, err := callContract(wasmx.Bech32String(info.ConsensusContract), calldSetup, false, MODULE_NAME); err != nil {
			LoggerError("cannot setup next consensus contract", []string{"new_contract", info.ConsensusContract, "err", err.Error()})
		} else if resp.Success > 0 {
			LoggerError("cannot setup next consensus contract", []string{"new_contract", info.ConsensusContract, "err", resp.Data})
		} else {
			LoggerInfo("next consensus contract is set", []string{"new_contract", info.ConsensusContract})
			newContractSetup = true

			// best-effort: stop this contract
			calldStop := `{"run":{"event":{"type":"stop","params":[]}}}`
			if resp2, err2 := callContract(wasmx.GetAddress(), calldStop, false, MODULE_NAME); err2 != nil || resp2.Success > 0 {
				if err2 != nil {
					LoggerError("cannot stop previous consensus contract", []string{"err", err2.Error(), "data", resp2.Data})
				} else {
					LoggerError("cannot stop previous consensus contract", []string{"err", resp2.Data})
				}
			} else {
				LoggerInfo("stopped current consensus contract", nil)
			}
		}
	}

	// AS: Commit (line 1315)
	// TODO commitResponse.retainHeight
	if _, err := consensuswrap.Commit(); err != nil {
		return false, err
	}
	LoggerInfo("block finalized", []string{"height", Int64ToString(entryobj.Index), "hash", strings.ToUpper(hex.EncodeToString(processReq.Hash))})

	// AS: Make sure termId is synced (line 1321)
	if err := SetTermId(entryobj.TermID); err != nil {
		return false, err
	}

	// AS: Check if we became a validator (lines 1323-1334)
	if len(info.CreatedValidators) > 0 {
		selfNode, err := getSelfNodeInfo()
		if err == nil {
			ouraddr := selfNode.Address
			for _, cv := range info.CreatedValidators {
				if cv.OperatorAddress == ouraddr {
					LoggerInfo("node is validator", []string{"height", Int64ToString(entryobj.Index), "address", string(ouraddr)})
					// AS: Call consensus contract with "becomeValidator" transition
					calldatastr := `{"run":{"event": {"type": "becomeValidator", "params": []}}}`
					_, _ = callContract(wasmx.GetAddress(), calldatastr, false, MODULE_NAME)
				}
			}
		}
	}

	// TODO if we cannot start with the new contract we should remove its consensus role
	// but we are already after Commit(), so restart is not really feasible with this mechanism
	// we may need another mechanism where nodes can trigger transactions

	// After Commit: start new consensus contract if changed
	if info.ConsensusContract != "" && newContractSetup {
		LoggerInfo("starting new consensus contract", []string{"address", info.ConsensusContract})
		calldPrestart := `{"run":{"event":{"type":"prestart","params":[]}}}`
		if resp, err := callContract(wasmx.Bech32String(info.ConsensusContract), calldPrestart, false, MODULE_NAME); err != nil || resp.Success > 0 {
			if err != nil {
				LoggerError("cannot start next consensus contract", []string{"new_contract", info.ConsensusContract, "err", err.Error()})
			} else {
				LoggerError("cannot start next consensus contract", []string{"new_contract", info.ConsensusContract, "err", resp.Data})
			}
			// attempt to restart previous contract so chain does not stop
			myaddr := wasmx.GetAddress()
			calldRestart := `{"run":{"event":{"type":"restart","params":[]}}}`
			if resp2, err2 := callContract(myaddr, calldRestart, false, MODULE_NAME); err2 != nil || resp2.Success > 0 {
				if err2 != nil {
					LoggerError("cannot restart previous consensus contract", []string{"err", err2.Error()})
				} else {
					LoggerError("cannot restart previous consensus contract", []string{"err", resp2.Data})
				}
			} else {
				LoggerInfo("restarted current consensus contract", nil)
			}
		} else {
			LoggerInfo("next consensus contract is started", []string{"new_contract", info.ConsensusContract})
			return true, nil // consensus changed
		}
	}

	return false, nil
}

func StartBlockFinalizationLeader(index int64) (bool, error) {
	LoggerInfo("start block finalization", []string{"height", Int64ToString(index)})
	entryobj, err := GetLogEntryAggregate(index)
	if err != nil {
		return false, err
	}
	if entryobj == nil {
		LoggerInfo("cannot start block finalization", []string{"height", Int64ToString(index), "reason", "block empty"})
		return false, nil
	}
	LoggerDebug("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID)})
	bz, _ := json.Marshal(entryobj.Data)
	LoggerDebugExtended("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID), "data", string(bz)})
	currentTerm, err := GetTermId()
	if err != nil {
		return false, err
	}
	if currentTerm == entryobj.TermID {
		return StartBlockFinalizationInternal(entryobj, false)
	}
	nodeId, err := GetCurrentNodeId()
	if err != nil {
		return false, err
	}
	// one case why we could have a term mismatch here is if the Leader went down
	// after proposing a block and this node is the new Leader
	// but we had this block stored, from the old Leader
	if entryobj.LeaderID != nodeId {
		return StartBlockFinalizationInternal(entryobj, false)
	}

	LoggerError("entry has current term mismatch", []string{"nodeType", "Leader", "currentTerm", Int32ToString(currentTerm), "entryTermId", Int32ToString(entryobj.TermID), "leaderId", Int32ToString(entryobj.LeaderID), "ourId", Int32ToString(nodeId)})
	return false, nil
}

// StartBlockFinalizationFollower mirrors leader finalization on follower side
func StartBlockFinalizationFollower(index int64) (bool, error) {
	LoggerInfo("start block finalization", []string{"height", Int64ToString(index)})
	entryobj, err := GetLogEntryAggregate(index)
	if err != nil {
		return false, err
	}
	if entryobj == nil {
		LoggerInfo("cannot start block finalization", []string{"height", Int64ToString(index), "reason", "block empty"})
		return false, nil
	}
	LoggerDebug("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID)})
	bz, _ := json.Marshal(entryobj.Data)
	LoggerDebugExtended("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID), "data", string(bz)})
	return StartBlockFinalizationInternal(entryobj, false)
}

// bootstrapAfterStateSync updates the current state after state sync completion
func BootstrapAfterStateSync(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract state parameter
	var stateStr string
	for _, p := range params {
		if p.Key == "state" {
			stateStr = p.Value
			break
		}
	}
	if stateStr == "" {
		for _, p := range event.Params {
			if p.Key == "state" {
				stateStr = p.Value
				break
			}
		}
	}
	if stateStr == "" {
		return errors.New("no state found")
	}

	// Decode base64 state
	stateBytes, err := base64.StdEncoding.DecodeString(stateStr)
	if err != nil {
		return fmt.Errorf("failed to decode state: %v", err)
	}

	// Parse state
	var state typestnd.State
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		return fmt.Errorf("failed to parse state: %v", err)
	}

	// Create last block ID
	lastBlockId := typestnd.BlockID{
		Hash: wasmx.HexString(strings.ToLower(string(state.LastBlockID.Hash))),
		Parts: typestnd.PartSetHeader{
			Total: state.LastBlockID.Parts.Total,
			Hash:  wasmx.HexString(strings.ToLower(string(state.LastBlockID.Parts.Hash))),
		},
	}

	// Update current state
	currentState, err := GetCurrentState()
	if err != nil {
		return fmt.Errorf("failed to get current state: %v", err)
	}

	currentState.ChainID = state.ChainID
	currentState.Version = state.Version
	currentState.AppHash = state.AppHash
	currentState.LastBlockID = lastBlockId
	currentState.LastResultsHash = state.LastResultsHash
	currentState.LastTime = state.LastBlockTime
	currentState.LastRound = 0
	currentState.NextHeight = state.LastBlockHeight + 1
	currentState.NextHash = []byte{}

	if err := SetCurrentState(currentState); err != nil {
		return fmt.Errorf("failed to set current state: %v", err)
	}

	// Update storage contract - bootstrap after state sync
	if err := storageBootstrapAfterStateSync(state.LastBlockHeight, state.LastHeightConsensusParamsChanged, state.ConsensusParams); err != nil {
		return fmt.Errorf("failed to bootstrap storage: %v", err)
	}

	// Update last log index
	if err := SetLastLogIndex(state.LastBlockHeight); err != nil {
		return fmt.Errorf("failed to set last log index: %v", err)
	}

	LoggerInfo("bootstrap after state sync completed", []string{
		"chain_id", state.ChainID,
		"last_height", Int64ToString(state.LastBlockHeight),
		"last_committed", Int64ToString(state.LastBlockHeight),
		"next_height", Int64ToString(currentState.NextHeight),
		"app_hash", hex.EncodeToString(currentState.AppHash),
	})

	return nil
}

// commitAfterStateSync processes block commit information after state sync
// not used
func CommitAfterStateSync(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract commit parameter
	var commitStr string
	for _, p := range params {
		if p.Key == "commit" {
			commitStr = p.Value
			break
		}
	}
	if commitStr == "" {
		for _, p := range event.Params {
			if p.Key == "commit" {
				commitStr = p.Value
				break
			}
		}
	}
	if commitStr == "" {
		return errors.New("no commit found")
	}
	commitbz, err := base64.StdEncoding.DecodeString(commitStr)
	if err != nil {
		return fmt.Errorf("failed to parse base64 block commit: %v", err)
	}

	// Parse block commit
	var commit typestnd.BlockCommit
	if err := json.Unmarshal(commitbz, &commit); err != nil {
		return fmt.Errorf("failed to parse block commit: %v", err)
	}

	LoggerDebug("commit after state sync", []string{
		"height", Int64ToString(commit.Height),
		"round", Int64ToString(commit.Round),
		"signatures_count", Int32ToString(int32(len(commit.Signatures))),
	})

	// TODO: In RAFT, we might want to update our understanding of validator signatures
	// For now, this is a placeholder as the current implementation doesn't use prevote/precommit arrays
	// like Tendermint does, but this function maintains compatibility with the interface

	return nil
}

func VerifyCommitLight(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract commit parameter
	var dataBase64 string
	for _, p := range params {
		if p.Key == "data" {
			dataBase64 = p.Value
			break
		}
	}
	if dataBase64 == "" {
		for _, p := range event.Params {
			if p.Key == "data" {
				dataBase64 = p.Value
				break
			}
		}
	}
	if dataBase64 == "" {
		return errors.New("no data found")
	}

	databz, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return fmt.Errorf("failed to parse base64 VerifyCommitLightRequest: %v", err)
	}

	// Parse block commit
	var data typestnd.VerifyCommitLightRequest
	if err := json.Unmarshal(databz, &data); err != nil {
		return fmt.Errorf("failed to parse VerifyCommitLightRequest: %v", err)
	}

	// TODO verify signatures & calculate voting threshold
	respbz, err := json.Marshal(&typestnd.VerifyCommitLightResponse{
		Valid: true,
		Error: "",
	})
	wasmx.SetFinishData(respbz)
	return nil
}
