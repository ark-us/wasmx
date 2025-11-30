package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-mcp-execute/lib"
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
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}
