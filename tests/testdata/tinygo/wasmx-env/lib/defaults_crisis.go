package wasmx

import (
	"encoding/json"

	sdkmath "cosmossdk.io/math"
)

const CONSTANT_FEE uint64 = 1000

type CrisisGenesisState struct {
	ConstantFee Coin `json:"constant_fee"`
}

func GetDefaultCrisisGenesis(bondBaseDenom string) []byte {
	amt := sdkmath.NewIntFromUint64(CONSTANT_FEE)
	gs := CrisisGenesisState{ConstantFee: Coin{Denom: bondBaseDenom, Amount: amt}}
	bz, _ := json.Marshal(&gs)
	return bz
}
