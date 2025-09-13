package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	blocks "github.com/loredanacirstea/wasmx-blocks/lib"
	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
)

// NodeInfoResponse mirrors the AS parse result
type NodeInfoResponse struct {
	NodeInfo *p2p.NodeInfo `json:"node_info,omitempty"`
	Error    string        `json:"error"`
}

// updateNodeAndReturn: Leader receives a node update, sends updated node list to requester
func UpdateNodeAndReturn(params []fsm.ActionParam, event fsm.EventObject) error {
	// Extract entry and signature from event/params
	entryB64 := ""
	signature := ""
	if len(event.Params) > 0 {
		for _, p := range event.Params {
			if p.Key == "entry" {
				entryB64 = p.Value
			} else if p.Key == "signature" {
				signature = p.Value
			}
		}
	}
	for _, p := range params {
		if p.Key == "entry" && entryB64 == "" {
			entryB64 = p.Value
		}
		if p.Key == "signature" && signature == "" {
			signature = p.Value
		}
	}
	if entryB64 == "" {
		return fmt.Errorf("updateNodeAndReturn: empty entry")
	}
	if signature == "" {
		return fmt.Errorf("updateNodeAndReturn: empty signature")
	}
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	var entry raftlib.NodeUpdate
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}
	LoggerDebug("updateNodeAndReturn", []string{"entry", string(data), "signature", signature})
	LoggerInfo("update node", []string{"ip", entry.Node.Node.IP, "type", fmt.Sprint(entry.Type), "index", fmt.Sprint(entry.Index), "address", string(entry.Node.Address)})
	if entry.Node.Address == "" {
		return fmt.Errorf("node update failed, address missing")
	}
	ok, err := verifyMessageByAddr(entry.Node.Address, signature, data)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature verification failed: address %s", entry.Node.Address)
	}
	// Update node list
	ips, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	switch entry.Type {
	case raftlib.NODE_UPDATE_ADD:
		if entry.Node.Node.IP == "" && entry.Node.Node.Host == "" {
			return fmt.Errorf("validator info missing from node update")
		}
		ndx := -1
		for i := range ips {
			if ips[i].Address == entry.Node.Address {
				ndx = i
				break
			}
		}
		if ndx > -1 {
			ips[ndx].Node = entry.Node.Node
		} else {
			ips = append(ips, entry.Node)
		}
	case raftlib.NODE_UPDATE_REMOVE:
		if int(entry.Index) >= 0 && int(entry.Index) < len(ips) {
			ips[entry.Index].Node.IP = ""
			ips[entry.Index].Node.Host = ""
			ips[entry.Index].Node.Port = ""
		}
	default:
		// ignore
	}
	if err := raftlib.SetNodeIPs(ips); err != nil {
		return err
	}
	ourId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	lastIndex, err := raftlib.GetLastLogIndex()
	if err != nil {
		return err
	}
	nextArr, err := raftlib.GetNextIndexArray()
	if err != nil {
		return err
	}
	// choose a synced node different than us (AS parity)
	syncNodeID := GetRandomSynced(nextArr, lastIndex, ourId)
	resp := raftlib.UpdateNodeResponse{Nodes: ips, SyncNodeID: syncNodeID, LastEntryIndex: lastIndex}
	// if sync target is different than us, send an appendEntry to that node first
	if syncNodeID != ourId {
		if err := sendAppendEntry(syncNodeID, ips[syncNodeID], ips); err != nil {
			// best-effort; continue even if it fails
			LoggerError("failed to send append entry to sync provider", []string{"id", fmt.Sprint(syncNodeID), "error", err.Error()})
		}
	}
	// build receiveUpdateNodeResponse payload and send only to requester (entry.Node)
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
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveUpdateNodeResponse"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(datastr)},
		{Key: "signature", Value: sigResp},
		{Key: "sender", Value: string(ips[ourId].Address)},
	}
	msg, _ := json.Marshal(&payload)
	contract := wasmx.GetAddress()
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	pid := GetProtocolIdFromState(st)
	peer := GetP2PAddress(entry.Node)
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: pid, Peers: []string{peer}})
	return err
}

