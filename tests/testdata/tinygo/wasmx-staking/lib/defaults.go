package lib

// Defaults for staking module (converted from AS defaults.ts)

const (
	UnbondingTime     = "1814400s"
	MaxValidators     = uint32(100)
	MaxEntries        = uint32(7)
	HistoricalEntries = uint32(10000)
	MinCommissionRate = "0.000000000000000000"
)

// GetDefaultParams returns default Params for staking with given bond denom
func GetDefaultParams(bondDenom string) Params {
	return Params{
		UnbondingTime:     UnbondingTime,
		MaxValidators:     MaxValidators,
		MaxEntries:        MaxEntries,
		HistoricalEntries: HistoricalEntries,
		BondDenom:         bondDenom,
		MinCommissionRate: MinCommissionRate,
	}
}

// GetDefaultGenesis builds a minimal staking genesis with params and base denom
func GetDefaultGenesis(baseDenom string, bondBaseDenom string) GenesisState {
	params := GetDefaultParams(bondBaseDenom)
	return GenesisState{
		Params:               params,
		LastTotalPower:       "0",
		LastValidatorPowers:  []LastValidatorPower{},
		Validators:           []Validator{},
		Delegations:          []Delegation{},
		UnbondingDelegations: []UnbondingDelegation{},
		Redelegations:        []Redelegation{},
		BaseDenom:            baseDenom,
	}
}
