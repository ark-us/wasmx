package main

import (
	"encoding/json"
	"os"

	httpserver "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-mcp/lib"
)

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
	// these entrypoints are internal by default
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

	// Handle empty data (e.g., during bootstrap/initialization)
	if len(databz) == 0 {
		wasmx.SetFinishData([]byte(`{"success": true}`))
		return
	}

	calldata := &lib.CallData{}
	if err := json.Unmarshal(databz, calldata); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(databz))
	}

	// Public operations
	switch {
	case calldata.GetParams != nil:
		res := lib.GetParams(*calldata.GetParams)
		wasmx.SetFinishData(res)
		return
	case calldata.ConnectDatabase != nil:
		res := lib.ConnectDatabase(calldata.ConnectDatabase)
		wasmx.SetFinishData(res)
		return
	case calldata.StartServer != nil:
		lib.StartServer(calldata.StartServer)
		wasmx.SetFinishData([]byte(`{"success": true}`))
		return
	}

	// Internal operations
	switch {
	case calldata.InitGenesis != nil:
		wasmx.OnlyInternal(lib.MODULE_NAME, "InitGenesis")
		res := lib.InitGenesis(*calldata.InitGenesis)
		wasmx.SetFinishData(res)
		return
	case calldata.RoleChanged != nil:
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_ROLES, "RoleChanged")
		// Role changed - initialize database tables
		res := lib.InitializeTables()
		wasmx.SetFinishData(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}

func handleHttpRequest() {
	// Get the HTTP request data from call data
	databz := wasmx.GetCallData()
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

	wasmx.SetFinishData(respBz)
}
