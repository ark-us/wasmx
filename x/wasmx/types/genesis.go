package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/tendermint/tendermint/libs/rand"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, systemContracts []SystemContract, bootstrapAccount sdk.AccAddress) GenesisState {
	return GenesisState{
		Params:                  params,
		SystemContracts:         systemContracts,
		BootstrapAccountAddress: bootstrapAccount.String(),
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                  DefaultParams(),
		SystemContracts:         DefaultSystemContracts(),
		BootstrapAccountAddress: sdk.AccAddress(rand.Bytes(address.Len)).String(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	for _, contract := range gs.SystemContracts {
		if err := contract.Validate(); err != nil {
			return err
		}
	}
	return nil
}
