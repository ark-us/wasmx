package gov

import (
	"encoding/json"
	"math/big"

	wasmx "github.com/loredanacirstea/wasmx-env"
	utils "github.com/loredanacirstea/wasmx-utils"
)

// Vote options
type VoteOption int32

const (
	VOTE_OPTION_UNSPECIFIED  VoteOption = 0
	VOTE_OPTION_YES          VoteOption = 1
	VOTE_OPTION_ABSTAIN      VoteOption = 2
	VOTE_OPTION_NO           VoteOption = 3
	VOTE_OPTION_NO_WITH_VETO VoteOption = 4
)

const (
	OptionUnspecified = "VOTE_OPTION_UNSPECIFIED"
	OptionYes         = "VOTE_OPTION_YES"
	OptionAbstain     = "VOTE_OPTION_ABSTAIN"
	OptionNo          = "VOTE_OPTION_NO"
	OptionNoVeto      = "VOTE_OPTION_NO_WITH_VETO"
)

var VoteOptionMap = map[string]VoteOption{
	OptionUnspecified: VOTE_OPTION_UNSPECIFIED,
	OptionYes:         VOTE_OPTION_YES,
	OptionAbstain:     VOTE_OPTION_ABSTAIN,
	OptionNo:          VOTE_OPTION_NO,
	OptionNoVeto:      VOTE_OPTION_NO_WITH_VETO,
}

// Proposal status
type ProposalStatus int32

const (
	PROPOSAL_STATUS_UNSPECIFIED    ProposalStatus = 0
	PROPOSAL_STATUS_DEPOSIT_PERIOD ProposalStatus = 1
	PROPOSAL_STATUS_VOTING_PERIOD  ProposalStatus = 2
	PROPOSAL_STATUS_PASSED         ProposalStatus = 3
	PROPOSAL_STATUS_REJECTED       ProposalStatus = 4
	PROPOSAL_STATUS_FAILED         ProposalStatus = 5
)

// Big represents a big integer serialized as string in JSON
type Big struct{ Int *big.Int }

func NewBigZero() Big { return Big{Int: new(big.Int)} }
func NewBigFromString(s string) Big {
	z := new(big.Int)
	_ = z.UnmarshalText([]byte(s))
	return Big{Int: z}
}
func (b Big) String() string {
	if b.Int == nil {
		return "0"
	}
	return b.Int.String()
}
func (b Big) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return json.Marshal("0")
	}
	return json.Marshal(b.Int.String())
}
func (b *Big) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return err
	}
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	_, ok := b.Int.SetString(s, 10)
	if !ok {
		b.Int.SetInt64(0)
	}
	return nil
}
func (b Big) Add(o Big) Big  { r := new(big.Int).Add(b.Int, o.Int); return Big{Int: r} }
func (b Big) Mul(o Big) Big  { r := new(big.Int).Mul(b.Int, o.Int); return Big{Int: r} }
func (b Big) Div(b2 Big) Big { r := new(big.Int).Div(b.Int, b2.Int); return Big{Int: r} }
func (b Big) Cmp(o Big) int  { return b.Int.Cmp(o.Int) }

// Helpers to build scaled integers
func NewBigPow10(dec int) Big {
	r := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	return Big{Int: r}
}

// Coin
type Coin struct {
	Denom  string `json:"denom"`
	Amount Big    `json:"amount"`
}

// WeightedVoteOption
type WeightedVoteOption struct {
	Option VoteOption `json:"option"`
	Weight string     `json:"weight"`
}

// Vote
type Vote struct {
	ProposalID utils.StringUint64   `json:"proposal_id"`
	Voter      wasmx.Bech32String   `json:"voter"`
	Options    []WeightedVoteOption `json:"options"`
	Metadata   string               `json:"metadata"`
}

// TallyResult
type TallyResult struct {
	YesCount        Big `json:"yes_count"`
	AbstainCount    Big `json:"abstain_count"`
	NoCount         Big `json:"no_count"`
	NoWithVetoCount Big `json:"no_with_veto_count"`
}

// Proposal
type Proposal struct {
	ID               utils.StringUint64 `json:"id"`
	Messages         []string           `json:"messages"`
	Status           ProposalStatus     `json:"status"`
	FinalTallyResult TallyResult        `json:"final_tally_result"`
	SubmitTime       string             `json:"submit_time"`
	DepositEndTime   string             `json:"deposit_end_time"`
	TotalDeposit     []Coin             `json:"total_deposit"`
	VotingStartTime  string             `json:"voting_start_time"`
	VotingEndTime    string             `json:"voting_end_time"`
	Metadata         string             `json:"metadata"`
	Title            string             `json:"title"`
	Summary          string             `json:"summary"`
	Proposer         wasmx.Bech32String `json:"proposer"`
	Expedited        bool               `json:"expedited"`
	FailedReason     string             `json:"failed_reason"`
}

