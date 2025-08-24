package gov

import (
	"encoding/json"
	"strconv"

	wasmx "github.com/loredanacirstea/wasmx-env"
	gov "github.com/loredanacirstea/wasmx-gov/gov"
	utils "github.com/loredanacirstea/wasmx-utils"
)

const PARAM_LOCAL_KEY = "local_params"

// Storage keys - reuse the same keys as base gov
const (
	SPLIT                       = "."
	PARAM_KEY                   = "params"
	PROPOSAL_ID_LAST_KEY        = "proposal_id_last"
	PROPOSAL_ID_FIRST_KEY       = "proposal_id_first"
	PROPOSAL_ID_COUNT_KEY       = "proposal_count"
	PROPOSAL_KEY                = "proposal."
	PROPOSAL_VOTE_COUNT_KEY     = "proposal_vote_count."
	PROPOSAL_VOTE_KEY           = "proposal_vote."
	PROPOSAL_VOTER_KEY          = "proposal_voter."
	PROPOSAL_DEPOSIT_COUNT_KEY  = "proposal_deposit_count."
	PROPOSAL_DEPOSIT_KEY        = "proposal_deposit."
	PROPOSAL_DEPOSITOR_KEY      = "proposal_depositer."
	PROPOSAL_ACTIVE_DEPOSIT_KEY = "proposal_active_deposit"
	PROPOSAL_ACTIVE_VOTING_KEY  = "proposal_active_voting"
)

// Proposal vote storage functions for continuous voting (using base gov functions)
func getProposalVote(proposalID, voteID uint64) (*DepositVote, bool) {
	key := gov.GetProposalVoteKey(proposalID, voteID)
	value := sload(key)
	if value == "" {
		return nil, false
	}
	var vote DepositVote
	if err := json.Unmarshal([]byte(value), &vote); err != nil {
		return nil, false
	}
	return &vote, true
}

func setProposalVote(proposalID, voteID uint64, vote DepositVote) {
	key := gov.GetProposalVoteKey(proposalID, voteID)
	data, _ := json.Marshal(vote)
	sstore(key, string(data))
}

func addProposalVote(proposalID uint64, vote DepositVote) uint64 {
	voteID := gov.GetProposalVoteCount(proposalID)
	setProposalVote(proposalID, voteID, vote)
	gov.SetProposalVoteCount(proposalID, voteID+1)
	return voteID
}

// Deposits (reuse base gov keys; types are from base gov)
func getProposalDeposit(proposalID, depositID uint64) (*gov.Deposit, bool) {
	key := PROPOSAL_DEPOSIT_KEY + utils.U64toa(proposalID) + SPLIT + utils.U64toa(depositID)
	value := sload(key)
	if value == "" {
		return nil, false
	}
	var d gov.Deposit
	if err := json.Unmarshal([]byte(value), &d); err != nil {
		return nil, false
	}
	return &d, true
}

func setProposalDeposit(proposalID, depositID uint64, d gov.Deposit) {
	key := PROPOSAL_DEPOSIT_KEY + utils.U64toa(proposalID) + SPLIT + utils.U64toa(depositID)
	data, _ := json.Marshal(d)
	sstore(key, string(data))
}

func getProposalDepositCount(proposalID uint64) uint64 {
	key := PROPOSAL_DEPOSIT_COUNT_KEY + utils.U64toa(proposalID)
	v := sload(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseUint(v, 10, 64)
	return n
}

func setProposalDepositCount(proposalID, count uint64) {
	key := PROPOSAL_DEPOSIT_COUNT_KEY + utils.U64toa(proposalID)
	sstore(key, utils.U64toa(count))
}

func addProposalDeposit(proposalID uint64, d gov.Deposit) uint64 {
	id := getProposalDepositCount(proposalID)
	setProposalDeposit(proposalID, id, d)
	setProposalDepositCount(proposalID, id+1)
	return id
}

// // removeProposalDeposits deletes all deposits for a proposal and resets the counter
// func removeProposalDeposits(proposalID uint64) {
// 	count := getProposalDepositCount(proposalID)
// 	for i := uint64(0); i < count; i++ {
// 		key := PROPOSAL_DEPOSIT_KEY + utils.U64toa(proposalID) + SPLIT + utils.U64toa(i)
// 		sstore(key, "")
// 	}
// 	setProposalDepositCount(proposalID, 0)
// }

// Proposal storage functions (using base gov keys and functions where possible)
func getProposal(id uint64) (*Proposal, bool) {
	key := PROPOSAL_KEY + utils.U64toa(id)
	value := sload(key)
	if value == "" {
		return nil, false
	}
	var proposal Proposal
	if err := json.Unmarshal([]byte(value), &proposal); err != nil {
		return nil, false
	}
	return &proposal, true
}

func removeProposal(id uint64) {
	key := PROPOSAL_KEY + utils.U64toa(id)
	sstore(key, "")
}

func setProposal(id uint64, proposal Proposal) {
	key := PROPOSAL_KEY + utils.U64toa(id)
	data, _ := json.Marshal(proposal)
	sstore(key, string(data))
}

func addProposal(proposal Proposal) uint64 {
	lastID := gov.GetProposalIdLast()
	id := lastID + 1
	proposal.ID = utils.StringUint64(id)
	setProposal(id, proposal)
	gov.SetProposalIdLast(int64(id))
	gov.SetProposalIdCount(gov.GetProposalIdCount() + 1)
	return id
}

// Local params storage (continuous voting specific)
func getParams() Params {
	value := sload(PARAM_LOCAL_KEY)
	if value == "" {
		// Return default params if none exist
		return Params{
			ArbitrationDenom: "arbitration",
			Coefs:            []uint64{1000000, 100, 150, 1, 1, 1000000, 100, 100}, // default coefs
			DefaultX:         100,
			DefaultY:         100,
		}
	}
	var params Params
	json.Unmarshal([]byte(value), &params)
	return params
}

func getParamsInternal() string {
	return sload(PARAM_LOCAL_KEY)
}

func SetParams(params Params) {
	data, _ := json.Marshal(params)
	sstore(PARAM_LOCAL_KEY, string(data))
}

// Storage helper functions
func sstore(key, value string) {
	wasmx.StorageStore([]byte(key), []byte(value))
}

func sload(key string) string {
	bz := wasmx.StorageLoad([]byte(key))
	return string(bz)
}
