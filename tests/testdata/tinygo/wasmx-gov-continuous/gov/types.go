package gov

import (
	"math/big"

	wasmx "github.com/loredanacirstea/wasmx-env"
	utils "github.com/loredanacirstea/wasmx-utils"
)

const MODULE_NAME = "gov-cont"
const VERSION = "0.0.1"

// Coefs enum
type Coefs int

const (
	Precision Coefs = iota
	CAL
	MaxWinnerCoef
	MaxWinnerToggle
	MaxWinnerToggleArbiter
	OptionRegisterAmount
	ProposalRatioX
	ProposalRatioY
)

// VoteStatus types
type VoteStatus int32

const (
	VOTE_STATUS_UNSPECIFIED VoteStatus = 0
	VOTE_STATUS_X           VoteStatus = 1 // x won
	VOTE_STATUS_Y           VoteStatus = 2 // y won
	VOTE_STATUS_Xu          VoteStatus = 3 // x undecided
	VOTE_STATUS_Yu          VoteStatus = 4 // y undecided
)

// ProposalVoteStatus represents vote status for continuous voting
type ProposalVoteStatus struct {
	Status  VoteStatus `json:"status"`  // vote status based on xi, yi
	Xi      uint32     `json:"xi"`      // the option index with the highest weight at current time
	Yi      uint32     `json:"yi"`      // the option index with the second highest weight
	Changed bool       `json:"changed"` // last change triggered option execution
}

type ProposalVoteStatusExtended struct {
	Status VoteStatus `json:"status"`
	Xi     uint32     `json:"xi"` // the option index with the highest weight at current time
	Yi     uint32     `json:"yi"` // the option index with the second highest weight
	X      *big.Int   `json:"x"`  // the highest amount
	Y      *big.Int   `json:"y"`  // the second highest amount
}

// ProposalOption represents an option in a continuous voting proposal
type ProposalOption struct {
	Proposer          wasmx.Bech32String `json:"proposer"`
	Messages          []string           `json:"messages"` // base64 encoded messages
	Amount            *big.Int           `json:"amount"`
	ArbitrationAmount *big.Int           `json:"arbitration_amount"`
	Weight            *big.Int           `json:"weight"`
	Title             string             `json:"title"`
	Summary           string             `json:"summary"`
	Metadata          string             `json:"metadata"`
}

// Proposal represents a continuous voting proposal (extends the base gov Proposal)
type Proposal struct {
	// Base gov.Proposal fields
	ID              utils.StringUint64 `json:"id"`
	Status          int32              `json:"status"` // ProposalStatus
	SubmitTime      string             `json:"submit_time"`
	DepositEndTime  string             `json:"deposit_end_time"`
	VotingStartTime string             `json:"voting_start_time"`
	VotingEndTime   string             `json:"voting_end_time"`
	Metadata        string             `json:"metadata"`
	Title           string             `json:"title"`
	Summary         string             `json:"summary"`
	Proposer        wasmx.Bech32String `json:"proposer"`
	FailedReason    string             `json:"failed_reason"`

	// Continuous voting specific fields
	X          uint64             `json:"x"`           // curve parameter
	Y          uint64             `json:"y"`           // curve parameter
	Denom      string             `json:"denom"`       // denomination for voting
	Options    []ProposalOption   `json:"options"`     // voting options
	VoteStatus ProposalVoteStatus `json:"vote_status"` // current vote status
	Winner     uint32             `json:"winner"`      // current winner (may differ from vote_status)
}

// CoefProposal for coefficient proposals
type CoefProposal struct {
	Key   *big.Int `json:"key"`
	Value *big.Int `json:"value"`
}

// DepositVote represents a deposit vote in continuous voting
type DepositVote struct {
	ProposalID        utils.StringUint64 `json:"proposal_id"`
	OptionID          int32              `json:"option_id"`
	Voter             wasmx.Bech32String `json:"voter"`
	Amount            *big.Int           `json:"amount"`
	ArbitrationAmount *big.Int           `json:"arbitration_amount"`
	Metadata          string             `json:"metadata"`
}

// Params for the continuous voting module
type Params struct {
	ArbitrationDenom string   `json:"arbitrationDenom"`
	Coefs            []uint64 `json:"coefs"`
	DefaultX         uint64   `json:"defaultX"`
	DefaultY         uint64   `json:"defaultY"`
}

// Messages

// MsgInitGenesis for genesis initialization
type MsgInitGenesis struct {
	StartingProposalID utils.StringUint64 `json:"starting_proposal_id"`
	Deposits           []Deposit          `json:"deposits"` // from gov
	Votes              []DepositVote      `json:"votes"`    // continuous voting specific
	Proposals          []Proposal         `json:"proposals"`
	Params             Params             `json:"params"` // from gov (extended)
	Constitution       string             `json:"constitution"`
}

// MsgSubmitProposalExtended extends the base SubmitProposal with continuous voting parameters
type MsgSubmitProposalExtended struct {
	// Base MsgSubmitProposal fields
	Messages       []string           `json:"messages"`
	InitialDeposit []wasmx.Coin       `json:"initial_deposit"`
	Proposer       wasmx.Bech32String `json:"proposer"`
	Metadata       string             `json:"metadata"`
	Title          string             `json:"title"`
	Summary        string             `json:"summary"`
	Expedited      bool               `json:"expedited"`

	// Continuous voting extensions
	X              uint64 `json:"x"`
	Y              uint64 `json:"y"`
	OptionTitle    string `json:"optionTitle"`
	OptionSummary  string `json:"optionSummary"`
	OptionMetadata string `json:"optionMetadata"`
}

// MsgAddProposalOption adds a new option to an existing proposal
type MsgAddProposalOption struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Option     ProposalOption     `json:"option"`
}

// Query types

