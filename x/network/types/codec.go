package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgPing2{}, "network/MsgPing2", nil)
	cdc.RegisterConcrete(&MsgGrpcSendRequest{}, "network/MsgGrpcSendRequest", nil)

	cdc.RegisterConcrete(&RequestPing{}, "network/RequestPing", nil)
	cdc.RegisterConcrete(&RequestBroadcastTx{}, "network/RequestBroadcastTx", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgPing2{},
		&MsgGrpcSendRequest{},

		&RequestPing{},
		&RequestBroadcastTx{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// TODO better solution?
var Network_Msg_serviceDesc = _Msg_serviceDesc
