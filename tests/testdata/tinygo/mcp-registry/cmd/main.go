package main

import (
	"encoding/json"
	"fmt"
	"os"

	httpserver "github.com/loredanacirstea/wasmx-env-httpserver/lib"
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

//go:wasm-module httpserver
//export wasmx_httpserver_i64_1
func Wasmx_httpserver_i64_1() {}

//go:wasm-module oauth2server
//export wasmx_oauth2server_i64_1
func Wasmx_oauth2server_i64_1() {}

//go:wasm-module postgresql
//export wasmx_postgresql_i64_1
func Wasmx_postgresql_i64_1() {}

func main() {
	// Handle internal entry points
	entrypoint := os.Getenv("ENTRY_POINT")
	switch entrypoint {
	case "http_request_incoming":
		handleHttpRequest()
		return
	case "instantiate":
		lib.InitializeTables()
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
	case calldata.ConnectDatabase != nil:
		res := lib.ConnectDatabase(calldata.ConnectDatabase)
		wasmx.Finish(res)
		return
	case calldata.InitTables != nil:
		res := lib.InitializeTables()
		wasmx.Finish(res)
		return
	case calldata.StartServer != nil:
		lib.StartServer(calldata.StartServer)
		wasmx.Finish([]byte(`{"success": true}`))
		return

	// Internal operations
	case calldata.InitGenesis != nil:
		res := lib.InitGenesis(*calldata.InitGenesis)
		wasmx.Finish(res)
		return
	case calldata.RoleChanged != nil:
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_ROLES, "RoleChanged")
		res := lib.InitializeTables()
		wasmx.Finish(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}

func handleHttpRequest() {
	fmt.Println("--handleHttpRequest--")
	// Get the HTTP request data from call data
	databz := wasmx.GetCallData()
	fmt.Println("--handleHttpRequest--", string(databz))
	var req httpserver.HttpRequestIncoming
	if err := json.Unmarshal(databz, &req); err != nil {
		lib.LoggerError("Failed to parse HTTP request", []string{"error", err.Error()})
		lib.Revert("invalid HTTP request: " + err.Error())
	}

	// Handle the request
	resp := lib.HandleHttpRequest(&req)

	// Marshal and return the response
	respBz, err := json.Marshal(resp)
	if err != nil {
		lib.LoggerError("Failed to marshal HTTP response", []string{"error", err.Error()})
		lib.Revert("failed to marshal response: " + err.Error())
	}

	wasmx.Finish(respBz)
}
