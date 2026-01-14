package lib

import (
	"encoding/json"
	"strconv"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

// Storage key prefixes
const (
	STORAGE_CONFIG           = "config"           // -> GroupConfig
	STORAGE_GROUP_PREFIX     = "group:"           // group:<group_id> -> Group
	STORAGE_GROUP_COUNTER    = "group_counter"    // -> uint64
	STORAGE_MEMBER_PREFIX    = "member:"          // member:<group_id>:<user_id> -> GroupMember
	STORAGE_MEMBER_COUNT     = "member_count:"    // member_count:<group_id> -> uint64
	STORAGE_MEMBER_INDEX     = "member_idx:"      // member_idx:<group_id>:<index> -> user_id
	STORAGE_USER_GROUPS      = "user_groups:"     // user_groups:<user_id>:<group_id> -> "1"
	STORAGE_USER_GROUP_COUNT = "user_group_count:" // user_group_count:<user_id> -> uint64
)

// =============================================================================
// CONFIG STORAGE
// =============================================================================

func LoadConfig() *GroupConfig {
	data := wasmx.StorageLoad([]byte(STORAGE_CONFIG))
	if len(data) == 0 {
		return nil
	}
	var config GroupConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}
	return &config
}

func SaveConfig(config *GroupConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	wasmx.StorageStore([]byte(STORAGE_CONFIG), data)
	return nil
}

// =============================================================================
// GROUP STORAGE
// =============================================================================

