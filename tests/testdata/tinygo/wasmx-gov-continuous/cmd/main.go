package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	gov "github.com/loredanacirstea/wasmx-gov-continuous/gov"
	gov1 "github.com/loredanacirstea/wasmx-gov/gov"
)

//go:wasm-module wasmx-gov-continuous
//export instantiate
func Instantiate() {
	// Initialize with default params if needed
	databz := wasmx.GetCallData()
	var params gov.Params
	err := json.Unmarshal(databz, &params)
	if err != nil {
		gov.Revert("invalid initiate with params: " + err.Error() + ": " + string(databz))
	}
	gov.SetParams(params)
}

func main() {
	databz := wasmx.GetCallData()
	calldata := &gov.CallData{}
	if err := json.Unmarshal(databz, calldata); err != nil {
		gov.Revert("invalid call data: " + err.Error() + ": " + string(databz))
	}

	// Public operations
	switch {
	case calldata.SubmitProposal != nil:
		res := gov.SubmitProposal(*calldata.SubmitProposal)
		wasmx.SetFinishData(res)
		return
	case calldata.SubmitProposalExtended != nil:
		res := gov.SubmitProposalExtended(*calldata.SubmitProposalExtended)
		wasmx.SetFinishData(res)
		return
	case calldata.AddProposalOption != nil:
		res := gov.AddProposalOption(*calldata.AddProposalOption)
		wasmx.SetFinishData(res)
		return
	case calldata.VoteWeighted != nil:
		res := gov.VoteWeighted(*calldata.VoteWeighted)
		wasmx.SetFinishData(res)
		return
	case calldata.Vote != nil:
		res := gov.DoVote(*calldata.Vote)
		wasmx.SetFinishData(res)
		return
	case calldata.Deposit != nil:
		res := gov.DoDeposit(*calldata.Deposit)
		wasmx.SetFinishData(res)
		return
	case calldata.DepositVote != nil:
		res := gov.DoDepositVote(*calldata.DepositVote)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposal != nil:
		res := gov.GetProposal(*calldata.GetProposal)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposals != nil:
		res := gov.GetProposals(*calldata.GetProposals)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposalsExtended != nil:
		res := gov.GetProposalsExtended(*calldata.GetProposalsExtended)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposalExtended != nil:
		res := gov.GetProposalExtended(*calldata.GetProposalExtended)
		wasmx.SetFinishData(res)
		return
	case calldata.GetTallyResult != nil:
		res := gov.GetTallyResult(*calldata.GetTallyResult)
		wasmx.SetFinishData(res)
		return
	case calldata.GetNextWinnerThreshold != nil:
		res := gov.GetNextWinnerThreshold(*calldata.GetNextWinnerThreshold)
		wasmx.SetFinishData(res)
		return
	case calldata.GetParams != nil:
		res := gov1.GetParams(*calldata.GetParams)
		wasmx.SetFinishData(res)
		return
	}

	// Internal operations
	switch {
	case calldata.EndBlock != nil:
		wasmx.OnlyInternal(gov.MODULE_NAME, "EndBlock")
		res := gov.EndBlock(*calldata.EndBlock)
		wasmx.SetFinishData(res)
		return
	case calldata.InitGenesis != nil:
		wasmx.OnlyInternal(gov.MODULE_NAME, "InitGenesis")
		res := gov.InitGenesis(*calldata.InitGenesis)
		wasmx.SetFinishData(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}

// Exported hook wrappers for direct host calls
//
//go:wasm-module wasmx-gov-continuous
//export end_block
func EndBlockExport() {
	databz := wasmx.GetCallData()
	req := gov.MsgEndBlock{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(gov.EndBlock(req))
}

//go:wasm-module wasmx-gov-continuous
//export init_genesis
func InitGenesisExport() {
	databz := wasmx.GetCallData()
	req := gov.MsgInitGenesis{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(gov.InitGenesis(req))
}
