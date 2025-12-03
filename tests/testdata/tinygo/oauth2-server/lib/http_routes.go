package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func registerHttpRoutes(registryAddr wasmx.Bech32String) {
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
		"/oauth/authorize",
		"/oauth/token",
		"/auth/register",
		"/auth/login",
		"/auth/logout",
		"/auth/me",
		"/login",
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

	LoggerInfo("OAuth2 HTTP routes registered", []string{"registry", string(httpRegistryAddr)})
}
