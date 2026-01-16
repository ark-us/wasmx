package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

type denomUnit struct {
	Denom    string   `json:"denom"`
	Exponent int32    `json:"exponent"`
	Aliases  []string `json:"aliases"`
}

type denomMetadata struct {
	Description string      `json:"description"`
	DenomUnits  []denomUnit `json:"denom_units"`
	Base        string      `json:"base"`
	Display     string      `json:"display"`
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
	URI         string      `json:"uri"`
	URIHash     string      `json:"uri_hash"`
}

// OnRoleChanged registers the token denom with the bank when it receives ROLE_DENOM.
func OnRoleChanged() []byte {
	callData := wasmx.GetCallData()
	if len(callData) == 0 {
		return MarshalJSON(map[string]bool{"success": true})
	}

	var hook struct {
		RoleChanged *wasmx.RolesChangedHook `json:"RoleChanged"`
	}
	if err := json.Unmarshal(callData, &hook); err != nil {
		LoggerError("Failed to unmarshal RoleChanged hook", []string{"error", err.Error()})
		return MarshalJSON(map[string]bool{"success": true})
	}
	if hook.RoleChanged == nil || hook.RoleChanged.RoleChanged == nil {
		return MarshalJSON(map[string]bool{"success": true})
	}

	role := hook.RoleChanged.RoleChanged.Role
	LoggerInfo("Role changed", []string{"role", role, "label", hook.RoleChanged.RoleChanged.Label})

	if role != wasmx.ROLE_DENOM {
		return MarshalJSON(map[string]bool{"success": true})
	}

	info := GetInfo()
	if info.Symbol == "" {
		Revert("token info not initialized")
	}

	base := info.Symbol
	units := []denomUnit{
		{Denom: base, Exponent: 0, Aliases: []string{}},
	}

	meta := denomMetadata{
		Description: "",
		DenomUnits:  units,
		Base:        base,
		Display:     base,
		Name:        info.Name,
		Symbol:      info.Symbol,
		URI:         "",
		URIHash:     "",
	}

	bankAddr := wasmx.GetAddressByRole(wasmx.ROLE_BANK)
	if bankAddr == "" {
		Revert("bank role not set")
	}

	payload := map[string]interface{}{
		"RegisterDenom": map[string]interface{}{
			"contract": string(wasmx.GetAddress()),
			"metadata": meta,
		},
	}
	bz, _ := json.Marshal(payload)
	ok, resp := wasmx.CallSimple(bankAddr, bz, false, MODULE_NAME)
	if !ok {
		Revert("register denom failed: " + string(resp))
	}

	return MarshalJSON(map[string]bool{"success": true})
}
