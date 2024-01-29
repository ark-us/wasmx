package types

import (
	stakingmod "github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(staking stakingtypes.GenesisState) *GenesisState {
	return &GenesisState{
		Staking: staking,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Staking: *stakingtypes.DefaultGenesisState(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := stakingmod.ValidateGenesis(&gs.Staking); err != nil {
		return err
	}
	return nil
}
