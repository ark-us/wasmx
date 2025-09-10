package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	consensuswrap "github.com/loredanacirstea/wasmx-env-consensus/lib"
	typestnd "github.com/loredanacirstea/wasmx-env-consensus/lib"
	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
	stakinglib "github.com/loredanacirstea/wasmx-staking/lib"
)

// NodeInfoResponse mirrors the AS parse result
type NodeInfoResponse struct {
	NodeInfo *raftlib.NodeInfo `json:"node_info,omitempty"`
	Error    string            `json:"error"`
}

// ConnectPeers starts local p2p node and connects to peers using protocolId
func ConnectPeersInternal(protocolId string) error {
	state, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	idx, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	if int(idx) >= len(nodes) {
		return nil
	}
	// derive protocol id like AS
	pid := GetProtocolIdFromState(state)
	// start node with identity
	req := p2p.StartNodeWithIdentityRequest{Port: nodes[idx].Node.Port, ProtocolId: pid, PK: state.ValidatorPrivkey}
	if _, err := p2p.StartNodeWithIdentity(req); err != nil {
		return err
	}
	// connect peers
	for i := range nodes {
		if int32(i) == idx {
			continue
		}
		peer := GetP2PAddress(nodes[i])
		if _, err := p2p.ConnectPeer(p2p.ConnectPeerRequest{ProtocolId: pid, Peer: peer}); err != nil {
			return err
		}
	}
	return nil
}

// ParseNodeAddress parses "<address>@/ip4/<host>/tcp/<port>/p2p/<id>" into NodeInfo
func ParseNodeAddress(peeraddr string) NodeInfoResponse {
	resp := NodeInfoResponse{}
	parts1 := strings.Split(peeraddr, "@")
	if len(parts1) != 2 {
		resp.Error = "invalid node format; found: " + peeraddr
		return resp
	}
	addr := parts1[0]
	parts2 := strings.Split(parts1[1], "/")
	if len(parts2) != 7 {
		resp.Error = "invalid node format; found: " + peeraddr
		return resp
	}
	host := parts2[2]
	port := parts2[4]
	p2pid := parts2[6]
	info := raftlib.NodeInfo{Address: addr, Node: raftlib.NetworkNode{ID: p2pid, Host: host, Port: port, IP: parts1[1]}, OutOfSync: false}
	resp.NodeInfo = &info
	return resp
}

// GetP2PAddress builds "/ip4/<host>/tcp/<port>/p2p/<id>"
func GetP2PAddress(node raftlib.NodeInfo) string {
	return "/ip4/" + node.Node.Host + "/tcp/" + node.Node.Port + "/p2p/" + node.Node.ID
}

// GetRandomSynced returns a random synced node id (or leader if <3 nodes)
func GetRandomSynced(arr []int64, lastIndex int64, leaderId int32) int32 {
	if len(arr) < 3 {
		return leaderId
	}
	synced := make([]int32, 0, len(arr))
	for i := range arr {
		if arr[i] >= lastIndex && int32(arr[i]) != leaderId {
			synced = append(synced, int32(i))
		}
	}
	if len(synced) == 0 {
		return leaderId
	}
	// choose random synced id like AS
	// use raftlib.GetRandomInRangeI32(0, len(synced)-1)
	idx, err := raftlib.GetRandomInRangeI32(0, int32(len(synced)-1))
	if err != nil {
		return synced[0]
	}
	if int(idx) < 0 || int(idx) >= len(synced) {
		return synced[0]
	}
	return synced[idx]
}

