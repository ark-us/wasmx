package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/types/tx"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgStoreCode{}, "wasmx/MsgStoreCode", nil)
	cdc.RegisterConcrete(&MsgInstantiateContract{}, "wasmx/MsgInstantiateContract", nil)
	cdc.RegisterConcrete(&MsgInstantiateContract2{}, "wasmx/MsgInstantiateContract2", nil)
	cdc.RegisterConcrete(&MsgExecuteContract{}, "wasmx/MsgExecuteContract", nil)
	cdc.RegisterConcrete(&MsgCompileContract{}, "wasmx/MsgCompileContract", nil)
	cdc.RegisterConcrete(&MsgExecuteEth{}, "wasmx/MsgExecuteEth", nil)
	cdc.RegisterConcrete(&MsgRegisterRole{}, "wasmx/MsgRegisterRole", nil)
	cdc.RegisterConcrete(&MsgDeregisterRole{}, "wasmx/MsgDeregisterRole", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgStoreCode{},
		&MsgInstantiateContract{},
		&MsgInstantiateContract2{},
		&MsgExecuteContract{},
		&MsgCompileContract{},
		&MsgExecuteEth{},
		&MsgRegisterRole{},
		&MsgDeregisterRole{},
	)

	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&RegisterRoleProposal{},
		&DeregisterRoleProposal{},
	)

	registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&ExtensionOptionEthereumTx{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
