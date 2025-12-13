package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// RouteCallData routes incoming call data to the appropriate handler
func RouteCallData(data []byte) []byte {
	var callData CallData
	if err := json.Unmarshal(data, &callData); err != nil {
		LoggerError("Failed to unmarshal calldata", []string{"error", err.Error()})
		return MarshalJSON(map[string]string{"error": "invalid calldata"})
	}

	// Handle initialization
	if callData.InitGenesis != nil {
		return InitGenesis(callData.InitGenesis)
	}

	// Handle user registration and management
	if callData.RegisterUser != nil {
		return RegisterUser(callData.RegisterUser)
	}

	if callData.AddAddress != nil {
		return AddAddress(callData.AddAddress)
	}

	if callData.RemoveAddress != nil {
		return RemoveAddress(callData.RemoveAddress)
	}

	if callData.UpdatePermissions != nil {
		return UpdatePermissions(callData.UpdatePermissions)
	}

	if callData.SetPrimaryAddress != nil {
		return SetPrimaryAddress(callData.SetPrimaryAddress)
	}

	// Handle queries
	if callData.QueryUserByID != nil {
		return QueryUserByID(callData.QueryUserByID)
	}

	if callData.QueryUserByAddress != nil {
		return QueryUserByAddress(callData.QueryUserByAddress)
	}

	if callData.QueryAddressInfo != nil {
		return QueryAddressInfo(callData.QueryAddressInfo)
	}

	if callData.QueryValidatePermission != nil {
		return QueryValidatePermission(callData.QueryValidatePermission)
	}

	LoggerError("Unknown calldata", nil)
	return MarshalJSON(map[string]string{"error": "unknown calldata"})
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

	// Identity contract can be notified when assigned ROLE_IDENTITY
	if role == wasmx.ROLE_IDENTITY {
		LoggerInfo("Identity contract role assigned", nil)
	}

	return MarshalJSON(map[string]bool{"success": true})
}
