package types

import (
	"errors"

	"google.golang.org/protobuf/proto"

	sdkerr "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"cosmossdk.io/x/tx/signing"

	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

var (
	TypeURL_ExtensionOptionEthereumTx         = "/mythos.wasmx.v1.ExtensionOptionEthereumTx"
	TypeURL_ExtensionOptionAtomicMultiChainTx = "/mythos.network.v1.ExtensionOptionAtomicMultiChainTx"
	TypeURL_ExtensionOptionMultiChainTx       = "/mythos.network.v1.ExtensionOptionMultiChainTx"
)

type RawContractMessage = wasmxtypes.RawContractMessage

func (msg MsgReentry) Route() string {
	return RouterKey
}

func (msg MsgReentry) Type() string {
	return "reentry"
}

func (msg MsgReentry) ValidateBasic() error {
	return nil
}

func (msg MsgGrpcSendRequest) Route() string {
	return RouterKey
}

func (msg MsgGrpcSendRequest) Type() string {
	return "grpc-request"
}

func (msg MsgGrpcSendRequest) ValidateBasic() error {
	if len(msg.Data) == 0 {
		return sdkerr.Wrapf(sdkerrors.ErrInvalidRequest, "empty request")
	}
	return nil
}

func (msg MsgMultiChainWrap) Route() string {
	return RouterKey
}

func (msg MsgMultiChainWrap) Type() string {
	return "multi-chain-wrap"
}

func (msg MsgMultiChainWrap) ValidateBasic() error {
	if msg.Data == nil {
		return sdkerr.Wrapf(sdkerrors.ErrInvalidRequest, "empty request")
	}
	// TODO address validator with AddressCodec
	// if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
	// 	return sdkerr.Wrap(err, "sender")
	// }
	return nil
}

func (msg *MsgMultiChainWrap) SetExtensionOptions(
	txBuilder client.TxBuilder,
	chainId string,
	index int32,
	txcount int32,
) (client.TxBuilder, error) {
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, errors.New("unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&ExtensionOptionMultiChainTx{ChainId: chainId, Index: index, TxCount: txcount})
	if err != nil {
		return nil, err
	}

	builder.SetExtensionOptions(option)
	return builder, nil
}

func (msg *MsgExecuteAtomicTxRequest) SetExtensionOptions(
	txBuilder client.TxBuilder,
	chainIds []string,
	leaderChainId string,
) (client.TxBuilder, error) {
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, errors.New("unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&ExtensionOptionAtomicMultiChainTx{LeaderChainId: leaderChainId, ChainIds: chainIds})
	if err != nil {
		return nil, err
	}

	builder.SetExtensionOptions(option)
	return builder, nil
}

func ProvideExecuteAtomicTxGetSigners() signing.CustomGetSigner {
	return signing.CustomGetSigner{
		MsgType: proto.MessageName(&MsgExecuteAtomicTxRequest{}),
		Fn: func(msg proto.Message) ([][]byte, error) {
			msg2 := msg.(*MsgExecuteAtomicTxRequest)
			return [][]byte{msg2.Sender}, nil
		},
	}
}
