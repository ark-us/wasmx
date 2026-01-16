package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func LoggerInfo(msg string, parts []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, parts)
}

func LoggerError(msg string, parts []string) {
	wasmx.LoggerError(MODULE_NAME, msg, parts)
}

func LoggerDebug(msg string, parts []string) {
	wasmx.LoggerDebug(MODULE_NAME, msg, parts)
}

func LoggerDebugExtended(msg string, parts []string) {
	wasmx.LoggerDebugExtended(MODULE_NAME, msg, parts)
}

func Revert(message string) {
	LoggerDebug("revert", []string{"err", message, "module", MODULE_NAME})
	wasmx.RevertWithModule(MODULE_NAME, message)
}

func StringInList(value string, list []string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// QueryUserByAddress queries the identity contract (by role) to get user_id from an address.
func QueryUserByAddress(address string) (string, error) {
	identityContract := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	if identityContract == "" {
		return "", nil
	}
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

func ResolveUserIDOrAddress(value string) string {
	if value == "" {
		return value
	}
	if userID, _ := QueryUserByAddress(value); userID != "" {
		return userID
	}
	return value
}

func GetCallerUserID() string {
	caller := wasmx.GetCaller()
	userID, _ := QueryUserByAddress(string(caller))
	return userID
}

func toFixed32(b []byte) [32]byte {
	var out [32]byte
	copy(out[:], b)
	return out
}

func userIDToTopic(userID string) [32]byte {
	if userID == "" {
		return [32]byte{}
	}
	return toFixed32(wasmx.Sha256([]byte(userID)))
}
