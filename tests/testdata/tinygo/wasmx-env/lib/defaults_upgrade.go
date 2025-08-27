package wasmx

import "encoding/json"

type UpgradeGenesisState struct{}

func GetDefaultUpgradeGenesis() []byte { bz, _ := json.Marshal(&UpgradeGenesisState{}); return bz }
