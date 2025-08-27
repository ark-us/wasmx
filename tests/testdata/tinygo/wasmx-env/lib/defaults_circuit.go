package wasmx

import "encoding/json"

// GenesisAccountPermissions placeholder
type GenesisAccountPermissions struct{}

type CircuitGenesisState struct {
	AccountPermissions []GenesisAccountPermissions `json:"account_permissions"`
	DisabledTypeUrls   []string                    `json:"disabled_type_urls"`
}

func GetDefaultCircuitGenesis() []byte {
	gs := CircuitGenesisState{AccountPermissions: []GenesisAccountPermissions{}, DisabledTypeUrls: []string{}}
	bz, _ := json.Marshal(&gs)
	return bz
}
