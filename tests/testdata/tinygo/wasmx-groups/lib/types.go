package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

const MODULE_NAME = "groups"
const VERSION = "0.0.1"

// =============================================================================
// VOTING PROTOCOL ABSTRACTION
// =============================================================================
// This abstraction allows flexible governance configuration for different use cases.
// Each group can use a different voting protocol, or multiple protocols for different
// types of decisions.

// EligibilityType defines who can vote
type EligibilityType int32

const (
	ELIGIBILITY_OPEN          EligibilityType = 0 // Anyone can vote
	ELIGIBILITY_STAKERS       EligibilityType = 1 // Must have stake in the staking module
	ELIGIBILITY_TOKEN_HOLDERS EligibilityType = 2 // Must hold/deposit specific token
	ELIGIBILITY_GROUP_MEMBERS EligibilityType = 3 // Must be member of a specific group
	ELIGIBILITY_IDENTITY      EligibilityType = 4 // Must have registered identity (user_id)
)

// WeightAlgorithm defines how votes are weighted
type WeightAlgorithm int32

const (
	WEIGHT_EQUAL     WeightAlgorithm = 0 // 1 vote per eligible voter (one person one vote)
	WEIGHT_STAKE     WeightAlgorithm = 1 // Vote weight = staked amount
	WEIGHT_DEPOSIT   WeightAlgorithm = 2 // Vote weight = deposited amount
	WEIGHT_TOKEN     WeightAlgorithm = 3 // Vote weight = token balance
	WEIGHT_QUADRATIC WeightAlgorithm = 4 // Vote weight = sqrt(tokens) for quadratic voting
	WEIGHT_DELEGATED WeightAlgorithm = 5 // Votes can be delegated to others
)

// VoteOptionsType defines the type of voting options available
type VoteOptionsType int32

const (
	OPTIONS_BINARY   VoteOptionsType = 0 // Yes/No only
	OPTIONS_STANDARD VoteOptionsType = 1 // Yes/No/Abstain/NoWithVeto (Cosmos-style)
	OPTIONS_MULTI    VoteOptionsType = 2 // Multiple custom options (continuous voting style)
	OPTIONS_RANKED   VoteOptionsType = 3 // Ranked choice voting
)

// ResultAlgorithm defines how the winning option is determined
type ResultAlgorithm int32

const (
	RESULT_SIMPLE_MAJORITY       ResultAlgorithm = 0 // >50% wins
	RESULT_SUPERMAJORITY         ResultAlgorithm = 1 // Configurable threshold (e.g., 2/3)
	RESULT_THRESHOLD_WITH_QUORUM ResultAlgorithm = 2 // Cosmos-style: quorum + threshold + veto
	RESULT_CONTINUOUS_CURVE      ResultAlgorithm = 3 // Continuous voting with X/Y curve
	RESULT_PLURALITY             ResultAlgorithm = 4 // Highest vote count wins
	RESULT_UNANIMOUS             ResultAlgorithm = 5 // All must agree
)

// TimingType defines when/how voting ends
type TimingType int32

const (
	TIMING_FIXED_PERIODS TimingType = 0 // Deposit period + voting period
	TIMING_DEADLINE      TimingType = 1 // Single deadline
	TIMING_CONTINUOUS    TimingType = 2 // Ongoing until conditions met
	TIMING_BLOCK_COUNT   TimingType = 3 // Number of blocks
)

