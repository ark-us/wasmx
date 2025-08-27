package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	gov "github.com/loredanacirstea/wasmx-gov/gov"
)

//go:wasm-module wasmx-gov
//export instantiate
func Instantiate() {
	databz := wasmx.GetCallData()
	calld := &gov.MsgInitialize{}
	if err := json.Unmarshal(databz, calld); err != nil {
		gov.Revert("invalid call data: " + err.Error() + ": " + string(databz))
	}
	gov.SetVotingDenom(calld.BondBaseDenom)
}

func main() {
	databz := wasmx.GetCallData()
	calld := &gov.Calldata{}
	if err := json.Unmarshal(databz, calld); err != nil {
		gov.Revert("invalid call data: " + err.Error() + ": " + string(databz))
	}

	// public operations
	switch {
	case calld.SubmitProposal != nil:
		res := gov.SubmitProposal(*calld.SubmitProposal)
		wasmx.SetFinishData(res)
		return
	case calld.VoteWeighted != nil:
		res := gov.VoteWeighted(*calld.VoteWeighted)
		wasmx.SetFinishData(res)
		return
	case calld.Vote != nil:
		res := gov.DoVote(*calld.Vote)
		wasmx.SetFinishData(res)
		return
	case calld.Deposit != nil:
		res := gov.DoDeposit(*calld.Deposit)
		wasmx.SetFinishData(res)
		return
	case calld.GetProposal != nil:
		res := gov.GetProposal(*calld.GetProposal)
		wasmx.SetFinishData(res)
		return
	case calld.GetProposals != nil:
		res := gov.GetProposals(*calld.GetProposals)
		wasmx.SetFinishData(res)
		return
	case calld.GetTallyResult != nil:
		res := gov.GetTallyResult(*calld.GetTallyResult)
		wasmx.SetFinishData(res)
		return
	case calld.GetParams != nil:
		res := gov.GetParams(*calld.GetParams)
		wasmx.SetFinishData(res)
		return
	case calld.GetVote != nil:
		res := gov.GetVote(*calld.GetVote)
		wasmx.SetFinishData(res)
		return
	case calld.GetVotes != nil:
		res := gov.GetVotes(*calld.GetVotes)
		wasmx.SetFinishData(res)
		return
	case calld.GetDeposit != nil:
		res := gov.GetDeposit(*calld.GetDeposit)
		wasmx.SetFinishData(res)
		return
	case calld.GetDeposits != nil:
		res := gov.GetDeposits(*calld.GetDeposits)
		wasmx.SetFinishData(res)
		return
	}

	// internal operations
	switch {
	case calld.EndBlock != nil:
		wasmx.OnlyInternal(gov.MODULE_NAME, "EndBlock")
		res := gov.EndBlock(*calld.EndBlock)
		wasmx.SetFinishData(res)
		return
	case calld.InitGenesis != nil:
		wasmx.OnlyInternal(gov.MODULE_NAME, "InitGenesis")
		res := gov.InitGenesis(*calld.InitGenesis)
		wasmx.SetFinishData(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}

// Exported hook wrappers for direct host calls, similar to emailchain's extra exports
//
//go:wasm-module wasmx-gov
//export end_block
func EndBlockExport() {
	databz := wasmx.GetCallData()
	req := gov.MsgEndBlock{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(gov.EndBlock(req))
}

//go:wasm-module wasmx-gov
//export init_genesis
func InitGenesisExport() {
	databz := wasmx.GetCallData()
	req := gov.GenesisState{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(gov.InitGenesis(req))
}
