package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// cdc.RegisterConcrete(MsgGrpcSendRequest{}, "network/MsgGrpcSendRequest", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		// &MsgGrpcSendRequest{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_MsgStaking_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_MsgBank_serviceDesc)
}
