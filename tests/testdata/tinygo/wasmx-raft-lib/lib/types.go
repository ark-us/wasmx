package lib

import (
    staking "github.com/loredanacirstea/wasmx-staking/lib"
)

// types.ts (AS) equivalents

type CosmosmodGenesisState struct {
    Staking staking.GenesisState `json:"staking"`
}

type IsNodeValidator struct {
    IsValidator bool `json:"isvalidator"`
    NodeIndex   int32 `json:"nodeIndex"`
}
