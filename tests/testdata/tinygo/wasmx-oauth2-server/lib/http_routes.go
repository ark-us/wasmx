package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func registerHttpRoutes(registryAddr wasmx.Bech32String, staticRoutes []StaticRoute) {
	self := string(wasmx.GetAddress())

	// Get HTTP registry address by role if not provided
	var httpRegistryAddr wasmx.Bech32String
	if registryAddr != "" {
		httpRegistryAddr = registryAddr
	} else {
		httpRegistryAddr = wasmx.GetAddressByRole(wasmx.ROLE_HTTP_SERVER)
	}

	routes := []string{
		"/.well-known/oauth-authorization-server",
		"/.well-known/oauth-authorization-server/sse",
		"/.well-known/openid-configuration",
		"/.well-known/openid-configuration/sse",
		"/.well-known/oauth-protected-resource",
		"/.well-known/oauth-protected-resource/sse",
		"/oauth/authorize",
		"/oauth/token",
		"/oauth/userinfo",
		"/register",
		"/register_tx",
		"/login",
		"/logout",
		"/auth/register",
		"/auth/register_init",
		"/auth/account_info",
		"/auth/login",
		"/auth/logout",
		"/auth/me",
		"/auth/contract_addresses",
	}

	for _, rt := range routes {
		msg := map[string]interface{}{
			"set_route": map[string]interface{}{
				"route":            rt,
				"contract_address": self,
			},
		}
		bz, _ := json.Marshal(msg)
		ok, data := wasmx.CallSimple(httpRegistryAddr, bz, false, MODULE_NAME)
		if !ok {
			// log and continue; don't hard fail init
			LoggerError("failed to set route", []string{"route", rt, "error", string(data)})
		}
	}

	// Register static routes
	for _, sr := range staticRoutes {
		msg := map[string]interface{}{
			"set_static_route": map[string]interface{}{
				"route":       sr.Route,
				"folder_path": sr.FolderPath,
			},
		}
		bz, _ := json.Marshal(msg)
		ok, data := wasmx.CallSimple(httpRegistryAddr, bz, false, MODULE_NAME)
		if !ok {
			LoggerError("failed to set static route", []string{"route", sr.Route, "folder", sr.FolderPath, "error", string(data)})
		} else {
			LoggerInfo("Static route registered", []string{"route", sr.Route, "folder", sr.FolderPath})
		}
	}

	LoggerInfo("OAuth2 HTTP routes registered", []string{"registry", string(httpRegistryAddr)})
}
