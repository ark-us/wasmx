package wasmx

import "encoding/json"

type GroupInfo struct{}
type GroupMember struct{}
type GroupPolicyInfo struct{}
type Proposal struct{}
type Vote struct{}

type GroupGenesisState struct {
	GroupSeq       uint64            `json:"group_seq"`
	Groups         []GroupInfo       `json:"groups"`
	GroupMembers   []GroupMember     `json:"group_members"`
	GroupPolicySeq uint64            `json:"group_policy_seq"`
	GroupPolicies  []GroupPolicyInfo `json:"group_policies"`
	ProposalSeq    uint64            `json:"proposal_seq"`
	Proposals      []Proposal        `json:"proposals"`
	Votes          []Vote            `json:"votes"`
}

func GetDefaultGroupGenesis() []byte {
	gs := GroupGenesisState{GroupSeq: 0, Groups: []GroupInfo{}, GroupMembers: []GroupMember{}, GroupPolicySeq: 0, GroupPolicies: []GroupPolicyInfo{}, ProposalSeq: 0, Proposals: []Proposal{}, Votes: []Vote{}}
	bz, _ := json.Marshal(&gs)
	return bz
}
