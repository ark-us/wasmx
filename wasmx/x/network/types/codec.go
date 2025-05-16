package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgGrpcSendRequest{}, "network/MsgGrpcSendRequest", nil)
	cdc.RegisterConcrete(&MsgStartTimeoutRequest{}, "network/MsgStartTimeoutRequest", nil)
	cdc.RegisterConcrete(&MsgReentry{}, "network/MsgReentry", nil)

	cdc.RegisterConcrete(&RequestPing{}, "network/RequestPing", nil)
	cdc.RegisterConcrete(&RequestBroadcastTx{}, "network/RequestBroadcastTx", nil)

	cdc.RegisterConcrete(&MsgMultiChainWrap{}, "network/MsgMultiChainWrap", nil)
	cdc.RegisterConcrete(&MsgMultiChainWrapResponse{}, "network/MsgMultiChainWrapResponse", nil)

	cdc.RegisterConcrete(&MsgExecuteAtomicTxRequest{}, "network/MsgExecuteAtomicTxRequest", nil)
	cdc.RegisterConcrete(&MsgExecuteAtomicTxResponse{}, "network/MsgExecuteAtomicTxResponse", nil)
	cdc.RegisterConcrete(&AtomicTxCrossChainCallInfo{}, "network/AtomicTxCrossChainCallInfo", nil)

	cdc.RegisterConcrete(&ExtensionOptionAtomicMultiChainTx{}, "network/ExtensionOptionAtomicMultiChainTx", nil)
	cdc.RegisterConcrete(&ExtensionOptionMultiChainTx{}, "network/ExtensionOptionMultiChainTx", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgGrpcSendRequest{},
		&MsgStartTimeoutRequest{},
		&MsgReentry{},

		&RequestPing{},
		&RequestBroadcastTx{},

		&MsgMultiChainWrap{},
		&MsgMultiChainWrapResponse{},
		&AtomicTxCrossChainCallInfo{},

		&MsgExecuteAtomicTxRequest{},
		&MsgExecuteAtomicTxResponse{},
	)

	registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&ExtensionOptionAtomicMultiChainTx{},
		&ExtensionOptionMultiChainTx{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// TODO better solution?
var Network_Msg_serviceDesc = _Msg_serviceDesc