// ForwardTxsToLeader sends up to 5 mempool txs to the leader over p2p
func ForwardTxsToLeader() error {
	mp, err := raftlib.GetMempool()
	if err != nil {
		return err
	}
	if len(mp.Map) == 0 {
		return nil
	}
	last, err := raftlib.GetLogEntryObjIndexLast()
	if err != nil {
		return err
	}
	if last.Index == 0 {
		return nil
	}
	nodeId := last.LeaderID
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	if int(nodeId) >= len(nodes) {
		return nil
	}
	leader := nodes[nodeId]
	limit := 0
	contract := wasmx.GetAddress()
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	protocolId := GetProtocolIdFromState(st)
	for _, txhash := range mp.Order {
		if limit >= 5 {
			break
		}
		tx, ok := mp.Map[txhash]
		if !ok {
			continue
		}
		msg := struct {
			Run struct {
				Event struct {
					Type   string `json:"type"`
					Params []struct {
						Key, Value string `json:"key","value"`
					} `json:"params"`
				} `json:"event"`
			} `json:"run"`
		}{}
		msg.Run.Event.Type = "newTransaction"
		msg.Run.Event.Params = []struct {
			Key, Value string `json:"key","value"`
		}{{Key: "transaction", Value: base64.StdEncoding.EncodeToString(tx.Tx)}}
		bz, _ := json.Marshal(&msg)
		if _, err := p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: bz, ProtocolId: protocolId, Peers: []string{GetP2PAddress(leader)}}); err != nil {
			return err
		}
		// remove; keep order consistent
		delete(mp.Map, txhash)
		for i, h := range mp.Order {
			if h == txhash {
				mp.Order = append(mp.Order[:i], mp.Order[i+1:]...)
				break
			}
		}
		limit++
	}
	return raftlib.SetMempool(mp)
}

// SendHeartbeatResponseMessage sends AppendEntryResponse to leader
func SendHeartbeatResponseMessage(response raftlib.AppendEntryResponse, leaderId int32) error {
	datastr, err := json.Marshal(&response)
	if err != nil {
		return err
	}
	sig, err := raftlib.SignMessage(string(datastr))
	if err != nil {
		return err
	}
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	idx, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	sender := nodes[idx].Address
	payload := struct {
		Run struct {
			Event struct {
				Type   string `json:"type"`
				Params []struct {
					Key, Value string `json:"key","value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveAppendEntryResponse"
	payload.Run.Event.Params = []struct {
		Key, Value string `json:"key","value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(datastr)},
		{Key: "signature", Value: sig},
		{Key: "sender", Value: string(sender)},
	}
	peers := []string{GetP2PAddress(nodes[leaderId])}
	contract := wasmx.GetAddress()
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	pid := GetProtocolIdFromState(st)
	bz, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: bz, ProtocolId: pid, Peers: peers})
	return err
}

// SendHeartbeatResponse builds AppendEntryResponse from entry and sends it to leader
func SendHeartbeatResponse(entry raftlib.AppendEntry) error {
	resp, err := raftlib.PrepareHeartbeatResponse(entry)
	if err != nil {
		return err
	}
	return SendHeartbeatResponseMessage(resp, entry.LeaderID)
}

// Vote handles a vote request: verify signature, compute response, send back to candidate
func Vote(entryB64 string, signature string) error {
	if entryB64 == "" {
		return fmt.Errorf("vote: empty entry")
	}
	if signature == "" {
		return fmt.Errorf("vote: empty signature")
	}
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	var req raftlib.VoteRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	if int(req.CandidateID) >= len(nodes) {
		return fmt.Errorf("candidate out of range")
	}
	ok, err := verifyMessageByAddr(string(nodes[req.CandidateID].Address), signature, data)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature verification failed for VoteRequest")
	}
	resp, err := voteInternalLocal(req.TermID, req.CandidateID, req.LastLogIndex, req.LastLogTerm)
	if err != nil {
		return err
	}
	// send response to candidate
	contract := wasmx.GetAddress()
	datastr, _ := json.Marshal(&resp)
	sigResp, err := raftlib.SignMessage(string(datastr))
	if err != nil {
		return err
	}
	payload := struct {
		Run struct {
			Event struct {
				Type   string `json:"type"`
				Params []struct {
					Key, Value string `json:"key","value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveVoteResponse"
	payload.Run.Event.Params = []struct {
		Key, Value string `json:"key","value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(datastr)},
		{Key: "signature", Value: sigResp},
	}
	msg, _ := json.Marshal(&payload)
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	pid := GetProtocolIdFromState(st)
	peer := GetP2PAddress(nodes[req.CandidateID])
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: pid, Peers: []string{peer}})
	return err
}

// ReceiveVoteResponse marks votes from sender after signature verification
func ReceiveVoteResponse(entryB64 string, signature string, sender string) error {
	if entryB64 == "" {
		return fmt.Errorf("receiveVoteResponse: empty entry")
	}
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	ok, err := verifyMessageByAddr(sender, signature, data)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature verification failed for vote response")
	}
	var resp raftlib.VoteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	arr, err := raftlib.GetVoteIndexArray()
	if err != nil {
		return err
	}
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	id := -1
	for i := range nodes {
		if string(nodes[i].Address) == sender {
			id = i
			break
		}
	}
	if id >= 0 {
		if id >= len(arr) {
			tmp := make([]int32, id+1)
			copy(tmp, arr)
			arr = tmp
		}
		if resp.VoteGranted == true {
			arr[id] = 1
		}
		if err := raftlib.SetVoteIndexArray(arr); err != nil {
			return err
		}
	}
	return nil
}

