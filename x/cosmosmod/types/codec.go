package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	govtypes.RegisterLegacyAminoCodec(cdc)
	banktypes.RegisterLegacyAminoCodec(cdc)
	authtypes.RegisterLegacyAminoCodec(cdc)
	stakingtypes.RegisterLegacyAminoCodec(cdc)
	slashingtypes.RegisterLegacyAminoCodec(cdc)
	distributiontypes.RegisterLegacyAminoCodec(cdc)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	govtypes.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	authtypes.RegisterInterfaces(registry)
	stakingtypes.RegisterInterfaces(registry)
	slashingtypes.RegisterInterfaces(registry)
	distributiontypes.RegisterInterfaces(registry)
}