// SetupNode initializes local state from InitChainSetup and peers (AS parity)
func SetupNode(_ []fsm.ActionParam, event fsm.EventObject) error {
	// read base64-encoded InitChainSetup from event params (key: "data")
	dataB64 := ""
	for _, p := range event.Params {
		if p.Key == "data" {
			dataB64 = p.Value
			break
		}
	}
	if dataB64 == "" {
		return fmt.Errorf("no initChainSetup found")
	}
	raw, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return err
	}
	// TODO remove validator private key from logs in initChainSetup
	LoggerDebug("setupNode", []string{"initChainSetup", string(raw)})

	// Reset commit/applied
	if err := raftlib.SetCommitIndex(raftlib.LOG_START); err != nil {
		return err
	}
	if err := raftlib.SetLastApplied(raftlib.LOG_START); err != nil {
		return err
	}

	// Decode into raftlib.InitChainSetup
	var init raftlib.InitChainSetup
	if err := json.Unmarshal(raw, &init); err != nil {
		return err
	}
	if err := raftlib.SetCurrentNodeId(init.NodeIndex); err != nil {
		return err
	}

	// Parse peers
	peers := make([]p2p.NodeInfo, len(init.Peers))
	for i := range init.Peers {
		resp := ParseNodeAddress(init.Peers[i])
		if resp.Error != "" || resp.NodeInfo == nil {
			return fmt.Errorf("%s", resp.Error)
		}
		peers[i] = *resp.NodeInfo
	}
	if err := raftlib.SetNodeIPs(peers); err != nil {
		return err
	}
	// Initialize chain state and store consensus params for next height
	if err := raftlib.InitChain(init); err != nil {
		return err
	}
	return raftlib.InitializeIndexArrays(len(peers))
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
		// AS: revert on start failure
		LoggerError("start node with identity failed", []string{"err", err.Error()})
		return err
	}
	// connect peers
	for i := range nodes {
		if int32(i) == idx {
			// don't connect with ourselves
			continue
		}
		peer := GetP2PAddress(nodes[i])
		LoggerDebug("trying to connect to peer", []string{"p2paddress", peer, "address", string(nodes[i].Address)})
		if _, err := p2p.ConnectPeer(p2p.ConnectPeerRequest{ProtocolId: pid, Peer: peer}); err != nil {
			return err
		}
	}
	return nil
}

// ParseNodeAddress parses "<address>@/ip4/<host>/tcp/<port>/p2p/<id>" into NodeInfo
// <address>@/ip4/127.0.0.1/tcp/5001/p2p/12D3Koo... (AS format)
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
	info := p2p.NodeInfo{Address: wasmx.Bech32String(addr), Node: p2p.NetworkNode{ID: p2pid, Host: host, Port: port, IP: parts1[1]}, OutOfSync: false}
	resp.NodeInfo = &info
	return resp
}

// GetP2PAddress builds "/ip4/<host>/tcp/<port>/p2p/<id>"
func GetP2PAddress(node p2p.NodeInfo) string {
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
// forward transactions to leader (max 5 per tick)
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
	// compute batch count (max 5)
	countToSend := len(mp.Map)
	if countToSend > 5 {
		countToSend = 5
	}
	LoggerDebug("forwarding txs to leader", []string{"nodeId", fmt.Sprint(nodeId), "nodeIp", leader.Node.IP, "count", fmt.Sprint(countToSend)})
	sent := 0
	contract := wasmx.GetAddress()
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	protocolId := GetProtocolIdFromState(st)
	for _, txhash := range mp.Order {
		if sent >= countToSend {
			break
		}
		tx, ok := mp.Map[txhash]
		if !ok {
			continue
		}
		LoggerDebug("forwarding tx to leader", []string{"nodeId", fmt.Sprint(nodeId), "nodeIp", leader.Node.IP, "tx_batch_index", fmt.Sprint(sent)})
		msg := struct {
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
		msg.Run.Event.Type = "newTransaction"
		msg.Run.Event.Params = []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{{Key: "transaction", Value: base64.StdEncoding.EncodeToString(tx.Tx)}}
		bz, _ := json.Marshal(&msg)
		if _, err := p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: bz, ProtocolId: protocolId, Peers: []string{GetP2PAddress(leader)}}); err != nil {
			return err
		}
		// TODO we will remove the tx from mempool now
		// if it is an invalid transaction (it can happen), then the leader will
		// reject it, so this node will continue to send it to the Leader
		// we should remove it only if found invalid
		// remove; keep order consistent
		delete(mp.Map, txhash)
		for i, h := range mp.Order {
			if h == txhash {
				mp.Order = append(mp.Order[:i], mp.Order[i+1:]...)
				break
			}
		}
		sent++
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
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveAppendEntryResponse"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
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
	ok, err := verifyMessageByAddr(nodes[req.CandidateID].Address, signature, data)
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
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveVoteResponse"
	// Get our node address to include as sender
	ourId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	senderAddr := string(nodes[ourId].Address)

	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(datastr)},
		{Key: "signature", Value: sigResp},
		{Key: "sender", Value: senderAddr}, // AS: Missing sender parameter (line 710)
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

