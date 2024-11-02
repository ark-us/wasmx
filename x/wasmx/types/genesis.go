package types

import (
	mcodec "mythos/v1/codec"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, systemContracts []SystemContract, bootstrapAccountBech32 string) GenesisState {
	return GenesisState{
		Params:                  params,
		SystemContracts:         systemContracts,
		BootstrapAccountAddress: bootstrapAccountBech32,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState(accBech32Codec mcodec.AccBech32Codec, bootstrapAccountBech32 string, feeCollectorBech32 string, mintBech32 string, minValidatorCount int32, enableEIDCheck bool, initialPortValues string) *GenesisState {
	return &GenesisState{
		Params:                  DefaultParams(),
		SystemContracts:         DefaultSystemContracts(accBech32Codec, feeCollectorBech32, mintBech32, minValidatorCount, enableEIDCheck, initialPortValues),
		BootstrapAccountAddress: bootstrapAccountBech32,
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
