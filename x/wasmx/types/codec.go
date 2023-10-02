package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	// this line is used by starport scaffolding # 1
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgStoreCode{}, "wasmx/MsgStoreCode")
	legacy.RegisterAminoMsg(cdc, &MsgInstantiateContract{}, "wasmx/MsgInstantiateContract")
	legacy.RegisterAminoMsg(cdc, &MsgInstantiateContract2{}, "wasmx/MsgInstantiateContract2")
	legacy.RegisterAminoMsg(cdc, &MsgExecuteContract{}, "wasmx/MsgExecuteContract")
	legacy.RegisterAminoMsg(cdc, &MsgCompileContract{}, "wasmx/MsgCompileContract")
	legacy.RegisterAminoMsg(cdc, &MsgExecuteEth{}, "wasmx/MsgExecuteEth")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// this line is used by starport scaffolding # 3

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgStoreCode{},
		&MsgInstantiateContract{},
		&MsgInstantiateContract2{},
		&MsgExecuteContract{},
		&MsgCompileContract{},
		&MsgExecuteEth{},
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

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
