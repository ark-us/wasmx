package lib

import (
	"encoding/json"
	"strconv"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

// Storage key prefixes
const (
	STORAGE_CONFIG        = "config"        // -> GroupConfig
	STORAGE_GROUP_PREFIX  = "group:"        // group:<group_id> -> Group
	STORAGE_GROUP_COUNTER = "group_counter" // -> uint64
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
