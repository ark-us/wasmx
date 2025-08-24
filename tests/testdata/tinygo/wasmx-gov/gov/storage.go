package gov

import (
	"encoding/json"
	"strconv"
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env"
	utils "github.com/loredanacirstea/wasmx-utils"
)

const (
	MODULE_NAME = "gov"
)

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

func getProposalVoterKey(proposalID uint64, voter wasmx.Bech32String) string {
	return PROPOSAL_VOTER_KEY + u64toa(proposalID) + SPLIT + string(voter)
}

func getProposalDepositorKey(proposalID uint64, depositor wasmx.Bech32String) string {
	return PROPOSAL_DEPOSITOR_KEY + u64toa(proposalID) + SPLIT + string(depositor)
}

func getProposalVoteKey(proposalID, voteID uint64) string {
	return PROPOSAL_VOTE_KEY + u64toa(proposalID) + SPLIT + u64toa(voteID)
}

func getProposalDepositKey(proposalID, depositID uint64) string {
	return PROPOSAL_DEPOSIT_KEY + u64toa(proposalID) + SPLIT + u64toa(depositID)
}

func nextEndingDepositProposals(end time.Time) []Proposal {
	ids := getActiveDepositProposals()
	if len(ids) == 0 {
		return nil
	}
	i := 0
	props := []Proposal{}
	for ; i < len(ids); i++ {
		id := ids[i]
		p, ok := getProposal(id)
		if !ok {
			continue
		}
		t, _ := time.Parse(time.RFC3339Nano, p.DepositEndTime)
		if !t.After(end) { // t <= end
			props = append(props, p)
			continue
		}
		break
	}
	ids = ids[i:]
	setActiveDepositProposals(ids)
	return props
}

func nextEndingVotingProposals(end time.Time) []Proposal {
	ids := getActiveVotingProposals()
	if len(ids) == 0 {
		return nil
	}
	i := 0
	props := []Proposal{}
	for ; i < len(ids); i++ {
		id := ids[i]
		p, ok := getProposal(id)
		if !ok {
			continue
		}
		t, _ := time.Parse(time.RFC3339Nano, p.VotingEndTime)
		if !t.After(end) { // t <= end
			props = append(props, p)
			continue
		}
		break
	}
	ids = ids[i:]
	setActiveVotingProposals(ids)
	return props
}

func getActiveDepositProposals() []uint64 {
	v := sload(PROPOSAL_ACTIVE_DEPOSIT_KEY)
	if v == "" {
		return nil
	}
	var ids []uint64
	_ = json.Unmarshal([]byte(v), &ids)
	return ids
}

func addActiveDepositProposal(id uint64) {
	ids := getActiveDepositProposals()
	ids = append(ids, id)
	setActiveDepositProposals(ids)
}

func removeActiveDepositProposal(id uint64) {
	ids := getActiveDepositProposals()
	nxt := ids[:0]
	for _, v := range ids {
		if v != id {
			nxt = append(nxt, v)
		}
	}
	setActiveDepositProposals(nxt)
}

func setActiveDepositProposals(ids []uint64) {
	bz, _ := json.Marshal(ids)
	sstore(PROPOSAL_ACTIVE_DEPOSIT_KEY, string(bz))
}

func getActiveVotingProposals() []uint64 {
	v := sload(PROPOSAL_ACTIVE_VOTING_KEY)
	if v == "" {
		return nil
	}
	var ids []uint64
	_ = json.Unmarshal([]byte(v), &ids)
	return ids
}

func addActiveVotingProposal(id uint64) {
	ids := getActiveVotingProposals()
	ids = append(ids, id)
	setActiveVotingProposals(ids)
}

func removeActiveVotingProposal(id uint64) {
	ids := getActiveVotingProposals()
	nxt := ids[:0]
	for _, v := range ids {
		if v != id {
			nxt = append(nxt, v)
		}
	}
	setActiveVotingProposals(nxt)
}

func setActiveVotingProposals(ids []uint64) {
	bz, _ := json.Marshal(ids)
	sstore(PROPOSAL_ACTIVE_VOTING_KEY, string(bz))
}

func getProposalDeposit(proposalID, depositID uint64) (Deposit, bool) {
	v := sload(getProposalDepositKey(proposalID, depositID))
	if v == "" {
		return Deposit{}, false
	}
	var d Deposit
	_ = json.Unmarshal([]byte(v), &d)
	return d, true
}

func setProposalDeposit(proposalID, depositID uint64, d Deposit) {
	bz, _ := json.Marshal(d)
	sstore(getProposalDepositKey(proposalID, depositID), string(bz))
}

func addProposalDeposit(proposalID uint64, d Deposit) uint64 {
	id := getProposalDepositCount(proposalID)
	setProposalDeposit(proposalID, id, d)
	setProposalDepositCount(proposalID, id+1)
	return id
}

func removeProposalDeposits(proposalID uint64) {
	count := getProposalDepositCount(proposalID)
	for i := uint64(0); i < count; i++ {
		sstore(getProposalDepositKey(proposalID, i), "")
	}
	sstore(PROPOSAL_DEPOSIT_COUNT_KEY+u64toa(proposalID), "")
}

