package main

import (
	"encoding/json"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-httpserver-registry/lib"
)

// host modules
//
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

func main() {
	entrypoint := os.Getenv("ENTRY_POINT")
	switch entrypoint {
	case "http_request_incoming":
		lib.HandleHttpRequestIncoming()
		return
	case "instantiate":
		// nothing to do
		return
	}

	calldataBz := wasmx.GetCallData()
	if len(calldataBz) == 0 {
		wasmx.Finish([]byte(`{"success":true}`))
		return
	}

	var calldata lib.CallData
	if err := json.Unmarshal(calldataBz, &calldata); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(calldataBz))
	}

	switch {
	case calldata.ConnectDatabase != nil:
		res := lib.ConnectDatabase(*calldata.ConnectDatabase)
		wasmx.Finish(res)
		return
	case calldata.SetRoute != nil:
		wasmx.OnlyInternal(lib.MODULE_NAME, "SetRoute")
		res := lib.SetRoute(*calldata.SetRoute)
		wasmx.Finish(res)
		return
	case calldata.RemoveRoute != nil:
		wasmx.OnlyInternal(lib.MODULE_NAME, "SetRoute")
		res := lib.RemoveRoute(*calldata.RemoveRoute)
		wasmx.Finish(res)
		return
	case calldata.GetRoute != nil:
		res := lib.GetRoute(*calldata.GetRoute)
		wasmx.Finish(res)
		return
	case calldata.GetRoutes != nil:
		res := lib.GetRoutes()
		wasmx.Finish(res)
		return
	case calldata.StartWebServer != nil:
		lib.LoggerInfo("StartWebServer called", nil)
		res := lib.StartWebServer(*calldata.StartWebServer)
		wasmx.Finish(res)
		return
	case calldata.CloseServer != nil:
		res := lib.CloseServer()
		wasmx.Finish(res)
		return
	case calldata.RoleChanged != nil:
		// Only roles module can invoke
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_ROLES, "RoleChanged")
		wasmx.Finish([]byte(`{"success":true}`))
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), calldataBz...))
}