// Local vote logic (AS parity) without exporting raft voteInternal
func voteInternalLocal(termId int32, candidateId int32, lastLogIndex int64, lastLogTerm int32) (raftlib.VoteResponse, error) {
	myTerm, err := raftlib.GetTermId()
	if err != nil {
		return raftlib.VoteResponse{}, err
	}
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return raftlib.VoteResponse{}, err
	}
	if int(candidateId) >= len(nodes) || !raftlib.IsNodeActive(nodes[candidateId]) {
		return raftlib.VoteResponse{TermID: myTerm, VoteGranted: false}, nil
	}
	if termId < myTerm {
		return raftlib.VoteResponse{TermID: myTerm, VoteGranted: false}, nil
	}
	if termId > myTerm {
		if err := raftlib.SetTermId(termId); err != nil {
			return raftlib.VoteResponse{}, err
		}
		if err := raftlib.SetVotedFor(candidateId); err != nil {
			return raftlib.VoteResponse{}, err
		}
		return raftlib.VoteResponse{TermID: termId, VoteGranted: true}, nil
	}
	hv, err := raftlib.HasVotedFor()
	if err != nil {
		return raftlib.VoteResponse{}, err
	}
	if hv {
		return raftlib.VoteResponse{TermID: termId, VoteGranted: false}, nil
	}
	myLast, err := raftlib.GetLastLogIndex()
	if err != nil {
		return raftlib.VoteResponse{}, err
	}
	if lastLogIndex < myLast {
		return raftlib.VoteResponse{TermID: termId, VoteGranted: false}, nil
	}
	if err := raftlib.SetVotedFor(candidateId); err != nil {
		return raftlib.VoteResponse{}, err
	}
	return raftlib.VoteResponse{TermID: termId, VoteGranted: true}, nil
}

// GetProtocolIdFromState returns PROTOCOL_ID + "_" + chainID (AS parity)
func GetProtocolIdFromState(state raftlib.CurrentState) string {
	return PROTOCOL_ID + "_" + state.ChainID
}

// ReceiveAppendEntryResponse processes peer response and updates nextIndex for sender
func ReceiveAppendEntryResponse(entry string, signature string, sender string) error {
	// parse
	var resp raftlib.AppendEntryResponse
	bz, err := base64.StdEncoding.DecodeString(entry)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bz, &resp); err != nil {
		return err
	}
	ok, err := verifyMessageByAddr(sender, signature, bz)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature verification failed for receiveAppendEntryResponse from %s", sender)
	}
	// find node id by sender address
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	nodeId := -1
	for i := range nodes {
		if string(nodes[i].Address) == sender {
			nodeId = i
			break
		}
	}
	if nodeId < 0 {
		return nil
	}
	if resp.Success {
		arr, err := raftlib.GetNextIndexArray()
		if err != nil {
			return err
		}
		if nodeId >= 0 && nodeId < len(arr) {
			arr[nodeId] = resp.LastIndex + 1
		}
		if err := raftlib.SetNextIndexArray(arr); err != nil {
			return err
		}
		nodes[nodeId].OutOfSync = false
		if err := raftlib.SetNodeIPs(nodes); err != nil {
			return err
		}
	}
	return nil
}

