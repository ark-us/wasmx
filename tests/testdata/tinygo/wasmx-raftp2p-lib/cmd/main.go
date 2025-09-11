package main

import (
	"encoding/base64"
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	raft "github.com/loredanacirstea/wasmx-raft-lib/lib"
	lib "github.com/loredanacirstea/wasmx-raftp2p-lib/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module consensus
//export wasmx_consensus_json_i64_1
func Wasmx_consensus_json_i64_1() {}

//go:wasm-module wasmxcore
//export wasmx_env_core_i64_1
func Wasmx_env_core_i64_1() {}

//go:wasm-module crosschain
//export wasmx_crosschain_json_i64_1
func Wasmx_crosschain_json_i64_1() {}

//go:wasm-module p2p
//export wasmx_p2p_json_i64_1
func Wasmx_p2p_json_i64_1() {}

//go:wasm-module wasmx-raftp2p
//export instantiate
func Instantiate() {}

func main() {
	// Only internal
	wasmx.OnlyInternal(lib.MODULE_NAME, "")

	databz := wasmx.GetCallData()
	var calld fsm.ExternalActionCallData
	if err := json.Unmarshal(databz, &calld); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(databz))
		return
	}

	// Helper to read params from event
	get := func(key string) string {
		for _, p := range calld.Event.Params {
			if p.Key == key {
				return p.Value
			}
		}
		for _, p := range calld.Params {
			if p.Key == key {
				return p.Value
			}
		}
		return ""
	}

	// Route methods
	switch calld.Method {
	case "isVotedLeader":
		ok, err := raft.IsVotedLeader(calld.Params, calld.Event)
		if err != nil {
			lib.Revert(err.Error())
			return
		}
		res := raft.WrapGuard(ok)
		wasmx.Finish(res)
		return
	case "setupNode":
		if err := lib.SetupNode(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "connectPeers":
		st, err := raft.GetCurrentState()
		if err != nil {
			lib.Revert(err.Error())
			return
		}
		if err := lib.ConnectPeersInternal(lib.GetProtocolIdFromState(st)); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "forwardTxsToLeader":
		if err := lib.ForwardTxsToLeader(); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "requestNetworkSync":
		st, err := raft.GetCurrentState()
		if err != nil {
			lib.Revert(err.Error())
			return
		}
		if err := lib.RequestNetworkSync(lib.GetProtocolIdFromState(st)); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "receiveAppendEntryResponse":
		entry := get("entry")
		sig := get("signature")
		sender := get("sender")
		if err := lib.ReceiveAppendEntryResponse(entry, sig, sender); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "receiveStateSyncRequest":
		entry := get("entry")
		sig := get("signature")
		sender := get("sender")
		if err := lib.ReceiveStateSyncRequest(entry, sig, sender); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "receiveStateSyncResponse":
		entry := get("entry")
		sender := get("sender")
		if err := lib.ReceiveStateSyncResponse(entry, sender); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "vote":
		entry := get("entry")
		sig := get("signature")
		if err := lib.Vote(entry, sig); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "receiveVoteResponse":
		entry := get("entry")
		sig := get("signature")
		sender := get("sender")
		if err := lib.ReceiveVoteResponse(entry, sig, sender); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "receiveUpdateNodeResponse":
		entry := get("entry")
		sig := get("signature")
		sender := get("sender")
		if err := lib.ReceiveUpdateNodeResponse(entry, sig, sender); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "sendHeartbeatResponse":
		// Parse AppendEntry from event
		entryB64 := get("entry")
		if entryB64 == "" {
			lib.Revert("missing entry")
			return
		}
		bz, err := base64.StdEncoding.DecodeString(entryB64)
		if err != nil {
			lib.Revert(err.Error())
			return
		}
		var appendEntry raft.AppendEntry
		if err := json.Unmarshal(bz, &appendEntry); err != nil {
			lib.Revert(err.Error())
			return
		}
		if err := lib.SendHeartbeatResponse(appendEntry); err != nil {
			lib.Revert(err.Error())
			return
		}
	// Proxy to raft-lib for shared actions
	case "processAppendEntries":
		if err := raft.ProcessAppendEntries(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "sendAppendEntries":
		if err := lib.SendAppendEntries(); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "sendNewTransactionResponse":
		if err := raft.SendNewTransactionResponse(nil, fsm.EventObject{}); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "addToMempool":
		if err := raft.AddToMempool(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "commitBlocks":
		if err := lib.CommitBlocks(); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "setRandomElectionTimeout":
		if err := raft.SetRandomElectionTimeout(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "initializeNextIndex":
		if err := raft.InitializeNextIndex(nil, fsm.EventObject{}); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "initializeMatchIndex":
		if err := raft.InitializeMatchIndex(nil, fsm.EventObject{}); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "incrementCurrentTerm":
		if err := raft.IncrementCurrentTerm(nil, fsm.EventObject{}); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "selfVote":
		if err := raft.SelfVote(nil, fsm.EventObject{}); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "sendVoteRequests":
		if err := lib.SendVoteRequests(); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "updateNodeAndReturn":
		if err := lib.UpdateNodeAndReturn(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "proposeBlock":
		if err := raft.ProposeBlock(nil, fsm.EventObject{}); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "setup":
		if err := raft.Setup(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	default:
		wasmx.Revert(append([]byte("invalid function call data: "), databz...))
		return
	}
}
