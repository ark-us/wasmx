package main

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env"
	utils "github.com/loredanacirstea/wasmx-utils"
)

//go:wasm-module wasmx-gov
//export instantiate
func Instantiate() {
	// no-op; params must be set via InitGenesis
}

func main() {
	databz := wasmx.GetCallData()
	calld := &Calldata{}
	if err := json.Unmarshal(databz, calld); err != nil {
		wasmx.Revert([]byte("invalid call data: " + err.Error()))
	}

	// public operations
	switch {
	case calld.SubmitProposal != nil:
		res := SubmitProposal(*calld.SubmitProposal)
		wasmx.SetFinishData(res)
		return
	case calld.VoteWeighted != nil:
		res := VoteWeighted(*calld.VoteWeighted)
		wasmx.SetFinishData(res)
		return
	case calld.Vote != nil:
		res := DoVote(*calld.Vote)
		wasmx.SetFinishData(res)
		return
	case calld.Deposit != nil:
		res := DoDeposit(*calld.Deposit)
		wasmx.SetFinishData(res)
		return
	case calld.GetProposal != nil:
		res := GetProposal(*calld.GetProposal)
		wasmx.SetFinishData(res)
		return
	case calld.GetProposals != nil:
		res := GetProposals(*calld.GetProposals)
		wasmx.SetFinishData(res)
		return
	case calld.GetTallyResult != nil:
		res := GetTallyResult(*calld.GetTallyResult)
		wasmx.SetFinishData(res)
		return
	case calld.GetParams != nil:
		res := GetParams(*calld.GetParams)
		wasmx.SetFinishData(res)
		return
	case calld.GetVote != nil:
		res := GetVote(*calld.GetVote)
		wasmx.SetFinishData(res)
		return
	case calld.GetVotes != nil:
		res := GetVotes(*calld.GetVotes)
		wasmx.SetFinishData(res)
		return
	case calld.GetDeposit != nil:
		res := GetDeposit(*calld.GetDeposit)
		wasmx.SetFinishData(res)
		return
	case calld.GetDeposits != nil:
		res := GetDeposits(*calld.GetDeposits)
		wasmx.SetFinishData(res)
		return
	}

	// internal operations
	switch {
	case calld.EndBlock != nil:
		wasmx.OnlyInternal(MODULE_NAME, "EndBlock")
		res := EndBlock(*calld.EndBlock)
		wasmx.SetFinishData(res)
		return
	case calld.InitGenesis != nil:
		wasmx.OnlyInternal(MODULE_NAME, "InitGenesis")
		res := InitGenesis(*calld.InitGenesis)
		wasmx.SetFinishData(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}

// Handlers

func InitGenesis(req GenesisState) []byte {
	LoggerInfo("initiating genesis", nil)
	setProposalIdFirst(int64(req.StartingProposalID))
	setProposalIdLast(int64(req.StartingProposalID) + int64(len(req.Proposals)))
	setProposalIdCount(int64(len(req.Proposals)) - 1)
	setParams(req.Params)

	for _, p := range req.Proposals {
		setProposal(uint64(p.ID), p)
		if p.Status == PROPOSAL_STATUS_DEPOSIT_PERIOD {
			addActiveDepositProposal(uint64(p.ID))
		}
		if p.Status == PROPOSAL_STATUS_VOTING_PERIOD {
			addActiveVotingProposal(uint64(p.ID))
		}
	}
	for _, d := range req.Deposits {
		addProposalDeposit(uint64(d.ProposalID), d)
	}
	for _, v := range req.Votes {
		addProposalVote(uint64(v.ProposalID), v)
	}
	LoggerInfo("initiated genesis", []string{"proposals", itoa(len(req.Proposals)), "deposits", itoa(len(req.Deposits)), "votes", itoa(len(req.Votes))})
	return []byte{}
}

func EndBlock(req MsgEndBlock) []byte {
	// decode base64 block entry then header
	var block struct {
		Header string `json:"header"`
	}
	hbz, _ := base64.StdEncoding.DecodeString(req.Data)
	_ = json.Unmarshal(hbz, &block)
	var header struct {
		Time string `json:"time"`
	}
	hhbz, _ := base64.StdEncoding.DecodeString(block.Header)
	_ = json.Unmarshal(hhbz, &header)
	t, _ := time.Parse(time.RFC3339Nano, header.Time)

	// deposit period expirations
	expired := nextEndingDepositProposals(t)
	if len(expired) > 0 {
		LoggerDebug("gov proposals expired deposit", []string{"count", itoa(len(expired)), "block_time", t.Format(time.RFC3339Nano)})
		for _, p := range expired {
			LoggerInfo("deleting proposal", []string{"reason", "deposit period expired", "proposal_id", u64toa(uint64(p.ID))})
			removeProposal(uint64(p.ID))
			removeProposalDeposits(uint64(p.ID))
		}
	}

	// voting period endings
	params := getParams()
	ending := nextEndingVotingProposals(t)
	if len(ending) > 0 {
		LoggerDebug("gov proposals ending voting period", []string{"count", itoa(len(ending))})
	}
	for _, p := range ending {
		removeActiveVotingProposal(uint64(p.ID))

		totalStake := callGetTotalStake()
		dec := decimalCount(params.Quorum)
		quorumInt := strToScaledInt(params.Quorum, dec)
		quorumAmount := totalStake.Mul(quorumInt).Div(NewBigPow10(dec))

		voted := new(Big)
		voted.Int = new(big.Int).Add(p.FinalTallyResult.YesCount.Int, p.FinalTallyResult.NoCount.Int)
		voted.Int = voted.Int.Add(voted.Int, p.FinalTallyResult.AbstainCount.Int)
		voted.Int = voted.Int.Add(voted.Int, p.FinalTallyResult.NoWithVetoCount.Int)

		LoggerDebug("proposal quorum", []string{"id", u64toa(uint64(p.ID)), "total_stake", totalStake.String(), "quorum", quorumAmount.String(), "voted_state", voted.String()})

		if voted.Cmp(quorumAmount) < 0 {
			p.FailedReason = "lack of quorum"
			p.Status = PROPOSAL_STATUS_REJECTED
			setProposal(uint64(p.ID), p)
			LoggerDebug("proposal failed quorum", []string{"id", u64toa(uint64(p.ID))})
			continue
		}

		// no_with_veto
		decVeto := decimalCount(params.VetoThreshold)
		thresholdVeto := strToScaledInt(params.VetoThreshold, decVeto)
		thresholdVetoAmt := voted.Mul(thresholdVeto).Div(NewBigPow10(decVeto))
		LoggerDebug("proposal veto threshold", []string{"id", u64toa(uint64(p.ID)), "veto_count", p.FinalTallyResult.NoWithVetoCount.String(), "threshold", thresholdVetoAmt.String()})
		if p.FinalTallyResult.NoWithVetoCount.Cmp(thresholdVetoAmt) >= 0 {
			p.FailedReason = "vetoed"
			p.Status = PROPOSAL_STATUS_REJECTED
			setProposal(uint64(p.ID), p)
			LoggerDebug("proposal failed with veto", []string{"id", u64toa(uint64(p.ID))})
			continue
		}

		// yes threshold
		decThr := decimalCount(params.Threshold)
		thr := strToScaledInt(params.Threshold, decThr)
		thrAmt := voted.Mul(thr).Div(NewBigPow10(decThr))
		LoggerDebug("proposal yes threshold", []string{"id", u64toa(uint64(p.ID)), "yes_count", p.FinalTallyResult.YesCount.String(), "threshold", thrAmt.String()})
		if p.FinalTallyResult.YesCount.Cmp(thrAmt) < 0 {
			p.FailedReason = "not enough yes votes"
			p.Status = PROPOSAL_STATUS_REJECTED
			setProposal(uint64(p.ID), p)
			LoggerInfo("proposal rejected", []string{"id", u64toa(uint64(p.ID))})
			continue
		}

		// execute messages
		LoggerInfo("proposal passed", []string{"id", u64toa(uint64(p.ID))})
		result := executeProposal(p)
		if result.Success {
			p.Status = PROPOSAL_STATUS_PASSED
			LoggerInfo("proposal passed and execution succeeded", []string{"id", u64toa(uint64(p.ID))})
		} else {
			p.Status = PROPOSAL_STATUS_FAILED
			p.FailedReason = result.Data
			LoggerInfo("proposal passed and execution failed", []string{"id", u64toa(uint64(p.ID)), "error", result.Data})
		}
		setProposal(uint64(p.ID), p)
	}

	return []byte{}
}

func SubmitProposal(req MsgSubmitProposal) []byte {
	LoggerDebug("submit proposal", []string{"title", req.Title})
	params := getParams()
	now := wasmx.GetTimestamp().UTC()
	depositEnd := now.Add(time.Duration(params.MaxDepositPeriod) * time.Millisecond)
	deposit := req.InitialDeposit
	if len(deposit) == 0 {
		deposit = []Coin{{Denom: params.MinDeposit[0].Denom, Amount: NewBigZero()}}
	}
	if deposit[0].Denom != params.MinDeposit[0].Denom {
		Revert("invalid denom; expected " + params.MinDeposit[0].Denom + ", got " + deposit[0].Denom)
	}

	metadata := req.Metadata
	if len(metadata) > MaxMetadataLen {
		metadata = metadata[:MaxMetadataLen]
	}
	proposal := Proposal{
		ID:               0,
		Messages:         req.Messages,
		Status:           PROPOSAL_STATUS_DEPOSIT_PERIOD,
		FinalTallyResult: TallyResult{YesCount: NewBigZero(), AbstainCount: NewBigZero(), NoCount: NewBigZero(), NoWithVetoCount: NewBigZero()},
		SubmitTime:       now.Format(time.RFC3339Nano),
		DepositEndTime:   depositEnd.Format(time.RFC3339Nano),
		TotalDeposit:     deposit,
		VotingStartTime:  "0001-01-01T00:00:00Z",
		VotingEndTime:    "0001-01-01T00:00:00Z",
		Metadata:         metadata,
		Title:            req.Title,
		Summary:          req.Summary,
		Proposer:         req.Proposer,
		Expedited:        req.Expedited,
		FailedReason:     "",
	}
	// promote to voting if initial deposit >= min deposit
	if deposit[0].Amount.Cmp(params.MinDeposit[0].Amount) > 0 {
		proposal.Status = PROPOSAL_STATUS_VOTING_PERIOD
		vs := now
		ve := vs.Add(time.Duration(params.VotingPeriod) * time.Millisecond)
		proposal.VotingStartTime = vs.Format(time.RFC3339Nano)
		proposal.VotingEndTime = ve.Format(time.RFC3339Nano)
	}
	pid := addProposal(proposal)
	if proposal.Status == PROPOSAL_STATUS_DEPOSIT_PERIOD {
		addActiveDepositProposal(pid)
	} else if proposal.Status == PROPOSAL_STATUS_VOTING_PERIOD {
		addActiveVotingProposal(pid)
	}

	// transfer deposit from proposer to this contract/module
	bankSendCoinFromAccountToModule(req.Proposer, MODULE_NAME, req.InitialDeposit)

	addProposalDeposit(pid, Deposit{ProposalID: utils.StringUint64(pid), Depositor: req.Proposer, Amount: req.InitialDeposit})

	// Emit event, mirroring AS
	ev := wasmx.Event{
		Type: EventTypeSubmitProposal,
		Attributes: []wasmx.EventAttribute{
			{Key: AttributeKeyProposalID, Value: u64toa(pid), Index: true},
			{Key: AttributeKeyProposalMessages, Value: strings.Join(req.Messages, ","), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	resp, _ := json.Marshal(MsgSubmitProposalResponse{ProposalID: utils.StringUint64(pid)})
	return resp
}

func DoVote(req MsgVote) []byte {
	LoggerDebug("vote", []string{"proposal_id", u64toa(uint64(req.ProposalID)), "option", req.Option})
	proposal, ok := getProposal(uint64(req.ProposalID))
	if !ok {
		Revert("invalid proposal id: " + u64toa(uint64(req.ProposalID)))
	}
	optionID, ok := VoteOptionMap[req.Option]
	if !ok {
		Revert("invalid vote option: " + req.Option)
	}
	option := WeightedVoteOption{Option: optionID, Weight: "1.0"}
	metadata := truncate(req.Metadata, MaxMetadataLen)
	addProposalVote(uint64(req.ProposalID), Vote{ProposalID: utils.StringUint64(req.ProposalID), Voter: req.Voter, Options: []WeightedVoteOption{option}, Metadata: metadata})

	stake := getStake(req.Voter)
	switch optionID {
	case VOTE_OPTION_YES:
		proposal.FinalTallyResult.YesCount = proposal.FinalTallyResult.YesCount.Add(stake)
	case VOTE_OPTION_ABSTAIN:
		proposal.FinalTallyResult.AbstainCount = proposal.FinalTallyResult.AbstainCount.Add(stake)
	case VOTE_OPTION_NO:
		proposal.FinalTallyResult.NoCount = proposal.FinalTallyResult.NoCount.Add(stake)
	case VOTE_OPTION_NO_WITH_VETO:
		proposal.FinalTallyResult.NoWithVetoCount = proposal.FinalTallyResult.NoWithVetoCount.Add(stake)
	}
	setProposal(uint64(proposal.ID), proposal)
	// Emit vote event
	optbz, _ := json.Marshal([]WeightedVoteOption{option})
	ev := wasmx.Event{
		Type: EventTypeProposalVote,
		Attributes: []wasmx.EventAttribute{
			{Key: AttributeKeyVoter, Value: string(req.Voter), Index: true},
			{Key: AttributeKeyOption, Value: string(optbz), Index: true},
			{Key: AttributeKeyProposalID, Value: u64toa(uint64(proposal.ID)), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})
	resp, _ := json.Marshal(MsgVoteResponse{})
	return resp
}

func VoteWeighted(req MsgVoteWeighted) []byte {
	proposal, ok := getProposal(uint64(req.ProposalID))
	if !ok {
		Revert("invalid proposal id: " + u64toa(uint64(req.ProposalID)))
	}
	if proposal.Status != PROPOSAL_STATUS_VOTING_PERIOD {
		Revert("cannot vote, proposal is in status: " + itoa(int(proposal.Status)))
	}
	metadata := truncate(req.Metadata, MaxMetadataLen)
	addProposalVote(uint64(req.ProposalID), Vote{ProposalID: utils.StringUint64(req.ProposalID), Voter: req.Voter, Options: req.Option, Metadata: metadata})

	stake := getStake(req.Voter)
	for _, opt := range req.Option {
		dec := decimalCount(opt.Weight)
		weightInt := strToScaledInt(opt.Weight, dec)
		amount := stake.Mul(weightInt).Div(NewBigPow10(dec))
		switch opt.Option {
		case VOTE_OPTION_YES:
			proposal.FinalTallyResult.YesCount = proposal.FinalTallyResult.YesCount.Add(amount)
		case VOTE_OPTION_ABSTAIN:
			proposal.FinalTallyResult.AbstainCount = proposal.FinalTallyResult.AbstainCount.Add(amount)
		case VOTE_OPTION_NO:
			proposal.FinalTallyResult.NoCount = proposal.FinalTallyResult.NoCount.Add(amount)
		case VOTE_OPTION_NO_WITH_VETO:
			proposal.FinalTallyResult.NoWithVetoCount = proposal.FinalTallyResult.NoWithVetoCount.Add(amount)
		}
	}
	setProposal(uint64(proposal.ID), proposal)
	// Emit weighted vote event
	optbz, _ := json.Marshal(req.Option)
	ev := wasmx.Event{
		Type: EventTypeProposalVote,
		Attributes: []wasmx.EventAttribute{
			{Key: AttributeKeyVoter, Value: string(req.Voter), Index: true},
			{Key: AttributeKeyOption, Value: string(optbz), Index: true},
			{Key: AttributeKeyProposalID, Value: u64toa(uint64(proposal.ID)), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})
	resp, _ := json.Marshal(MsgVoteResponse{})
	return resp
}

func DoDeposit(req MsgDeposit) []byte {
	proposal, ok := getProposal(uint64(req.ProposalID))
	if !ok {
		Revert("invalid proposal id: " + u64toa(uint64(req.ProposalID)))
	}
	if proposal.Status != PROPOSAL_STATUS_DEPOSIT_PERIOD {
		Revert("cannot deposit, proposal is in status: " + itoa(int(proposal.Status)))
	}

	bankSendCoinFromAccountToModule(req.Depositor, MODULE_NAME, req.Amount)

	addProposalDeposit(uint64(req.ProposalID), Deposit{ProposalID: utils.StringUint64(req.ProposalID), Depositor: req.Depositor, Amount: req.Amount})

	// update total deposit (single denom)
	for _, c := range req.Amount {
		found := false
		for i := range proposal.TotalDeposit {
			if proposal.TotalDeposit[i].Denom == c.Denom {
				proposal.TotalDeposit[i].Amount = proposal.TotalDeposit[i].Amount.Add(c.Amount)
				found = true
				break
			}
		}
		if !found {
			proposal.TotalDeposit = append(proposal.TotalDeposit, c)
		}
	}

	params := getParams()
	if proposal.TotalDeposit[0].Amount.Cmp(params.MinDeposit[0].Amount) > 0 {
		proposal.Status = PROPOSAL_STATUS_VOTING_PERIOD
		vs := wasmx.GetTimestamp().UTC()
		proposal.VotingStartTime = vs.Format(time.RFC3339Nano)
		proposal.VotingEndTime = vs.Add(time.Duration(params.VotingPeriod) * time.Millisecond).Format(time.RFC3339Nano)
		addActiveVotingProposal(uint64(proposal.ID))
		removeActiveDepositProposal(uint64(proposal.ID))
		// emit voting period start event
		ev := wasmx.Event{
			Type: EventTypeProposalDeposit,
			Attributes: []wasmx.EventAttribute{
				{Key: AttributeKeyVotingPeriodStart, Value: u64toa(uint64(proposal.ID)), Index: true},
			},
		}
		wasmx.EmitCosmosEvents([]wasmx.Event{ev})
	}
	setProposal(uint64(proposal.ID), proposal)
	return []byte{}
}

// queries

func GetProposal(req QueryProposalRequest) []byte {
	proposal, ok := getProposal(uint64(req.ProposalID))
	var resp []byte
	if ok {
		bz, _ := json.Marshal(QueryProposalResponse{Proposal: &proposal})
		resp = bz
	} else {
		resp = []byte("{\"proposal\":null}")
	}
	return resp
}

func GetProposals(req QueryProposalsRequest) []byte {
	lastID := getProposalIdLast()
	firstID := getProposalIdFirst()
	count := getProposalIdCount()
	var props []Proposal
	for i := firstID; i <= lastID; i++ {
		if p, ok := getProposal(i); ok {
			props = append(props, p)
		}
	}
	bz, _ := json.Marshal(QueryProposalsResponse{Proposals: props, Pagination: PageResponse{Total: utils.StringUint64(uint64(count))}})
	return bz
}

func GetParams(_ QueryParamsRequest) []byte {
	p := getParams()
	bz, _ := json.Marshal(QueryParamsResponse{Params: p})
	return bz
}

func GetTallyResult(req QueryTallyResultRequest) []byte {
	proposal, ok := getProposal(uint64(req.ProposalID))
	var resp []byte
	if ok {
		bz, _ := json.Marshal(QueryTallyResultResponse{Tally: proposal.FinalTallyResult})
		resp = bz
	} else {
		resp = []byte("{\"tally\":null}")
	}
	return resp
}

// Stubs kept for parity with AssemblyScript
func GetVote(_ QueryVoteRequest) []byte         { return []byte{} }
func GetVotes(_ QueryVotesRequest) []byte       { return []byte{} }
func GetDeposit(_ QueryDepositRequest) []byte   { return []byte{} }
func GetDeposits(_ QueryDepositsRequest) []byte { return []byte{} }

// Exported hook wrappers for direct host calls, similar to emailchain's extra exports
//
//go:wasm-module wasmx-gov
//export end_block
func EndBlockExport() {
	databz := wasmx.GetCallData()
	req := MsgEndBlock{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(EndBlock(req))
}

//go:wasm-module wasmx-gov
//export init_genesis
func InitGenesisExport() {
	databz := wasmx.GetCallData()
	req := GenesisState{}
	if err := json.Unmarshal(databz, &req); err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	wasmx.SetFinishData(InitGenesis(req))
}
