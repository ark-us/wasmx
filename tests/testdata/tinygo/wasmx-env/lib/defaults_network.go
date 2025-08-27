package wasmx

import "encoding/json"

type NetworkGenesisState struct{}

func GetDefaultNetworkGenesis() []byte { bz, _ := json.Marshal(&NetworkGenesisState{}); return bz }
