package lib

import (
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "strconv"

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
func SetupNode(_ []fsm.ActionParam, _ fsm.EventObject) error                  { return nil }
func ProcessAppendEntries(_ []fsm.ActionParam, _ fsm.EventObject) error       { return nil }

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

func SendAppendEntries(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }

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
    tx := ""
    for _, p := range event.Params {
        if p.Key == "transaction" { tx = p.Value; break }
    }
    if tx == "" { return errors.New("no transaction found") }
    mp, err := GetMempool()
    if err != nil { return err }
    // Use hash as tx itself for now; caller should pass hash as base64 if needed
    // Here we assume p.Value is the base64-encoded tx bytes; use it as key too
    mp.Add(tx, tx, 0, "")
    if err := SetMempool(mp); err != nil { return err }
    LoggerDebug("new transaction received", []string{"transaction", tx})
    return nil
}

func CommitBlocks(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }

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

func Vote(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }

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

func ForwardTxsToLeader(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }
func UpdateNodeAndReturn(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }
func RegisteredCheck(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }

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

func ProposeBlock(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }
func Setup(_ []fsm.ActionParam, _ fsm.EventObject) error { return nil }

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