// SendVoteRequests iterates peers and sends VoteRequest over p2p (AS parity)
func SendVoteRequests() error {
	termId, err := raftlib.GetTermId()
	if err != nil {
		return err
	}
	candidateId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	lastLogIndex, err := raftlib.GetLastLogIndex()
	if err != nil {
		return err
	}
	lastEntry, err := raftlib.GetLogEntryObjIndexLast()
	if err != nil {
		return err
	}
	lastLogTerm := lastEntry.TermID
	req := raftlib.VoteRequest{TermID: termId, CandidateID: candidateId, LastLogIndex: lastLogIndex, LastLogTerm: lastLogTerm}
	ips, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	if len(ips) > 1 {
		LoggerInfo("sending vote requests...", []string{})
	}
	contract := wasmx.GetAddress()
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	pid := GetProtocolIdFromState(st)
	for i := range ips {
		if int32(i) == candidateId || !raftlib.IsNodeActive(ips[i]) {
			continue
		}
		// prepare request
		datastr, err := json.Marshal(&req)
		if err != nil {
			return err
		}
		sig, err := raftlib.SignMessage(string(datastr))
		if err != nil {
			return err
		}
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
		payload.Run.Event.Type = "receiveVoteRequest"
		payload.Run.Event.Params = []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{
			{Key: "entry", Value: base64.StdEncoding.EncodeToString(datastr)},
			{Key: "signature", Value: sig},
		}
		msg, _ := json.Marshal(&payload)
		peer := GetP2PAddress(ips[i])
		LoggerDebug("sending vote request", []string{"nodeId", fmt.Sprint(i), "nodeIp", ips[i].Node.IP, "termId", fmt.Sprint(termId), "data", string(datastr)})
		if _, err := p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: pid, Peers: []string{peer}}); err != nil {
			return err
		}
	}
	return nil
}

