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

// SetMinters stores the minter user IDs
func SetMinters(minters []string) {
	data, _ := json.Marshal(minters)
	wasmx.StorageStore([]byte(MINTERS_KEY), data)
}

// GetMinters retrieves the minter user IDs
func GetMinters() []string {
	value := wasmx.StorageLoad([]byte(MINTERS_KEY))
	if len(value) == 0 {
		return []string{}
	}
	var minters []string
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

// SetBalance stores the balance for an address
func SetBalance(addr string, amount sdkmath.Int) {
	key := GetBalanceKey(addr)
	wasmx.StorageStore([]byte(key), []byte(amount.String()))
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
