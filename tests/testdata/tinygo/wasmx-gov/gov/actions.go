package gov

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

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
		// Use 18 decimal places for precise percentage calculations
		quorumInt := parseDecimalToSdkInt(params.Quorum, 18)
		quorumAmount := totalStake.Mul(quorumInt).Quo(NewSdkIntPow10(18))

		voted := p.FinalTallyResult.YesCount.Add(p.FinalTallyResult.NoCount)
		voted = voted.Add(p.FinalTallyResult.AbstainCount)
		voted = voted.Add(p.FinalTallyResult.NoWithVetoCount)

		LoggerDebug("proposal quorum", []string{"id", u64toa(uint64(p.ID)), "total_stake", totalStake.String(), "quorum", quorumAmount.String(), "voted_state", voted.String()})

		if voted.LT(quorumAmount) {
			p.FailedReason = "lack of quorum"
			p.Status = PROPOSAL_STATUS_REJECTED
			setProposal(uint64(p.ID), p)
			LoggerDebug("proposal failed quorum", []string{"id", u64toa(uint64(p.ID))})
			continue
		}

		// no_with_veto
		thresholdVeto := parseDecimalToSdkInt(params.VetoThreshold, 18)
		thresholdVetoAmt := voted.Mul(thresholdVeto).Quo(NewSdkIntPow10(18))
		LoggerDebug("proposal veto threshold", []string{"id", u64toa(uint64(p.ID)), "veto_count", p.FinalTallyResult.NoWithVetoCount.String(), "threshold", thresholdVetoAmt.String()})
		if p.FinalTallyResult.NoWithVetoCount.GTE(thresholdVetoAmt) {
			p.FailedReason = "vetoed"
			p.Status = PROPOSAL_STATUS_REJECTED
			setProposal(uint64(p.ID), p)
			LoggerDebug("proposal failed with veto", []string{"id", u64toa(uint64(p.ID))})
			continue
		}

		// yes threshold
		thr := parseDecimalToSdkInt(params.Threshold, 18)
		thrAmt := voted.Mul(thr).Quo(NewSdkIntPow10(18))
		LoggerDebug("proposal yes threshold", []string{"id", u64toa(uint64(p.ID)), "yes_count", p.FinalTallyResult.YesCount.String(), "threshold", thrAmt.String()})
		if p.FinalTallyResult.YesCount.LT(thrAmt) {
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

	// TODO
	// Loop through all messages and confirm they are valid and the gov module account
	// is the only signer

	params := getParams()
	now := wasmx.GetTimestamp().UTC()
	depositEnd := now.Add(time.Duration(params.MaxDepositPeriod) * time.Millisecond)
	deposit := req.InitialDeposit
	if len(deposit) == 0 {
		amount := NewSdkIntZero()
		deposit = []wasmx.Coin{{Denom: params.MinDeposit[0].Denom, Amount: amount}}
	}
	if deposit[0].Denom != params.MinDeposit[0].Denom {
		Revert("invalid denom; expected " + params.MinDeposit[0].Denom + ", got " + deposit[0].Denom)
	}

	metadata := req.Metadata
	if len(metadata) > MaxMetadataLen {
		metadata = metadata[:MaxMetadataLen]
	}
	yescount := NewSdkIntZero()
	abstain := NewSdkIntZero()
	nocount := NewSdkIntZero()
	noveto := NewSdkIntZero()
	proposal := Proposal{
		ID:               0,
		Messages:         req.Messages,
		Status:           PROPOSAL_STATUS_DEPOSIT_PERIOD,
		FinalTallyResult: TallyResult{YesCount: yescount, AbstainCount: abstain, NoCount: nocount, NoWithVetoCount: noveto},
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
	if deposit[0].Amount.BigInt().Cmp(params.MinDeposit[0].Amount.BigInt()) > 0 {
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
		amount := proposal.FinalTallyResult.YesCount.Add(stake)
		proposal.FinalTallyResult.YesCount = amount
	case VOTE_OPTION_ABSTAIN:
		amount := proposal.FinalTallyResult.AbstainCount.Add(stake)
		proposal.FinalTallyResult.AbstainCount = amount
	case VOTE_OPTION_NO:
		amount := proposal.FinalTallyResult.NoCount.Add(stake)
		proposal.FinalTallyResult.NoCount = amount
	case VOTE_OPTION_NO_WITH_VETO:
		amount := proposal.FinalTallyResult.NoWithVetoCount.Add(stake)
		proposal.FinalTallyResult.NoWithVetoCount = amount
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
		weightInt := parseDecimalToSdkInt(opt.Weight, 18)
		amount := stake.Mul(weightInt).Quo(NewSdkIntPow10(18))
		switch opt.Option {
		case VOTE_OPTION_YES:
			amount := proposal.FinalTallyResult.YesCount.Add(amount)
			proposal.FinalTallyResult.YesCount = amount
		case VOTE_OPTION_ABSTAIN:
			amount := proposal.FinalTallyResult.AbstainCount.Add(amount)
			proposal.FinalTallyResult.AbstainCount = amount
		case VOTE_OPTION_NO:
			amount := proposal.FinalTallyResult.NoCount.Add(amount)
			proposal.FinalTallyResult.NoCount = amount
		case VOTE_OPTION_NO_WITH_VETO:
			amount := proposal.FinalTallyResult.NoWithVetoCount.Add(amount)
			proposal.FinalTallyResult.NoWithVetoCount = amount
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
				amount := proposal.TotalDeposit[i].Amount.Add(c.Amount)
				proposal.TotalDeposit[i].Amount = amount
				found = true
				break
			}
		}
		if !found {
			proposal.TotalDeposit = append(proposal.TotalDeposit, c)
		}
	}

	params := getParams()
	if proposal.TotalDeposit[0].Amount.BigInt().Cmp(params.MinDeposit[0].Amount.BigInt()) > 0 {
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
