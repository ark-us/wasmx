package lib

import (
	"encoding/json"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func GetBalanceKey(addr string) string {
	return BALANCE_KEY + addr
}

func GetAllowanceKey(owner, spender string) string {
	return ALLOWANCE_KEY + owner + "_" + spender
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

// SetAdmins stores the admin addresses
func SetAdmins(admins []wasmx.Bech32String) {
	data, _ := json.Marshal(admins)
	wasmx.StorageStore([]byte(ADMINS_KEY), data)
}

// GetAdmins retrieves the admin addresses
func GetAdmins() []wasmx.Bech32String {
	value := wasmx.StorageLoad([]byte(ADMINS_KEY))
	if len(value) == 0 {
		return []wasmx.Bech32String{}
	}
	var admins []wasmx.Bech32String
	json.Unmarshal(value, &admins)
	return admins
}

// SetMinters stores the minter addresses
func SetMinters(minters []wasmx.Bech32String) {
	data, _ := json.Marshal(minters)
	wasmx.StorageStore([]byte(MINTERS_KEY), data)
}

// GetMinters retrieves the minter addresses
func GetMinters() []wasmx.Bech32String {
	value := wasmx.StorageLoad([]byte(MINTERS_KEY))
	if len(value) == 0 {
		return []wasmx.Bech32String{}
	}
	var minters []wasmx.Bech32String
	json.Unmarshal(value, &minters)
	return minters
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

// GetNegativeBalanceThreshold retrieves the negative balance threshold
func GetNegativeBalanceThreshold() sdkmath.Int {
	value := wasmx.StorageLoad([]byte(NEGATIVE_BALANCE_THRESHOLD))
	if len(value) == 0 {
		return sdkmath.ZeroInt()
	}
	result, ok := sdkmath.NewIntFromString(string(value))
	if !ok {
		return sdkmath.ZeroInt()
	}
	return result
}

// SetNegativeBalanceThreshold stores the negative balance threshold
func SetNegativeBalanceThreshold(threshold sdkmath.Int) {
	wasmx.StorageStore([]byte(NEGATIVE_BALANCE_THRESHOLD), []byte(threshold.String()))
}

// GetBalance retrieves the balance for an address (ERC20 compatible - returns 0 for negative balances)
func GetBalance(addr string) sdkmath.Int {
	threshold := GetNegativeBalanceThreshold()
	key := GetBalanceKey(addr)
	value := wasmx.StorageLoad([]byte(key))
	if len(value) == 0 {
		return sdkmath.ZeroInt()
	}
	stored, ok := sdkmath.NewIntFromString(string(value))
	if !ok {
		return sdkmath.ZeroInt()
	}

	// stored = actual + threshold
	// actual = stored - threshold
	actual := stored.Sub(threshold)

	// For ERC20 compatibility, return 0 if balance is negative
	if actual.IsNegative() {
		return sdkmath.ZeroInt()
	}
	return actual
}

// GetSignedBalance retrieves the actual signed balance for an address (can be negative)
func GetSignedBalance(addr string) sdkmath.Int {
	threshold := GetNegativeBalanceThreshold()
	key := GetBalanceKey(addr)
	value := wasmx.StorageLoad([]byte(key))
	if len(value) == 0 {
		// No balance stored means actual balance is -threshold (minimum allowed)
		return threshold.Neg()
	}
	stored, ok := sdkmath.NewIntFromString(string(value))
	if !ok {
		return threshold.Neg()
	}

	// stored = actual + threshold
	// actual = stored - threshold
	return stored.Sub(threshold)
}

// SetBalance stores the balance for an address using offset storage
// amount is the actual balance (can be negative down to -threshold)
func SetBalance(addr string, amount sdkmath.Int) {
	threshold := GetNegativeBalanceThreshold()

	// stored = actual + threshold
	stored := amount.Add(threshold)

	// Validate that stored value is non-negative (i.e., actual >= -threshold)
	if stored.IsNegative() {
		Revert("balance cannot be less than negative threshold")
	}

	key := GetBalanceKey(addr)
	wasmx.StorageStore([]byte(key), []byte(stored.String()))
}

// GetAllowance retrieves the allowance for owner/spender pair
func GetAllowance(owner, spender string) sdkmath.Int {
	key := GetAllowanceKey(owner, spender)
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

// SetAllowance stores the allowance for owner/spender pair
func SetAllowance(owner, spender string, amount sdkmath.Int) {
	key := GetAllowanceKey(owner, spender)
	wasmx.StorageStore([]byte(key), []byte(amount.String()))
}