// VotingProtocol describes a complete voting protocol configuration
type VotingProtocol struct {
	Name        string `json:"name"`        // Human-readable name
	Description string `json:"description"` // Protocol description

	// Core protocol dimensions
	Eligibility     EligibilityType `json:"eligibility"`      // Who can vote
	WeightAlgorithm WeightAlgorithm `json:"weight_algorithm"` // How votes are weighted
	OptionsType     VoteOptionsType `json:"options_type"`     // Type of options
	ResultAlgorithm ResultAlgorithm `json:"result_algorithm"` // How winner is determined
	TimingType      TimingType      `json:"timing_type"`      // Timing rules

	// Eligibility parameters
	EligibilityGroupID string `json:"eligibility_group_id,omitempty"` // For GROUP_MEMBERS
	EligibilityDenom   string `json:"eligibility_denom,omitempty"`    // For TOKEN_HOLDERS/DEPOSIT

	// Weight parameters
	WeightDenom string `json:"weight_denom,omitempty"` // Token denom for weight calculation

	// Result parameters (decimals assumed to be 18)
	Quorum          string `json:"quorum,omitempty"`           // Minimum participation (e.g., "0.334")
	Threshold       string `json:"threshold,omitempty"`        // Winning threshold (e.g., "0.5")
	VetoThreshold   string `json:"veto_threshold,omitempty"`   // For THRESHOLD_WITH_QUORUM
	CurveX          uint64 `json:"curve_x,omitempty"`          // For CONTINUOUS_CURVE
	CurveY          uint64 `json:"curve_y,omitempty"`          // For CONTINUOUS_CURVE
	ArbitrationCoef uint64 `json:"arbitration_coef,omitempty"` // For CONTINUOUS_CURVE

	// Timing parameters (in milliseconds)
	DepositPeriod uint64 `json:"deposit_period,omitempty"` // For FIXED_PERIODS
	VotingPeriod  uint64 `json:"voting_period,omitempty"`  // For FIXED_PERIODS/DEADLINE
}

// VotingProtocolReference references a protocol by address or ID
type VotingProtocolReference struct {
	// The governance contract that implements this protocol
	GovernanceContract wasmx.Bech32String `json:"governance_contract"`
	// Optional: specific protocol ID if contract supports multiple
	ProtocolID string `json:"protocol_id,omitempty"`
}

// =============================================================================
// GROUP TYPES
// =============================================================================

// GroupMember represents a member of a group
type GroupMember struct {
	UserID   string `json:"user_id"`   // Identity from wasmx-identity
	JoinedAt int64  `json:"joined_at"` // Timestamp when joined
	Metadata string `json:"metadata"`  // Optional metadata (e.g., role, title)
}

// Group represents a group of members
type Group struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	CreatedAt   int64                `json:"created_at"`
	UpdatedAt   int64                `json:"updated_at"`
	MemberCount uint64               `json:"member_count"`
	Admins      []string             `json:"admins"` // user_ids that can manage directly (for bootstrap)
	Metadata    string                  `json:"metadata"`
	Protocol    VotingProtocolReference `json:"protocol"` // Governance protocol for membership changes
	TokenDenom  string                  `json:"token_denom,omitempty"`
	Token       wasmx.Bech32String      `json:"token,omitempty"`
}

// GroupConfig stores initialization parameters
type GroupConfig struct {
	IdentityContract wasmx.Bech32String `json:"identity_contract"` // wasmx-identity contract address
}

// =============================================================================
// MESSAGE TYPES
// =============================================================================

// MsgInitGenesis initializes the contract
type MsgInitGenesis struct {
	IdentityContract wasmx.Bech32String `json:"identity_contract"`
	Groups           []GroupGenesis     `json:"groups,omitempty"`
}

// GroupGenesis for genesis initialization
type GroupGenesis struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Admins      []string                `json:"admins"` // user_ids
	Members     []GroupMemberGenesis    `json:"members"`
	Protocol    VotingProtocolReference `json:"protocol"`
	TokenDenom  string                  `json:"token_denom,omitempty"`
	Token       wasmx.Bech32String      `json:"token,omitempty"`
	Metadata    string                  `json:"metadata,omitempty"`
}

// GroupMemberGenesis for genesis member initialization
type GroupMemberGenesis struct {
	UserID   string `json:"user_id"`
	Metadata string `json:"metadata,omitempty"`
}

// MsgCreateGroup creates a new group
type MsgCreateGroup struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Admins      []string                `json:"admins"` // Initial admins (user_ids)
	Protocol    VotingProtocolReference `json:"protocol"`
	TokenDenom  string                  `json:"token_denom,omitempty"`
	Token       wasmx.Bech32String      `json:"token,omitempty"`
	Metadata    string                  `json:"metadata,omitempty"`
}

