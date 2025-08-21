package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
)

//go:wasm-module wasmx-gov-continuous
//export instantiate
func Instantiate() {
	// Initialize with default params if needed
	databz := wasmx.GetCallData()
	if len(databz) > 0 {
		var params Params
		if err := json.Unmarshal(databz, &params); err == nil {
			setParams(params)
		}
	}
}

func main() {
	databz := wasmx.GetCallData()
	calldata := &CallData{}
	if err := json.Unmarshal(databz, calldata); err != nil {
		wasmx.Revert([]byte("invalid call data: " + err.Error()))
	}

	// Public operations
	switch {
	case calldata.SubmitProposal != nil:
		res := SubmitProposal(*calldata.SubmitProposal)
		wasmx.SetFinishData(res)
		return
	case calldata.SubmitProposalExtended != nil:
		res := SubmitProposalExtended(*calldata.SubmitProposalExtended)
		wasmx.SetFinishData(res)
		return
	case calldata.AddProposalOption != nil:
		res := AddProposalOption(*calldata.AddProposalOption)
		wasmx.SetFinishData(res)
		return
	case calldata.VoteWeighted != nil:
		res := VoteWeighted(*calldata.VoteWeighted)
		wasmx.SetFinishData(res)
		return
	case calldata.Vote != nil:
		res := DoVote(*calldata.Vote)
		wasmx.SetFinishData(res)
		return
	case calldata.Deposit != nil:
		res := DoDeposit(*calldata.Deposit)
		wasmx.SetFinishData(res)
		return
	case calldata.DepositVote != nil:
		res := DoDepositVote(*calldata.DepositVote)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposal != nil:
		res := GetProposal(*calldata.GetProposal)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposals != nil:
		res := GetProposals(*calldata.GetProposals)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposalsExtended != nil:
		res := GetProposalsExtended(*calldata.GetProposalsExtended)
		wasmx.SetFinishData(res)
		return
	case calldata.GetProposalExtended != nil:
		res := GetProposalExtended(*calldata.GetProposalExtended)
		wasmx.SetFinishData(res)
		return
	case calldata.GetTallyResult != nil:
		res := GetTallyResult(*calldata.GetTallyResult)
		wasmx.SetFinishData(res)
		return
	case calldata.GetNextWinnerThreshold != nil:
		res := GetNextWinnerThreshold(*calldata.GetNextWinnerThreshold)
		wasmx.SetFinishData(res)
		return
	case calldata.GetParams != nil:
		res := GetParams(*calldata.GetParams)
		wasmx.SetFinishData(res)
		return
	}

	// Internal operations
	switch {
	case calldata.EndBlock != nil:
		wasmx.OnlyInternal(MODULE_NAME, "EndBlock")
		res := EndBlock(*calldata.EndBlock)
		wasmx.SetFinishData(res)
		return
	case calldata.InitGenesis != nil:
		wasmx.OnlyInternal(MODULE_NAME, "InitGenesis")
		res := InitGenesis(*calldata.InitGenesis)
		wasmx.SetFinishData(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}

// GetParams returns the module parameters
func GetParams(req QueryParamsRequest) []byte {
	params := getParams()
	resp := struct {
		Params Params `json:"params"`
	}{
		Params: params,
	}
	result, _ := json.Marshal(resp)
	return result
}

// Exported hook wrappers for direct host calls
//
//go:wasm-module wasmx-gov-continuous
//export end_block
func EndBlockExport() {
	databz := wasmx.GetCallData()
	req := MsgEndBlock{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(EndBlock(req))
}

//go:wasm-module wasmx-gov-continuous
//export init_genesis
func InitGenesisExport() {
	databz := wasmx.GetCallData()
	req := MsgInitGenesis{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(InitGenesis(req))
}