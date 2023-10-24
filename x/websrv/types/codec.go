package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterOAuthClient{}, "websrv/MsgRegisterOAuthClient", nil)
	cdc.RegisterConcrete(&MsgDeregisterOAuthClient{}, "websrv/MsgDeregisterOAuthClient", nil)
	cdc.RegisterConcrete(&MsgEditOAuthClient{}, "websrv/MsgEditOAuthClient", nil)
	cdc.RegisterConcrete(&MsgRegisterRoute{}, "websrv/MsgRegisterRoute", nil)
	cdc.RegisterConcrete(&MsgDeregisterRoute{}, "websrv/MsgDeregisterRoute", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterOAuthClient{},
		&MsgDeregisterOAuthClient{},
		&MsgEditOAuthClient{},
		&MsgRegisterRoute{},
		&MsgDeregisterRoute{},
	)

	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&RegisterRouteProposal{},
		&DeregisterRouteProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
