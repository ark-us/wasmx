package main

import (
	"encoding/json"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-oauth2-server/lib"
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
	case "instantiate":
		// Nothing to initialize on instantiate
		return
	}

	databz := wasmx.GetCallData()

	// Handle empty data
	if len(databz) == 0 {
		wasmx.Finish([]byte(`{\"success\": true}`))
		return
	}

	calldata := &lib.CallData{}
	if err := json.Unmarshal(databz, calldata); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(databz))
	}

	// Handle operations
	switch {
	// OAuth client management
	case calldata.RegisterOAuthClient != nil:
		res := lib.RegisterOAuthClient(*calldata.RegisterOAuthClient)
		wasmx.Finish(res)
		return
	case calldata.UpdateOAuthClient != nil:
		res := lib.UpdateOAuthClient(*calldata.UpdateOAuthClient)
		wasmx.Finish(res)
		return
	case calldata.RevokeOAuthClient != nil:
		res := lib.RevokeOAuthClient(*calldata.RevokeOAuthClient)
		wasmx.Finish(res)
		return
	case calldata.GetOAuthClient != nil:
		res := lib.GetOAuthClient(*calldata.GetOAuthClient)
		wasmx.Finish(res)
		return
	case calldata.ListOAuthClients != nil:
		res := lib.ListOAuthClients(*calldata.ListOAuthClients)
		wasmx.Finish(res)
		return

	// User account management
	case calldata.RegisterUser != nil:
		res := lib.RegisterUser(*calldata.RegisterUser)
		wasmx.Finish(res)
		return
	case calldata.Login != nil:
		res := lib.Login(*calldata.Login)
		wasmx.Finish(res)
		return
	case calldata.Logout != nil:
		res := lib.Logout(*calldata.Logout)
		wasmx.Finish(res)
		return
	case calldata.GetCurrentUser != nil:
		res := lib.GetCurrentUser(*calldata.GetCurrentUser)
		wasmx.Finish(res)
		return

	// OAuth2 flow operations
	case calldata.CreateAuthorizationCode != nil:
		res := lib.CreateAuthorizationCode(*calldata.CreateAuthorizationCode)
		wasmx.Finish(res)
		return
	case calldata.ExchangeCodeForToken != nil:
		res := lib.ExchangeCodeForToken(*calldata.ExchangeCodeForToken)
		wasmx.Finish(res)
		return
	case calldata.ValidateAccessToken != nil:
		res := lib.ValidateAccessToken(*calldata.ValidateAccessToken)
		wasmx.Finish(res)
		return
	case calldata.RefreshAccessToken != nil:
		res := lib.RefreshAccessToken(*calldata.RefreshAccessToken)
		wasmx.Finish(res)
		return
	case calldata.ValidateSession != nil:
		res := lib.ValidateSession(*calldata.ValidateSession)
		wasmx.Finish(res)
		return

	// Internal operations
	case calldata.InitGenesis != nil:
		res := lib.InitGenesis(*calldata.InitGenesis)
		wasmx.Finish(res)
		return
	case calldata.RoleChanged != nil:
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_ROLES, "RoleChanged")
		wasmx.Finish([]byte(`{\"success\": true}`))
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}