// RegisteredCheck sends a NodeUpdate request to the leader(s)
func RegisteredCheck(protocolId string) error {
	ips, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	if len(ips) == 1 {
		return nil
	}
	nodeId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	// Build NodeUpdate message (add self)
	update := raftlib.NodeUpdate{Node: ips[nodeId], Index: nodeId, Type: raftlib.NODE_UPDATE_ADD}
	dataBz, err := json.Marshal(&update)
	if err != nil {
		return err
	}
	sig, err := raftlib.SignMessage(string(dataBz))
	if err != nil {
		return err
	}
	payload := struct {
		Run struct {
			Event struct {
				Type   string `json:"type"`
				Params []struct {
					Key, Value string `json:"key","value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "updateNode"
	payload.Run.Event.Params = []struct {
		Key, Value string `json:"key","value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(dataBz)},
		{Key: "signature", Value: sig},
	}
	contract := wasmx.GetAddress()
	msgBz, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	// send to all peers (except self)
	peers := make([]string, 0, len(ips))
	for i := range ips {
		if int32(i) != nodeId {
			peers = append(peers, GetP2PAddress(ips[i]))
		}
	}
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msgBz, ProtocolId: protocolId, Peers: peers})
	return err
}

// RequestNetworkSync invokes RegisteredCheck
func RequestNetworkSync(protocolId string) error { return RegisteredCheck(protocolId) }

// ReceiveStateSyncResponse processes incoming state sync batch
func ReceiveStateSyncResponse(entryB64 string, sender string) error {
	bz, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	var resp StateSyncResponse
	if err := json.Unmarshal(bz, &resp); err != nil {
		return err
	}
	if err := raftlib.SetTermId(resp.TermID); err != nil {
		return err
	}
	for i := range resp.Entries {
		if err := processAppendEntryAggregate(resp.Entries[i]); err != nil {
			return err
		}
	}
	if resp.LastBatchIndex >= resp.LastLogIndex {
		last, err := raftlib.GetLogEntryObjIndexLast()
		if err != nil {
			return err
		}
		msg := raftlib.AppendEntryResponse{TermID: resp.TermID, Success: true, LastIndex: resp.LastBatchIndex}
		if err := SendHeartbeatResponseMessage(msg, last.LeaderID); err != nil {
			return err
		}
	}
	return nil
}

const stateSyncBatch = 200

// ReceiveStateSyncRequest verifies and replies with batches
func ReceiveStateSyncRequest(entryB64 string, signature string, sender string) error {
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	ok, err := verifyMessageByAddr(sender, signature, data)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature verification failed for statesync request from %s", sender)
	}
	var req StateSyncRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}
	term, err := raftlib.GetTermId()
	if err != nil {
		return err
	}
	lastIndex, err := raftlib.GetLastLogIndex()
	if err != nil {
		return err
	}
	count := lastIndex - req.StartIndex
	if count == 0 {
		return nil
	}
	batches := int((count + stateSyncBatch - 1) / stateSyncBatch)
	startIdx := req.StartIndex
	lastTo := startIdx
	for i := 0; i < batches-1; i++ {
		lastTo = startIdx + stateSyncBatch
		if err := sendStateSyncBatch(startIdx, lastTo, lastIndex, term, sender); err != nil {
			return err
		}
		startIdx += stateSyncBatch
	}
	if lastTo < lastIndex {
		if err := sendStateSyncBatch(startIdx, lastIndex, lastIndex, term, sender); err != nil {
			return err
		}
	}
	return nil
}

