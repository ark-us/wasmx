package lib

import (
	"strings"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// isValidRoutePrefix checks if a route prefix is valid
func isValidRoutePrefix(prefix string) bool {
	return strings.HasPrefix(prefix, "/tools/") && len(prefix) > 7
}

// hasMcpRole checks if caller has mcp role
func hasGovRole(caller wasmx.Bech32String) bool {
	role := wasmx.GetRoleName(MODULE_NAME, caller)
	return role == wasmx.ROLE_GOVERNANCE
}

// formatToolName formats a tool name with route prefix
// Extract contract identifier from route: /tools/execute -> execute
func formatToolName(routePrefix, toolName string) string {
	parts := strings.Split(strings.TrimPrefix(routePrefix, "/tools/"), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0] + "." + toolName
	}
	return toolName
}

// parseToolName parses a namespaced tool name into contract and tool parts
// e.g., "execute.execute_py" -> ("execute", "execute_py")
func parseToolName(toolName string) (contractID string, tool string, ok bool) {
	parts := strings.SplitN(toolName, ".", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// bigInt helper
func bigInt(n int64) sdkmath.Int {
	return sdkmath.NewInt(n)
}

// Revert reverts the transaction with an error message
func Revert(msg string) {
	wasmx.Revert([]byte(msg))
}

// LoggerInfo logs info message
func LoggerInfo(msg string, args []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, args)
}

// LoggerError logs error message
func LoggerError(msg string, args []string) {
	wasmx.LoggerError(MODULE_NAME, msg, args)
}

// LoggerDebug logs debug message
func LoggerDebug(msg string, args []string) {
	wasmx.LoggerDebug(MODULE_NAME, msg, args)
}

// LoggerDebugExtended logs extended debug message
func LoggerDebugExtended(msg string, args []string) {
	wasmx.LoggerDebugExtended(MODULE_NAME, msg, args)
}
