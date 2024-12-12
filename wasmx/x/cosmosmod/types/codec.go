package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	govtypes.RegisterLegacyAminoCodec(cdc)
	banktypes.RegisterLegacyAminoCodec(cdc)
	authtypes.RegisterLegacyAminoCodec(cdc)
	stakingtypes.RegisterLegacyAminoCodec(cdc)
	slashingtypes.RegisterLegacyAminoCodec(cdc)
	distributiontypes.RegisterLegacyAminoCodec(cdc)

	cdc.RegisterInterface((*mcodec.ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*mcodec.AccountI)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmosmod/BaseAccount", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmosmod/ModuleAccount", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	govtypes.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	authtypes.RegisterInterfaces(registry)
	stakingtypes.RegisterInterfaces(registry)
	slashingtypes.RegisterInterfaces(registry)
	distributiontypes.RegisterInterfaces(registry)

	// auth module
	registry.RegisterInterface(
		"mythos.cosmosmod.v1.AccountI",
		(*mcodec.AccountI)(nil),
		&BaseAccount{},
		&ModuleAccount{},
	)
	registry.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*sdk.AccountI)(nil),
		&BaseAccount{},
		&ModuleAccount{},
	)
}
