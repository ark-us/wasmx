package main

import (
	"encoding/json"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-mcp-search/lib"
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
		// Store initialization data
		databz := wasmx.GetCallData()
		if len(databz) > 0 {
			var calldata lib.CallData
			if err := json.Unmarshal(databz, &calldata); err == nil && calldata.InitGenesis != nil {
				lib.InitGenesis(*calldata.InitGenesis)
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
	case calldata.ExecuteTool != nil:
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_MCP_REGISTRY, string(databz))
		res := lib.ExecuteTool(*calldata.ExecuteTool)
		wasmx.Finish(res)
		return
	case calldata.RoleChanged != nil:
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_ROLES, "RoleChanged")
		res := lib.OnRoleChanged()
		wasmx.Finish(res)
		return
	case calldata.HttpRequestHandler != nil:
		// Called by HTTP server registry contract with HTTP request data
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_HTTP_SERVER, string(databz))
		res := lib.HandleHttpRequest(*calldata.HttpRequestHandler)
		wasmx.Finish(res)
		return
	case calldata.Populate != nil:
		// Populate with initial/test data - no auth required
		// TODO: In production, consider adding admin role check or removing this endpoint
		res := lib.Populate(*calldata.Populate)
		wasmx.Finish(res)
		return
	case calldata.GenerateEmbeddings != nil:
		// Generate embeddings for data - no auth required for now
		// TODO: In production, consider adding admin role check
		res := lib.GenerateEmbeddings(*calldata.GenerateEmbeddings)
		wasmx.Finish(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}