func sendStateSyncBatch(startIndex, lastIndexToSend, lastIndex int64, termId int32, receiver string) error {
	entries := make([]raftlib.LogEntryAggregate, 0, lastIndexToSend-startIndex+1)
	for i := startIndex; i <= lastIndexToSend; i++ {
		if agg, ok, err := getLogEntryAggregateLocal(i); err == nil && ok {
			entries = append(entries, agg)
		} else if err != nil {
			return err
		}
	}
	resp := StateSyncResponse{StartBatchIndex: startIndex, LastBatchIndex: lastIndexToSend, LastLogIndex: lastIndex, TermID: termId, Entries: entries}
	bz, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	idx, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	sender := nodes[idx].Address
	payload := struct {
		Run struct {
			Event struct {
				Type   string `json:"type"`
				Params []struct {
					Key, Value string `json:"key","value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveStateSyncResponse"
	payload.Run.Event.Params = []struct {
		Key, Value string `json:"key","value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(bz)},
		{Key: "sender", Value: string(sender)},
	}
	msg, _ := json.Marshal(&payload)
	contract := wasmx.GetAddress()
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	pid := GetProtocolIdFromState(st)
	node := getNodeByAddress(receiver, nodes)
	if node == nil {
		return fmt.Errorf("cannot find node by address: %s", receiver)
	}
	peer := GetP2PAddress(*node)
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: pid, Peers: []string{peer}})
	return err
}

func getNodeByAddress(addr string, nodes []raftlib.NodeInfo) *raftlib.NodeInfo {
	for i := range nodes {
		if string(nodes[i].Address) == addr {
			return &nodes[i]
		}
	}
	return nil
}

func getLogEntryAggregateLocal(index int64) (raftlib.LogEntryAggregate, bool, error) {
	e, err := raftlib.GetLogEntryObj(index)
	if err != nil {
		return raftlib.LogEntryAggregate{}, false, err
	}
	if e.Index == 0 {
		return raftlib.LogEntryAggregate{}, false, nil
	}
	var data []byte
	if len(e.Data) > 0 {
		data = e.Data
	} else {
		s, err := getFinalBlock(index)
		if err != nil {
			return raftlib.LogEntryAggregate{}, false, err
		}
		data = s
	}
	agg := raftlib.LogEntryAggregate{Index: e.Index, TermID: e.TermID, LeaderID: e.LeaderID, Data: json.RawMessage(data)}
	return agg, true, nil
}

func callStorage(calldata string, isQuery bool) (wasmx.CallResponse, error) {
	addr := wasmx.GetAddressByRole("storage")
	ok, data := wasmx.CallSimple(addr, []byte(calldata), isQuery, MODULE_NAME)
	resp := wasmx.CallResponse{Success: 0, Data: string(data)}
	if !ok {
		resp.Success = 1
	}
	return resp, nil
}

func getFinalBlock(index int64) ([]byte, error) {
	calld := blocks.CallData{GetBlockByIndex: &blocks.CallDataGetBlockByIndex{Index: index}}
	bz, err := json.Marshal(&calld)
	if err != nil {
		return nil, err
	}
	resp, err := callStorage(string(bz), true)
	if err != nil {
		return nil, err
	}
	if resp.Success > 0 {
		return nil, fmt.Errorf("could not get finalized block: %d", index)
	}
	return []byte(resp.Data), nil
}

func processAppendEntryAggregate(entry raftlib.LogEntryAggregate) error {
	var wrap typestnd.RequestProcessProposalWithMetaInfo
	if err := json.Unmarshal(entry.Data, &wrap); err != nil {
		return err
	}
	resp, err := consensuswrap.ProcessProposal(wrap.Request)
	if err != nil {
		return err
	}
	if resp.Status != typestnd.ProposalStatus_ACCEPT {
		return nil
	}
	return raftlib.AppendLogEntry(entry)
}

// Signature verification helpers (by validator address)
func verifyMessageByAddr(addr string, signatureB64 string, msg []byte) (bool, error) {
	vals, err := getAllValidators()
	if err != nil {
		return false, err
	}
	var pub *wasmx.PublicKey
	for i := range vals {
		if string(vals[i].OperatorAddress) == addr {
			pub = vals[i].ConsensusPubkey
			break
		}
	}
	if pub == nil {
		return false, nil
	}
	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, err
	}
	return wasmx.Ed25519Verify(pub.GetKey().Key, sig, msg), nil
}

