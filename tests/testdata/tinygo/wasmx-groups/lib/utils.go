package lib

import (
	"encoding/json"
	"math/big"

	sdkmath "cosmossdk.io/math"
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

func getIdentityContract() wasmx.Bech32String {
	identityContract := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	if identityContract == "" {
		identityContract = wasmx.GetAddressByRole(wasmx.ROLE_IDENTITY)
	}
	if identityContract == "" {
		Revert("identity role not set")
	}
	return identityContract
}

// QueryUserByAddress queries the identity contract to get user_id from an address
func QueryUserByAddress(address string) (string, error) {
	identityContract := getIdentityContract()
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

// QueryAddressByUserID queries the identity contract to get address from a user_id
func QueryAddressByUserID(userID string) (string, error) {
	identityContract := getIdentityContract()
	query := map[string]interface{}{
		"query_user_by_id": map[string]string{
			"user_id": userID,
		},
	}
	queryBz, _ := json.Marshal(query)

	ok, respBz := wasmx.CallSimple(identityContract, queryBz, true, MODULE_NAME)
	if !ok {
		return "", nil
	}

	var resp struct {
		User struct {
			Addresses []string `json:"addresses"`
		} `json:"user"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBz, &resp); err != nil {
		return "", err
	}
	if resp.Error != "" || len(resp.User.Addresses) == 0 {
		return "", nil
	}
	return resp.User.Addresses[0], nil
}

// QueryUserExists queries if a user_id exists in the identity contract
func QueryUserExists(userID string) bool {
	identityContract := getIdentityContract()
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
	userID, err := QueryUserByAddress(string(caller))
	if err != nil || userID == "" {
		return false
	}

	return StringInList(userID, group.Admins)
}

// GetCallerUserID gets the user_id for the caller from the identity contract
func GetCallerUserID(config *GroupConfig) string {
	caller := wasmx.GetCaller()
	userID, _ := QueryUserByAddress(string(caller))
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

func getTokenSymbol(token wasmx.Bech32String) (string, bool) {
	payload := struct {
		Q struct{} `json:"symbol"`
	}{}
	bz, _ := json.Marshal(&payload)
	ok, resp := wasmx.CallSimple(token, bz, true, MODULE_NAME)
	if !ok {
		return "", false
	}
	var out struct {
		Symbol string `json:"symbol"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		return "", false
	}
	if out.Symbol == "" {
		return "", false
	}
	return out.Symbol, true
}

func getTokenBalance(token wasmx.Bech32String, owner string) (sdkmath.Int, bool) {
	payload := struct {
		Q struct {
			Owner string `json:"owner"`
		} `json:"balanceOf"`
	}{}
	payload.Q.Owner = owner
	bz, _ := json.Marshal(&payload)
	ok, resp := wasmx.CallSimple(token, bz, true, MODULE_NAME)
	if !ok {
		return sdkmath.ZeroInt(), false
	}
	var out struct {
		Balance wasmx.Coin `json:"balance"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		return sdkmath.ZeroInt(), false
	}
	return out.Balance.Amount, true
}

func getTokenBalanceWithAddress(group *Group, owner string, userID string) (wasmx.Bech32String, sdkmath.Int) {
	token := group.Token
	if token == "" {
		return "", sdkmath.ZeroInt()
	}
	if balance, ok := getTokenBalance(token, owner); ok && !balance.IsZero() {
		return token, balance
	}
	if userID == "" {
		if resolved, err := QueryUserByAddress(owner); err == nil {
			userID = resolved
		}
	}
	if userID != "" {
		if balance, ok := getTokenBalance(token, userID); ok {
			return token, balance
		}
	}
	return token, sdkmath.ZeroInt()
}

func hasMinTokenBalance(group *Group, owner string, userID string) bool {
	_, balance := getTokenBalanceWithAddress(group, owner, userID)
	return hasMinTokenBalanceWithValue(group, balance)
}

func hasMinTokenBalanceWithValue(group *Group, balance sdkmath.Int) bool {
	minBalance := parseMinBalance(group)
	bal := new(big.Int)
	bal.SetString(balance.String(), 10)
	return bal.Cmp(minBalance) >= 0
}

func parseMinBalance(group *Group) *big.Int {
	minBalance := new(big.Int)
	if group == nil || group.MinBalance == "" {
		return minBalance.SetInt64(1)
	}
	if _, ok := minBalance.SetString(group.MinBalance, 10); !ok {
		return minBalance.SetInt64(1)
	}
	if minBalance.Sign() <= 0 {
		return minBalance.SetInt64(1)
	}
	return minBalance
}
