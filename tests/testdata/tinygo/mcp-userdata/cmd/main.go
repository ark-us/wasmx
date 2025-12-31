package main

import (
	"encoding/json"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-mcp-userdata/lib"
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
		// Called either:
		// 1. Directly by HTTP server for GET/read requests
		// 2. Via transaction signed by user's ephemeral key for POST/write requests
		// OAuth token validation inside HandleHttpRequest ensures user authorization
		res := lib.HandleHttpRequest(*calldata.HttpRequestHandler)
		wasmx.Finish(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}
