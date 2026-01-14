package gov

// Governance event types and attribute keys (mirrors AssemblyScript events.ts)
const (
	EventTypeSubmitProposal      = "submit_proposal"
	AttributeKeyProposalID       = "proposal_id"
	AttributeKeyProposalMessages = "proposal_messages"

	EventTypeProposalDeposit      = "proposal_deposit"
	AttributeKeyVotingPeriodStart = "voting_period_start"

	EventTypeProposalVote = "proposal_vote"
	AttributeKeyVoter     = "voter"
	AttributeKeyOption    = "option"

	EventTypeCancelProposal = "cancel_proposal"
	AttributeKeySender      = "sender"
)
