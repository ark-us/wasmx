package lib

import (
    "encoding/json"
    "errors"
    "fmt"

    wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// Helper storage key not defined in config.ts
const LOGS_LAST_INDEX = "logs_last_index"

// Helpers for simple typed storage
func sGet(key string) string      { return string(wasmx.StorageLoad([]byte(key))) }
func sSet(key string, val string) { wasmx.StorageStore([]byte(key), []byte(val)) }

// ===== State getters/setters with error surfacing =====

// Node list
func GetNodeIPs() ([]NodeInfo, error) {
    v := sGet(NODE_IPS)
    if v == "" {
        return []NodeInfo{}, nil
    }
    var nodes []NodeInfo
    if err := json.Unmarshal([]byte(v), &nodes); err != nil {
        return nil, err
    }
    return nodes, nil
}

func SetNodeIPs(nodes []NodeInfo) error {
    bz, err := json.Marshal(nodes)
    if err != nil {
        return err
    }
    sSet(NODE_IPS, string(bz))
    return nil
}

func GetNodeCount() (int, error) {
    nodes, err := GetNodeIPs()
    if err != nil {
        return 0, err
    }
    c := 0
    for _, n := range nodes {
        if n.Node.IP != "" || (n.Node.Host != "" && n.Node.Port != "") {
            c++
        }
    }
    return c, nil
}

// Voting
func GetVoteIndexArray() ([]int32, error) {
    v := sGet(VOTE_INDEX_ARRAY)
    if v == "" {
        return []int32{}, nil
    }
    var arr []int32
    if err := json.Unmarshal([]byte(v), &arr); err != nil {
        return nil, err
    }
    return arr, nil
}

func SetVoteIndexArray(arr []int32) error {
    bz, err := json.Marshal(arr)
    if err != nil {
        return err
    }
    sSet(VOTE_INDEX_ARRAY, string(bz))
    return nil
}

// Indices and terms
func GetLastLogIndex() (int64, error) {
    v := sGet(LOGS_LAST_INDEX)
    if v == "" {
        return LOG_START, nil
    }
    return parseI64(v, LOG_START), nil
}

func SetLastLogIndex(val int64) error { sSet(LOGS_LAST_INDEX, i64ToStr(val)); return nil }

func GetCommitIndex() (int64, error) {
    v := sGet(COMMIT_INDEX)
    if v == "" {
        return LOG_START, nil
    }
    return parseI64(v, LOG_START), nil
}

func SetCommitIndex(val int64) error { sSet(COMMIT_INDEX, i64ToStr(val)); return nil }

func GetTermId() (int32, error) {
    v := sGet(TERM_ID)
    if v == "" {
        return 0, nil
    }
    return parseI32(v, 0), nil
}

func SetTermId(val int32) error { sSet(TERM_ID, i32ToStr(val)); return nil }

func GetCurrentNodeId() (int32, error) {
    v := sGet(CURRENT_NODE_ID)
    if v == "" {
        return 0, nil
    }
    return parseI32(v, 0), nil
}

func SetCurrentNodeId(idx int32) error { sSet(CURRENT_NODE_ID, i32ToStr(idx)); return nil }

// Election timeout
func GetElectionTimeout() (int64, error) {
    v := sGet(ELECTION_TIMEOUT_KEY)
    if v == "" {
        return 0, nil
    }
    return parseI64(v, 0), nil
}

func SetElectionTimeout(val int64) error { sSet(ELECTION_TIMEOUT_KEY, i64ToStr(val)); return nil }

// LastApplied
func GetLastApplied() (int64, error) {
    v := sGet(LAST_APPLIED)
    if v == "" {
        return LOG_START, nil
    }
    return parseI64(v, LOG_START), nil
}

func SetLastApplied(val int64) error { sSet(LAST_APPLIED, i64ToStr(val)); return nil }

// Mempool
func GetMempool() (Mempool, error) {
    v := sGet(MEMPOOL_KEY)
    if v == "" {
        return NewMempool(map[string]MempoolTx{}, 200), nil
    }
    var mp Mempool
    if err := json.Unmarshal([]byte(v), &mp); err != nil {
        return Mempool{}, err
    }
    if mp.Map == nil { mp.Map = map[string]MempoolTx{} }
    if mp.Temp == nil { mp.Temp = map[string]bool{} }
    if mp.ToBeSent == nil { mp.ToBeSent = map[string]bool{} }
    if mp.TempSize == 0 { mp.TempSize = 200 }
    return mp, nil
}

func SetMempool(m Mempool) error {
    bz, err := json.Marshal(m)
    if err != nil { return err }
    sSet(MEMPOOL_KEY, string(bz))
    return nil
}

// MaxTxBytes
func GetMaxTxBytes() (int64, error) {
    v := sGet(MAX_TX_BYTES)
    if v == "" {
        return 0, nil
    }
    return parseI64(v, 0), nil
}

func SetMaxTxBytes(val int64) error { sSet(MAX_TX_BYTES, i64ToStr(val)); return nil }

// Current state
func GetCurrentState() (CurrentState, error) {
    v := sGet(STATE_KEY)
    if v == "" {
        return CurrentState{}, errors.New("chain init setup was not ran")
    }
    var st CurrentState
    if err := json.Unmarshal([]byte(v), &st); err != nil {
        return CurrentState{}, err
    }
    return st, nil
}

func SetCurrentState(st CurrentState) error {
    bz, err := json.Marshal(&st)
    if err != nil { return err }
    sSet(STATE_KEY, string(bz))
    return nil
}

// VotedFor
func HasVotedFor() (bool, error) {
    v, err := GetVotedForInternal()
    if err != nil { return false, err }
    return v > 0, nil
}

func GetVotedForInternal() (int32, error) {
    v := sGet(VOTED_FOR_KEY)
    if v == "" { return 0, nil }
    return parseI32(v, 0), nil
}

func SetVotedFor(value int32) error {
    // mirror AS: setVotedForInternal(value+1)
    return SetVotedForInternal(value + 1)
}

func SetVotedForInternal(value int32) error { sSet(VOTED_FOR_KEY, i32ToStr(value)); return nil }

// Match/Next indexes
func GetMatchIndexArray() ([]int64, error) {
    v := sGet(MATCH_INDEX_ARRAY)
    if v == "" { return []int64{}, nil }
    var arr []int64
    if err := json.Unmarshal([]byte(v), &arr); err != nil { return nil, err }
    return arr, nil
}

func SetMatchIndexArray(arr []int64) error {
    bz, err := json.Marshal(arr)
    if err != nil { return err }
    sSet(MATCH_INDEX_ARRAY, string(bz))
    return nil
}

func GetNextIndexArray() ([]int64, error) {
    v := sGet(NEXT_INDEX_ARRAY)
    if v == "" { return []int64{}, nil }
    var arr []int64
    if err := json.Unmarshal([]byte(v), &arr); err != nil { return nil, err }
    return arr, nil
}

func SetNextIndexArray(arr []int64) error {
    bz, err := json.Marshal(arr)
    if err != nil { return err }
    sSet(NEXT_INDEX_ARRAY, string(bz))
    return nil
}

// Log entries
func AppendLogEntry(entry LogEntryAggregate) error {
    last, err := GetLastLogIndex()
    if err != nil { return err }
    index := last + 1
    if index > entry.Index {
        LoggerDebug("already appended entry", []string{"height", i64ToStr(entry.Index)})
        return nil
    }
    if index != entry.Index {
        return fmt.Errorf("mismatched index while appending log entry: expected %d, found %d", index, entry.Index)
    }
    if err := SetLastLogIndex(index); err != nil { return err }
    return SetLogEntryAggregate(entry)
}

func SetLogEntryAggregate(entry LogEntryAggregate) error {
    data, err := json.Marshal(entry.Data)
    if err != nil { return err }
    temp := LogEntry{Index: entry.Index, TermID: entry.TermID, LeaderID: entry.LeaderID, Data: data}
    valuestr, err := json.Marshal(temp)
    if err != nil { return err }
    return SetLogEntry(entry.Index, string(valuestr))
}

func SetLogEntryObj(entry LogEntry) error {
    valuestr, err := json.Marshal(entry)
    if err != nil { return err }
    return SetLogEntry(entry.Index, string(valuestr))
}

func SetLogEntry(index int64, entry string) error {
    LoggerDebugExtended("setting entry", []string{"height", i64ToStr(index), "value", entry})
    key := GetLogEntryKey(index)
    if key == "" { return errors.New("empty key for log entry") }
    sSet(key, entry)
    return nil
}

func GetLogEntry(index int64) (string, error) {
    key := GetLogEntryKey(index)
    return sGet(key), nil
}

func GetLogEntryObj(index int64) (LogEntry, error) {
    v, _ := GetLogEntry(index)
    if v == "" { return LogEntry{Index: 0, TermID: 0, LeaderID: 0, Data: ""}, nil }
    var e LogEntry
    if err := json.Unmarshal([]byte(v), &e); err != nil { return LogEntry{}, err }
    return e, nil
}

func RemoveLogEntry(index int64) error {
    e, err := GetLogEntryObj(index)
    if err != nil { return err }
    e.Data = nil
    return SetLogEntryObj(e)
}

func GetLogEntryKey(index int64) string { return "logs_" + i64ToStr(index) }

// Utilities used by actions
func GetMajority(count int) int64 { return int64(count/2) + 1 }

// Small parsing helpers
func parseI64(s string, d int64) int64 {
    var v int64
    if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
        return d
    }
    return v
}

func parseI32(s string, d int32) int32 {
    var v int32
    if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
        return d
    }
    return v
}

func i64ToStr(v int64) string { return fmt.Sprintf("%d", v) }
func i32ToStr(v int32) string { return fmt.Sprintf("%d", v) }
