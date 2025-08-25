package gov

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	wasmx "github.com/loredanacirstea/wasmx-env"
	gov "github.com/loredanacirstea/wasmx-gov/gov"
	utils "github.com/loredanacirstea/wasmx-utils"
)

const OPTION_ID_START = 2

// InitGenesis initializes the continuous voting module
func InitGenesis(req MsgInitGenesis) []byte {
	LoggerInfo("initiating genesis", nil)

	// Set base gov state using base gov storage functions
	gov.SetProposalIdFirst(int64(req.StartingProposalID))
	gov.SetProposalIdLast(int64(req.StartingProposalID) + int64(len(req.Proposals)) - 1)
	gov.SetProposalIdCount(int64(len(req.Proposals)))
	gov.SetParams(req.Params)

	// Store proposals
	for _, proposal := range req.Proposals {
		setProposal(uint64(proposal.ID), proposal)
	}

	// Store deposits using base gov-compatible keys
	for _, deposit := range req.Deposits {
		addProposalDeposit(uint64(deposit.ProposalID), deposit)
	}

	// Store votes (continuous voting specific)
	for _, vote := range req.Votes {
		addProposalVote(uint64(vote.ProposalID), vote)
	}

	LoggerInfo("initiated genesis", []string{
		"proposals", utils.Itoa(len(req.Proposals)),
		"deposits", utils.Itoa(len(req.Deposits)),
		"votes", utils.Itoa(len(req.Votes)),
	})
	return []byte{}
}

// EndBlock processes end of block for continuous voting
func EndBlock(req MsgEndBlock) []byte {
	// TODO: Implement continuous voting logic for end block
	// - Check if there are proposals to be executed
	// - Execute proposals based on continuous voting outcomes
	// - Update vote statuses

	return []byte{}
}

// SubmitProposal handles standard proposal submission
func SubmitProposal(req MsgSubmitProposal) []byte {
	localParams := getParams()
	LoggerDebug("submit proposal", []string{"title", req.Title})

	extended := MsgSubmitProposalExtended{
		Messages:       req.Messages,
		InitialDeposit: req.InitialDeposit,
		Proposer:       req.Proposer,
		Metadata:       req.Metadata,
		Title:          req.Title,
		Summary:        req.Summary,
		Expedited:      req.Expedited,
		X:              localParams.DefaultX,
		Y:              localParams.DefaultY,
		OptionTitle:    req.Title,
		OptionSummary:  req.Summary,
		OptionMetadata: req.Metadata,
	}

	return SubmitProposalInternal(extended, localParams)
}

// SubmitProposalExtended handles extended proposal submission with custom parameters
func SubmitProposalExtended(req MsgSubmitProposalExtended) []byte {
	LoggerDebug("submit proposal extended", []string{"title", req.Title})
	localParams := getParams()

	// Check if proposer can set custom coefficients
	hasArbitration := false
	for _, deposit := range req.InitialDeposit {
		if deposit.Denom == localParams.ArbitrationDenom {
			amount := NewBigFromString(string(deposit.Amount.String()))
			if amount.Cmp(big.NewInt(1)) > 0 {
				hasArbitration = true
				break
			}
		}
	}

	if !hasArbitration && (req.X != localParams.DefaultX || req.Y != localParams.DefaultY) {
		Revert("only arbiters can propose custom curves for proposals")
	}

	return SubmitProposalInternal(req, localParams)
}

