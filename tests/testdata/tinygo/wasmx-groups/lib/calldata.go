package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// RouteCallData routes incoming call data to the appropriate handler
func RouteCallData(data []byte) []byte {
	LoggerDebug("RouteCallData", []string{"data", string(data)})

	var callData CallData
	if err := json.Unmarshal(data, &callData); err != nil {
		LoggerError("Failed to unmarshal calldata", []string{"error", err.Error()})
		Revert("invalid calldata: " + err.Error())
	}

	// Handle initialization
	if callData.InitGenesis != nil {
		return InitGenesis(callData.InitGenesis)
	}

	// Handle group management
	if callData.CreateGroup != nil {
		return CreateGroup(callData.CreateGroup)
	}

	// Handle protocol management
	if callData.UpdateProtocol != nil {
		return UpdateProtocol(callData.UpdateProtocol)
	}

	// Handle admin management
	if callData.AddAdmin != nil {
		return AddAdmin(callData.AddAdmin)
	}

	if callData.RemoveAdmin != nil {
		return RemoveAdmin(callData.RemoveAdmin)
	}

	if callData.SubmitGroupProposal != nil {
		return SubmitGroupProposal(callData.SubmitGroupProposal)
	}

	if callData.SubmitGroupProposalExecute != nil {
		return SubmitGroupProposalExecute(callData.SubmitGroupProposalExecute)
	}

	if callData.VoteGroupProposal != nil {
		return VoteGroupProposal(callData.VoteGroupProposal)
	}

	if callData.VoteGroupProposalWeighted != nil {
		return VoteGroupProposalWeighted(callData.VoteGroupProposalWeighted)
	}

	if callData.DepositVoteGroupProposal != nil {
		return DepositVoteGroupProposal(callData.DepositVoteGroupProposal)
	}

	// Handle queries
	if callData.QueryIsMember != nil {
		return QueryIsMember(callData.QueryIsMember)
	}

	if callData.QueryGetGroup != nil {
		return QueryGetGroup(callData.QueryGetGroup)
	}

	if callData.QueryGetGroups != nil {
		return QueryGetGroups(callData.QueryGetGroups)
	}

	if callData.QueryGetVotingProtocol != nil {
		return QueryGetVotingProtocol(callData.QueryGetVotingProtocol)
	}

	if callData.QueryGetUserGroups != nil {
		return QueryGetUserGroups(callData.QueryGetUserGroups)
	}

	if callData.QueryIsAdmin != nil {
		return QueryIsAdmin(callData.QueryIsAdmin)
	}

	if callData.QueryGetVoterPower != nil {
		return QueryGetVoterPower(callData.QueryGetVoterPower)
	}

	if callData.QueryGetConfig != nil {
		return QueryGetConfig(callData.QueryGetConfig)
	}

	Revert("invalid call data: " + string(data))
	return []byte{}
}

// Execute is the main entry point for the contract
func Execute(data []byte) []byte {
	return RouteCallData(data)
}

// Query is the entry point for read-only operations
func Query(data []byte) []byte {
	return RouteCallData(data)
}

// OnRoleChanged is called when the contract is assigned a role
func OnRoleChanged() []byte {
	callData := wasmx.GetCallData()
	if len(callData) == 0 {
		return MarshalJSON(map[string]bool{"success": true})
	}
	var hook struct {
		RoleChanged *wasmx.RolesChangedHook `json:"RoleChanged"`
	}
	if err := json.Unmarshal(callData, &hook); err != nil {
		LoggerError("Failed to unmarshal RoleChanged hook", []string{"error", err.Error()})
		return MarshalJSON(map[string]bool{"success": true})
	}
	if hook.RoleChanged == nil || hook.RoleChanged.RoleChanged == nil {
		return MarshalJSON(map[string]bool{"success": true})
	}
	role := hook.RoleChanged.RoleChanged.Role
	LoggerInfo("Role changed", []string{"role", role, "label", hook.RoleChanged.RoleChanged.Label})

	// Groups contract can be notified when assigned a role
	if role == "groups" {
		LoggerInfo("Groups contract role assigned", nil)
	}

	return MarshalJSON(map[string]bool{"success": true})
}