type MsgCreateGroupResponse struct {
	GroupID string `json:"group_id"`
}

// MsgAddMember adds a member to a group (must be called by governance contract)
type MsgAddMember struct {
	GroupID  string `json:"group_id"`
	UserID   string `json:"user_id"`
	Metadata string `json:"metadata,omitempty"`
}

type MsgAddMemberResponse struct {
	Success bool `json:"success"`
}

// MsgRemoveMember removes a member from a group (must be called by governance contract)
type MsgRemoveMember struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

type MsgRemoveMemberResponse struct {
	Success bool `json:"success"`
}

// MsgUpdateProtocol updates the governance protocol for a group
type MsgUpdateProtocol struct {
	GroupID  string                  `json:"group_id"`
	Protocol VotingProtocolReference `json:"protocol"`
}

type MsgUpdateProtocolResponse struct {
	Success bool `json:"success"`
}

// MsgAddAdmin adds an admin to a group (must be called by governance contract)
type MsgAddAdmin struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

type MsgAddAdminResponse struct {
	Success bool `json:"success"`
}

// MsgRemoveAdmin removes an admin from a group (must be called by governance contract)
type MsgRemoveAdmin struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

type MsgRemoveAdminResponse struct {
	Success bool `json:"success"`
}

// =============================================================================
// GOVERNANCE FORWARDING
// =============================================================================

type MsgSubmitGroupProposal struct {
	GroupID        string       `json:"group_id"`
	Messages       []string     `json:"messages"`
	InitialDeposit []wasmx.Coin `json:"initial_deposit"`
	Metadata       string       `json:"metadata"`
	Title          string       `json:"title"`
	Summary        string       `json:"summary"`
	Expedited      bool         `json:"expedited"`
}

type MsgVoteGroupProposal struct {
	GroupID    string `json:"group_id"`
	ProposalID string `json:"proposal_id"`
	Option     string `json:"option"`
	Metadata   string `json:"metadata"`
}

type GroupWeightedVoteOption struct {
	Option string `json:"option"`
	Weight string `json:"weight"`
}

type MsgVoteGroupProposalWeighted struct {
	GroupID    string                    `json:"group_id"`
	ProposalID string                    `json:"proposal_id"`
	Option     []GroupWeightedVoteOption `json:"option"`
	Metadata   string                    `json:"metadata"`
}

type MsgDepositVoteGroupProposal struct {
	GroupID            string `json:"group_id"`
	ProposalID         string `json:"proposal_id"`
	OptionID           int32  `json:"option_id"`
	Amount             string `json:"amount"`
	ArbitrationAmount  string `json:"arbitration_amount"`
	Metadata           string `json:"metadata"`
}

// =============================================================================
// QUERY TYPES
// =============================================================================

// MsgQueryIsMember checks if a user is a member of a group
type MsgQueryIsMember struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

type QueryIsMemberResponse struct {
	IsMember bool         `json:"is_member"`
	Member   *GroupMember `json:"member,omitempty"` // Present if is_member is true
}

// MsgQueryGetAllMembers returns all members of a group
type MsgQueryGetAllMembers struct {
	GroupID    string `json:"group_id"`
	Pagination PageRequest `json:"pagination,omitempty"`
}

type QueryGetAllMembersResponse struct {
	Members    []GroupMember `json:"members"`
	Pagination PageResponse  `json:"pagination"`
}

// MsgQueryGetGroup returns group info
type MsgQueryGetGroup struct {
	GroupID string `json:"group_id"`
}

type QueryGetGroupResponse struct {
	Group *Group `json:"group"`
}

// MsgQueryGetGroups returns all groups
type MsgQueryGetGroups struct {
	Pagination PageRequest `json:"pagination,omitempty"`
}

type QueryGetGroupsResponse struct {
	Groups     []Group      `json:"groups"`
	Pagination PageResponse `json:"pagination"`
}