type QueryNextWinnerThreshold struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
}

type QueryNextWinnerThresholdResponse struct {
	Weight *big.Int `json:"weight"`
}

type QueryProposalExtendedResponse struct {
	Proposal Proposal `json:"proposal"`
}

type QueryProposalsExtendedResponse struct {
	Proposals  []Proposal   `json:"proposals"`
	Pagination PageResponse `json:"pagination"`
}

// Re-export types from base gov and wasmx-env that we need
type Coin = wasmx.Coin
type Bech32String = wasmx.Bech32String
type Base64String = wasmx.Base64String

// Base gov types (simplified, we'll import what we need)
type Deposit struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Depositor  Bech32String       `json:"depositor"`
	Amount     []Coin             `json:"amount"`
}

type PageResponse struct {
	Total utils.StringUint64 `json:"total"`
}

// Proposal statuses (from base gov)
const (
	PROPOSAL_STATUS_UNSPECIFIED    = 0
	PROPOSAL_STATUS_DEPOSIT_PERIOD = 1
	PROPOSAL_STATUS_VOTING_PERIOD  = 2
	PROPOSAL_STATUS_PASSED         = 3
	PROPOSAL_STATUS_REJECTED       = 4
	PROPOSAL_STATUS_FAILED         = 5
)

// Event types
const (
	EventTypeAddProposalOption = "add_proposal_option"
	EventTypeExecuteProposal   = "execute_proposal"
	EventTypeProposalOutcome   = "proposal_outcome"
	EventTypeProposalOutcome1  = "proposal_outcome1"
)

// Event attribute keys
const (
	AttributeKeyOptionID      = "proposal_option_id"
	AttributeKeyOptionWeights = "proposal_option_weights"
)

// Calldata structure
type CallData struct {
	// Base gov operations
	InitGenesis    *MsgInitGenesis    `json:"InitGenesis"`
	SubmitProposal *MsgSubmitProposal `json:"SubmitProposal"`
	Vote           *MsgVote           `json:"Vote"`
	VoteWeighted   *MsgVoteWeighted   `json:"VoteWeighted"`
	Deposit        *MsgDeposit        `json:"Deposit"`

	// Continuous voting extensions
	SubmitProposalExtended *MsgSubmitProposalExtended `json:"SubmitProposalExtended"`
	AddProposalOption      *MsgAddProposalOption      `json:"AddProposalOption"`
	DepositVote            *DepositVote               `json:"DepositVote"`

	// Hooks
	BeginBlock *MsgEmpty    `json:"BeginBlock"`
	EndBlock   *MsgEndBlock `json:"EndBlock"`

	// Queries
	GetProposal    *QueryProposalRequest    `json:"GetProposal"`
	GetProposals   *QueryProposalsRequest   `json:"GetProposals"`
	GetVote        *QueryVoteRequest        `json:"GetVote"`
	GetVotes       *QueryVotesRequest       `json:"GetVotes"`
	GetParams      *MsgEmpty                `json:"GetParams"`
	GetDeposit     *QueryDepositRequest     `json:"GetDeposit"`
	GetDeposits    *QueryDepositsRequest    `json:"GetDeposits"`
	GetTallyResult *QueryTallyResultRequest `json:"GetTallyResult"`

	// Extended queries
	GetProposalExtended    *QueryProposalRequest     `json:"GetProposalExtended"`
	GetProposalsExtended   *QueryProposalsRequest    `json:"GetProposalsExtended"`
	GetNextWinnerThreshold *QueryNextWinnerThreshold `json:"GetNextWinnerThreshold"`
}

type MsgEmpty struct{}

// Re-export types we need from base gov (these should be imported from wasmx-gov)
type MsgSubmitProposal struct {
	Messages       []string     `json:"messages"`
	InitialDeposit []Coin       `json:"initial_deposit"`
	Proposer       Bech32String `json:"proposer"`
	Metadata       string       `json:"metadata"`
	Title          string       `json:"title"`
	Summary        string       `json:"summary"`
	Expedited      bool         `json:"expedited"`
}

type MsgVote struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Voter      Bech32String       `json:"voter"`
	Option     string             `json:"option"`
	Metadata   string             `json:"metadata"`
}

type MsgVoteWeighted struct {
	ProposalID utils.StringUint64   `json:"proposal_id"`
	Voter      Bech32String         `json:"voter"`
	Option     []WeightedVoteOption `json:"option"`
	Metadata   string               `json:"metadata"`
}

type MsgDeposit struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Depositor  Bech32String       `json:"depositor"`
	Amount     []Coin             `json:"amount"`
}

type MsgEndBlock struct {
	Data string `json:"data"` // base64
}

type WeightedVoteOption struct {
	Option int32  `json:"option"`
	Weight string `json:"weight"`
}

// Query request types
type QueryProposalRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
}

type QueryProposalsRequest struct {
	ProposalStatus string       `json:"proposal_status"`
	Voter          Bech32String `json:"voter"`
	Depositor      Bech32String `json:"depositor"`
	Pagination     PageRequest  `json:"pagination"`
}

type QueryVoteRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Voter      Bech32String       `json:"voter"`
}

type QueryVotesRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Pagination PageRequest        `json:"pagination"`
}

type QueryParamsRequest struct {
	ParamsType string `json:"params_type"`
}

type QueryDepositRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Depositor  Bech32String       `json:"depositor"`
}

type QueryDepositsRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Pagination PageRequest        `json:"pagination"`
}

type QueryTallyResultRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
}

type PageRequest struct {
	Key        uint8              `json:"key"`
	Offset     utils.StringUint64 `json:"offset"`
	Limit      utils.StringUint64 `json:"limit"`
	CountTotal bool               `json:"count_total"`
	Reverse    bool               `json:"reverse"`
}