func getAllValidators() ([]stakinglib.Validator, error) {
	payload := map[string]any{"QueryValidators": map[string]any{}}
	bz, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	addr := wasmx.GetAddressByRole(wasmx.ROLE_STAKING)
	ok, data := wasmx.CallSimple(addr, bz, true, MODULE_NAME)
	resp := wasmx.CallResponse{Success: 0, Data: string(data)}
	if !ok {
		resp.Success = 1
	}
	if resp.Success > 0 || resp.Data == "" {
		return []stakinglib.Validator{}, nil
	}
	var out struct {
		Validators []stakinglib.Validator `json:"validators"`
	}
	if err := json.Unmarshal([]byte(resp.Data), &out); err != nil {
		return []stakinglib.Validator{}, nil
	}
	return out.Validators, nil
}

// ReceiveUpdateNodeResponse verifies and applies node list, then triggers state sync request
func ReceiveUpdateNodeResponse(entryB64 string, signature string, sender string) error {
	if entryB64 == "" {
		return fmt.Errorf("receiveUpdateNodeResponse: empty entry")
	}
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	// verify signature by sender
	ok, err := verifyMessageByAddr(sender, signature, data)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature verification failed for UpdateNodeResponse from %s", sender)
	}
	var resp raftlib.UpdateNodeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	// find our index in provided nodes by address
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	ourId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	if int(ourId) >= len(nodes) {
		return fmt.Errorf("current node id out of range")
	}
	// locate ourId in resp.Nodes
	idInResp := -1
	for i := range resp.Nodes {
		if resp.Nodes[i].Address == nodes[ourId].Address {
			idInResp = i
			break
		}
	}
	if idInResp == -1 {
		return fmt.Errorf("node list does not contain our node")
	}
	node := resp.Nodes[idInResp]
	// consistency checks
	cur := nodes[ourId]
	if node.Node.Host != cur.Node.Host {
		return fmt.Errorf("node list wrong host data for %s", cur.Address)
	}
	if node.Node.IP != cur.Node.IP {
		return fmt.Errorf("node list wrong ip data for %s", cur.Address)
	}
	if node.Node.Port != cur.Node.Port {
		return fmt.Errorf("node list wrong port data for %s", cur.Address)
	}
	if node.Node.ID != cur.Node.ID {
		return fmt.Errorf("node list wrong id data for %s", cur.Address)
	}
	// apply node list and current id
	if err := raftlib.SetNodeIPs(resp.Nodes); err != nil {
		return err
	}
	if err := raftlib.SetCurrentNodeId(int32(idInResp)); err != nil {
		return err
	}
	// trigger state sync to sync_node_id
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	return SendStateSyncRequest(GetProtocolIdFromState(st), resp.SyncNodeID)
}

// SendStateSyncRequest sends a StateSyncRequest to a specified node
func SendStateSyncRequest(protocolId string, nodeId int32) error {
	lastIndex, err := raftlib.GetLastLogIndex()
	if err != nil {
		return err
	}
	ourId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	if int(nodeId) >= len(nodes) {
		return fmt.Errorf("state sync node out of range")
	}
	receiver := nodes[nodeId]
	req := StateSyncRequest{StartIndex: lastIndex + 1}
	bz, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	sig, err := raftlib.SignMessage(string(bz))
	if err != nil {
		return err
	}
	sender := nodes[ourId].Address
	payload := struct {
		Run struct {
			Event struct {
				Type   string `json:"type"`
				Params []struct {
					Key, Value string `json:"key","value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveStateSyncRequest"
	payload.Run.Event.Params = []struct {
		Key, Value string `json:"key","value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(bz)},
		{Key: "signature", Value: sig},
		{Key: "sender", Value: string(sender)},
	}
	msg, _ := json.Marshal(&payload)
	contract := wasmx.GetAddress()
	peer := GetP2PAddress(receiver)
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: protocolId, Peers: []string{peer}})
	return err
}
