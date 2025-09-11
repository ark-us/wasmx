package lib

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	consutils "github.com/loredanacirstea/wasmx-consensus-utils/lib"
	consensuswrap "github.com/loredanacirstea/wasmx-env-consensus/lib"
	typestnd "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	multichain "github.com/loredanacirstea/wasmx-env-multichain/lib"
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
	ncount, err := GetNodeCount()
	if err != nil {
		return false, err
	}
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
	if err := SetCommitIndex(LOG_START); err != nil {
		return err
	}
	if err := SetLastApplied(LOG_START); err != nil {
		return err
	}

	// Current node id
	if err := SetCurrentNodeId(init.NodeIndex); err != nil {
		return err
	}

	// Parse peers as address@host:port -> NodeInfo
	peers := make([]NodeInfo, len(init.Peers))
	for i, peer := range init.Peers {
		parts := strings.Split(peer, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid node format; found: %s", peer)
		}
		peers[i] = NodeInfo{Address: parts[0], Node: NetworkNode{IP: parts[1]}, OutOfSync: false}
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
func ProcessAppendEntries(_ []fsm.ActionParam, event fsm.EventObject) error {
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
	for _, e := range appendEntry.Entries {
		if err := processAppendEntry(e); err != nil {
			return err
		}
	}
	LoggerDebug("new entries processing finished", []string{"leaderId", Int32ToString(appendEntry.LeaderID), "leaderCommit", Int64ToString(appendEntry.LeaderCommit), "prevLogIndex", Int64ToString(appendEntry.PrevLogIndex), "count", Int32ToString(int32(len(appendEntry.Entries)))})
	return nil
}

// processAppendEntry decodes the proposal from entry and processes it before appending
func processAppendEntry(entry LogEntryAggregate) error {
	// decode wrap from entry data
	var wrap typestnd.RequestProcessProposalWithMetaInfo
	if err := json.Unmarshal(entry.Data, &wrap); err != nil {
		return err
	}
	resp, err := consensuswrap.ProcessProposal(wrap.Request)
	if err != nil {
		return err
	}
	if resp.Status == typestnd.ProposalStatus_REJECT {
		LoggerError("new block rejected", []string{"height", Int64ToString(wrap.Request.Height), "node type", "Follower"})
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
	// Stateless validation (CheckTx) similar to AS
	// If not OK, treat as invalid and surface error text
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

	req := typestnd.RequestCheckTx{Tx: txBytes, Type: typestnd.CheckTxTypeNew}
	resp, err := consensuswrap.CheckTx(req)
	if err != nil {
		return err
	}
	if resp.Code != uint32(typestnd.CodeTypeOk) {
		return fmt.Errorf("%s; code %d; %s", ERROR_INVALID_TX, resp.Code, resp.Log)
	}

	// Decode the Cosmos tx for gas and atomic extension handling
	txDecoded, err := decodeTx(txBytes)
	if err != nil {
		return errors.New(ERROR_INVALID_TX)
	}
	// Determine gas from fee if present, default to 1_000_000
	var txGas uint64 = 1000000
	if txDecoded.AuthInfo != nil && txDecoded.AuthInfo.Fee != nil && txDecoded.AuthInfo.Fee.GasLimit > 0 {
		txGas = uint64(txDecoded.AuthInfo.Fee.GasLimit)
	}
	// Enforce consensus max gas if configured
	if cparams, err := getConsensusParams(0); err == nil && cparams != nil {
		if cparams.Block.MaxGas > -1 && uint64(cparams.Block.MaxGas) < txGas {
			return fmt.Errorf("out of gas: %d; max %d", txGas, cparams.Block.MaxGas)
		}
	}
	// Atomic extension parsing: determine leader and ensure our chain participates if provided
	leaderChain := ""
	if len(txDecoded.Body.ExtensionOptions) > 0 {
		for _, any := range txDecoded.Body.ExtensionOptions {
			if any.TypeURL == typestnd.TypeUrl_ExtensionOptionAtomicMultiChainTx {
				ext, err := typestnd.ExtensionOptionAtomicMultiChainTxFromAnyWrap(any)
				if err != nil {
					return err
				}
				ourchain := wasmx.GetChainId()
				found := false
				for _, cid := range ext.ChainIDs {
					if cid == ourchain {
						found = true
						break
					}
				}
				if !found {
					// not for us; do not add to local mempool
					return nil
				}
				computed := multichain.GetLeaderChain(ext.ChainIDs)
				if ext.LeaderChainID != computed {
					return fmt.Errorf("atomic transaction wrong leader: expected %s, got %s", computed, ext.LeaderChainID)
				}
				leaderChain = ext.LeaderChainID
				break
			}
		}
	}
	mp.Add(txhash, txBytes, txGas, leaderChain)
	if err := SetMempool(mp); err != nil {
		return err
	}
	LoggerDebug("new transaction received", []string{"txhash", txhash})
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
		LoggerInfo("register request", []string{"IP", node.Node.IP, "address", node.Address})
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
	lastCommit, err := GetCommitIndex()
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
	if cparams, err2 := getConsensusParams(0); err2 == nil && cparams != nil {
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
	// Build prepare proposal request
	prep := typestnd.RequestPrepareProposal{
		MaxTxBytes: maxBytes,
		Txs:        batch.Txs,
		Height:     height + 1,
		Time:       wasmx.GetTimestamp().UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	_, err = consensuswrap.PrepareProposal(prep)
	if err != nil {
		return err
	}
	// Process proposal
	proc := typestnd.RequestProcessProposal{
		Txs:    batch.Txs,
		Height: height + 1,
		Time:   prep.Time,
		Hash:   consutils.GetTxsHash(batch.Txs),
	}
	resp, err := consensuswrap.ProcessProposal(proc)
	if err != nil {
		return err
	}
	if resp.Status != typestnd.ProposalStatus_ACCEPT {
		return nil
	}
	// optimistic execution if atomic tx leader
	optimistic := false
	if batch.IsAtomicTx && batch.IsLeader {
		optimistic = true
	}
	var metainfo map[string][]byte
	if optimistic {
		oe, err := doOptimisticExecution(proc, resp)
		if err == nil {
			metainfo = oe.Metainfo
		}
	}
	// Save to logs as uncommitted entry
	wrap := typestnd.RequestProcessProposalWithMetaInfo{Request: proc, OptimisticExecution: optimistic, Metainfo: metainfo}
	data, err := json.Marshal(&wrap)
	if err != nil {
		return err
	}
	term, err := GetTermId()
	if err != nil {
		return err
	}
	leader, err := GetCurrentNodeId()
	if err != nil {
		return err
	}
	entry := LogEntryAggregate{Index: height + 1, TermID: term, LeaderID: leader, Data: json.RawMessage(data)}
	return AppendLogEntry(entry)
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
	var nodeIps []NodeInfo
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
	lastIndex, err := getLastBlockIndex()
	if err != nil {
		return err
	}
	LoggerInfo("setting up last log index", []string{"index", Int64ToString(lastIndex)})
	if err := SetLastLogIndex(lastIndex); err != nil {
		return err
	}
	if err := SetCommitIndex(lastIndex); err != nil {
		return err
	}
	// after we set last log index
	if err := InitializeIndexArrays(len(nodeIps)); err != nil {
		return err
	}
	return nil
}

// Helpers
func IsNodeActive(node NodeInfo) bool {
	return !node.OutOfSync && (node.Node.IP != "" || (node.Node.Host != "" && node.Node.Port != ""))
}

func sendVoteRequest(nodeId int32, node NodeInfo, request VoteRequest, termId int32) error {
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

// getLogEntryAggregate returns a LogEntryAggregate with Data populated from
// uncommitted entry bytes if present, or from finalized storage block JSON.
func getLogEntryAggregate(index int64) (*LogEntryAggregate, error) {
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
	agg := LogEntryAggregate{Index: e.Index, TermID: e.TermID, LeaderID: e.LeaderID, Data: json.RawMessage(data)}
	return &agg, nil
}

// appendLogInternalVerified mirrors AS helper: packages request + meta then appends as log entry
func appendLogInternalVerified(processReq typestnd.RequestProcessProposal, header typestnd.Header, blockCommit typestnd.BlockCommit, optimisticExecution bool, meta map[string][]byte, validatorSet typestnd.TendermintValidators) error {
	if meta == nil {
		meta = map[string][]byte{}
	}
	hbz, _ := json.Marshal(&header)
	cbz, _ := json.Marshal(&blockCommit)
	meta["header"] = hbz
	meta["commit"] = cbz
	wrap := typestnd.RequestProcessProposalWithMetaInfo{Request: processReq, OptimisticExecution: optimisticExecution, Metainfo: meta}
	// encode wrap into LogEntryAggregate
	bz, err := json.Marshal(&wrap)
	if err != nil {
		return err
	}
	last, err := GetLastLogIndex()
	if err != nil {
		return err
	}
	term, err := GetTermId()
	if err != nil {
		return err
	}
	leader, err := GetCurrentNodeId()
	if err != nil {
		return err
	}
	entry := LogEntryAggregate{Index: last + 1, TermID: term, LeaderID: leader, Data: json.RawMessage(bz)}
	return AppendLogEntry(entry)
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

// prepareAppendEntry builds AppendEntry with log entries between nextIndex..lastIndex
func prepareAppendEntry(nodeIps []NodeInfo, nextIndex int64, lastIndex int64) (AppendEntry, error) {
	entries := make([]LogEntryAggregate, 0)
	for i := nextIndex; i <= lastIndex; i++ {
		e, err := GetLogEntryObj(i)
		if err != nil {
			return AppendEntry{}, err
		}
		if e.Index == 0 {
			continue
		}
		agg := LogEntryAggregate{Index: e.Index, TermID: e.TermID, LeaderID: e.LeaderID, Data: json.RawMessage(e.Data)}
		entries = append(entries, agg)
	}
	prev, err := GetLogEntryObj(nextIndex - 1)
	if err != nil {
		return AppendEntry{}, err
	}
	lastCommit, err := GetCommitIndex()
	if err != nil {
		return AppendEntry{}, err
	}
	term, err := GetTermId()
	if err != nil {
		return AppendEntry{}, err
	}
	leader, err := GetCurrentNodeId()
	if err != nil {
		return AppendEntry{}, err
	}
	data := AppendEntry{
		TermID:       term,
		LeaderID:     leader,
		PrevLogIndex: nextIndex - 1,
		PrevLogTerm:  prev.TermID,
		Entries:      entries,
		LeaderCommit: lastCommit,
		NodeIPs:      nodeIps,
	}
	return data, nil
}

// registeredCheckNeeded logic from AS
func registeredCheckNeeded(ips []NodeInfo) (bool, error) {
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

func registeredCheckMessage(ips []NodeInfo, nodeId int32) (string, error) {
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
	// TODO: verify signature
	LoggerDebug("updateNodeAndReturn", []string{"entry", string(data), "signature", sig})
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

func removeNode(nodes []NodeInfo, index int) []NodeInfo {
	nodes[index].Node.IP = ""
	nodes[index].Node.Host = ""
	nodes[index].Node.Port = ""
	return nodes
}

// checkCommits minimal implementation
func checkCommits() (bool, error) {
	lastCommit, err := GetCommitIndex()
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
	ncount, err := GetNodeCount()
	if err != nil {
		return false, err
	}
	committing := int64(count) >= GetMajority(ncount)
	LoggerDebug("committing diseminated block", []string{"height", Int64ToString(nextCommit)})
	if committing {
		changed, err2 := startBlockFinalizationLeader(nextCommit)
		if err2 != nil {
			return false, err2
		}
		if err := SetCommitIndex(nextCommit); err != nil {
			return false, err
		}
		if err := SetLastApplied(nextCommit); err != nil {
			return false, err
		}
		return changed, nil
	}
	return false, nil
}

// startBlockFinalizationLeader finalizes and commits a block (minimal flow)
// startBlockFinalizationInternal mirrors common finalize logic (without Commit)
func startBlockFinalizationInternal(entryobj *LogEntryAggregate, retry bool) (bool, error) {
	// entry data may be wrap JSON or BlockEntry JSON
	var wrap typestnd.RequestProcessProposalWithMetaInfo
	if err := json.Unmarshal(entryobj.Data, &wrap); err != nil {
		var be blocks.BlockEntry
		if err2 := json.Unmarshal(entryobj.Data, &be); err2 != nil {
			return false, err
		}
		if len(be.Data) == 0 {
			return false, errors.New("block entry data empty")
		}
		if err3 := json.Unmarshal(be.Data, &wrap); err3 != nil {
			return false, err3
		}
	}
	if err := verifyBlockProposalMeta(wrap); err != nil {
		return false, err
	}
	finReq := typestnd.RequestFinalizeBlock{Txs: wrap.Request.Txs, Height: entryobj.Index, Time: wrap.Request.Time, ProposerAddress: wrap.Request.ProposerAddress}
	w := typestnd.WrapRequestFinalizeBlock{Request: finReq, Metainfo: wrap.Metainfo}
	// Optimistic execution check: skip BeginBlock if already executed optimistically by proposer
	optimisticRan := wrap.OptimisticExecution && string(wrap.Request.ProposerAddress) == string(getCurrentValidator().Address)
	if !optimisticRan {
		resbegin, err := consensuswrap.BeginBlock(finReq)
		if err != nil {
			return false, err
		}
		if resbegin.Error != "" {
			// Retry on height mismatch by rolling back to height-1
			mismatch := fmt.Sprintf("expected: %d", finReq.Height+1)
			if strings.Contains(resbegin.Error, "invalid height") && strings.Contains(resbegin.Error, mismatch) {
				if err := consensuswrap.RollbackToVersion(finReq.Height - 1); err != nil {
					return false, fmt.Errorf("consensus break: %s; %v", resbegin.Error, err)
				}
				if !retry {
					return startBlockFinalizationInternal(entryobj, true)
				}
				return false, fmt.Errorf(resbegin.Error)
			} else {
				return false, fmt.Errorf(resbegin.Error)
			}
		}
	}
	finResp, err := consensuswrap.FinalizeBlock(w)
	if err != nil {
		return false, err
	}
	if finResp.Error != "" {
		return false, errors.New(finResp.Error)
	}
	if err := verifyBlockProposal(wrap, finResp.Data); err != nil {
		return false, err
	}

	// Build and store block, update state, EndBlock, parse events, Commit
	// and remove tx from mempool
	mp, err := GetMempool()
	if err != nil {
		return false, err
	}
	txhashes := make([][]byte, len(wrap.Request.Txs))
	for i := range wrap.Request.Txs {
		hash := wasmx.Sha256(wrap.Request.Txs[i])
		txhashes[i] = hash
		mp.Remove(base64.StdEncoding.EncodeToString(hash))
	}
	SetMempool(mp)
	topics := extractIndexedTopics(*finResp.Data, txhashes)
	st, _ := GetCurrentState()
	var vset *typestnd.TendermintValidators
	if vlist, err := GetAllValidators(); err == nil {
		if tvals, err2 := consutils.GetActiveValidatorInfo(vlist); err2 == nil && len(tvals) > 0 {
			vset = &typestnd.TendermintValidators{Validators: tvals}
		}
	}
	blockEntryJSON, err := buildBlockEntry(entryobj.Index, wrap, finResp.Data, string(st.ValidatorAddress), vset)
	if err != nil {
		return false, err
	}
	hashStr := base64.StdEncoding.EncodeToString(wrap.Request.Hash)
	if err := setFinalizedBlock(blockEntryJSON, hashStr, txhashes, topics); err != nil {
		return false, err
	}
	stUp, _ := GetCurrentState()
	stUp.LastBlockID = getBlockID(wrap.Request.Hash)
	if cbz, ok := wrap.Metainfo["commit"]; ok && len(cbz) > 0 {
		var commit typestnd.BlockCommit
		if err := json.Unmarshal(cbz, &commit); err == nil {
			stUp.LastCommitHash = consutils.GetCommitHash(commit)
		}
	}
	stUp.LastResultsHash = consutils.GetResultsHash(finResp.Data.TxResults)
	_ = SetCurrentState(stUp)
	respend, err := consensuswrap.EndBlock(string(blockEntryJSON))
	if err == nil && respend.Error == "" && respend.Data != nil {
		st2, _ := GetCurrentState()
		st2.AppHash = respend.Data.AppHash
		_ = SetCurrentState(st2)
	}
	aggregated := consensuswrap.AggregateEvents(respend.Data.TxResults, respend.Data.Events)
	info := consensuswrap.DefaultFinalizeResponseEventsParse(finResp.Data.TxResults, aggregated)
	if len(info.CreatedValidators) > 0 {
		for _, cv := range info.CreatedValidators {
			if int(cv.TxIndex) >= len(wrap.Request.Txs) {
				continue
			}
			decodedTx, _ := decodeTx(wrap.Request.Txs[cv.TxIndex])
			memo := decodedTx.Body.Memo
			resp := parseNodeAddress(memo)
			if resp.Error != "" {
				continue
			}
			nodeInfo := resp.NodeInfo
			if nodeInfo.Address == "" {
				nodeInfo.Address = string(cv.OperatorAddress)
			}
			updateNodeEntry(NodeUpdate{Node: nodeInfo, Index: 0, Type: NODE_UPDATE_ADD})
			bz, _ := json.Marshal(&nodeInfo)
			_ = callHookContract(wasmx.HOOK_CREATE_VALIDATOR, string(bz))
		}
	}
	if finResp.Data.ConsensusParamUpdates != nil {
		if err := updateConsensusParams(entryobj.Index, finResp.Data.ConsensusParamUpdates); err != nil {
			return false, err
		}
	}
	if _, err := consensuswrap.Commit(); err != nil {
		return false, err
	}
	return false, nil
}

func startBlockFinalizationLeader(index int64) (bool, error) {
	LoggerInfo("start block finalization", []string{"height", Int64ToString(index)})
	entryobj, err := getLogEntryAggregate(index)
	if err != nil {
		return false, err
	}
	if entryobj == nil {
		LoggerInfo("cannot start block finalization", []string{"height", Int64ToString(index), "reason", "block empty"})
		return false, nil
	}
	LoggerDebug("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID)})
	LoggerDebugExtended("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID), "data", string(entryobj.Data)})
	currentTerm, err := GetTermId()
	if err != nil {
		return false, err
	}
	if currentTerm == entryobj.TermID {
		return startBlockFinalizationInternal(entryobj, false)
	}
	LoggerError("entry has current term mismatch", []string{"nodeType", "Leader", "currentTerm", Int32ToString(currentTerm), "entryTermId", Int32ToString(entryobj.TermID)})
	return false, nil
}

// startBlockFinalizationFollower mirrors leader finalization on follower side
func startBlockFinalizationFollower(index int64) (bool, error) {
	LoggerInfo("start block finalization", []string{"height", Int64ToString(index)})
	entryobj, err := getLogEntryAggregate(index)
	if err != nil {
		return false, err
	}
	if entryobj == nil {
		LoggerInfo("cannot start block finalization", []string{"height", Int64ToString(index), "reason", "block empty"})
		return false, nil
	}
	LoggerDebug("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID)})
	LoggerDebugExtended("start block finalization", []string{"height", Int64ToString(index), "leaderId", Int32ToString(entryobj.LeaderID), "termId", Int32ToString(entryobj.TermID), "data", string(entryobj.Data)})
	return startBlockFinalizationInternal(entryobj, false)
}
