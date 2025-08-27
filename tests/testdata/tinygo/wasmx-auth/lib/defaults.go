package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env"
)

// Defaults for auth module (converted from AS defaults.ts)

const (
	MaxMemoCharacters      = uint64(256)
	TxSigLimit             = uint64(7)
	TxSizeCostPerByte      = uint64(10)
	SigVerifyCostEd25519   = uint64(590)
	SigVerifyCostSecp256k1 = uint64(1000)
)

func GetDefaultParams() Params {
	return Params{
		MaxMemoCharacters:      MaxMemoCharacters,
		TxSigLimit:             TxSigLimit,
		TxSizeCostPerByte:      TxSizeCostPerByte,
		SigVerifyCostEd25519:   SigVerifyCostEd25519,
		SigVerifyCostSecp256k1: SigVerifyCostSecp256k1,
	}
}

func GetDefaultGenesis() GenesisState {
	params := GetDefaultParams()
	return GenesisState{
		Params:               params,
		Accounts:             []wasmx.AnyWrap{},
		BaseAccountTypeurl:   TypeUrl_BaseAccount,
		ModuleAccountTypeurl: TypeUrl_ModuleAccount,
	}
}
