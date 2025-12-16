package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	wasmxhttp "github.com/loredanacirstea/wasmx-env-httpserver/lib"
)

// RouteCallData routes incoming call data to the appropriate handler
func RouteCallData(data []byte) []byte {
	// Check if this is an HTTP request
	var httpReq wasmxhttp.HttpRequestIncoming
	if err := json.Unmarshal(data, &httpReq); err == nil && httpReq.RequestURI != "" {
		return HandleHTTPRequest(&httpReq)
	}

	var callData CallData
	if err := json.Unmarshal(data, &callData); err != nil {
		LoggerError("Failed to unmarshal calldata", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "invalid calldata"})
	}

	// Handle initialization
	if callData.InitGenesis != nil {
		return InitGenesis(callData.InitGenesis)
	}

	// Handle key generation and management
	if callData.GenerateEphemeralKey != nil {
		return GenerateEphemeralKey(callData.GenerateEphemeralKey)
	}

	if callData.RegisterExternalKey != nil {
		return RegisterExternalKey(callData.RegisterExternalKey)
	}

	if callData.RevokeKey != nil {
		return RevokeKey(callData.RevokeKey)
	}

	if callData.DeleteExpiredKeys != nil {
		return DeleteExpiredKeys(callData.DeleteExpiredKeys)
	}

	if callData.InitAccount != nil {
		return InitAccount(callData.InitAccount)
	}

	// Handle queries
	if callData.QueryGetPublicKey != nil {
		return QueryGetPublicKey(callData.QueryGetPublicKey)
	}

	if callData.QueryGetKeyInfo != nil {
		return QueryGetKeyInfo(callData.QueryGetKeyInfo)
	}

	if callData.QueryValidateAndGetKey != nil {
		return QueryValidateAndGetKey(callData.QueryValidateAndGetKey)
	}

	Revert("invalid call data: " + string(data))
	return []byte{}
}

// Main entry point for the contract
func Execute(data []byte) []byte {
	return RouteCallData(data)
}

// Query entry point (read-only operations)
func Query(data []byte) []byte {
	return RouteCallData(data)
}

// OnRoleChanged is called when the contract is assigned a role
func OnRoleChanged() []byte {
	role := string(wasmx.GetEnv())
	LoggerInfo("Role changed", []string{"role", role})

	// OAuth2 Keys contract can be notified when assigned ROLE_OAUTH2_KEYS
	if role == wasmx.ROLE_OAUTH2_KEYS {
		LoggerInfo("OAuth2 Keys contract role assigned", nil)

		// Register HTTP routes with HTTP registry
		httpRegistryAddr := wasmx.GetAddressByRole(wasmx.ROLE_HTTP_SERVER)
		registerHttpRoutes(httpRegistryAddr)
	}

	return MarshalJSON(map[string]bool{"success": true})
}
