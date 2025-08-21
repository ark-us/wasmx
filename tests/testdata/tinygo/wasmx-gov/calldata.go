package main

type MsgEmpty struct{}

// Calldata mirrors the AssemblyScript CallData union
type Calldata struct {
	InitGenesis    *GenesisState      `json:"InitGenesis,omitempty"`
	SubmitProposal *MsgSubmitProposal `json:"SubmitProposal,omitempty"`
	Vote           *MsgVote           `json:"Vote,omitempty"`
	VoteWeighted   *MsgVoteWeighted   `json:"VoteWeighted,omitempty"`
	Deposit        *MsgDeposit        `json:"Deposit,omitempty"`

	// hooks
	BeginBlock *MsgEmpty    `json:"BeginBlock,omitempty"`
	EndBlock   *MsgEndBlock `json:"EndBlock,omitempty"`

	// queries
	GetProposal    *QueryProposalRequest    `json:"GetProposal,omitempty"`
	GetProposals   *QueryProposalsRequest   `json:"GetProposals,omitempty"`
	GetVote        *QueryVoteRequest        `json:"GetVote,omitempty"`
	GetVotes       *QueryVotesRequest       `json:"GetVotes,omitempty"`
	GetParams      *QueryParamsRequest      `json:"GetParams,omitempty"`
	GetDeposit     *QueryDepositRequest     `json:"GetDeposit,omitempty"`
	GetDeposits    *QueryDepositsRequest    `json:"GetDeposits,omitempty"`
	GetTallyResult *QueryTallyResultRequest `json:"GetTallyResult,omitempty"`
}