// MsgQueryGetVotingProtocol returns the voting protocol for a group
type MsgQueryGetVotingProtocol struct {
	GroupID string `json:"group_id"`
}

type QueryGetVotingProtocolResponse struct {
	Protocol VotingProtocolReference `json:"protocol"`
}

// MsgQueryGetUserGroups returns all groups a user belongs to
type MsgQueryGetUserGroups struct {
	UserID     string `json:"user_id"`
	Pagination PageRequest `json:"pagination,omitempty"`
}

type QueryGetUserGroupsResponse struct {
	GroupIDs   []string     `json:"group_ids"`
	Pagination PageResponse `json:"pagination"`
}

// MsgQueryIsAdmin checks if a user is an admin of a group
type MsgQueryIsAdmin struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

type QueryIsAdminResponse struct {
	IsAdmin bool `json:"is_admin"`
}

// MsgQueryGetVoterPower returns membership and voting power for a voter
type MsgQueryGetVoterPower struct {
	GroupID string `json:"group_id"`
	Voter   string `json:"voter"`
}

type QueryGetVoterPowerResponse struct {
	IsMember bool   `json:"is_member"`
	UserID   string `json:"user_id,omitempty"`
	Power    string `json:"power"`
	Denom    string `json:"denom,omitempty"`
	Token    string `json:"token,omitempty"`
}

// MsgQueryGetConfig returns the contract configuration
type MsgQueryGetConfig struct{}

type QueryGetConfigResponse struct {
	Config GroupConfig `json:"config"`
}

// =============================================================================
// PAGINATION
// =============================================================================

type PageRequest struct {
	Offset utils.StringUint64 `json:"offset,omitempty"`
	Limit  utils.StringUint64 `json:"limit,omitempty"`
}

type PageResponse struct {
	Total utils.StringUint64 `json:"total"`
}

// =============================================================================
// CALLDATA
// =============================================================================

type CallData struct {
	// Initialization
	InitGenesis *MsgInitGenesis `json:"init_genesis,omitempty"`

	// Group management
	CreateGroup    *MsgCreateGroup    `json:"create_group,omitempty"`
	AddMember      *MsgAddMember      `json:"add_member,omitempty"`
	RemoveMember   *MsgRemoveMember   `json:"remove_member,omitempty"`
	UpdateProtocol *MsgUpdateProtocol `json:"update_protocol,omitempty"`
	AddAdmin       *MsgAddAdmin       `json:"add_admin,omitempty"`
	RemoveAdmin    *MsgRemoveAdmin    `json:"remove_admin,omitempty"`
	SubmitGroupProposal       *MsgSubmitGroupProposal       `json:"submit_group_proposal,omitempty"`
	VoteGroupProposal         *MsgVoteGroupProposal         `json:"vote_group_proposal,omitempty"`
	VoteGroupProposalWeighted *MsgVoteGroupProposalWeighted `json:"vote_group_proposal_weighted,omitempty"`
	DepositVoteGroupProposal  *MsgDepositVoteGroupProposal  `json:"deposit_vote_group_proposal,omitempty"`

	// Queries
	QueryIsMember          *MsgQueryIsMember          `json:"query_is_member,omitempty"`
	QueryGetAllMembers     *MsgQueryGetAllMembers     `json:"query_get_all_members,omitempty"`
	QueryGetGroup          *MsgQueryGetGroup          `json:"query_get_group,omitempty"`
	QueryGetGroups         *MsgQueryGetGroups         `json:"query_get_groups,omitempty"`
	QueryGetVotingProtocol *MsgQueryGetVotingProtocol `json:"query_get_voting_protocol,omitempty"`
	QueryGetUserGroups     *MsgQueryGetUserGroups     `json:"query_get_user_groups,omitempty"`
	QueryIsAdmin           *MsgQueryIsAdmin           `json:"query_is_admin,omitempty"`
	QueryGetVoterPower     *MsgQueryGetVoterPower     `json:"query_get_voter_power,omitempty"`
	QueryGetConfig         *MsgQueryGetConfig         `json:"query_get_config,omitempty"`
}
