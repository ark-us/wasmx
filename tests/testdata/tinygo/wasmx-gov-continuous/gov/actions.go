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

	// Store votes (continuous voting specific)
	for _, vote := range req.Votes {
		addProposalVote(uint64(vote.ProposalID), vote)
	}

	LoggerInfo("initiated genesis", []string{
		"proposals", utils.Itoa(len(req.Proposals)),
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

	// Transfer deposit from proposer to this module
	// This would typically call a bank module function
	// For now, we'll assume the transfer is handled externally

	var proposalCoin wasmx.Coin
	arbitrationAmount := sdkmath.ZeroInt()

	for _, deposit := range req.InitialDeposit {
		if deposit.Denom == localParams.ArbitrationDenom {
			arbitrationAmount = *deposit.Amount
		} else {
			proposalCoin = deposit
		}
	}

	// Create the first proposal option
	firstOption := ProposalOption{
		Proposer:          req.Proposer,
		Messages:          req.Messages,
		Amount:            proposalCoin.Amount,
		ArbitrationAmount: &arbitrationAmount,
		Weight:            proposalCoin.Amount,
		Title:             req.OptionTitle,
		Summary:           req.OptionSummary,
		Metadata:          req.OptionMetadata,
	}

	// Create proposal
	proposal := Proposal{
		Status:          PROPOSAL_STATUS_VOTING_PERIOD, // Continuous voting starts immediately
		SubmitTime:      now.Format(time.RFC3339Nano),
		DepositEndTime:  depositEndTime.Format(time.RFC3339Nano),
		VotingStartTime: now.Format(time.RFC3339Nano),
		VotingEndTime:   now.Add(365 * 24 * time.Hour).Format(time.RFC3339Nano), // Long voting period
		Metadata:        req.Metadata,
		Title:           req.Title,
		Summary:         req.Summary,
		Proposer:        req.Proposer,
		FailedReason:    "",
		X:               req.X,
		Y:               req.Y,
		Denom:           proposalCoin.Denom,
		Options:         []ProposalOption{firstOption},
		VoteStatus: ProposalVoteStatus{
			Status:  VOTE_STATUS_UNSPECIFIED,
			Xi:      0,
			Yi:      0,
			Changed: false,
		},
		Winner: 0,
	}

	proposalID := addProposal(proposal)

	// Add initial deposit vote
	initialVote := DepositVote{
		ProposalID:        utils.StringUint64(proposalID),
		OptionID:          0, // First option
		Voter:             req.Proposer,
		Amount:            proposalCoin.Amount,
		ArbitrationAmount: &arbitrationAmount,
		Metadata:          "",
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
	setProposal(uint64(req.ProposalID), *proposal)

	// Emit event
	ev := wasmx.Event{
		Type: EventTypeAddProposalOption,
		Attributes: []wasmx.EventAttribute{
			{Key: AttributeKeyOptionID, Value: utils.Itoa(len(proposal.Options) - 1), Index: true},
			{Key: "proposal_id", Value: req.ProposalID.ToString(), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	tryExecuteProposal(proposal)

	return []byte{}
}

// VoteWeighted handles weighted voting (delegated to DoVote for simplicity)
func VoteWeighted(req MsgVoteWeighted) []byte {
	// For continuous voting, we convert weighted vote to simple vote
	// TODO: Implement proper weighted voting logic
	return []byte{}
}

// DoVote handles voting logic
func DoVote(req MsgVote) []byte {
	// TODO: Implement continuous voting logic
	LoggerDebug("vote", []string{"proposal_id", req.ProposalID.ToString(), "option", req.Option})
	return []byte{}
}

// DoDeposit handles deposit logic for continuous voting
func DoDeposit(req MsgDeposit) []byte {
	// TODO: Implement deposit logic for continuous voting
	LoggerDebug("deposit", []string{"proposal_id", req.ProposalID.ToString()})
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

	// Add vote
	addProposalVote(uint64(req.ProposalID), req)

	// Update proposal option weight
	weight := proposal.Options[req.OptionID].Weight.Add(*req.Amount)
	proposal.Options[req.OptionID].Weight = &weight

	// TODO: Update vote status and check for winner changes
	// This would involve complex continuous voting calculations

	setProposal(uint64(req.ProposalID), *proposal)

	// Emit vote event
	ev := wasmx.Event{
		Type: "proposal_vote",
		Attributes: []wasmx.EventAttribute{
			{Key: "voter", Value: string(req.Voter), Index: true},
			{Key: "proposal_id", Value: req.ProposalID.ToString(), Index: true},
			{Key: "option_id", Value: utils.Itoa(int(req.OptionID)), Index: true},
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

	// For now, return empty list
	resp := QueryProposalsExtendedResponse{
		Proposals:  proposals,
		Pagination: PageResponse{Total: utils.StringUint64(0)},
	}

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

	resp := QueryProposalsExtendedResponse{
		Proposals:  proposals,
		Pagination: PageResponse{Total: utils.StringUint64(0)},
	}

	result, _ := json.Marshal(resp)
	return result
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

	// Calculate: amount + arbitration_amount * cAL
	arbitrationWeight := option.ArbitrationAmount.Mul(calCoef)
	return option.Amount.Add(arbitrationWeight)
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

		response := wasmx.ExecuteCosmosMsg(string(msgBytes))
		if response.Success > 0 {
			return &gov.Response{Success: false, Data: response.Data}
		}
	}
	return &gov.Response{Success: true, Data: ""}
}