func getProposalDepositCount(proposalID uint64) uint64 {
	v := sload(PROPOSAL_DEPOSIT_COUNT_KEY + u64toa(proposalID))
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseUint(v, 10, 64)
	return n
}

func setProposalDepositCount(proposalID uint64, n uint64) {
	sstore(PROPOSAL_DEPOSIT_COUNT_KEY+u64toa(proposalID), u64toa(n))
}

func getProposalVote(proposalID, voteID uint64) (Vote, bool) {
	v := sload(getProposalVoteKey(proposalID, voteID))
	if v == "" {
		return Vote{}, false
	}
	var d Vote
	_ = json.Unmarshal([]byte(v), &d)
	return d, true
}

func setProposalVote(proposalID, voteID uint64, v Vote) {
	bz, _ := json.Marshal(v)
	sstore(getProposalVoteKey(proposalID, voteID), string(bz))
}

func addProposalVote(proposalID uint64, v Vote) uint64 {
	id := getProposalVoteCount(proposalID)
	setProposalVote(proposalID, id, v)
	setProposalVoteCount(proposalID, id+1)
	return id
}

func getProposalVoteCount(proposalID uint64) uint64 {
	v := sload(PROPOSAL_VOTE_COUNT_KEY + u64toa(proposalID))
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseUint(v, 10, 64)
	return n
}

func setProposalVoteCount(proposalID uint64, n uint64) {
	sstore(PROPOSAL_VOTE_COUNT_KEY+u64toa(proposalID), u64toa(n))
}

func getProposal(id uint64) (Proposal, bool) {
	v := sload(PROPOSAL_KEY + u64toa(id))
	if v == "" {
		return Proposal{}, false
	}
	var p Proposal
	_ = json.Unmarshal([]byte(v), &p)
	return p, true
}

func removeProposal(id uint64) {
	sstore(PROPOSAL_KEY+u64toa(id), "")
}

func setProposal(id uint64, p Proposal) {
	bz, _ := json.Marshal(p)
	sstore(PROPOSAL_KEY+u64toa(id), string(bz))
}

func addProposal(p Proposal) uint64 {
	last := getProposalIdLast()
	id := uint64(int64(last) + 1)
	p.ID = utils.StringUint64(id)
	setProposal(id, p)
	setProposalIdLast(int64(id))
	setProposalIdCount(getProposalIdCount() + 1)
	return id
}

func getProposalIdCount() int64 {
	v := sload(PROPOSAL_ID_COUNT_KEY)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}

func setProposalIdCount(n int64) {
	sstore(PROPOSAL_ID_COUNT_KEY, strconv.FormatInt(n, 10))
}

func getProposalIdFirst() uint64 {
	v := sload(PROPOSAL_ID_FIRST_KEY)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseUint(v, 10, 64)
	return n
}

func setProposalIdFirst(n int64) {
	sstore(PROPOSAL_ID_FIRST_KEY, strconv.FormatInt(n, 10))
}

func getProposalIdLast() uint64 {
	v := sload(PROPOSAL_ID_LAST_KEY)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseUint(v, 10, 64)
	return n
}

func setProposalIdLast(n int64) {
	sstore(PROPOSAL_ID_LAST_KEY, strconv.FormatInt(n, 10))
}

func getParams() Params {
	v := sload(PARAM_KEY)
	var p Params
	_ = json.Unmarshal([]byte(v), &p)
	return p
}

func setParams(p Params) {
	bz, _ := json.Marshal(p)
	sstore(PARAM_KEY, string(bz))
}

// storage helpers (string-based)
func sstore(key, value string) {
	wasmx.StorageStore([]byte(key), []byte(value))
}

func sload(key string) string {
	bz := wasmx.StorageLoad([]byte(key))
	return string(bz)
}

// Exported functions for use by other modules like wasmx-gov-continuous

// GetProposalIdLast returns the last proposal ID
func GetProposalIdLast() uint64 {
	return getProposalIdLast()
}

// SetProposalIdLast sets the last proposal ID
func SetProposalIdLast(id int64) {
	setProposalIdLast(id)
}

// GetProposalIdFirst returns the first proposal ID
func GetProposalIdFirst() uint64 {
	return getProposalIdFirst()
}

// SetProposalIdFirst sets the first proposal ID
func SetProposalIdFirst(id int64) {
	setProposalIdFirst(id)
}

// GetProposalIdCount returns the proposal count
func GetProposalIdCount() int64 {
	return getProposalIdCount()
}

// SetProposalIdCount sets the proposal count
func SetProposalIdCount(count int64) {
	setProposalIdCount(count)
}

// GetProposalVoteKey returns the storage key for a proposal vote
func GetProposalVoteKey(proposalID, voteID uint64) string {
	return getProposalVoteKey(proposalID, voteID)
}

// GetProposalVoteCount returns the vote count for a proposal
func GetProposalVoteCount(proposalID uint64) uint64 {
	return getProposalVoteCount(proposalID)
}

// SetProposalVoteCount sets the vote count for a proposal
func SetProposalVoteCount(proposalID uint64, count uint64) {
	setProposalVoteCount(proposalID, count)
}

func SetParams(params Params) {
	setParams(params)
}