// ReceiveVoteResponse marks votes from sender after signature verification
func ReceiveVoteResponse(entryB64 string, signature string, sender wasmx.Bech32String) error {
	if entryB64 == "" {
		return fmt.Errorf("receiveVoteResponse: empty entry")
	}
	data, err := base64.StdEncoding.DecodeString(entryB64)
	if err != nil {
		return err
	}
	LoggerDebug("received vote response", []string{"sender", string(sender), "data", string(data)})
	ok, err := verifyMessageByAddr(sender, signature, data)
	if err != nil {
		return err
	}
	if !ok {
		LoggerError("signature verification failed for vote response", []string{"sender", string(sender)})
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
		if nodes[i].Address == sender {
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
func ReceiveAppendEntryResponse(entry string, signature string, sender wasmx.Bech32String) error {
	// parse
	var resp raftlib.AppendEntryResponse
	bz, err := base64.StdEncoding.DecodeString(entry)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bz, &resp); err != nil {
		return err
	}
	LoggerDebug("received append entry response", []string{"sender", string(sender), "data", string(bz)})
	ok, err := verifyMessageByAddr(sender, signature, bz)
	if err != nil {
		return err
	}
	if !ok {
		LoggerError("signature verification failed for receiveAppendEntryResponse", []string{"sender", string(sender)})
		return fmt.Errorf("signature verification failed for receiveAppendEntryResponse from %s", sender)
	}
	// find node id by sender address
	nodes, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	nodeId := -1
	for i := range nodes {
		if nodes[i].Address == sender {
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

// sendAppendEntry: we eliminate out of sync nodes until they get back online (AS parity)
func sendAppendEntry(nodeId int32, node p2p.NodeInfo, nodeIps []p2p.NodeInfo) error {
	nextIndexPerNode, err := raftlib.GetNextIndexArray()
	if err != nil {
		return err
	}
	var nextIndex int64 = raftlib.LOG_START
	if int(nodeId) < len(nextIndexPerNode) {
		nextIndex = nextIndexPerNode[nodeId]
	}
	lastIndex, err := raftlib.GetLastLogIndex()
	if err != nil {
		return err
	}
	if (lastIndex - nextIndex) > int64(stateSyncBatch) {
		// this node is out of sync
		nodeIps[nodeId].OutOfSync = true
		if err := raftlib.SetNodeIPs(nodeIps); err != nil {
			return err
		}
		return nil
	}
	// build AppendEntry and message via helpers
	data, err := raftlib.PrepareAppendEntry(nodeIps, nextIndex, lastIndex)
	if err != nil {
		return err
	}
	// AS also logs a short entry dissemination here
	LoggerDebug("diseminate entry", []string{"count", fmt.Sprint(len(data.Entries)), "address", string(node.Address)})
	msgStr, err := raftlib.PrepareAppendEntryMessage(nodeId, nextIndex, lastIndex, lastIndex, node, data)
	if err != nil {
		return err
	}
	msg := []byte(msgStr)
	// send to peer
	contract := wasmx.GetAddress()
	st, err := raftlib.GetCurrentState()
	if err != nil {
		return err
	}
	protocolId := GetProtocolIdFromState(st)
	peer := GetP2PAddress(node)
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: protocolId, Peers: []string{peer}})
	return err
}

// SendAppendEntries goes through each node and sends AppendEntry if needed (AS parity)
func SendAppendEntries() error {
	nodeId, err := raftlib.GetCurrentNodeId()
	if err != nil {
		return err
	}
	ips, err := raftlib.GetNodeIPs()
	if err != nil {
		return err
	}
	if bz, err := json.Marshal(&ips); err == nil {
		LoggerDebug("diseminate entries...", []string{"nodeId", fmt.Sprint(nodeId), "ips", string(bz)})
	} else {
		LoggerDebug("diseminate entries...", []string{"nodeId", fmt.Sprint(nodeId)})
	}
	for i := range ips {
		if int32(i) == nodeId || !raftlib.IsNodeActive(ips[i]) {
			continue
		}
		if err := sendAppendEntry(int32(i), ips[i], ips); err != nil {
			return err
		}
	}
	return nil
}

// CommitBlocks: run commit checks, and if consensus changed, disseminate via p2p
func CommitBlocks() error {
	// AS parity: call raft's checkCommits (no dissemination)
	changed, err := raftlib.CheckCommits(nil, fsm.EventObject{})
	if err != nil {
		return err
	}
	if changed {
		// propagate commit so peers can enter new consensus
		return SendAppendEntries()
	}
	return nil
}

// Helpers: build AppendEntry and message (AS parity)
// removed local prepareAppendEntry; use raftlib.PrepareAppendEntry

// removed local prepareAppendEntryMessage; use raftlib.PrepareAppendEntryMessage

// we just send a NodeUpdateRequest
// the node will receive a UpdateNodeResponse from the leader and then proceed to do state sync
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
	// when a node starts, it needs to add itself to the Leader's node list
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
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "updateNode"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(dataBz)},
		{Key: "signature", Value: sig},
	}
	contract := wasmx.GetAddress()
	msgBz, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	LoggerInfo("register request", []string{"req", string(msgBz)})
	// send to all peers (except self)
	peers := make([]string, 0, len(ips))
	for i := range ips {
		if int32(i) != nodeId {
			peers = append(peers, GetP2PAddress(ips[i]))
		}
	}
	LoggerDebug("sending node registration", []string{"peers", strings.Join(peers, ","), "data", string(msgBz)})
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msgBz, ProtocolId: protocolId, Peers: peers})
	return err
}