// SubmitProposalInternal handles the core proposal submission logic
func SubmitProposalInternal(req MsgSubmitProposalExtended, localParams Params) []byte {
	now := wasmx.GetTimestamp().UTC()
	depositEndTime := now // For continuous voting, proposals can start voting immediately

	if len(req.InitialDeposit) == 0 {
		Revert("proposal must contain a deposit")
	}
	if len(req.InitialDeposit) >= 2 {
		Revert("proposal contains too many denoms")
	}

	// Transfer deposit from proposer to this module (lock funds)
	bankSendCoinFromAccountToModule(req.Proposer, wasmx.Bech32String(wasmx.ROLE_GOVERNANCE), req.InitialDeposit)

	var proposalCoin wasmx.Coin
	arbitrationAmount := sdkmath.ZeroInt()

	for _, deposit := range req.InitialDeposit {
		if deposit.Denom == localParams.ArbitrationDenom {
			arbitrationAmount = *deposit.Amount
		} else {
			proposalCoin = deposit
		}
	}
	w := sdkmath.NewInt(0)

	// Create the proposer's option
	firstOption := ProposalOption{
		Proposer:          req.Proposer,
		Messages:          req.Messages,
		Amount:            proposalCoin.Amount,
		ArbitrationAmount: &arbitrationAmount,
		Weight:            &w,
		Title:             req.OptionTitle,
		Summary:           req.OptionSummary,
		Metadata:          req.OptionMetadata,
	}

	// default options + proposer option
	options := defaultProposalOptions()
	options = append(options, firstOption)

	// Truncate metadata similar to AS MaxMetadataLen
	metadata := req.Metadata
	if len(metadata) > MaxMetadataLen {
		metadata = metadata[:MaxMetadataLen]
	}

	// Create proposal
	proposal := Proposal{
		Status:          PROPOSAL_STATUS_VOTING_PERIOD, // Continuous voting starts immediately
		SubmitTime:      now.Format(time.RFC3339Nano),
		DepositEndTime:  depositEndTime.Format(time.RFC3339Nano),
		VotingStartTime: now.Format(time.RFC3339Nano),
		VotingEndTime:   now.Add(365 * 24 * time.Hour).Format(time.RFC3339Nano), // Long voting period
		Metadata:        metadata,
		Title:           req.Title,
		Summary:         req.Summary,
		Proposer:        req.Proposer,
		FailedReason:    "",
		X:               req.X,
		Y:               req.Y,
		Denom:           proposalCoin.Denom,
		Options:         options,
		VoteStatus: ProposalVoteStatus{
			Status:  VOTE_STATUS_UNSPECIFIED,
			Xi:      0,
			Yi:      0,
			Changed: false,
		},
		Winner: 0,
	}

	// fill in proposalVoteStatus
	proposal = setProposalVoteStatus(proposal, localParams)

	proposalID := addProposal(proposal)

	// Add initial deposit vote
	initialVote := DepositVote{
		ProposalID:        utils.StringUint64(proposalID),
		OptionID:          int32(OPTION_ID_START),
		Voter:             req.Proposer,
		Amount:            proposalCoin.Amount,
		ArbitrationAmount: &arbitrationAmount,
		Metadata:          firstOption.Metadata,
	}
	addProposalVote(proposalID, initialVote)

	// Emit event
	ev := wasmx.Event{
		Type: "submit_proposal",
		Attributes: []wasmx.EventAttribute{
			{Key: "proposal_id", Value: utils.U64toa(proposalID), Index: true},
			{Key: "proposal_messages", Value: strings.Join(req.Messages, ","), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	// Emit add_proposal_option event for the proposer's option, mirroring AS
	addOptEv := wasmx.Event{
		Type: EventTypeAddProposalOption,
		Attributes: []wasmx.EventAttribute{
			{Key: AttributeKeyOptionID, Value: utils.Itoa(int(OPTION_ID_START)), Index: true},
			{Key: gov.AttributeKeyProposalID, Value: utils.U64toa(proposalID), Index: true},
			{Key: gov.AttributeKeyOption, Value: marshalJSON(firstOption), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{addOptEv})

	tryExecuteProposal(&proposal)

	resp := struct {
		ProposalID utils.StringUint64 `json:"proposal_id"`
	}{
		ProposalID: utils.StringUint64(proposalID),
	}

	result, _ := json.Marshal(resp)
	return result
}

// AddProposalOption adds a new option to an existing proposal
func AddProposalOption(req MsgAddProposalOption) []byte {
	proposal, exists := getProposal(uint64(req.ProposalID))
	if !exists {
		Revert("proposal not found: " + req.ProposalID.ToString())
	}

	if proposal.Status != PROPOSAL_STATUS_VOTING_PERIOD {
		Revert("proposal is not in voting period")
	}

	// Validate option deposit
	localParams := getParams()
	optionRegisterAmount := NewBigFromUint64(localParams.Coefs[OptionRegisterAmount])
	if req.Option.Amount.BigInt().Cmp(optionRegisterAmount) < 0 {
		Revert("insufficient option registration amount")
	}

	// Add option to proposal
	proposal.Options = append(proposal.Options, req.Option)

	// lock funds for the new option (proposal denom + arbitration denom)
	// ensure non-nil arbitration amount
	arbAmt := req.Option.ArbitrationAmount
	if arbAmt == nil {
		z := sdkmath.NewInt(0)
		arbAmt = &z
	}
	deposit := []wasmx.Coin{
		{Denom: proposal.Denom, Amount: req.Option.Amount},
		{Denom: localParams.ArbitrationDenom, Amount: arbAmt},
	}
	bankSendCoinFromAccountToModule(req.Option.Proposer, wasmx.Bech32String(wasmx.ROLE_GOVERNANCE), deposit)

	// Recalculate voting status and weights
	proposalLocalParams := getParams()
	*proposal = setProposalVoteStatus(*proposal, proposalLocalParams)

	// Save proposal and store an initial vote record for the new option
	setProposal(uint64(req.ProposalID), *proposal)
	vote := DepositVote{
		ProposalID:        req.ProposalID,
		OptionID:          int32(len(proposal.Options) - 1),
		Voter:             req.Option.Proposer,
		Amount:            req.Option.Amount,
		ArbitrationAmount: req.Option.ArbitrationAmount,
		Metadata:          req.Option.Metadata,
	}
	addProposalVote(uint64(req.ProposalID), vote)

	// Emit event
	ev := wasmx.Event{
		Type: EventTypeAddProposalOption,
		Attributes: []wasmx.EventAttribute{
			{Key: AttributeKeyOptionID, Value: utils.Itoa(len(proposal.Options) - 1), Index: true},
			{Key: gov.AttributeKeyProposalID, Value: req.ProposalID.ToString(), Index: true},
			{Key: gov.AttributeKeyOption, Value: marshalJSON(req.Option), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	// Emit a proposal_vote event mirroring AS
	voteEv := wasmx.Event{
		Type: gov.EventTypeProposalVote,
		Attributes: []wasmx.EventAttribute{
			{Key: gov.AttributeKeyVoter, Value: string(req.Option.Proposer), Index: true},
			{Key: gov.AttributeKeyOption, Value: marshalJSON(vote), Index: true},
			{Key: gov.AttributeKeyProposalID, Value: req.ProposalID.ToString(), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{voteEv})

	tryExecuteProposal(proposal)

	return []byte{}
}

// VoteWeighted handles weighted voting (delegated to DoVote for simplicity)
func VoteWeighted(req MsgVoteWeighted) []byte {
	Revert("VoteWeighted not available; use \"depositVote\"")
	return []byte{}
}

// DoVote handles voting logic
func DoVote(req MsgVote) []byte {
	Revert("Vote not available; use \"depositVote\"")
	return []byte{}
}

// DoDeposit handles deposit logic for continuous voting
func DoDeposit(req MsgDeposit) []byte {
	Revert("Deposit not available; use \"depositVote\"")
	return []byte{}
}

// DoDepositVote handles deposit voting specific to continuous voting
func DoDepositVote(req DepositVote) []byte {
	proposal, exists := getProposal(uint64(req.ProposalID))
	if !exists {
		Revert("proposal not found: " + req.ProposalID.ToString())
	}

	if proposal.Status != PROPOSAL_STATUS_VOTING_PERIOD {
		Revert("proposal is not in voting period")
	}

	// Validate option ID
	if int(req.OptionID) >= len(proposal.Options) {
		Revert("invalid option ID")
	}

	// lock funds for this deposit vote
	localParams := getParams()
	arb := req.ArbitrationAmount
	if arb == nil {
		z := sdkmath.NewInt(0)
		arb = &z
	}
	deposit := []wasmx.Coin{
		{Denom: proposal.Denom, Amount: req.Amount},
		{Denom: localParams.ArbitrationDenom, Amount: arb},
	}
	bankSendCoinFromAccountToModule(req.Voter, wasmx.Bech32String(wasmx.ROLE_GOVERNANCE), deposit)

	// Add vote
	addProposalVote(uint64(req.ProposalID), req)

	// Update proposal option deposit amounts
	opt := proposal.Options[req.OptionID]
	amt := opt.Amount.Add(*req.Amount)
	opt.Amount = &amt
	if req.ArbitrationAmount != nil {
		arb := opt.ArbitrationAmount.Add(*req.ArbitrationAmount)
		opt.ArbitrationAmount = &arb
	}
	proposal.Options[req.OptionID] = opt

	// Recompute vote status and winner
	*proposal = setProposalVoteStatus(*proposal, getParams())

	setProposal(uint64(req.ProposalID), *proposal)

	// Emit vote event mirroring AS
	ev := wasmx.Event{
		Type: gov.EventTypeProposalVote,
		Attributes: []wasmx.EventAttribute{
			{Key: gov.AttributeKeyVoter, Value: string(req.Voter), Index: true},
			{Key: AttributeKeyOptionID, Value: utils.Itoa(int(req.OptionID)), Index: true},
			{Key: gov.AttributeKeyProposalID, Value: req.ProposalID.ToString(), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	tryExecuteProposal(proposal)

	return []byte{}
}

// Query functions

// GetProposal returns a single proposal
func GetProposal(req QueryProposalRequest) []byte {
	proposal, exists := getProposal(uint64(req.ProposalID))
	if !exists || proposal == nil {
		return []byte("{\"proposal\":null}")
	}

	govprop := proposalToExternal(*proposal)
	resp := &gov.QueryProposalResponse{Proposal: &govprop}
	result, _ := json.Marshal(resp)
	return result
}

// GetProposals returns multiple proposals
func GetProposals(req QueryProposalsRequest) []byte {
	// TODO: Implement pagination and filtering
	var proposals []Proposal
	total := uint64(0)
	first := gov.GetProposalIdFirst()
	count := gov.GetProposalIdCount()
	for i := uint64(0); i < uint64(count); i++ {
		id := uint64(first) + i
		if p, ok := getProposal(id); ok && p != nil {
			proposals = append(proposals, *p)
			total++
		}
	}
	resp := QueryProposalsExtendedResponse{Proposals: proposals, Pagination: PageResponse{Total: utils.StringUint64(total)}}
	result, _ := json.Marshal(resp)
	return result
}

// GetProposalExtended returns an extended proposal
func GetProposalExtended(req QueryProposalRequest) []byte {
	proposal, exists := getProposal(uint64(req.ProposalID))
	if !exists {
		return []byte("{\"proposal\":null}")
	}

	resp := QueryProposalExtendedResponse{
		Proposal: *proposal,
	}

	result, _ := json.Marshal(resp)
	return result
}

// GetProposalsExtended returns multiple extended proposals
func GetProposalsExtended(req QueryProposalsRequest) []byte {
	// TODO: Implement pagination and filtering
	var proposals []Proposal
	total := uint64(0)
	first := gov.GetProposalIdFirst()
	count := gov.GetProposalIdCount()
	for i := uint64(0); i < uint64(count); i++ {
		id := uint64(first) + i
		if p, ok := getProposal(id); ok && p != nil {
			proposals = append(proposals, *p)
			total++
		}
	}
	resp := QueryProposalsExtendedResponse{Proposals: proposals, Pagination: PageResponse{Total: utils.StringUint64(total)}}
	result, _ := json.Marshal(resp)
	return result
}

// GetVote stub to mirror AS presence
func GetVote(req QueryVoteRequest) []byte {
	count := gov.GetProposalVoteCount(uint64(req.ProposalID))
	var found *DepositVote
	for i := uint64(0); i < count; i++ {
		if v, ok := getProposalVote(uint64(req.ProposalID), i); ok && v != nil {
			if v.Voter == req.Voter {
				found = v
				break
			}
		}
	}
	resp := struct {
		Vote *DepositVote `json:"vote"`
	}{Vote: found}
	bz, _ := json.Marshal(resp)
	return bz
}

func GetVotes(req QueryVotesRequest) []byte {
	count := gov.GetProposalVoteCount(uint64(req.ProposalID))
	votes := make([]DepositVote, 0, count)
	for i := uint64(0); i < count; i++ {
		if v, ok := getProposalVote(uint64(req.ProposalID), i); ok && v != nil {
			votes = append(votes, *v)
		}
	}
	resp := struct {
		Votes      []DepositVote `json:"votes"`
		Pagination PageResponse  `json:"pagination"`
	}{Votes: votes, Pagination: PageResponse{Total: utils.StringUint64(uint64(len(votes)))}}
	bz, _ := json.Marshal(resp)
	return bz
}

func GetDeposit(req QueryDepositRequest) []byte {
	// scan deposits for depositor match
	var found *gov.Deposit
	count := getProposalDepositCount(uint64(req.ProposalID))
	for i := uint64(0); i < count; i++ {
		if d, ok := getProposalDeposit(uint64(req.ProposalID), i); ok && d != nil {
			if d.Depositor == req.Depositor {
				found = d
				break
			}
		}
	}
	resp := struct {
		Deposit *gov.Deposit `json:"deposit"`
	}{Deposit: found}
	bz, _ := json.Marshal(resp)
	return bz
}

func GetDeposits(req QueryDepositsRequest) []byte {
	// list all deposits for proposal
	count := getProposalDepositCount(uint64(req.ProposalID))
	deps := make([]gov.Deposit, 0, count)
	for i := uint64(0); i < count; i++ {
		if d, ok := getProposalDeposit(uint64(req.ProposalID), i); ok && d != nil {
			deps = append(deps, *d)
		}
	}
	resp := struct {
		Deposits   []gov.Deposit `json:"deposits"`
		Pagination PageResponse  `json:"pagination"`
	}{Deposits: deps, Pagination: PageResponse{Total: utils.StringUint64(uint64(len(deps)))}}
	bz, _ := json.Marshal(resp)
	return bz
}

// GetTallyResult returns tally results for a proposal
func GetTallyResult(req QueryTallyResultRequest) []byte {
	proposal, exists := getProposal(uint64(req.ProposalID))
	if !exists {
		return []byte("{\"tally\":null}")
	}

	// Calculate tally from options
	totalVotes := sdkmath.ZeroInt()
	for _, option := range proposal.Options {
		totalVotes = totalVotes.Add(*option.Weight)
	}

	resp := struct {
		Tally struct {
			YesCount        string `json:"yes_count"`
			AbstainCount    string `json:"abstain_count"`
			NoCount         string `json:"no_count"`
			NoWithVetoCount string `json:"no_with_veto_count"`
		} `json:"tally"`
	}{
		Tally: struct {
			YesCount        string `json:"yes_count"`
			AbstainCount    string `json:"abstain_count"`
			NoCount         string `json:"no_count"`
			NoWithVetoCount string `json:"no_with_veto_count"`
		}{
			YesCount:        totalVotes.String(),
			AbstainCount:    "0",
			NoCount:         "0",
			NoWithVetoCount: "0",
		},
	}

	result, _ := json.Marshal(resp)
	return result
}

// GetNextWinnerThreshold calculates the threshold for the next winner
func GetNextWinnerThreshold(req QueryNextWinnerThreshold) []byte {
	proposal, exists := getProposal(uint64(req.ProposalID))
	if !exists {
		Revert("proposal not found: " + req.ProposalID.ToString())
	}

	weight := sdkmath.ZeroInt()
	params := getParams()
	normalizedWeights := normalizeTally(*proposal, params)

	if len(normalizedWeights) > 0 {
		// Find max weight index
		index := getMaxFromArray(normalizedWeights)
		highestWeight := normalizedWeights[index]

		// Calculate threshold: highestWeight * proposal.x / proposal.y
		weight = highestWeight.Mul(sdkmath.NewIntFromUint64(proposal.X))
		weight = weight.Quo(sdkmath.NewIntFromUint64(proposal.Y))
	}

	resp := QueryNextWinnerThresholdResponse{
		Weight: &weight,
	}

	result, _ := json.Marshal(resp)
	return result
}

// GetParams returns the module parameters
func GetParams() []byte {
	params := getParams()
	resp := struct {
		Params Params `json:"params"`
	}{
		Params: params,
	}
	result, _ := json.Marshal(resp)
	return result
}

// Helper functions for continuous voting calculations

// normalizeTally calculates normalized weights for all proposal options
func normalizeTally(proposal Proposal, params Params) []sdkmath.Int {
	tally := make([]sdkmath.Int, len(proposal.Options))
	for i, option := range proposal.Options {
		weight := normalizeOptionTally(option, params)
		tally[i] = weight
	}
	return tally
}

// normalizeOptionTally calculates normalized weight for a single option
// Formula: option.amount + option.arbitration_amount * params.coefs[CAL]
func normalizeOptionTally(option ProposalOption, params Params) sdkmath.Int {
	// TODO
	// _WL + _AL * tasks[taskid].amount * coefs[uint256(Coefs.cAL)] / (10 ** decimals);
	// Get CAL coefficient from params
	calCoef := sdkmath.NewIntFromUint64(params.Coefs[CAL])

	arb := sdkmath.ZeroInt()
	if option.ArbitrationAmount != nil {
		arb = *option.ArbitrationAmount
	}
	amount := sdkmath.ZeroInt()
	if option.Amount != nil {
		amount = *option.Amount
	}

	// Calculate: amount + arbitration_amount * cAL
	arbitrationWeight := arb.Mul(calCoef)
	return amount.Add(arbitrationWeight)
}

// getMaxFromArray finds the index of the maximum value in a big.Int array
func getMaxFromArray(arr []sdkmath.Int) int {
	index := 0
	for i := 1; i < len(arr); i++ {
		if arr[i].BigInt().Cmp(arr[index].BigInt()) > 0 {
			index = i
		}
	}
	return index
}

// getMaxFromArrayExcept finds the index of the maximum value excluding a specific position
func getMaxFromArrayExcept(arr []*sdkmath.Int, excludePos int) int {
	index := 0
	if index == excludePos {
		index = 1
		if index >= len(arr) {
			return 0
		}
	}

	for i := 0; i < len(arr); i++ {
		if i != excludePos && arr[i].BigInt().Cmp(arr[index].BigInt()) > 0 {
			index = i
		}
	}
	return index
}

// defaultProposalOptions returns the two base options: status quo and unenforceable
func defaultProposalOptions() []ProposalOption {
	thisaddr := wasmx.GetAddress()
	zero := sdkmath.NewInt(0)
	statusQuo := ProposalOption{
		Proposer:          thisaddr,
		Messages:          []string{},
		Amount:            &zero,
		ArbitrationAmount: &zero,
		Weight:            &zero,
		Title:             "status quo",
		Summary:           "The outcome of this proposal is the current status quo",
		Metadata:          "",
	}
	unenforceable := ProposalOption{
		Proposer:          thisaddr,
		Messages:          []string{},
		Amount:            &zero,
		ArbitrationAmount: &zero,
		Weight:            &zero,
		Title:             "unenforceable",
		Summary:           "This proposal is unenforceable for reasons including but not limited to: being unclear, too broad, already covered by previous proposals, illegal.",
		Metadata:          "",
	}
	return []ProposalOption{statusQuo, unenforceable}
}

// setProposalVoteStatus recomputes weights, updates vote status, and sets winner/changed
func setProposalVoteStatus(proposal Proposal, params Params) Proposal {
	proposal.VoteStatus.Changed = false

	// recompute weights
	normalizedWeights := normalizeTally(proposal, params)
	for i := range normalizedWeights {
		w := normalizedWeights[i]
		proposal.Options[i].Weight = &w
	}

	next := getProposalVoteStatus(proposal, params, normalizedWeights)
	proposal.VoteStatus = next
	winner := getWinner(proposal.Winner, next)
	if proposal.Winner != winner {
		proposal.Winner = winner
		proposal.VoteStatus.Changed = true
	}
	return proposal
}

// getProposalVoteStatus determines xi/yi and overall status
func getProposalVoteStatus(proposal Proposal, params Params, normalizedWeights []sdkmath.Int) ProposalVoteStatus {
	if len(normalizedWeights) == 0 {
		return ProposalVoteStatus{Status: VOTE_STATUS_UNSPECIFIED, Xi: 0, Yi: 0, Changed: false}
	}
	xi := uint32(getMaxFromArray(normalizedWeights))
	// compute yi as second max
	tmp := make([]*sdkmath.Int, len(normalizedWeights))
	for i := range normalizedWeights {
		tmp[i] = &normalizedWeights[i]
	}
	yi := uint32(getMaxFromArrayExcept(tmp, int(xi)))
	status := getVoteStatus(normalizedWeights[xi], normalizedWeights[yi], proposal, params)
	return ProposalVoteStatus{Status: status, Xi: xi, Yi: yi, Changed: false}
}

// getWinner decides which option index is the current winner based on status
func getWinner(prevWinner uint32, nextStatus ProposalVoteStatus) uint32 {
	if nextStatus.Status == VOTE_STATUS_X {
		return nextStatus.Xi
	}
	if nextStatus.Status == VOTE_STATUS_Y {
		return nextStatus.Yi
	}
	return prevWinner
}

// getVoteStatus mirrors the AS logic for deciding region/status
// 0: nobody won
// 1: x option wins (x_threshold <-> x axis)
// 2: y option wins (y_threshold <-> y axis)
// 3: undecidable, leaning towards x (middleline <-> x_threshold)
// 4: undecidable, leaning towards y (middleline <-> y_threshold)
func getVoteStatus(x sdkmath.Int, y sdkmath.Int, p Proposal, params Params) VoteStatus {
	if x.IsZero() && y.IsZero() {
		return VOTE_STATUS_UNSPECIFIED
	}
	if y.IsZero() {
		return VOTE_STATUS_X
	}
	if x.IsZero() {
		return VOTE_STATUS_Y
	}

	precision := sdkmath.NewIntFromUint64(params.Coefs[Precision])
	r1 := sdkmath.NewIntFromUint64(p.X).Mul(precision).Quo(sdkmath.NewIntFromUint64(p.Y))

	// if x*PRECISION / y >= r1 => 1
	left := x.Mul(precision).Quo(y)
	if left.GTE(r1) {
		return VOTE_STATUS_X
	}
	// if y*PRECISION / x >= r1 => 2
	left2 := y.Mul(precision).Quo(x)
	if left2.GTE(r1) {
		return VOTE_STATUS_Y
	}

	// midline == PRECISION
	if x.Mul(precision).Quo(y).GTE(precision) {
		return VOTE_STATUS_Xu
	}
	return VOTE_STATUS_Yu
}

func proposalToExternal(proposal Proposal) gov.Proposal {
	localparams := getParams()
	depositA := sdkmath.NewInt(0)
	deposit := wasmx.Coin{Denom: proposal.Denom, Amount: &depositA}

	arbCoinA := sdkmath.NewInt(0)
	arbCoin := wasmx.Coin{Denom: localparams.ArbitrationDenom, Amount: &arbCoinA}

	// i := 1 to mirror the AssemblyScript loop
	for i := 1; i < len(proposal.Options); i++ {
		opt := proposal.Options[i]
		depositA := deposit.Amount.Add(*opt.Amount)
		deposit = wasmx.NewCoin(deposit.Denom, depositA)
		arbCoinA := arbCoin.Amount.Add(*opt.ArbitrationAmount)
		arbCoin = wasmx.NewCoin(arbCoin.Denom, arbCoinA)
	}

	govprop := gov.Proposal{
		ID:               proposal.ID,
		Messages:         proposal.Options[proposal.Winner].Messages, // ensure type matches gov.Proposal
		Status:           gov.ProposalStatus(proposal.Status),
		FinalTallyResult: tallyToExternal(proposal, localparams),
		SubmitTime:       proposal.SubmitTime,
		DepositEndTime:   proposal.DepositEndTime,
		TotalDeposit:     []wasmx.Coin{deposit, arbCoin},
		VotingStartTime:  proposal.VotingStartTime,
		VotingEndTime:    proposal.VotingEndTime,
		Metadata:         proposal.Metadata,
		Title:            proposal.Title,
		Summary:          proposal.Summary,
		Proposer:         proposal.Proposer,
		// Expedited / IsExpedited: false, // if your gov.Proposal has this bool field
		FailedReason: proposal.FailedReason,
	}

	return govprop
}

func tallyToExternal(proposal Proposal, localparams Params) gov.TallyResult {
	yes := normalizeOptionTally(proposal.Options[proposal.VoteStatus.Xi], localparams)

	no := normalizeOptionTally(proposal.Options[proposal.VoteStatus.Yi], localparams)

	abstaincount := sdkmath.ZeroInt()
	noveto := sdkmath.ZeroInt()

	return gov.TallyResult{
		YesCount:        &yes,
		AbstainCount:    &abstaincount,
		NoCount:         &no,
		NoWithVetoCount: &noveto,
	}
}

// tryExecuteProposal attempts to execute a proposal if conditions are met
func tryExecuteProposal(proposal *Proposal) {
	LoggerDebug("try execute proposal", []string{"proposal_id", utils.U64toa(uint64(proposal.ID)), "changed", fmt.Sprintf("%t", proposal.VoteStatus.Changed)})
	if !proposal.VoteStatus.Changed {
		return
	}

	params := getParams()
	normalizedWeights := normalizeTally(*proposal, params)
	weightsStr := make([]string, len(normalizedWeights))
	for i, weight := range normalizedWeights {
		weightsStr[i] = weight.String()
	}

	// Emit proposal outcome event
	ev := wasmx.Event{
		Type: EventTypeProposalOutcome,
		Attributes: []wasmx.EventAttribute{
			{Key: gov.AttributeKeyProposalID, Value: utils.U64toa(uint64(proposal.ID)), Index: true},
			{Key: AttributeKeyOptionID, Value: utils.Itoa(int(proposal.Winner)), Index: true},
			{Key: AttributeKeyOptionWeights, Value: strings.Join(weightsStr, ","), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	result := executeProposal(*proposal)
	if result != nil {
		LoggerInfo("execute proposal", []string{
			"success", fmt.Sprintf("%t", result.Success),
			"proposal_id", utils.U64toa(uint64(proposal.ID)),
			"winner", utils.Itoa(int(proposal.Winner)),
			"data", result.Data,
		})

		if !result.Success {
			proposal.FailedReason = result.Data
		} else {
			proposal.FailedReason = ""
			// TODO: when we have a better initial threshold
			// proposal.Status = PROPOSAL_STATUS_PASSED

			// Emit execution success event
			execEv := wasmx.Event{
				Type: EventTypeExecuteProposal,
				Attributes: []wasmx.EventAttribute{
					{Key: gov.AttributeKeyProposalID, Value: utils.U64toa(uint64(proposal.ID)), Index: true},
					{Key: gov.AttributeKeyOption, Value: utils.Itoa(int(proposal.Winner)), Index: true},
				},
			}
			wasmx.EmitCosmosEvents([]wasmx.Event{execEv})
		}
		setProposal(uint64(proposal.ID), *proposal)
	}
}

// executeProposal executes the messages in the winning proposal option
func executeProposal(proposal Proposal) *gov.Response {
	messages := proposal.Options[proposal.Winner].Messages
	for _, msg := range messages {
		// Decode base64 message
		msgBytes, err := base64.StdEncoding.DecodeString(msg)
		if err != nil {
			return &gov.Response{Success: false, Data: "invalid message encoding"}
		}

		response := wasmx.ExecuteCosmosMsg(string(msgBytes), MODULE_NAME)
		if response.Success > 0 {
			return &gov.Response{Success: false, Data: response.Data}
		}
	}
	return &gov.Response{Success: true, Data: ""}
}

// marshalJSON marshals v to a compact JSON string; returns "" on error
func marshalJSON(v interface{}) string {
	bz, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bz)
}

// Simple bank call helper
func callBank(calldata string, isQuery bool) (bool, []byte) {
	return wasmx.CallInternal(wasmx.Bech32String(wasmx.ROLE_BANK), nil, []byte(calldata), big.NewInt(50_000_000), isQuery, MODULE_NAME)
}

// bankSendCoinFromAccountToModule locks coins by calling bank module
func bankSendCoinFromAccountToModule(from wasmx.Bech32String, to wasmx.Bech32String, coins []wasmx.Coin) {
	// Envelope mirrors base gov: {"SendCoinsFromAccountToModule": {"from_address":"...","to_address":"...","amount":[...]}}
	payload := struct {
		Send struct {
			From   string       `json:"from_address"`
			To     string       `json:"to_address"`
			Amount []wasmx.Coin `json:"amount"`
		} `json:"SendCoinsFromAccountToModule"`
	}{}
	payload.Send.From = string(from)
	payload.Send.To = string(to)
	payload.Send.Amount = coins
	bz, _ := json.Marshal(&payload)
	ok, resp := callBank(string(bz), false)
	if !ok {
		Revert("could not transfer coins by bank: " + string(resp))
	}
}
