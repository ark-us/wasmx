package types

import (
	"encoding/json"

	codec "github.com/cosmos/cosmos-sdk/codec"
	exported "github.com/cosmos/cosmos-sdk/x/bank/exported"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

type QueryAddressByDenomResponse struct {
	Address mcodec.AccAddressPrefixed `json:"address"`
}

// GenesisBalancesIterator implements genesis account iteration.
type GenesisBalancesIterator struct{}

// IterateGenesisBalances iterates over all the genesis balances found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisBalancesIterator) IterateGenesisBalances(
	cdc codec.JSONCodec, appState map[string]json.RawMessage, cb func(exported.GenesisBalance) (stop bool),
) {
	for _, balance := range GetGenesisStateFromAppState(cdc, appState).Bank.Balances {
		if cb(balance) {
			break
		}
	}
}

// GetGenesisStateFromAppState returns x/bank GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}
