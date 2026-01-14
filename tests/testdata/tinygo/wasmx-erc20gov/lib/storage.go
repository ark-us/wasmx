package lib

import (
	"encoding/json"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func GetBalanceKey(addr string) string {
	return BALANCE_KEY + addr
}

// SetInfo stores token information
func SetInfo(value TokenInfo) {
	data, _ := json.Marshal(value)
	wasmx.StorageStore([]byte(INFO_KEY), data)
}

// GetInfo retrieves token information
func GetInfo() TokenInfo {
	value := wasmx.StorageLoad([]byte(INFO_KEY))
	if len(value) == 0 {
		return TokenInfo{Name: "", Symbol: "", Decimals: 0}
	}
	var info TokenInfo
	json.Unmarshal(value, &info)
	return info
}

// SetAdmins stores the admin user IDs
func SetAdmins(admins []string) {
	data, _ := json.Marshal(admins)
	wasmx.StorageStore([]byte(ADMINS_KEY), data)
}

// GetAdmins retrieves the admin user IDs
func GetAdmins() []string {
	value := wasmx.StorageLoad([]byte(ADMINS_KEY))
	if len(value) == 0 {
		return []string{}
	}
	var admins []string
	json.Unmarshal(value, &admins)
	return admins
}

// SetGovernanceAddress stores the governance contract address
func SetGovernanceAddress(address string) {
	wasmx.StorageStore([]byte(GOVERNANCE_ADDRESS_KEY), []byte(address))
}

// GetGovernanceAddress retrieves the governance contract address
func GetGovernanceAddress() string {
	value := wasmx.StorageLoad([]byte(GOVERNANCE_ADDRESS_KEY))
	if len(value) == 0 {
		return ""
	}
	return string(value)
}

// SetTotalSupply stores the total supply
func SetTotalSupply(value sdkmath.Int) {
	wasmx.StorageStore([]byte(TOTAL_SUPPLY_KEY), []byte(value.String()))
}

// GetTotalSupply retrieves the total supply
func GetTotalSupply() sdkmath.Int {
	value := wasmx.StorageLoad([]byte(TOTAL_SUPPLY_KEY))
	if len(value) == 0 {
		return sdkmath.ZeroInt()
	}
	result, ok := sdkmath.NewIntFromString(string(value))
	if !ok {
		return sdkmath.ZeroInt()
	}
	return result
}

// GetBalance retrieves the balance for an address
func GetBalance(addr string) sdkmath.Int {
	key := GetBalanceKey(addr)
	value := wasmx.StorageLoad([]byte(key))
	if len(value) == 0 {
		return sdkmath.ZeroInt()
	}
	result, ok := sdkmath.NewIntFromString(string(value))
	if !ok {
		return sdkmath.ZeroInt()
	}
	return result
}

// SetBalance stores the balance for an address and updates member count
func SetBalance(addr string, amount sdkmath.Int) {
	key := GetBalanceKey(addr)
	oldBalance := GetBalance(addr)

	// Update member count based on balance transition
	if oldBalance.IsZero() && amount.IsPositive() {
		// New member: balance 0 -> positive
		IncrementMemberCount()
	} else if oldBalance.IsPositive() && amount.IsZero() {
		// Member removed: balance positive -> 0
		DecrementMemberCount()
	}

	wasmx.StorageStore([]byte(key), []byte(amount.String()))
}

// GetMemberCount retrieves the count of members with balance > 0
func GetMemberCount() uint64 {
	value := wasmx.StorageLoad([]byte(MEMBER_COUNT_KEY))
	if len(value) == 0 {
		return 0
	}
	result, ok := sdkmath.NewIntFromString(string(value))
	if !ok {
		return 0
	}
	return result.Uint64()
}

// SetMemberCount stores the member count
func SetMemberCount(count uint64) {
	wasmx.StorageStore([]byte(MEMBER_COUNT_KEY), []byte(sdkmath.NewIntFromUint64(count).String()))
}

// IncrementMemberCount increases the member count by 1
func IncrementMemberCount() {
	count := GetMemberCount()
	SetMemberCount(count + 1)
}

// DecrementMemberCount decreases the member count by 1
func DecrementMemberCount() {
	count := GetMemberCount()
	if count > 0 {
		SetMemberCount(count - 1)
	}
}
