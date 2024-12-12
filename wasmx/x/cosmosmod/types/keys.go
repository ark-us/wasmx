package types

import (
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "cosmosmod"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const (
	paramsPrefix = iota + 1
	otherPrefix
)

var (
	KeyParamsPrefix = []byte{paramsPrefix}
	KeyOtherPrefix  = []byte{otherPrefix}
)

func BankModuleName() string {
	return ModuleName + "_" + wasmxtypes.ROLE_BANK
}

func AuthModuleName() string {
	return ModuleName + "_" + wasmxtypes.ROLE_AUTH
}

func GovModuleName() string {
	return ModuleName + "_" + wasmxtypes.ROLE_GOVERNANCE
}

func StakingModuleName() string {
	return ModuleName + "_" + wasmxtypes.ROLE_STAKING
}

func SlashingModuleName() string {
	return ModuleName + "_" + wasmxtypes.ROLE_SLASHING
}

func DistributionModuleName() string {
	return ModuleName + "_" + wasmxtypes.ROLE_DISTRIBUTION
}
