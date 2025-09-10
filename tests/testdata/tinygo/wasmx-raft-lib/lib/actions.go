package lib

import (
    "encoding/base64"
    "fmt"
    "encoding/json"
    "errors"
    "encoding/hex"
    "strings"
    "strconv"

    consutils "github.com/loredanacirstea/wasmx-consensus-utils/lib"
    consensus "github.com/loredanacirstea/wasmx-env-consensus/lib"
    wasmx "github.com/loredanacirstea/wasmx-env/lib"
    wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
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
func SetupNode(params []fsm.ActionParam, event fsm.EventObject) error {
    // Pass-through to non-consensus hook; if params exist, forward payload
    var data string
    if len(params) > 0 {
        // key/value pairs; wrap into simple JSON
        m := map[string]string{}
        for _, p := range params { m[p.Key] = p.Value }
        bz, _ := json.Marshal(&m)
        data = string(bz)
    } else if len(event.Params) > 0 {
        m := map[string]string{}
        for _, p := range event.Params { m[p.Key] = p.Value }
        bz, _ := json.Marshal(&m)
        data = string(bz)
    }
    return callHookNonCContract(wasmx.HOOK_SETUP_NODE, data)
}
func ProcessAppendEntries(params []fsm.ActionParam, event fsm.EventObject) error {
    entry, err := ExtractAppendEntry(params, event)
    if err != nil { return err }
    for _, e := range entry.Entries {
        if err := processAppendEntry(e); err != nil { return err }
    }
    LoggerDebug("new entries processing finished", []string{"leaderId", Int32ToString(entry.LeaderID), "leaderCommit", Int64ToString(entry.LeaderCommit), "prevLogIndex", Int64ToString(entry.PrevLogIndex), "count", Int32ToString(int32(len(entry.Entries)))})
    return nil
}

// processAppendEntry decodes the proposal from entry and processes it before appending
func processAppendEntry(entry LogEntryAggregate) error {
    // decode wrap from entry data
    var wrap consensus.RequestProcessProposalWithMetaInfo
    if err := json.Unmarshal(entry.Data, &wrap); err != nil { return err }
    resp, err := consensus.ProcessProposal(wrap.Request)
    if err != nil { return err }
    if resp.Status == consensus.ProposalStatus_REJECT {
        LoggerError("new block rejected", []string{"height", Int64ToString(wrap.Request.Height), "node type", "Follower"})
        return nil
    }
    return AppendLogEntry(entry)
}

func SendHeartbeatResponse(params []fsm.ActionParam, event fsm.EventObject) error {
    entry, err := ExtractAppendEntry(params, event)
    if err != nil { return err }
    resp, err := PrepareHeartbeatResponse(entry)
    if err != nil { return err }
    bz, err := json.Marshal(&resp)
    if err != nil { return err }
    wasmx.SetFinishData(bz)
    return nil
}

func SendAppendEntries(_ []fsm.ActionParam, _ fsm.EventObject) error {
    leaderId, err := GetCurrentNodeId()
    if err != nil { return err }
    ips, err := GetNodeIPs()
    if err != nil { return err }
    LoggerDebug("diseminate entries...", []string{"nodeId", Int32ToString(leaderId)})
    for i, ip := range ips {
        if int32(i) == leaderId || !isNodeActive(ip) { continue }
        // determine next index for this node
        nextArr, err := GetNextIndexArray()
        if err != nil { return err }
        var nextIndex int64 = LOG_START
        if i < len(nextArr) { nextIndex = nextArr[i] } else { nextIndex = LOG_START }
        lastIndex, err := GetLastLogIndex()
        if err != nil { return err }
        // batch limit similar to AS
        lastIndexToSend := lastIndex
        if (lastIndex - nextIndex) > int64(STATE_SYNC_BATCH) {
            lastIndexToSend = nextIndex + int64(STATE_SYNC_BATCH)
        }
        data, err := prepareAppendEntry(ips, nextIndex, lastIndexToSend)
        if err != nil { return err }
        msg, err := PrepareAppendEntryMessage(int32(i), nextIndex, lastIndex, lastIndexToSend, ip, data)
        if err != nil { return err }
        // send
        contract := wasmx.GetAddress()
        resp, err := SendGrpcJSONBase64(ip.Node.IP, contract, msg)
        if err != nil { continue }
        if resp.Error != "" || resp.Data == "" { continue }
        var vresp AppendEntryResponse
        if err := json.Unmarshal([]byte(resp.Data), &vresp); err != nil { continue }
        if vresp.Success {
            nextArr, err := GetNextIndexArray()
            if err != nil { return err }
            // update next index with number of entries we sent
            sent := int64(len(data.Entries))
            if i < len(nextArr) { nextArr[i] = nextIndex + sent } else {
                // grow
                needed := i - len(nextArr) + 1
                for n:=0;n<needed;n++{ nextArr = append(nextArr, LOG_START) }
                nextArr[i] = nextIndex + sent
            }
            if err := SetNextIndexArray(nextArr); err != nil { return err }
        }
    }
    return nil
}

func SendNewTransactionResponse(_ []fsm.ActionParam, _ fsm.EventObject) error {
    e, err := GetLogEntryObjIndexLast()
    if err != nil { return err }
    resp := TransactionResponse{TermID: e.TermID, LeaderID: e.LeaderID, Index: e.Index}
    bz, err := json.Marshal(&resp)
    if err != nil { return err }
    wasmx.SetFinishData(bz)
    return nil
}

func AddToMempool(_ []fsm.ActionParam, event fsm.EventObject) error {
    txB64 := ""
    for _, p := range event.Params {
        if p.Key == "transaction" { txB64 = p.Value; break }
    }
    if txB64 == "" { return errors.New("no transaction found") }
    // Decode transaction payload (base64 => []byte)
    txBytes, err := base64.StdEncoding.DecodeString(txB64)
    if err != nil { return err }
    mp, err := GetMempool()
    if err != nil { return err }
    // Use the base64 value as key, store raw bytes as payload
    mp.Add(txB64, txBytes, 0, "")
    if err := SetMempool(mp); err != nil { return err }
    LoggerDebug("new transaction received", []string{"transaction", txB64})
    return nil
}

func CommitBlocks(_ []fsm.ActionParam, _ fsm.EventObject) error {
    changed, err := checkCommits()
    if err != nil { return err }
    if changed {
        // propagate commit to others
        return SendAppendEntries(nil, fsm.EventObject{})
    }
    return nil
}

func SetRandomElectionTimeout(params []fsm.ActionParam, event fsm.EventObject) error {
    var minS, maxS string
    if len(params) == 0 { params = event.Params }
    for _, p := range params {
        if p.Key == "min" { minS = p.Value }
        if p.Key == "max" { maxS = p.Value }
    }
    if minS == "" { return errors.New("no min found") }
    if maxS == "" { return errors.New("no max found") }
    min, err := strconv.ParseInt(minS, 10, 64)
    if err != nil { return err }
    max, err := strconv.ParseInt(maxS, 10, 64)
    if err != nil { return err }
    v, err := GetRandomInRange(min, max)
    if err != nil { return err }
    return SetElectionTimeout(v)
}

func InitializeNextIndex(_ []fsm.ActionParam, _ fsm.EventObject) error {
    last, err := GetLastLogIndex()
    if err != nil { return err }
    next := last + 1
    arr, err := GetNextIndexArray()
    if err != nil { return err }
    if len(arr) == 0 {
        // init from node count
        n, err := GetNodeCount()
        if err != nil { return err }
        arr = make([]int64, n)
    }
    for i := range arr { arr[i] = next }
    return SetNextIndexArray(arr)
}

func InitializeMatchIndex(_ []fsm.ActionParam, _ fsm.EventObject) error {
    arr, err := GetMatchIndexArray()
    if err != nil { return err }
    if len(arr) == 0 {
        n, err := GetNodeCount()
        if err != nil { return err }
        arr = make([]int64, n)
    }
    for i := range arr { arr[i] = 0 }
    return SetMatchIndexArray(arr)
}

func IncrementCurrentTerm(_ []fsm.ActionParam, _ fsm.EventObject) error {
    t, err := GetTermId()
    if err != nil { return err }
    return SetTermId(t + 1)
}

func Vote(_ []fsm.ActionParam, event fsm.EventObject) error {
    entryB64 := ""
    sig := ""
    for _, p := range event.Params {
        if p.Key == "entry" { entryB64 = p.Value }
        if p.Key == "signature" { sig = p.Value }
    }
    if entryB64 == "" { return errors.New("vote: empty entry") }
    if sig == "" { return errors.New("vote: empty signature") }
    data, err := base64.StdEncoding.DecodeString(entryB64)
    if err != nil { return err }
    var req VoteRequest
    if err := json.Unmarshal(data, &req); err != nil { return err }
    // verify signature against candidate's pubkey
    ok, err := verifyMessage(req.CandidateID, sig, string(data))
    if err != nil { return err }
    if !ok {
        LoggerError("signature verification failed for VoteRequest", []string{"candidateId", Int32ToString(req.CandidateID), "termId", Int32ToString(req.TermID)})
        return nil
    }
    resp, err := voteInternal(req.TermID, req.CandidateID, req.LastLogIndex, req.LastLogTerm)
    if err != nil { return err }
    bz, err := json.Marshal(&resp)
    if err != nil { return err }
    wasmx.SetFinishData(bz)
    return nil
}

func SelfVote(_ []fsm.ActionParam, _ fsm.EventObject) error {
    myId, err := GetCurrentNodeId()
    if err != nil { return err }
    if err := SetVotedFor(myId); err != nil { return err }
    ips, err := GetNodeIPs()
    if err != nil { return err }
    voteArray := make([]int32, len(ips))
    for i := range voteArray { voteArray[i] = 0 }
    if int(myId) < len(voteArray) { voteArray[myId] = 1 }
    return SetVoteIndexArray(voteArray)
}

func ForwardTxsToLeader(_ []fsm.ActionParam, _ fsm.EventObject) error {
    mp, err := GetMempool()
    if err != nil { return err }
    if len(mp.Map) == 0 { return nil }
    // get leader from last log
    last, err := GetLogEntryObjIndexLast()
    if err != nil { return err }
    if last.Index == 0 { return nil }
    nodeId := last.LeaderID
    ips, err := GetNodeIPs()
    if err != nil { return err }
    if int(nodeId) >= len(ips) { return nil }
    target := ips[nodeId]
    contract := wasmx.GetAddress()
    limit := 0
    for txhash, tx := range mp.Map {
        if limit >= 5 { break }
        msg := fmt.Sprintf(`{"run":{"event":{"type":"newTransaction","params":[{"key":"transaction","value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(tx.Tx))
        resp, err := SendGrpcJSONBase64(target.Node.IP, contract, msg)
        LoggerDebug("forwarding tx to leader", []string{"nodeId", Int32ToString(nodeId), "nodeIp", target.Node.IP, "batch", Int32ToString(int32(limit)), "error", resp.Error})
        if err == nil && resp.Error == "" {
            delete(mp.Map, txhash)
            limit++
            continue
        }
        if err == nil && resp.Error != "" && resp.Error == ERROR_INVALID_TX {
            delete(mp.Map, txhash)
            LoggerDebug("forwarded invalid transaction", []string{"tx", txhash})
        } else {
            break
        }
    }
    return SetMempool(mp)
}
func UpdateNodeAndReturn(params []fsm.ActionParam, event fsm.EventObject) error {
    entry, err := extractUpdateNodeEntryAndVerify(params, event)
    if err != nil { return err }
    resp, err := updateNodeEntry(entry)
    if err != nil { return err }
    bz, err := json.Marshal(&resp)
    if err != nil { return err }
    wasmx.SetFinishData(bz)
    return nil
}
func RegisteredCheck(_ []fsm.ActionParam, _ fsm.EventObject) error {
    ips, err := GetNodeIPs()
    if err != nil { return err }
    needed, err := registeredCheckNeeded(ips)
    if err != nil { return err }
    if !needed { return nil }
    nodeId, err := GetCurrentNodeId()
    if err != nil { return err }
    msgstr, err := registeredCheckMessage(ips, nodeId)
    if err != nil { return err }
    LoggerInfo("register request", []string{"req", msgstr})
    contract := wasmx.GetAddress()
    for i, node := range ips {
        if int32(i) == nodeId || node.Node.IP == "" { continue }
        LoggerInfo("register request", []string{"IP", node.Node.IP, "address", node.Address})
        resp, err := SendGrpcJSONBase64(node.Node.IP, contract, msgstr)
        if err != nil || resp.Error != "" || resp.Data == "" { continue }
        var update UpdateNodeResponse
        if err := json.Unmarshal([]byte(resp.Data), &update); err != nil { continue }
        ourId := -1
        for j, n := range update.Nodes {
            if n.Address == ips[nodeId].Address { ourId = j; break }
        }
        if ourId != -1 {
            if err := SetCurrentNodeId(int32(ourId)); err != nil { return err }
            if err := SetNodeIPs(update.Nodes); err != nil { return err }
            break
        }
    }
    return nil
}

func SendVoteRequests(_ []fsm.ActionParam, _ fsm.EventObject) error {
    termId, err := GetTermId()
    if err != nil { return err }
    candidateId, err := GetCurrentNodeId()
    if err != nil { return err }
    lastLogIndex, err := GetLastLogIndex()
    if err != nil { return err }
    lastEntry, err := GetLogEntryObjIndexLast()
    if err != nil { return err }
    lastLogTerm := lastEntry.TermID
    request := VoteRequest{TermID: termId, CandidateID: candidateId, LastLogIndex: lastLogIndex, LastLogTerm: lastLogTerm}
    ips, err := GetNodeIPs()
    if err != nil { return err }
    for i, ip := range ips {
        if int32(i) == candidateId || !isNodeActive(ip) { continue }
        if err := sendVoteRequest(int32(i), ip, request, termId); err != nil { return err }
    }
    return nil
}

func ProposeBlock(_ []fsm.ActionParam, _ fsm.EventObject) error {
    // Only propose if last block is committed
    height, err := GetLastLogIndex()
    if err != nil { return err }
    lastCommit, err := GetCommitIndex()
    if err != nil { return err }
    if lastCommit < height {
        LoggerInfo("cannot propose new block, last block not committed", []string{"height", Int64ToString(height), "lastCommitIndex", Int64ToString(lastCommit)})
        return nil
    }
    // gather batch
    mp, err := GetMempool()
    if err != nil { return err }
    maxBytes, err := GetMaxTxBytes()
    if err != nil { return err }
    if maxBytes == 0 { maxBytes = MaxBlockSizeBytes }
    batch, err := mp.Batch(-1, maxBytes, wasmx.GetChainId())
    if err != nil { return err }
    // Build prepare proposal request
    prep := consensus.RequestPrepareProposal{
        MaxTxBytes: maxBytes,
        Txs:        batch.Txs,
        Height:     height + 1,
        Time:       wasmx.GetTimestamp().UTC().Format("2006-01-02T15:04:05Z07:00"),
    }
    _, err = consensus.PrepareProposal(prep)
    if err != nil { return err }
    // Process proposal
    proc := consensus.RequestProcessProposal{
        Txs:     batch.Txs,
        Height:  height + 1,
        Time:    prep.Time,
        Hash:    consutils.GetTxsHash(batch.Txs),
    }
    resp, err := consensus.ProcessProposal(proc)
    if err != nil { return err }
    if resp.Status != consensus.ProposalStatus_ACCEPT { return nil }
    // optimistic execution if atomic tx leader
    optimistic := false
    if batch.IsAtomicTx && batch.IsLeader { optimistic = true }
    var metainfo map[string][]byte
    if optimistic {
        oe, err := doOptimisticExecution(proc, resp)
        if err == nil { metainfo = oe.Metainfo }
    }
    // Save to logs as uncommitted entry
    wrap := consensus.RequestProcessProposalWithMetaInfo{Request: proc, OptimisticExecution: optimistic, Metainfo: metainfo}
    data, err := json.Marshal(&wrap)
    if err != nil { return err }
    term, err := GetTermId()
    if err != nil { return err }
    leader, err := GetCurrentNodeId()
    if err != nil { return err }
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
    if oldContract == "" { return errors.New("previous contract address not provided") }
    // helper to call getContextValue
    getCtx := func(key string) (string, error) {
        payload := struct{ GetContextValue struct{ Key string `json:"key"` } `json:"getContextValue"` }{}
        payload.GetContextValue.Key = key
        bz, _ := json.Marshal(&payload)
        addr := wasmx.Bech32String(oldContract)
        resp, err := callContract(addr, string(bz), true, MODULE_NAME)
        if err != nil { return "", err }
        if resp.Success > 0 { return "", fmt.Errorf("cannot get %s from previous contract", key) }
        // Data may be base64 or raw JSON; try base64 first
        if dbz, err := base64.StdEncoding.DecodeString(resp.Data); err == nil {
            return string(dbz), nil
        }
        return resp.Data, nil
    }
    // NODE_IPS
    data, err := getCtx(NODE_IPS)
    if err != nil { return err }
    LoggerInfo("setting up nodeIPs", []string{"ips", data})
    var nodeIps []NodeInfo
    if err := json.Unmarshal([]byte(data), &nodeIps); err != nil { return err }
    if err := SetNodeIPs(nodeIps); err != nil { return err }
    // state
    data, err = getCtx(STATE_KEY)
    if err != nil { return err }
    LoggerInfo("setting up state", []string{"data", data})
    var state CurrentState
    if err := json.Unmarshal([]byte(data), &state); err != nil { return err }
    if err := SetCurrentState(state); err != nil { return err }
    // mempool
    data, err = getCtx(MEMPOOL_KEY)
    if err != nil { return err }
    LoggerInfo("setting up mempool", []string{"data", data})
    var mempool Mempool
    if err := json.Unmarshal([]byte(data), &mempool); err != nil { return err }
    if err := SetMempool(mempool); err != nil { return err }
    // currentNodeId
    data, err = getCtx(CURRENT_NODE_ID)
    if err != nil { return err }
    LoggerInfo("setting up currentNodeId", []string{"data", data})
    if v, err := ParseI32(data); err == nil { if err := SetCurrentNodeId(v); err != nil { return err } }
    // currentTerm
    data, err = getCtx(TERM_ID)
    if err != nil { return err }
    LoggerInfo("setting up currentTerm", []string{"data", data})
    if v, err := ParseI32(data); err == nil { if err := SetTermId(v); err != nil { return err } }
    // last block index from storage contract
    lastIndex, err := getLastBlockIndex()
    if err != nil { return err }
    LoggerInfo("setting up last log index", []string{"index", Int64ToString(lastIndex)})
    if err := SetLastLogIndex(lastIndex); err != nil { return err }
    if err := SetCommitIndex(lastIndex); err != nil { return err }
    // after we set last log index
    if err := initializeIndexArrays(len(nodeIps)); err != nil { return err }
    return nil
}

// Helpers
func isNodeActive(node NodeInfo) bool {
    return !node.OutOfSync && (node.Node.IP != "" || (node.Node.Host != "" && node.Node.Port != ""))
}

func sendVoteRequest(nodeId int32, node NodeInfo, request VoteRequest, termId int32) error {
    datastrBz, err := json.Marshal(&request)
    if err != nil { return err }
    datastr := string(datastrBz)
    signature, err := SignMessage(datastr)
    if err != nil { return err }
    dataBase64 := base64.StdEncoding.EncodeToString([]byte(datastr))
    msgstr := fmt.Sprintf(`{"run":{"event":{"type":"receiveVoteRequest","params":[{"key":"entry","value":"%s"},{"key":"signature","value":"%s"}]}}}`, dataBase64, signature)
    contract := wasmx.GetAddress()
    wasmx.LoggerDebug(MODULE_NAME, "sending vote request", []string{"nodeId", Int32ToString(nodeId), "nodeIp", node.Node.IP, "termId", Int32ToString(termId), "data", datastr})
    resp, err := wasmxcore.GrpcRequest(node.Node.IP, contract, base64.StdEncoding.EncodeToString([]byte(msgstr)))
    if err != nil { return err }
    wasmx.LoggerDebug(MODULE_NAME, "vote request response", []string{"nodeId", Int32ToString(nodeId), "nodeIp", node.Node.IP, "termId", Int32ToString(termId), "data", resp.Data, "error", resp.Error})
    if resp.Error != "" || resp.Data == "" { return nil }
    var vresp VoteResponse
    if err := json.Unmarshal([]byte(resp.Data), &vresp); err != nil { return err }
    if vresp.VoteGranted {
        voteArray, err := GetVoteIndexArray()
        if err != nil { return err }
        if int(nodeId) >= 0 && int(nodeId) < len(voteArray) {
            voteArray[nodeId] = 1
        }
        return SetVoteIndexArray(voteArray)
    }
    if vresp.TermID > termId { return SetTermId(vresp.TermID) }
    return nil
}

// ExtractAppendEntry similar to AS
func ExtractAppendEntry(_ []fsm.ActionParam, event fsm.EventObject) (AppendEntry, error) {
    entryBase64 := ""
    for _, p := range event.Params {
        if p.Key == "entry" { entryBase64 = p.Value; break }
    }
    if entryBase64 == "" { return AppendEntry{}, errors.New("update node: empty entry") }
    data, err := base64.StdEncoding.DecodeString(entryBase64)
    if err != nil { return AppendEntry{}, err }
    var entry AppendEntry
    if err := json.Unmarshal(data, &entry); err != nil { return AppendEntry{}, err }
    return entry, nil
}

func PrepareHeartbeatResponse(entry AppendEntry) (AppendEntryResponse, error) {
    termId, err := GetTermId()
    if err != nil { return AppendEntryResponse{}, err }
    lastLogIndex, err := GetLastLogIndex()
    if err != nil { return AppendEntryResponse{}, err }
    successful := true
    for _, e := range entry.Entries {
        if e.Index > lastLogIndex { successful = false; break }
    }
    resp := AppendEntryResponse{TermID: termId, Success: successful, LastIndex: lastLogIndex}
    LoggerDebug("send heartbeat response", []string{"termId", Int32ToString(termId), "success", "true", "lastLogIndex", Int64ToString(lastLogIndex)})
    return resp, nil
}

// Get latest log entry
func GetLogEntryObjIndexLast() (LogEntry, error) {
    idx, err := GetLastLogIndex()
    if err != nil { return LogEntry{}, err }
    return GetLogEntryObj(idx)
}

// appendLogInternalVerified mirrors AS helper: packages request + meta then appends as log entry
func appendLogInternalVerified(processReq consensus.RequestProcessProposal, header consensus.Header, blockCommit consensus.BlockCommit, optimisticExecution bool, meta map[string][]byte, validatorSet consensus.TendermintValidators) error {
    wrap := consensus.RequestProcessProposalWithMetaInfo{Request: processReq, OptimisticExecution: optimisticExecution, Metainfo: meta}
    // encode wrap into LogEntryAggregate
    bz, err := json.Marshal(&wrap)
    if err != nil { return err }
    last, err := GetLastLogIndex()
    if err != nil { return err }
    term, err := GetTermId()
    if err != nil { return err }
    leader, err := GetCurrentNodeId()
    if err != nil { return err }
    entry := LogEntryAggregate{Index: last + 1, TermID: term, LeaderID: leader, Data: json.RawMessage(bz)}
    return AppendLogEntry(entry)
}

// voteInternal simplified version with error surfacing
func voteInternal(termId int32, candidateId int32, lastLogIndex int64, lastLogTerm int32) (VoteResponse, error) {
    myTerm, err := GetTermId()
    if err != nil { return VoteResponse{}, err }
    nodes, err := GetNodeIPs()
    if err != nil { return VoteResponse{}, err }
    if int(candidateId) >= len(nodes) || !isNodeActive(nodes[candidateId]) {
        return VoteResponse{TermID: myTerm, VoteGranted: false}, nil
    }
    if termId < myTerm {
        return VoteResponse{TermID: myTerm, VoteGranted: false}, nil
    }
    if termId > myTerm {
        if err := SetTermId(termId); err != nil { return VoteResponse{}, err }
    }
    hv, err := HasVotedFor()
    if err != nil { return VoteResponse{}, err }
    if hv {
        return VoteResponse{TermID: termId, VoteGranted: false}, nil
    }
    // simple log freshness check
    myLast, err := GetLastLogIndex()
    if err != nil { return VoteResponse{}, err }
    if lastLogIndex < myLast {
        return VoteResponse{TermID: termId, VoteGranted: false}, nil
    }
    if err := SetVotedFor(candidateId); err != nil { return VoteResponse{}, err }
    return VoteResponse{TermID: termId, VoteGranted: true}, nil
}

// prepareAppendEntry builds AppendEntry with log entries between nextIndex..lastIndex
func prepareAppendEntry(nodeIps []NodeInfo, nextIndex int64, lastIndex int64) (AppendEntry, error) {
    entries := make([]LogEntryAggregate, 0)
    for i := nextIndex; i <= lastIndex; i++ {
        e, err := GetLogEntryObj(i)
        if err != nil { return AppendEntry{}, err }
        if e.Index == 0 { continue }
        agg := LogEntryAggregate{Index: e.Index, TermID: e.TermID, LeaderID: e.LeaderID, Data: json.RawMessage(e.Data)}
        entries = append(entries, agg)
    }
    prev, err := GetLogEntryObj(nextIndex - 1)
    if err != nil { return AppendEntry{}, err }
    lastCommit, err := GetCommitIndex()
    if err != nil { return AppendEntry{}, err }
    term, err := GetTermId()
    if err != nil { return AppendEntry{}, err }
    leader, err := GetCurrentNodeId()
    if err != nil { return AppendEntry{}, err }
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
    if err != nil { return false, err }
    if lastIndex > LOG_START { return false, nil }
    if len(ips) == 1 { return false, nil }
    term, err := GetTermId()
    if err != nil { return false, err }
    if term < 2 { return false, nil }
    return true, nil
}

func registeredCheckMessage(ips []NodeInfo, nodeId int32) (string, error) {
    if int(nodeId) >= len(ips) { return "", errors.New("invalid node id") }
    LoggerInfo("trying to register node IP with Leader", nil)
    update := NodeUpdate{Node: ips[nodeId], Index: nodeId, Type: NODE_UPDATE_ADD}
    dataBz, err := json.Marshal(&update)
    if err != nil { return "", err }
    signature, err := SignMessage(string(dataBz))
    if err != nil { return "", err }
    dataBase64 := base64.StdEncoding.EncodeToString(dataBz)
    msg := fmt.Sprintf(`{"run":{"event":{"type":"updateNode","params":[{"key":"entry","value":"%s"},{"key":"signature","value":"%s"}]}}}`, dataBase64, signature)
    LoggerInfo("register request", []string{"req", msg})
    return msg, nil
}

func extractUpdateNodeEntryAndVerify(_ []fsm.ActionParam, event fsm.EventObject) (NodeUpdate, error) {
    entryB64 := ""
    sig := ""
    for _, p := range event.Params {
        if p.Key == "entry" { entryB64 = p.Value }
        if p.Key == "signature" { sig = p.Value }
    }
    if entryB64 == "" { return NodeUpdate{}, errors.New("updateNodeAndReturn: empty entry") }
    if sig == "" { return NodeUpdate{}, errors.New("updateNodeAndReturn: empty signature") }
    data, err := base64.StdEncoding.DecodeString(entryB64)
    if err != nil { return NodeUpdate{}, err }
    var entry NodeUpdate
    if err := json.Unmarshal(data, &entry); err != nil { return NodeUpdate{}, err }
    // TODO: verify signature
    LoggerDebug("updateNodeAndReturn", []string{"entry", string(data), "signature", sig})
    return entry, nil
}

func updateNodeEntry(entry NodeUpdate) (UpdateNodeResponse, error) {
    ips, err := GetNodeIPs()
    if err != nil { return UpdateNodeResponse{}, err }
    if entry.Type == NODE_UPDATE_ADD {
        if entry.Node.Node.IP == "" && entry.Node.Node.Host == "" {
            return UpdateNodeResponse{}, errors.New("validator info missing from node update")
        }
        ndx := -1
        for i, n := range ips {
            if n.Address == entry.Node.Address { ndx = i; break }
        }
        if ndx > -1 { ips[ndx].Node = entry.Node.Node } else { ips = append(ips, entry.Node) }
    } else if entry.Type == NODE_UPDATE_REMOVE {
        if int(entry.Index) >= 0 && int(entry.Index) < len(ips) {
            ips[entry.Index] = removeNode(ips, int(entry.Index))[entry.Index]
        }
    }
    if err := SetNodeIPs(ips); err != nil { return UpdateNodeResponse{}, err }
    nodeId, err := GetCurrentNodeId()
    if err != nil { return UpdateNodeResponse{}, err }
    last, err := GetLastLogIndex()
    if err != nil { return UpdateNodeResponse{}, err }
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
    if err != nil { return false, err }
    lastLog, err := GetLastLogIndex()
    if err != nil { return false, err }
    nextCommit := lastCommit + 1
    if lastLog < nextCommit { return false, nil }
    nextArr, err := GetNextIndexArray()
    if err != nil { return false, err }
    count := 1 // leader
    for _, v := range nextArr {
        if v > nextCommit { count++ }
    }
    ncount, err := GetNodeCount()
    if err != nil { return false, err }
    committing := int64(count) >= GetMajority(ncount)
    LoggerDebug("committing diseminated block", []string{"height", Int64ToString(nextCommit)})
    if committing {
        changed, err2 := startBlockFinalizationLeader(nextCommit)
        if err2 != nil { return false, err2 }
        if err := SetCommitIndex(nextCommit); err != nil { return false, err }
        if err := SetLastApplied(nextCommit); err != nil { return false, err }
        return changed, nil
    }
    return false, nil
}

// startBlockFinalizationLeader finalizes and commits a block (minimal flow)
// startBlockFinalizationInternal mirrors common finalize logic (without Commit)
func startBlockFinalizationInternal(height int64) (consensus.RequestProcessProposalWithMetaInfo, *consensus.ResponseFinalizeBlock, error) {
    e, err := GetLogEntryObj(height)
    if err != nil { return consensus.RequestProcessProposalWithMetaInfo{}, nil, err }
    if e.Index == 0 { return consensus.RequestProcessProposalWithMetaInfo{}, nil, errors.New("missing log entry") }
    var wrap consensus.RequestProcessProposalWithMetaInfo
    if err := json.Unmarshal(e.Data, &wrap); err != nil { return consensus.RequestProcessProposalWithMetaInfo{}, nil, err }
    if err := verifyBlockProposalMeta(wrap); err != nil { return consensus.RequestProcessProposalWithMetaInfo{}, nil, err }
    finReq := consensus.RequestFinalizeBlock{Txs: wrap.Request.Txs, Height: height, Time: wrap.Request.Time, ProposerAddress: wrap.Request.ProposerAddress}
    w := consensus.WrapRequestFinalizeBlock{Request: finReq, Metainfo: wrap.Metainfo}
    // Optimistic execution check: skip BeginBlock if already executed optimistically by proposer
    optimisticRan := wrap.OptimisticExecution && string(wrap.Request.ProposerAddress) == string(getCurrentValidator().Address)
    if !optimisticRan {
        resbegin, err := consensus.BeginBlock(finReq)
        if err != nil { return consensus.RequestProcessProposalWithMetaInfo{}, nil, err }
        if resbegin.Error != "" {
            // Retry on height mismatch by rolling back to height-1
            mismatch := fmt.Sprintf("expected: %d", finReq.Height+1)
            if strings.Contains(resbegin.Error, "invalid height") && strings.Contains(resbegin.Error, mismatch) {
                if err := consensus.RollbackToVersion(finReq.Height - 1); err != nil { return consensus.RequestProcessProposalWithMetaInfo{}, nil, fmt.Errorf("consensus break: %s; %v", resbegin.Error, err) }
                // retry once
                resbegin2, err2 := consensus.BeginBlock(finReq)
                if err2 != nil || resbegin2.Error != "" { return consensus.RequestProcessProposalWithMetaInfo{}, nil, fmt.Errorf("BeginBlock failed after rollback: %v %s", err2, resbegin2.Error) }
            } else {
                return consensus.RequestProcessProposalWithMetaInfo{}, nil, fmt.Errorf(resbegin.Error)
            }
        }
    }
    finResp, err := consensus.FinalizeBlock(w)
    if err != nil { return consensus.RequestProcessProposalWithMetaInfo{}, nil, err }
    if finResp.Error != "" { return consensus.RequestProcessProposalWithMetaInfo{}, nil, errors.New(finResp.Error) }
    if err := verifyBlockProposal(wrap, finResp.Data); err != nil { return consensus.RequestProcessProposalWithMetaInfo{}, nil, err }
    return wrap, finResp.Data, nil
}

func startBlockFinalizationLeader(height int64) (bool, error) {
    wrap, finData, err := startBlockFinalizationInternal(height)
    if err != nil { return false, err }
    // index topics & store block
    txhashes := make([][]byte, len(wrap.Request.Txs))
    for i := range wrap.Request.Txs { txhashes[i] = wrap.Request.Txs[i] }
    topics := extractIndexedTopics(*finData, txhashes)
    st, _ := GetCurrentState()
    // include validator set if available
    var vset *consensus.TendermintValidators
    if vlist, err := getAllValidators(); err == nil {
        if tvals, err2 := consutils.GetActiveValidatorInfo(vlist); err2 == nil && len(tvals) > 0 {
            vset = &consensus.TendermintValidators{Validators: tvals}
        }
    }
    blockEntryJSON, err := buildBlockEntry(height, wrap, finData, string(st.ValidatorAddress), vset)
    if err != nil { return false, err }
    // Use header hash from proposal for storage entry (base64 like AS)
    hashStr := base64.StdEncoding.EncodeToString(wrap.Request.Hash)
    if err := setFinalizedBlock(blockEntryJSON, hashStr, txhashes, topics); err != nil { return false, err }
    // EndBlock: execute governance proposals; update app_hash from response
    respend, err := consensus.EndBlock(string(blockEntryJSON))
    if err == nil && respend.Error == "" && respend.Data != nil {
        st2, _ := GetCurrentState()
        st2.AppHash = respend.Data.AppHash
        _ = SetCurrentState(st2)
    }
    if finData.ConsensusParamUpdates != nil {
        if err := updateConsensusParams(height, finData.ConsensusParamUpdates); err != nil { return false, err }
    }
    if _, err := consensus.Commit(); err != nil { return false, err }
    return false, nil
}

// startBlockFinalizationFollower mirrors leader finalization on follower side
func startBlockFinalizationFollower(height int64) (bool, error) {
    _, _, err := startBlockFinalizationInternal(height)
    if err != nil { return false, err }
    if _, err := consensus.Commit(); err != nil { return false, err }
    return false, nil
}