// this is executed each time the node is started in Follower or Candidate state if needed
// and the first time the node is started
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
	LoggerInfo("received statesync response", []string{"count", fmt.Sprint(len(resp.Entries)), "from", fmt.Sprint(resp.StartBatchIndex), "to", fmt.Sprint(resp.LastBatchIndex), "last_log_index", fmt.Sprint(resp.LastLogIndex)})
	if err := raftlib.SetTermId(resp.TermID); err != nil {
		return err
	}
	for i := range resp.Entries {
		if err := raftlib.ProcessAppendEntry(resp.Entries[i]); err != nil {
			return err
		}
	}
	if resp.LastBatchIndex >= resp.LastLogIndex {
		last, err := raftlib.GetLogEntryObjIndexLast()
		if err != nil {
			return err
		}
		msg := raftlib.AppendEntryResponse{TermID: resp.TermID, Success: true, LastIndex: resp.LastBatchIndex}
		LoggerInfo("send heartbeat response", []string{"termId", fmt.Sprint(resp.TermID), "success", "true", "lastLogIndex", fmt.Sprint(resp.LastBatchIndex), "leaderId", fmt.Sprint(last.LeaderID)})
		if err := SendHeartbeatResponseMessage(msg, last.LeaderID); err != nil {
			return err
		}
	}
	return nil
}

const stateSyncBatch = 200

// received statesync request
// ReceiveStateSyncRequest verifies and replies with batches
func ReceiveStateSyncRequest(entryB64 string, signature string, sender wasmx.Bech32String) error {
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
	LoggerInfo("statesync request", []string{"data", string(data), "sender", string(sender)})
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
	batches := int(math.Ceil(float64(count) / float64(stateSyncBatch)))
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
	LoggerInfo("state sync messages finished", []string{"lastIndex", fmt.Sprint(lastIndex)})
	return nil
}

func sendStateSyncBatch(startIndex, lastIndexToSend, lastIndex int64, termId int32, receiver wasmx.Bech32String) error {
	entries := make([]raftlib.LogEntryAggregate, 0, lastIndexToSend-startIndex+1)
	for i := startIndex; i <= lastIndexToSend; i++ {
		if agg, err := raftlib.GetLogEntryAggregate(i); err == nil {
			entries = append(entries, *agg)
		} else {
			return err
		}
	}
	// we do not sign this message, because the receiver does not have our publicKey
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
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveStateSyncResponse"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
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
	LoggerDebug("sending state sync chunk", []string{"to", string(receiver), "count", fmt.Sprint(len(entries)), "from", fmt.Sprint(startIndex), "to", fmt.Sprint(lastIndexToSend), "last_index", fmt.Sprint(lastIndex)})
	if node == nil {
		LoggerError("node not found for address", []string{"receiver", string(receiver)})
		return fmt.Errorf("cannot find node by address: %s", receiver)
	}
	peer := GetP2PAddress(*node)
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: pid, Peers: []string{peer}})
	return err
}

func getNodeByAddress(addr wasmx.Bech32String, nodes []p2p.NodeInfo) *p2p.NodeInfo {
	for i := range nodes {
		if nodes[i].Address == addr {
			return &nodes[i]
		}
	}
	return nil
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

// Signature verification helpers (by validator address)
func verifyMessageByAddr(addr wasmx.Bech32String, signatureB64 string, msg []byte) (bool, error) {
	vals, err := raftlib.GetAllValidators()
	if err != nil {
		return false, err
	}
	var pub *wasmx.PublicKey
	for i := range vals {
		if vals[i].OperatorAddress == addr {
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

// received an updated node list from Leader
// ReceiveUpdateNodeResponse verifies and applies node list, then triggers state sync request
func ReceiveUpdateNodeResponse(entryB64 string, signature string, sender wasmx.Bech32String) error {
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
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"event"`
		} `json:"run"`
	}{}
	payload.Run.Event.Type = "receiveStateSyncRequest"
	payload.Run.Event.Params = []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		{Key: "entry", Value: base64.StdEncoding.EncodeToString(bz)},
		{Key: "signature", Value: sig},
		{Key: "sender", Value: string(sender)},
	}
	msg, _ := json.Marshal(&payload)
	contract := wasmx.GetAddress()
	peer := GetP2PAddress(receiver)
	LoggerDebug("sending statesync request", []string{"nodeId", fmt.Sprint(nodeId), "address", string(receiver.Address), "data", string(bz)})
	_, err = p2p.SendMessageToPeers(p2p.SendMessageToPeersRequest{Contract: contract, Sender: contract, Msg: msg, ProtocolId: protocolId, Peers: []string{peer}})
	return err
}
