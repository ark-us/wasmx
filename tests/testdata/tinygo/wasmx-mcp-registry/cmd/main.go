package main

import (
	"encoding/json"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-mcp-registry/lib"
)

//go:wasm-module wasmxcore
//export wasmx_nondeterministic_1
func Wasmx_nondeterministic_1() {}

//go:wasm-module wasmxcore
//export wasmx_env_core_i64_1
func Wasmx_env_core_i64_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module postgresql
//export wasmx_postgresql_i64_1
func Wasmx_postgresql_i64_1() {}

func main() {
	// Handle internal entry points
	entrypoint := os.Getenv("ENTRY_POINT")
	switch entrypoint {
	case "instantiate":
		// Check if there's initialization data in the call data
		databz := wasmx.GetCallData()
		if len(databz) > 0 {
			var calldata lib.CallData
			if err := json.Unmarshal(databz, &calldata); err == nil && calldata.InitGenesis != nil {
				// Store the init data first
				lib.InitGenesis(*calldata.InitGenesis)
				// Initialize tables and register HTTP routes
				// This is called during chain initialization when the contract is activated
				lib.InitializeTables()
			}
		}
		return
	}

	databz := wasmx.GetCallData()

	// Handle empty data
	if len(databz) == 0 {
		wasmx.Finish([]byte(`{"success": true}`))
		return
	}

	calldata := &lib.CallData{}
	if err := json.Unmarshal(databz, calldata); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(databz))
	}

	// Handle operations
	switch {
	// Registry operations
	case calldata.RegisterMCPContract != nil:
		res := lib.RegisterMCPContract(*calldata.RegisterMCPContract)
		wasmx.Finish(res)
		return
	case calldata.DeregisterMCPContract != nil:
		res := lib.DeregisterMCPContract(*calldata.DeregisterMCPContract)
		wasmx.Finish(res)
		return
	case calldata.UpdateMCPTools != nil:
		res := lib.UpdateMCPTools(*calldata.UpdateMCPTools)
		wasmx.Finish(res)
		return

	// Query operations
	case calldata.GetRegisteredContracts != nil:
		res := lib.GetRegisteredContracts(*calldata.GetRegisteredContracts)
		wasmx.Finish(res)
		return
	case calldata.GetContractInfo != nil:
		res := lib.GetContractInfo(*calldata.GetContractInfo)
		wasmx.Finish(res)
		return
	case calldata.GetAllTools != nil:
		res := lib.GetAllTools(*calldata.GetAllTools)
		wasmx.Finish(res)
		return

	// Server management operations
	case calldata.GetParams != nil:
		res := lib.GetParams(*calldata.GetParams)
		wasmx.Finish(res)
		return
	case calldata.RegisterHttpRoutes != nil:
		res := lib.RegisterHttpRoutes(calldata.RegisterHttpRoutes)
		wasmx.Finish(res)
		return

	// Internal operations
	case calldata.RoleChanged != nil:
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_ROLES, "RoleChanged")
		res := lib.InitializeTables()
		wasmx.Finish(res)
		return
	case calldata.HttpRequestHandler != nil:
		// Called by HTTP server registry contract with HTTP request data
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_HTTP_SERVER, string(databz))
		res := lib.HandleHttpRequest(*calldata.HttpRequestHandler)
		wasmx.Finish(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}
