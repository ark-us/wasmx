package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	lib "github.com/loredanacirstea/wasmx-raft-lib/lib"
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

//go:wasm-module wasmx-raft
//export instantiate
func Instantiate() {}

func main() {
	// Only allow internal calls like the AS contract
	wasmx.OnlyInternal(lib.MODULE_NAME, "")
	databz := wasmx.GetCallData()

	var calld fsm.ExternalActionCallData
	if err := json.Unmarshal(databz, &calld); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(databz))
		return
	}

	// Dispatch public operations compatible with AS index.ts
	switch calld.Method {
	case "isVotedLeader":
		ok, err := lib.IsVotedLeader(calld.Params, calld.Event)
		if err != nil {
			lib.Revert("isVotedLeader failed: " + err.Error())
			return
		}
		res := lib.WrapGuard(ok)
		wasmx.Finish(res)
		return
	case "setupNode":
		if err := lib.SetupNode(calld.Params, calld.Event); err != nil {
			lib.Revert("setupNode failed: " + err.Error())
			return
		}
	case "processAppendEntries":
		if err := lib.ProcessAppendEntries(calld.Params, calld.Event); err != nil {
			lib.Revert("processAppendEntries failed: " + err.Error())
			return
		}
	case "sendHeartbeatResponse":
		if err := lib.SendHeartbeatResponse(calld.Params, calld.Event); err != nil {
			lib.Revert("sendHeartbeatResponse failed: " + err.Error())
			return
		}
	case "sendAppendEntries":
		if err := lib.SendAppendEntries(calld.Params, calld.Event); err != nil {
			lib.Revert("sendAppendEntries failed: " + err.Error())
			return
		}
	case "sendNewTransactionResponse":
		if err := lib.SendNewTransactionResponse(calld.Params, calld.Event); err != nil {
			lib.Revert("sendNewTransactionResponse failed: " + err.Error())
			return
		}
	case "addToMempool":
		if err := lib.AddToMempool(calld.Params, calld.Event); err != nil {
			lib.Revert("addToMempool failed: " + err.Error())
			return
		}
	case "commitBlocks":
		if err := lib.CommitBlocks(calld.Params, calld.Event); err != nil {
			lib.Revert("commitBlocks failed: " + err.Error())
			return
		}
	case "setRandomElectionTimeout":
		if err := lib.SetRandomElectionTimeout(calld.Params, calld.Event); err != nil {
			lib.Revert("setRandomElectionTimeout failed: " + err.Error())
			return
		}
	case "initializeNextIndex":
		if err := lib.InitializeNextIndex(calld.Params, calld.Event); err != nil {
			lib.Revert("initializeNextIndex failed: " + err.Error())
			return
		}
	case "initializeMatchIndex":
		if err := lib.InitializeMatchIndex(calld.Params, calld.Event); err != nil {
			lib.Revert("initializeMatchIndex failed: " + err.Error())
			return
		}
	case "incrementCurrentTerm":
		if err := lib.IncrementCurrentTerm(calld.Params, calld.Event); err != nil {
			lib.Revert("incrementCurrentTerm failed: " + err.Error())
			return
		}
	case "vote":
		if err := lib.Vote(calld.Params, calld.Event); err != nil {
			lib.Revert("vote failed: " + err.Error())
			return
		}
	case "selfVote":
		if err := lib.SelfVote(calld.Params, calld.Event); err != nil {
			lib.Revert("selfVote failed: " + err.Error())
			return
		}
	case "forwardTxsToLeader":
		if err := lib.ForwardTxsToLeader(calld.Params, calld.Event); err != nil {
			lib.Revert("forwardTxsToLeader failed: " + err.Error())
			return
		}
	case "updateNodeAndReturn":
		if err := lib.UpdateNodeAndReturn(calld.Params, calld.Event); err != nil {
			lib.Revert("updateNodeAndReturn failed: " + err.Error())
			return
		}
	case "registeredCheck":
		if err := lib.RegisteredCheck(calld.Params, calld.Event); err != nil {
			lib.Revert("registeredCheck failed: " + err.Error())
			return
		}
	case "sendVoteRequests":
		if err := lib.SendVoteRequests(calld.Params, calld.Event); err != nil {
			lib.Revert("sendVoteRequests failed: " + err.Error())
			return
		}
	case "proposeBlock":
		if err := lib.ProposeBlock(calld.Params, calld.Event); err != nil {
			lib.Revert("proposeBlock failed: " + err.Error())
			return
		}
	case "setup":
		if err := lib.Setup(calld.Params, calld.Event); err != nil {
			lib.Revert("setup failed: " + err.Error())
			return
		}
	default:
		wasmx.Revert(append([]byte("invalid function call data: "), databz...))
		return
	}

	wasmx.Finish(wasmx.GetFinishData())
}