func LoadGroup(groupID string) (*Group, error) {
	key := []byte(STORAGE_GROUP_PREFIX + groupID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil, nil
	}
	var group Group
	if err := json.Unmarshal(data, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

func SaveGroup(group *Group) error {
	key := []byte(STORAGE_GROUP_PREFIX + group.ID)
	data, err := json.Marshal(group)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

func DeleteGroup(groupID string) {
	key := STORAGE_GROUP_PREFIX + groupID
	wasmx.StorageDelete(key)
}

func GetNextGroupID() string {
	key := []byte(STORAGE_GROUP_COUNTER)
	data := wasmx.StorageLoad(key)

	var counter uint64
	if len(data) > 0 {
		counter, _ = strconv.ParseUint(string(data), 10, 64)
	}

	counter++
	wasmx.StorageStore(key, []byte(utils.U64toa(counter)))

	return "group_" + utils.U64toa(counter)
}

func GetGroupCount() uint64 {
	key := []byte(STORAGE_GROUP_COUNTER)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return 0
	}
	counter, _ := strconv.ParseUint(string(data), 10, 64)
	return counter
}

// =============================================================================
// MEMBER STORAGE
// =============================================================================

func LoadMember(groupID, userID string) (*GroupMember, error) {
	key := []byte(STORAGE_MEMBER_PREFIX + groupID + ":" + userID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil, nil
	}
	var member GroupMember
	if err := json.Unmarshal(data, &member); err != nil {
		return nil, err
	}
	return &member, nil
}

func SaveMember(groupID string, member *GroupMember) error {
	key := []byte(STORAGE_MEMBER_PREFIX + groupID + ":" + member.UserID)
	data, err := json.Marshal(member)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

func DeleteMember(groupID, userID string) {
	key := STORAGE_MEMBER_PREFIX + groupID + ":" + userID
	wasmx.StorageDelete(key)
}

// GetMemberCount returns the count of members in a group
func GetMemberCount(groupID string) uint64 {
	key := []byte(STORAGE_MEMBER_COUNT + groupID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return 0
	}
	count, _ := strconv.ParseUint(string(data), 10, 64)
	return count
}

func SetMemberCount(groupID string, count uint64) {
	key := []byte(STORAGE_MEMBER_COUNT + groupID)
	wasmx.StorageStore(key, []byte(utils.U64toa(count)))
}

func IncrementMemberCount(groupID string) uint64 {
	count := GetMemberCount(groupID)
	count++
	SetMemberCount(groupID, count)
	return count
}

func DecrementMemberCount(groupID string) uint64 {
	count := GetMemberCount(groupID)
	if count > 0 {
		count--
	}
	SetMemberCount(groupID, count)
	return count
}

// Member index storage for enumeration
func SaveMemberIndex(groupID string, index uint64, userID string) {
	key := []byte(STORAGE_MEMBER_INDEX + groupID + ":" + utils.U64toa(index))
	wasmx.StorageStore(key, []byte(userID))
}

func LoadMemberIndex(groupID string, index uint64) string {
	key := []byte(STORAGE_MEMBER_INDEX + groupID + ":" + utils.U64toa(index))
	data := wasmx.StorageLoad(key)
	return string(data)
}

func DeleteMemberIndex(groupID string, index uint64) {
	key := STORAGE_MEMBER_INDEX + groupID + ":" + utils.U64toa(index)
	wasmx.StorageDelete(key)
}

// AddMemberToIndex adds a user to the member index and increments count
func AddMemberToIndex(groupID, userID string) uint64 {
	count := GetMemberCount(groupID)
	SaveMemberIndex(groupID, count, userID)
	return IncrementMemberCount(groupID)
}

// RemoveMemberFromIndex removes a user from the index by swapping with last
func RemoveMemberFromIndex(groupID, userID string) {
	count := GetMemberCount(groupID)
	if count == 0 {
		return
	}

	// Find the index of the user
	var foundIndex int64 = -1
	for i := uint64(0); i < count; i++ {
		if LoadMemberIndex(groupID, i) == userID {
			foundIndex = int64(i)
			break
		}
	}

	if foundIndex < 0 {
		return
	}

	// If not the last element, swap with last
	lastIndex := count - 1
	if uint64(foundIndex) != lastIndex {
		lastUserID := LoadMemberIndex(groupID, lastIndex)
		SaveMemberIndex(groupID, uint64(foundIndex), lastUserID)
	}

	// Delete the last index
	DeleteMemberIndex(groupID, lastIndex)
	DecrementMemberCount(groupID)
}

// GetAllMemberIDs returns all member user_ids in a group
func GetAllMemberIDs(groupID string, offset, limit uint64) []string {
	count := GetMemberCount(groupID)
	if offset >= count {
		return []string{}
	}

	end := offset + limit
	if end > count {
		end = count
	}

	result := make([]string, 0, end-offset)
	for i := offset; i < end; i++ {
		userID := LoadMemberIndex(groupID, i)
		if userID != "" {
			result = append(result, userID)
		}
	}
	return result
}

// =============================================================================
// USER GROUPS STORAGE (reverse index: user_id -> group_ids)
// =============================================================================

func AddUserGroup(userID, groupID string) {
	key := []byte(STORAGE_USER_GROUPS + userID + ":" + groupID)
	wasmx.StorageStore(key, []byte("1"))
	IncrementUserGroupCount(userID)
}

func RemoveUserGroup(userID, groupID string) {
	key := STORAGE_USER_GROUPS + userID + ":" + groupID
	wasmx.StorageDelete(key)
	DecrementUserGroupCount(userID)
}

func IsUserInGroup(userID, groupID string) bool {
	key := []byte(STORAGE_USER_GROUPS + userID + ":" + groupID)
	data := wasmx.StorageLoad(key)
	return len(data) > 0
}

func GetUserGroupCount(userID string) uint64 {
	key := []byte(STORAGE_USER_GROUP_COUNT + userID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return 0
	}
	count, _ := strconv.ParseUint(string(data), 10, 64)
	return count
}

func SetUserGroupCount(userID string, count uint64) {
	key := []byte(STORAGE_USER_GROUP_COUNT + userID)
	wasmx.StorageStore(key, []byte(utils.U64toa(count)))
}

func IncrementUserGroupCount(userID string) {
	count := GetUserGroupCount(userID)
	SetUserGroupCount(userID, count+1)
}

func DecrementUserGroupCount(userID string) {
	count := GetUserGroupCount(userID)
	if count > 0 {
		SetUserGroupCount(userID, count-1)
	}
}
