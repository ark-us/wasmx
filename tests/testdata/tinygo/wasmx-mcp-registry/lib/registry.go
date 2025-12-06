package lib

import (
	"encoding/json"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// RegisterMCPContract registers a new MCP contract with the registry
func RegisterMCPContract(req RegisterMCPContractRequest) []byte {
	caller := wasmx.GetCaller()

	// Check mcp role - only contracts with mcp role can register themselves
	if !hasMCPRole(caller) {
		Revert("unauthorized: only mcp role can register mcp tool")
	}

	// Validate route prefix
	if !isValidRoutePrefix(req.RoutePrefix) {
		Revert("invalid route prefix: must start with /tools/ and be unique")
	}

	// Check route prefix not already taken
	if routeExists(req.RoutePrefix) {
		existingContract := getContractByRoute(req.RoutePrefix)
		if existingContract != req.ContractAddress {
			Revert("route prefix already registered by another contract")
		}
	}

	// Parse and validate tools JSON
	var tools []MCPToolDefinition
	if err := json.Unmarshal([]byte(req.ToolsJSON), &tools); err != nil {
		Revert("invalid tools JSON: " + err.Error())
	}

	// Validate tools
	if len(tools) == 0 {
		Revert("must register at least one tool")
	}
	for _, tool := range tools {
		if tool.Name == "" || tool.Description == "" {
			Revert("tool must have name and description")
		}
	}

	// Get current block height
	currentBlock := wasmx.GetCurrentBlock()

	// Create or update registration
	existingReg := getContractRegistration(req.ContractAddress)
	var registration ContractRegistration

	if existingReg != nil {
		// Update existing registration
		registration = *existingReg
		registration.RoutePrefix = req.RoutePrefix
		registration.Tools = tools
		registration.LastUpdatedAt = int64(currentBlock.Height)
		registration.Active = true
		registration.UseOAuth2 = req.UseOAuth2
	} else {
		// Create new registration
		registration = ContractRegistration{
			Address:       req.ContractAddress,
			RoutePrefix:   req.RoutePrefix,
			Tools:         tools,
			RegisteredAt:  int64(currentBlock.Height),
			LastUpdatedAt: int64(currentBlock.Height),
			Active:        true,
			UseOAuth2:     req.UseOAuth2,
		}
	}

	// Store registration
	storeContractRegistration(registration)
	addToRegisteredList(req.ContractAddress)
	storeRouteMapping(req.RoutePrefix, req.ContractAddress)

	// Update HTTP routing via SetRouteHandler
	updateHttpRoute(req.RoutePrefix, req.ContractAddress)

	// Emit event
	LoggerInfo("MCP contract registered", []string{
		"contract_address", req.ContractAddress,
		"route_prefix", req.RoutePrefix,
		"tool_count", fmt.Sprintf("%d", len(tools)),
	})

	return []byte(fmt.Sprintf(`{"success": true, "contract_address": "%s", "route_prefix": "%s"}`, req.ContractAddress, req.RoutePrefix))
}

// DeregisterMCPContract removes an MCP contract from the registry
func DeregisterMCPContract(req DeregisterMCPContractRequest) []byte {
	wasmx.OnlyRole(MODULE_NAME, wasmx.ROLE_MCP_REGISTRY, "DeregisterMCPContract")

	registration := getContractRegistration(req.ContractAddress)
	if registration == nil {
		Revert("contract not registered")
	}

	// Remove HTTP route
	removeHttpRoute(registration.RoutePrefix)

	// Remove storage
	deleteContractRegistration(req.ContractAddress)
	removeFromRegisteredList(req.ContractAddress)
	deleteRouteMapping(registration.RoutePrefix)

	// Emit event
	LoggerInfo("MCP contract deregistered", []string{
		"contract_address", req.ContractAddress,
	})

	return []byte(`{"success": true}`)
}

// UpdateMCPTools updates the tool list for a registered contract
func UpdateMCPTools(req UpdateMCPToolsRequest) []byte {
	caller := wasmx.GetCaller()

	registration := getContractRegistration(req.ContractAddress)
	if registration == nil {
		Revert("contract not registered")
	}

	// Authorization
	if !hasGovRole(caller) {
		Revert("unauthorized: only governance role can update tools")
	}

	// Parse new tools
	var tools []MCPToolDefinition
	if err := json.Unmarshal([]byte(req.ToolsJSON), &tools); err != nil {
		Revert("invalid tools JSON: " + err.Error())
	}

	if len(tools) == 0 {
		Revert("must have at least one tool")
	}

	// Update registration
	registration.Tools = tools
	currentBlock := wasmx.GetCurrentBlock()
	registration.LastUpdatedAt = int64(currentBlock.Height)

	storeContractRegistration(*registration)

	// Emit event
	LoggerInfo("MCP tools updated", []string{
		"contract_address", req.ContractAddress,
		"tool_count", fmt.Sprintf("%d", len(tools)),
	})

	return []byte(`{"success": true}`)
}

// GetRegisteredContracts returns all registered MCP contracts
func GetRegisteredContracts(_ GetRegisteredContractsRequest) []byte {
	addresses := getRegisteredList()
	contracts := []ContractRegistration{}

	for _, addr := range addresses {
		registration := getContractRegistration(addr)
		if registration != nil && registration.Active {
			contracts = append(contracts, *registration)
		}
	}

	data, _ := json.Marshal(contracts)
	return data
}

// GetContractInfo returns info for a specific contract
func GetContractInfo(req GetContractInfoRequest) []byte {
	registration := getContractRegistration(req.ContractAddress)
	if registration == nil {
		return []byte(`{"error": "contract not registered"}`)
	}

	data, _ := json.Marshal(registration)
	return data
}

// GetAllTools returns aggregated tool list (for MCP /tools/list)
func GetAllTools(_ GetAllToolsRequest) []byte {
	addresses := getRegisteredList()
	allTools := []ToolsListEntry{}

	for _, addr := range addresses {
		registration := getContractRegistration(addr)
		if registration == nil || !registration.Active {
			continue
		}

		for _, tool := range registration.Tools {
			// Prefix tool name with contract route to make it unique
			prefixedName := formatToolName(registration.RoutePrefix, tool.Name)

			allTools = append(allTools, ToolsListEntry{
				Name:            prefixedName,
				Description:     tool.Description,
				InputSchema:     tool.InputSchema,
				ContractAddress: registration.Address,
				RoutePrefix:     registration.RoutePrefix,
			})
		}
	}

	data, _ := json.Marshal(allTools)
	return data
}

// updateHttpRoute registers an HTTP route for a contract
func updateHttpRoute(routePrefix, contractAddress string) {
	// Use wildcard route to match all paths under the prefix
	route := routePrefix + "/*"

	// Get HTTP registry address by role
	httpRegistryAddr := wasmx.GetAddressByRole(wasmx.ROLE_HTTP_SERVER)

	// Call HTTP server registry contract to set the route
	msg := map[string]interface{}{
		"set_route": map[string]interface{}{
			"route":            route,
			"contract_address": contractAddress,
		},
	}
	msgBz, _ := json.Marshal(msg)
	ok, data := wasmx.CallSimple(httpRegistryAddr, msgBz, false, MODULE_NAME)
	if !ok {
		LoggerError("Failed to register HTTP route", []string{"route", route, "error", string(data)})
		return
	}

	LoggerInfo("HTTP route registered", []string{
		"route", route,
		"contract", contractAddress,
	})
}

// removeHttpRoute removes an HTTP route
func removeHttpRoute(routePrefix string) {
	route := routePrefix + "/*"

	// Get HTTP registry address by role
	httpRegistryAddr := wasmx.GetAddressByRole(wasmx.ROLE_HTTP_SERVER)

	// Call HTTP server registry contract to remove the route
	msg := map[string]interface{}{
		"remove_route": map[string]interface{}{
			"route": route,
		},
	}
	msgBz, _ := json.Marshal(msg)
	ok, data := wasmx.CallSimple(httpRegistryAddr, msgBz, false, MODULE_NAME)
	if !ok {
		LoggerError("Failed to remove HTTP route", []string{"route", route, "error", string(data)})
		return
	}

	LoggerInfo("HTTP route removed", []string{"route", route})
}

// Helper function to call a contract
func callContract(contractAddress string, callData []byte, gasLimit int64) (bool, []byte) {
	value := sdkmath.NewInt(0)
	gas := big.NewInt(gasLimit)
	success, responseData := wasmx.Call(
		wasmx.Bech32String(contractAddress),
		&value,
		callData,
		gas,
		MODULE_NAME,
	)
	return success, responseData
}
