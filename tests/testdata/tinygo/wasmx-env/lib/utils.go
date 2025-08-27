package wasmx

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

// IsAuthorized returns true if caller is explicitly allowed or matches any provided role.
func IsAuthorized(caller Bech32String, authorities []Bech32String) bool {
	for _, a := range authorities {
		if a == caller {
			return true
		}
	}
	// also allow when any authority entry is a role name that resolves to caller address
	for _, a := range authorities {
		addr := GetAddressByRole(string(a))
		if addr == caller {
			return true
		}
	}
	return false
}

// ToUpperCase uppercases ASCII letters.
func ToUpperCase(str string) string {
	b := []byte(str)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] = b[i] - 32
		}
	}
	return string(b)
}

// QueryRoleContract queries the roles module.
func QueryRoleContract(calldata string, moduleName string) (bool, []byte) {
	return CallStatic(Bech32String(ROLE_ROLES), []byte(calldata), bigInt(DEFAULT_GAS_TX), moduleName)
}

// GetRoleName returns the role name for an address, or empty string.
func GetRoleName(moduleName string, addr Bech32String) string {
	calldata := fmt.Sprintf("{\"GetRoleNameByAddress\":{\"address\":\"%s\"}}", addr)
	ok, data := QueryRoleContract(calldata, moduleName)
	if !ok {
		msg := "role name by address failed: " + string(data)
		LoggerDebug(moduleName, "revert", []string{"err", msg, "module", moduleName})
		Revert([]byte(msg))
	}
	return string(data)
}

// CallerHasRole returns true if caller has any role under moduleName.
func CallerHasRole(moduleName string) bool {
	caller := GetCaller()
	role := GetRoleName(moduleName, caller)
	return len(role) > 0
}

// HasRole returns true if address has a role bound in module.
func HasRole(addr Bech32String, moduleName string) bool {
	role := GetRoleName(moduleName, addr)
	return len(role) > 0
}

// IsGoCoreModule checks if addr bytes match a known GoCore module.
func IsGoCoreModule(addr []byte) bool {
	hexstr := hex.EncodeToString(addr)
	if _, ok := GOCORE_MODULE_ADDRESSES[hexstr]; ok {
		return true
	}
	return false
}

// IsInternalContract returns true if addr is a role address for moduleName or a GoCore module address.
func IsInternalContract(moduleName string, addr Bech32String) bool {
	if HasRole(addr, moduleName) {
		return true
	}
	if IsGoCoreModule(AddrCanonicalize(string(addr))) {
		return true
	}
	return false
}

// OnlyRole allows only a specific role to call.
func OnlyRole(moduleName string, roleName string, message string) {
	caller := GetCaller()
	role := GetRoleName(moduleName, caller)
	if role == roleName {
		return
	}
	msg := fmt.Sprintf("unauthorized caller: %s, expected role %s: %s", caller, roleName, message)
	LoggerDebug(moduleName, "revert", []string{"err", msg, "module", moduleName})
	Revert([]byte(msg))
}

// OnlyInternal restricts calls to internal/core contract callers.
func OnlyInternal(moduleName string, message string) {
	caller := GetCaller()
	if HasRole(caller, moduleName) {
		return
	}
	if IsGoCoreModule(AddrCanonicalize(string(caller))) {
		return
	}
	addr := GetAddress()
	if caller == addr {
		return
	}
	msg := fmt.Sprintf("%s: unauthorized caller: %s: %s", moduleName, caller, message)
	LoggerDebug(moduleName, "revert", []string{"err", msg, "module", moduleName})
	Revert([]byte(msg))
}

// internal helper to build gas limit
func bigInt(v int64) *big.Int { return new(big.Int).SetInt64(v) }
