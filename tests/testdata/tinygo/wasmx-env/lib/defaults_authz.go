package wasmx

import "encoding/json"

// GrantAuthorization mirrors AS placeholder
type GrantAuthorization struct{}

// AuthzGenesisState mirrors AS GenesisState for authz module
type AuthzGenesisState struct {
	Authorization []GrantAuthorization `json:"authorization"`
}

func GetDefaultAuthzGenesis() []byte {
	gs := AuthzGenesisState{Authorization: []GrantAuthorization{}}
	bz, _ := json.Marshal(&gs)
	return bz
}