// Params
type Params struct {
	MinDeposit                 []Coin             `json:"min_deposit"`
	MaxDepositPeriod           utils.StringUint64 `json:"max_deposit_period"`
	VotingPeriod               utils.StringUint64 `json:"voting_period"`
	Quorum                     string             `json:"quorum"`
	Threshold                  string             `json:"threshold"`
	VetoThreshold              string             `json:"veto_threshold"`
	MinInitialDepositRatio     string             `json:"min_initial_deposit_ratio"`
	ProposalCancelRatio        string             `json:"proposal_cancel_ratio"`
	ProposalCancelDest         wasmx.Bech32String `json:"proposal_cancel_dest"`
	ExpeditedVotingPeriod      utils.StringUint64 `json:"expedited_voting_period"`
	ExpeditedThreshold         string             `json:"expedited_threshold"`
	ExpeditedMinDeposit        []Coin             `json:"expedited_min_deposit"`
	BurnVoteQuorum             bool               `json:"burn_vote_quorum"`
	BurnProposalDepositPrevote bool               `json:"burn_proposal_deposit_prevote"`
	BurnVoteVeto               bool               `json:"burn_vote_veto"`
	MinDepositRatio            string             `json:"min_deposit_ratio"`
}

// Deposits
type Deposit struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Depositor  wasmx.Bech32String `json:"depositor"`
	Amount     []Coin             `json:"amount"`
}

// Genesis
type GenesisState struct {
	StartingProposalID utils.StringUint64 `json:"starting_proposal_id"`
	Deposits           []Deposit          `json:"deposits"`
	Votes              []Vote             `json:"votes"`
	Proposals          []Proposal         `json:"proposals"`
	Params             Params             `json:"params"`
	Constitution       string             `json:"constitution"`
}

// Messages and responses
type MsgSubmitProposal struct {
	Messages       []string           `json:"messages"`
	InitialDeposit []Coin             `json:"initial_deposit"`
	Proposer       wasmx.Bech32String `json:"proposer"`
	Metadata       string             `json:"metadata"`
	Title          string             `json:"title"`
	Summary        string             `json:"summary"`
	Expedited      bool               `json:"expedited"`
}

type MsgSubmitProposalResponse struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
}

type MsgVote struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Voter      wasmx.Bech32String `json:"voter"`
	Option     string             `json:"option"`
	Metadata   string             `json:"metadata"`
}

type MsgVoteResponse struct{}

type MsgVoteWeighted struct {
	ProposalID utils.StringUint64   `json:"proposal_id"`
	Voter      wasmx.Bech32String   `json:"voter"`
	Option     []WeightedVoteOption `json:"option"`
	Metadata   string               `json:"metadata"`
}

type MsgDeposit struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Depositor  wasmx.Bech32String `json:"depositor"`
	Amount     []Coin             `json:"amount"`
}

type MsgDepositResponse struct{}

type MsgEndBlock struct {
	Data string `json:"data"` // base64
}

// Queries
type PageRequest struct {
	Key        uint8              `json:"key"`
	Offset     utils.StringUint64 `json:"offset"`
	Limit      utils.StringUint64 `json:"limit"`
	CountTotal bool               `json:"count_total"`
	Reverse    bool               `json:"reverse"`
}

type PageResponse struct {
	Total utils.StringUint64 `json:"total"`
}

type QueryProposalRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
}

type QueryProposalResponse struct {
	Proposal *Proposal `json:"proposal"`
}

type QueryProposalsRequest struct {
	ProposalStatus string             `json:"proposal_status"`
	Voter          wasmx.Bech32String `json:"voter"`
	Depositor      wasmx.Bech32String `json:"depositor"`
	Pagination     PageRequest        `json:"pagination"`
}

type QueryProposalsResponse struct {
	Proposals  []Proposal   `json:"proposals"`
	Pagination PageResponse `json:"pagination"`
}

type QueryVoteRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Voter      wasmx.Bech32String `json:"voter"`
}

type QueryVotesRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Pagination PageRequest        `json:"pagination"`
}

type QueryParamsRequest struct {
	ParamsType string `json:"params_type"`
}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryDepositRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Depositor  wasmx.Bech32String `json:"depositor"`
}

type QueryDepositsRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
	Pagination PageRequest        `json:"pagination"`
}

type QueryTallyResultRequest struct {
	ProposalID utils.StringUint64 `json:"proposal_id"`
}

type QueryTallyResultResponse struct {
	Tally TallyResult `json:"tally"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
}

type ResponseBz struct {
	Success bool   `json:"success"`
	Data    []byte `json:"data"`
}
