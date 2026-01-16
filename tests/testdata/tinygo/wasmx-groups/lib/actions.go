package lib

import (
	"encoding/json"
	"strings"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

// =============================================================================
// INITIALIZATION
// =============================================================================

func InitGenesis(msg *MsgInitGenesis) []byte {
	LoggerInfo("InitGenesis called", nil)

	// Save config
	config := &GroupConfig{}
	if err := SaveConfig(config); err != nil {
		Revert("failed to save config: " + err.Error())
	}

	// Create initial groups
	if msg.Groups != nil {
		for _, groupGenesis := range msg.Groups {
			groupID := GetNextGroupID()
			timestamp := GetBlockTime()

			group := &Group{
				ID:          groupID,
				Name:        groupGenesis.Name,
				Description: groupGenesis.Description,
				CreatedAt:   timestamp,
				UpdatedAt:   timestamp,
				Admins:      groupGenesis.Admins,
				Metadata:    groupGenesis.Metadata,
				Protocol:    groupGenesis.Protocol,
				Token:       groupGenesis.Token,
				MinBalance:  groupGenesis.MinBalance,
			}
			if group.Token == "" {
				Revert("group token required")
			}
			if symbol, ok := getTokenSymbol(group.Token); ok {
				group.TokenDenom = symbol
			} else {
				Revert("failed to query token symbol")
			}
			if group.MinBalance == "" {
				group.MinBalance = "1"
			}

			if err := SaveGroup(group); err != nil {
				LoggerError("Failed to save group in genesis", []string{"error", err.Error()})
				continue
			}

			LoggerInfo("Group created in genesis", []string{
				"group_id", groupID,
				"name", group.Name,
				"token", string(group.Token),
			})
		}
	}

	return MarshalJSON(map[string]bool{"success": true})
}

// =============================================================================
// GROUP MANAGEMENT
// =============================================================================

// CreateGroup creates a new group
func CreateGroup(msg *MsgCreateGroup) []byte {
	LoggerInfo("CreateGroup called", []string{"name", msg.Name})

	config := LoadConfig()
	if config == nil {
		Revert("contract not initialized")
	}

	// Verify caller has identity
	callerUserID := GetCallerUserID(config)
	if callerUserID == "" {
		Revert("caller does not have a registered identity")
	}

	timestamp := GetBlockTime()
	groupID := GetNextGroupID()

	// If no admins specified, use the creator
	admins := msg.Admins
	if len(admins) == 0 {
		admins = []string{callerUserID}
	}

	group := &Group{
		ID:          groupID,
		Name:        msg.Name,
		Description: msg.Description,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		Admins:      admins,
		Metadata:    msg.Metadata,
		Protocol:    msg.Protocol,
		Token:       msg.Token,
		MinBalance:  msg.MinBalance,
	}
	if group.Token == "" {
		Revert("group token required")
	}
	if symbol, ok := getTokenSymbol(group.Token); ok {
		group.TokenDenom = symbol
	} else {
		Revert("failed to query token symbol")
	}
	if group.MinBalance == "" {
		group.MinBalance = "1"
	}

	if err := SaveGroup(group); err != nil {
		Revert("failed to save group: " + err.Error())
	}

	// Emit event
	ev := wasmx.Event{
		Type: "create_group",
		Attributes: []wasmx.EventAttribute{
			{Key: "group_id", Value: groupID, Index: true},
			{Key: "name", Value: msg.Name, Index: true},
			{Key: "creator", Value: callerUserID, Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	LoggerInfo("Group created", []string{"group_id", groupID, "name", msg.Name})

	return MarshalJSON(MsgCreateGroupResponse{GroupID: groupID})
}

// =============================================================================
// PROTOCOL MANAGEMENT
// =============================================================================

// UpdateProtocol updates the governance protocol for a group
// Must be called by the CURRENT governance contract
func UpdateProtocol(msg *MsgUpdateProtocol) []byte {
	LoggerInfo("UpdateProtocol called", []string{"group_id", msg.GroupID})

	config := LoadConfig()
	if config == nil {
		Revert("contract not initialized")
	}

	group, err := LoadGroup(msg.GroupID)
	if err != nil || group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	// Authorization: must be current governance contract
	if !IsCallerGovernanceContract(group) {
		Revert("unauthorized: must be current governance contract")
	}

	oldProtocol := group.Protocol.GovernanceContract
	group.Protocol = msg.Protocol
	group.UpdatedAt = GetBlockTime()

	if err := SaveGroup(group); err != nil {
		Revert("failed to save group: " + err.Error())
	}

	// Emit event
	ev := wasmx.Event{
		Type: "update_protocol",
		Attributes: []wasmx.EventAttribute{
			{Key: "group_id", Value: msg.GroupID, Index: true},
			{Key: "old_governance", Value: string(oldProtocol), Index: true},
			{Key: "new_governance", Value: string(msg.Protocol.GovernanceContract), Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	LoggerInfo("Protocol updated", []string{
		"group_id", msg.GroupID,
		"new_governance", string(msg.Protocol.GovernanceContract),
	})

	return MarshalJSON(MsgUpdateProtocolResponse{Success: true})
}

// =============================================================================
// ADMIN MANAGEMENT
// =============================================================================

// AddAdmin adds an admin to a group
// Must be called by the governance contract
func AddAdmin(msg *MsgAddAdmin) []byte {
	LoggerInfo("AddAdmin called", []string{"group_id", msg.GroupID, "user_id", msg.UserID})

	config := LoadConfig()
	if config == nil {
		Revert("contract not initialized")
	}

	group, err := LoadGroup(msg.GroupID)
	if err != nil || group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	// Authorization: must be governance contract
	if !IsCallerGovernanceContract(group) {
		Revert("unauthorized: must be governance contract")
	}

	// Check if already admin
	if StringInList(msg.UserID, group.Admins) {
		Revert("user is already an admin: " + msg.UserID)
	}

	group.Admins = append(group.Admins, msg.UserID)
	group.UpdatedAt = GetBlockTime()

	if err := SaveGroup(group); err != nil {
		Revert("failed to save group: " + err.Error())
	}

	// Emit event
	ev := wasmx.Event{
		Type: "add_admin",
		Attributes: []wasmx.EventAttribute{
			{Key: "group_id", Value: msg.GroupID, Index: true},
			{Key: "user_id", Value: msg.UserID, Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	return MarshalJSON(MsgAddAdminResponse{Success: true})
}

// RemoveAdmin removes an admin from a group
// Must be called by the governance contract
func RemoveAdmin(msg *MsgRemoveAdmin) []byte {
	LoggerInfo("RemoveAdmin called", []string{"group_id", msg.GroupID, "user_id", msg.UserID})

	config := LoadConfig()
	if config == nil {
		Revert("contract not initialized")
	}

	group, err := LoadGroup(msg.GroupID)
	if err != nil || group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	// Authorization: must be governance contract
	if !IsCallerGovernanceContract(group) {
		Revert("unauthorized: must be governance contract")
	}

	// Check if is admin
	if !StringInList(msg.UserID, group.Admins) {
		Revert("user is not an admin: " + msg.UserID)
	}

	group.Admins = RemoveStringFromList(msg.UserID, group.Admins)
	group.UpdatedAt = GetBlockTime()

	if err := SaveGroup(group); err != nil {
		Revert("failed to save group: " + err.Error())
	}

	// Emit event
	ev := wasmx.Event{
		Type: "remove_admin",
		Attributes: []wasmx.EventAttribute{
			{Key: "group_id", Value: msg.GroupID, Index: true},
			{Key: "user_id", Value: msg.UserID, Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	return MarshalJSON(MsgRemoveAdminResponse{Success: true})
}

// =============================================================================
// GOVERNANCE FORWARDING
// =============================================================================

func SubmitGroupProposal(msg *MsgSubmitGroupProposal) []byte {
	config := LoadConfig()
	if config == nil {
		Revert("contract not initialized")
	}

	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	caller := wasmx.GetCaller()
	respBz := QueryGetVoterPower(&MsgQueryGetVoterPower{GroupID: msg.GroupID, Voter: string(caller)})
	var powerResp QueryGetVoterPowerResponse
	if err := json.Unmarshal(respBz, &powerResp); err != nil || !powerResp.IsMember {
		Revert("caller is not a group member")
	}

	payload := map[string]interface{}{
		"SubmitProposal": map[string]interface{}{
			"messages":        msg.Messages,
			"initial_deposit": msg.InitialDeposit,
			"proposer":        string(caller),
			"metadata":        msg.Metadata,
			"title":           msg.Title,
			"summary":         msg.Summary,
			"expedited":       msg.Expedited,
			"group_id":        msg.GroupID,
			"group_contract":  string(group.Protocol.GovernanceContract),
		},
	}
	bz, _ := json.Marshal(payload)
	ok, resp := CallInternal(group.Protocol.GovernanceContract, bz, false)
	if !ok {
		Revert("gov submit failed: " + string(resp))
	}
	return resp
}

func VoteGroupProposal(msg *MsgVoteGroupProposal) []byte {
	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	caller := wasmx.GetCaller()
	respBz := QueryGetVoterPower(&MsgQueryGetVoterPower{GroupID: msg.GroupID, Voter: string(caller)})
	var powerResp QueryGetVoterPowerResponse
	if err := json.Unmarshal(respBz, &powerResp); err != nil || !powerResp.IsMember {
		Revert("caller is not a group member")
	}

	payload := map[string]interface{}{
		"Vote": map[string]interface{}{
			"proposal_id": msg.ProposalID,
			"voter":       string(caller),
			"option":      msg.Option,
			"metadata":    msg.Metadata,
		},
	}
	bz, _ := json.Marshal(payload)
	ok, resp := CallInternal(group.Protocol.GovernanceContract, bz, false)
	if !ok {
		Revert("gov vote failed: " + string(resp))
	}
	return resp
}

func VoteGroupProposalWeighted(msg *MsgVoteGroupProposalWeighted) []byte {
	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	caller := wasmx.GetCaller()
	respBz := QueryGetVoterPower(&MsgQueryGetVoterPower{GroupID: msg.GroupID, Voter: string(caller)})
	var powerResp QueryGetVoterPowerResponse
	if err := json.Unmarshal(respBz, &powerResp); err != nil || !powerResp.IsMember {
		Revert("caller is not a group member")
	}

	payload := map[string]interface{}{
		"VoteWeighted": map[string]interface{}{
			"proposal_id": msg.ProposalID,
			"voter":       string(caller),
			"option":      msg.Option,
			"metadata":    msg.Metadata,
		},
	}
	bz, _ := json.Marshal(payload)
	ok, resp := CallInternal(group.Protocol.GovernanceContract, bz, false)
	if !ok {
		Revert("gov vote weighted failed: " + string(resp))
	}
	return resp
}

func DepositVoteGroupProposal(msg *MsgDepositVoteGroupProposal) []byte {
	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	caller := wasmx.GetCaller()
	respBz := QueryGetVoterPower(&MsgQueryGetVoterPower{GroupID: msg.GroupID, Voter: string(caller)})
	var powerResp QueryGetVoterPowerResponse
	if err := json.Unmarshal(respBz, &powerResp); err != nil || !powerResp.IsMember {
		Revert("caller is not a group member")
	}

	payload := map[string]interface{}{
		"DepositVote": map[string]interface{}{
			"proposal_id":        msg.ProposalID,
			"option_id":          msg.OptionID,
			"voter":              string(caller),
			"amount":             msg.Amount,
			"arbitration_amount": msg.ArbitrationAmount,
			"metadata":           msg.Metadata,
		},
	}
	bz, _ := json.Marshal(payload)
	ok, resp := CallInternal(group.Protocol.GovernanceContract, bz, false)
	if !ok {
		Revert("gov deposit vote failed: " + string(resp))
	}
	return resp
}

// =============================================================================
// QUERIES
// =============================================================================

// QueryIsMember checks if a user is a member of a group
func QueryIsMember(msg *MsgQueryIsMember) []byte {
	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		return MarshalJSON(QueryIsMemberResponse{IsMember: false})
	}

	if group.Token == "" {
		return MarshalJSON(QueryIsMemberResponse{IsMember: false})
	}

	addr, _ := QueryAddressByUserID(msg.UserID)
	if addr == "" {
		return MarshalJSON(QueryIsMemberResponse{IsMember: false})
	}

	return MarshalJSON(QueryIsMemberResponse{
		IsMember: hasMinTokenBalance(group, addr, msg.UserID),
	})
}

// QueryGetGroup returns group information
func QueryGetGroup(msg *MsgQueryGetGroup) []byte {
	group, _ := LoadGroup(msg.GroupID)
	return MarshalJSON(QueryGetGroupResponse{Group: group})
}

// QueryGetGroups returns all groups
func QueryGetGroups(msg *MsgQueryGetGroups) []byte {
	count := GetGroupCount()

	offset := uint64(0)
	limit := uint64(100)
	if msg.Pagination.Offset > 0 {
		offset = uint64(msg.Pagination.Offset)
	}
	if msg.Pagination.Limit > 0 {
		limit = uint64(msg.Pagination.Limit)
	}

	groups := make([]Group, 0)
	for i := uint64(1); i <= count && uint64(len(groups)) < limit; i++ {
		if i <= offset {
			continue
		}
		groupID := "group_" + utils.U64toa(i)
		group, _ := LoadGroup(groupID)
		if group != nil {
			groups = append(groups, *group)
		}
	}

	return MarshalJSON(QueryGetGroupsResponse{
		Groups: groups,
		Pagination: PageResponse{
			Total: utils.StringUint64(count),
		},
	})
}

// QueryGetVotingProtocol returns the voting protocol for a group
func QueryGetVotingProtocol(msg *MsgQueryGetVotingProtocol) []byte {
	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		return MarshalJSON(map[string]string{"error": "group not found"})
	}

	return MarshalJSON(QueryGetVotingProtocolResponse{Protocol: group.Protocol})
}

// QueryGetUserGroups returns all groups a user belongs to
func QueryGetUserGroups(msg *MsgQueryGetUserGroups) []byte {
	count := GetGroupCount()
	groupIDs := make([]string, 0)
	config := LoadConfig()
	if config == nil {
		return MarshalJSON(QueryGetUserGroupsResponse{
			GroupIDs: []string{},
			Pagination: PageResponse{
				Total: utils.StringUint64(0),
			},
		})
	}
	addr, _ := QueryAddressByUserID(msg.UserID)

	for i := uint64(1); i <= count; i++ {
		groupID := "group_" + utils.U64toa(i)
		group, _ := LoadGroup(groupID)
		if group == nil {
			continue
		}
		if addr != "" && hasMinTokenBalance(group, addr, msg.UserID) {
			groupIDs = append(groupIDs, groupID)
		}
	}

	return MarshalJSON(QueryGetUserGroupsResponse{
		GroupIDs: groupIDs,
		Pagination: PageResponse{
			Total: utils.StringUint64(uint64(len(groupIDs))),
		},
	})
}

// QueryIsAdmin checks if a user is an admin of a group
func QueryIsAdmin(msg *MsgQueryIsAdmin) []byte {
	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		return MarshalJSON(QueryIsAdminResponse{IsAdmin: false})
	}

	return MarshalJSON(QueryIsAdminResponse{
		IsAdmin: StringInList(msg.UserID, group.Admins),
	})
}

// QueryGetVoterPower returns membership and voting power for a voter
func QueryGetVoterPower(msg *MsgQueryGetVoterPower) []byte {
	config := LoadConfig()
	if config == nil {
		return MarshalJSON(map[string]string{"error": "contract not initialized"})
	}

	group, _ := LoadGroup(msg.GroupID)
	if group == nil {
		return MarshalJSON(QueryGetVoterPowerResponse{IsMember: false, Power: "0"})
	}

	userID := ""
	if strings.HasPrefix(msg.Voter, "user_") {
		userID = msg.Voter
	}

	if group.Token == "" {
		return MarshalJSON(QueryGetVoterPowerResponse{IsMember: false, UserID: userID, Power: "0"})
	}

	token, balance := getTokenBalanceWithAddress(group, msg.Voter, userID)
	isMember := hasMinTokenBalanceWithValue(group, balance)
	return MarshalJSON(QueryGetVoterPowerResponse{
		IsMember: isMember,
		UserID:   userID,
		Power:    balance.String(),
		Denom:    group.TokenDenom,
		Token:    string(token),
	})
}

// QueryGetConfig returns the contract configuration
func QueryGetConfig(_ *MsgQueryGetConfig) []byte {
	config := LoadConfig()
	if config == nil {
		return MarshalJSON(map[string]string{"error": "contract not initialized"})
	}
	return MarshalJSON(QueryGetConfigResponse{Config: *config})
}

// =============================================================================
// VOTING PROTOCOL HELPERS
// =============================================================================

// GetDefaultVotingProtocols returns predefined voting protocol configurations
func GetDefaultVotingProtocols() map[string]VotingProtocol {
	return map[string]VotingProtocol{
		"cosmos_standard": {
			Name:            "Cosmos Standard",
			Description:     "Standard Cosmos SDK governance with stake-weighted voting",
			Eligibility:     ELIGIBILITY_STAKERS,
			WeightAlgorithm: WEIGHT_STAKE,
			OptionsType:     OPTIONS_STANDARD,
			ResultAlgorithm: RESULT_THRESHOLD_WITH_QUORUM,
			TimingType:      TIMING_FIXED_PERIODS,
			Quorum:          "0.334",
			Threshold:       "0.5",
			VetoThreshold:   "0.334",
			DepositPeriod:   86400000,  // 1 day
			VotingPeriod:    259200000, // 3 days
		},
		"continuous_deposit": {
			Name:            "Continuous Deposit Voting",
			Description:     "Continuous voting with token deposits and curve-based thresholds",
			Eligibility:     ELIGIBILITY_TOKEN_HOLDERS,
			WeightAlgorithm: WEIGHT_DEPOSIT,
			OptionsType:     OPTIONS_MULTI,
			ResultAlgorithm: RESULT_CONTINUOUS_CURVE,
			TimingType:      TIMING_CONTINUOUS,
			CurveX:          10,
			CurveY:          8,
		},
		"group_equal_vote": {
			Name:            "Group Equal Vote",
			Description:     "Group members with equal voting power (one person one vote)",
			Eligibility:     ELIGIBILITY_GROUP_MEMBERS,
			WeightAlgorithm: WEIGHT_EQUAL,
			OptionsType:     OPTIONS_BINARY,
			ResultAlgorithm: RESULT_SIMPLE_MAJORITY,
			TimingType:      TIMING_DEADLINE,
			Quorum:          "0.5",
			VotingPeriod:    604800000, // 7 days
		},
		"quadratic_vote": {
			Name:            "Quadratic Voting",
			Description:     "Token-weighted voting with quadratic vote power",
			Eligibility:     ELIGIBILITY_TOKEN_HOLDERS,
			WeightAlgorithm: WEIGHT_QUADRATIC,
			OptionsType:     OPTIONS_MULTI,
			ResultAlgorithm: RESULT_PLURALITY,
			TimingType:      TIMING_DEADLINE,
			VotingPeriod:    604800000, // 7 days
		},
		"identity_vote": {
			Name:            "Identity-Based Vote",
			Description:     "Equal votes for verified identities",
			Eligibility:     ELIGIBILITY_IDENTITY,
			WeightAlgorithm: WEIGHT_EQUAL,
			OptionsType:     OPTIONS_BINARY,
			ResultAlgorithm: RESULT_SUPERMAJORITY,
			TimingType:      TIMING_DEADLINE,
			Threshold:       "0.667",
			VotingPeriod:    259200000, // 3 days
		},
	}
}

// MarshalVotingProtocol marshals a voting protocol to JSON
func MarshalVotingProtocol(protocol VotingProtocol) string {
	data, _ := json.Marshal(protocol)
	return string(data)
}
