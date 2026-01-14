package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

// =============================================================================
// INITIALIZATION
// =============================================================================

func InitGenesis(msg *MsgInitGenesis) []byte {
	LoggerInfo("InitGenesis called", nil)

	// Save config
	config := &GroupConfig{
		IdentityContract: msg.IdentityContract,
	}
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
				MemberCount: 0,
				Admins:      groupGenesis.Admins,
				Metadata:    groupGenesis.Metadata,
				Protocol:    groupGenesis.Protocol,
			}

			if err := SaveGroup(group); err != nil {
				LoggerError("Failed to save group in genesis", []string{"error", err.Error()})
				continue
			}

			// Add initial members
			for _, memberGenesis := range groupGenesis.Members {
				addMemberInternal(groupID, memberGenesis.UserID, memberGenesis.Metadata, timestamp)
			}

			// Update member count
			group.MemberCount = GetMemberCount(groupID)
			SaveGroup(group)

			LoggerInfo("Group created in genesis", []string{
				"group_id", groupID,
				"name", group.Name,
				"members", utils.U64toa(group.MemberCount),
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
		MemberCount: 0,
		Admins:      admins,
		Metadata:    msg.Metadata,
		Protocol:    msg.Protocol,
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
// MEMBER MANAGEMENT
// =============================================================================

// AddMember adds a member to a group
// Must be called by the governance contract OR an admin during bootstrap
func AddMember(msg *MsgAddMember) []byte {
	LoggerInfo("AddMember called", []string{"group_id", msg.GroupID, "user_id", msg.UserID})

	config := LoadConfig()
	if config == nil {
		Revert("contract not initialized")
	}

	group, err := LoadGroup(msg.GroupID)
	if err != nil || group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	// Authorization: must be governance contract or admin
	isGov := IsCallerGovernanceContract(group)
	isAdmin := IsCallerAdmin(group, config)

	if !isGov && !isAdmin {
		Revert("unauthorized: must be governance contract or admin")
	}

	// Verify user exists in identity contract
	if !QueryUserExists(config.IdentityContract, msg.UserID) {
		Revert("user does not exist in identity contract: " + msg.UserID)
	}

	// Check if already a member
	existingMember, _ := LoadMember(msg.GroupID, msg.UserID)
	if existingMember != nil {
		Revert("user is already a member: " + msg.UserID)
	}

	timestamp := GetBlockTime()
	addMemberInternal(msg.GroupID, msg.UserID, msg.Metadata, timestamp)

	// Update group member count
	group.MemberCount = GetMemberCount(msg.GroupID)
	group.UpdatedAt = timestamp
	SaveGroup(group)

	// Emit event
	ev := wasmx.Event{
		Type: "add_member",
		Attributes: []wasmx.EventAttribute{
			{Key: "group_id", Value: msg.GroupID, Index: true},
			{Key: "user_id", Value: msg.UserID, Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	LoggerInfo("Member added", []string{"group_id", msg.GroupID, "user_id", msg.UserID})

	return MarshalJSON(MsgAddMemberResponse{Success: true})
}

// RemoveMember removes a member from a group
// Must be called by the governance contract OR an admin during bootstrap
func RemoveMember(msg *MsgRemoveMember) []byte {
	LoggerInfo("RemoveMember called", []string{"group_id", msg.GroupID, "user_id", msg.UserID})

	config := LoadConfig()
	if config == nil {
		Revert("contract not initialized")
	}

	group, err := LoadGroup(msg.GroupID)
	if err != nil || group == nil {
		Revert("group not found: " + msg.GroupID)
	}

	// Authorization: must be governance contract or admin
	isGov := IsCallerGovernanceContract(group)
	isAdmin := IsCallerAdmin(group, config)

	if !isGov && !isAdmin {
		Revert("unauthorized: must be governance contract or admin")
	}

	// Check if member exists
	existingMember, _ := LoadMember(msg.GroupID, msg.UserID)
	if existingMember == nil {
		Revert("user is not a member: " + msg.UserID)
	}

	timestamp := GetBlockTime()
	removeMemberInternal(msg.GroupID, msg.UserID)

	// Update group member count
	group.MemberCount = GetMemberCount(msg.GroupID)
	group.UpdatedAt = timestamp
	SaveGroup(group)

	// Emit event
	ev := wasmx.Event{
		Type: "remove_member",
		Attributes: []wasmx.EventAttribute{
			{Key: "group_id", Value: msg.GroupID, Index: true},
			{Key: "user_id", Value: msg.UserID, Index: true},
		},
	}
	wasmx.EmitCosmosEvents([]wasmx.Event{ev})

	LoggerInfo("Member removed", []string{"group_id", msg.GroupID, "user_id", msg.UserID})

	return MarshalJSON(MsgRemoveMemberResponse{Success: true})
}

// addMemberInternal is the internal implementation for adding a member
func addMemberInternal(groupID, userID, metadata string, timestamp int64) {
	member := &GroupMember{
		UserID:   userID,
		JoinedAt: timestamp,
		Metadata: metadata,
	}

	if err := SaveMember(groupID, member); err != nil {
		Revert("failed to save member: " + err.Error())
	}

	// Add to index and user groups
	AddMemberToIndex(groupID, userID)
	AddUserGroup(userID, groupID)
}

// removeMemberInternal is the internal implementation for removing a member
func removeMemberInternal(groupID, userID string) {
	DeleteMember(groupID, userID)
	RemoveMemberFromIndex(groupID, userID)
	RemoveUserGroup(userID, groupID)
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
// QUERIES
// =============================================================================

// QueryIsMember checks if a user is a member of a group
func QueryIsMember(msg *MsgQueryIsMember) []byte {
	member, _ := LoadMember(msg.GroupID, msg.UserID)

	response := QueryIsMemberResponse{
		IsMember: member != nil,
	}
	if member != nil {
		response.Member = member
	}

	return MarshalJSON(response)
}

// QueryGetAllMembers returns all members of a group
func QueryGetAllMembers(msg *MsgQueryGetAllMembers) []byte {
	totalCount := GetMemberCount(msg.GroupID)

	offset := uint64(0)
	limit := uint64(100) // default limit
	if msg.Pagination.Offset > 0 {
		offset = uint64(msg.Pagination.Offset)
	}
	if msg.Pagination.Limit > 0 {
		limit = uint64(msg.Pagination.Limit)
	}

	userIDs := GetAllMemberIDs(msg.GroupID, offset, limit)

	members := make([]GroupMember, 0, len(userIDs))
	for _, userID := range userIDs {
		member, _ := LoadMember(msg.GroupID, userID)
		if member != nil {
			members = append(members, *member)
		}
	}

	return MarshalJSON(QueryGetAllMembersResponse{
		Members: members,
		Pagination: PageResponse{
			Total: utils.StringUint64(totalCount),
		},
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
	// This is a simplified implementation - in production you might want
	// to maintain a reverse index for efficiency
	count := GetGroupCount()
	groupIDs := make([]string, 0)

	for i := uint64(1); i <= count; i++ {
		groupID := "group_" + utils.U64toa(i)
		if IsUserInGroup(msg.UserID, groupID) {
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
