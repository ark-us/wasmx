package wasmx

import "encoding/json"

type EvidenceGenesisState struct {
	Evidence []AnyWrap `json:"evidence"`
}

func GetDefaultEvidenceGenesis() []byte {
	gs := EvidenceGenesisState{Evidence: []AnyWrap{}}
	bz, _ := json.Marshal(&gs)
	return bz
}
