package lib

import (
	"encoding/json"
	"math/big"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// =============================================================================
// LOGGING
// =============================================================================

func LoggerInfo(message string, keyvals []string) {
	wasmx.LoggerInfo(MODULE_NAME, message, keyvals)
}

func LoggerError(message string, keyvals []string) {
	wasmx.LoggerError(MODULE_NAME, message, keyvals)
}

func LoggerDebug(message string, keyvals []string) {
	wasmx.LoggerDebug(MODULE_NAME, message, keyvals)
}

func Revert(message string) {
	LoggerDebug("revert", []string{"err", message, "module", MODULE_NAME})
	wasmx.RevertWithModule(MODULE_NAME, message)
}

// =============================================================================
// TIME HELPERS
// =============================================================================

func GetBlockTime() int64 {
	block := wasmx.GetCurrentBlock()
	return int64(block.Timestamp)
}

// =============================================================================
// LIST HELPERS
// =============================================================================

func StringInList(item string, list []string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

func RemoveStringFromList(item string, list []string) []string {
	result := make([]string, 0, len(list))
	for _, i := range list {
		if i != item {
			result = append(result, i)
		}
	}
	return result
}

// =============================================================================
// IDENTITY CONTRACT CALLS
// =============================================================================

// QueryUserByAddress queries the identity contract to get user_id from an address
func QueryUserByAddress(identityContract wasmx.Bech32String, address string) (string, error) {
	query := map[string]interface{}{
		"query_user_by_address": map[string]string{
			"address": address,
		},
	}
	queryBz, _ := json.Marshal(query)

	ok, respBz := wasmx.CallSimple(identityContract, queryBz, true, MODULE_NAME)
	if !ok {
		return "", nil
	}

	var resp struct {
		UserID string `json:"user_id"`
		Error  string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBz, &resp); err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", nil
	}

	return resp.UserID, nil
}

// QueryUserExists queries if a user_id exists in the identity contract
func QueryUserExists(identityContract wasmx.Bech32String, userID string) bool {
	query := map[string]interface{}{
		"query_user_by_id": map[string]string{
			"user_id": userID,
		},
	}
	queryBz, _ := json.Marshal(query)

	ok, respBz := wasmx.CallSimple(identityContract, queryBz, true, MODULE_NAME)
	if !ok {
		return false
	}

	var resp struct {
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBz, &resp); err != nil {
		return false
	}

	return resp.Error == ""
}

// =============================================================================
// AUTHORIZATION HELPERS
// =============================================================================

// IsCallerGovernanceContract checks if the caller is the governance contract for the group
func IsCallerGovernanceContract(group *Group) bool {
	caller := wasmx.GetCaller()
	return caller == group.Protocol.GovernanceContract
}

// IsCallerAdmin checks if the caller (by user_id) is an admin of the group
func IsCallerAdmin(group *Group, config *GroupConfig) bool {
	caller := wasmx.GetCaller()

	// Get user_id from caller address
	userID, err := QueryUserByAddress(config.IdentityContract, string(caller))
	if err != nil || userID == "" {
		return false
	}

	return StringInList(userID, group.Admins)
}

// GetCallerUserID gets the user_id for the caller from the identity contract
func GetCallerUserID(config *GroupConfig) string {
	caller := wasmx.GetCaller()
	userID, _ := QueryUserByAddress(config.IdentityContract, string(caller))
	return userID
}

// =============================================================================
// JSON HELPERS
// =============================================================================

func MarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		LoggerError("Failed to marshal JSON", []string{"error", err.Error()})
		return []byte("{}")
	}
	return data
}

// =============================================================================
// INTERNAL CALL HELPER
// =============================================================================

func CallInternal(target wasmx.Bech32String, calldata []byte, isQuery bool) (bool, []byte) {
	return wasmx.CallInternal(target, nil, calldata, big.NewInt(50_000_000), isQuery, MODULE_NAME)
}
