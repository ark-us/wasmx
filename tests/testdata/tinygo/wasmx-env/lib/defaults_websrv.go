package wasmx

import "encoding/json"

type WebsrvParams struct {
	OAuthClientRegistrationOnlyEID bool `json:"oauth_client_registration_only_e_id"`
}

type WebsrvGenesisState struct {
	Params WebsrvParams `json:"params"`
}

func GetDefaultWebsrvGenesis() []byte {
	gs := WebsrvGenesisState{Params: WebsrvParams{OAuthClientRegistrationOnlyEID: false}}
	bz, _ := json.Marshal(&gs)
	return bz
}
