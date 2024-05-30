package types

import (
	"google.golang.org/protobuf/proto"

	sdkerr "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"cosmossdk.io/x/tx/signing"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

type RawContractMessage = wasmxtypes.RawContractMessage

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

func ProvideExecuteAtomicTxGetSigners() signing.CustomGetSigner {
	return signing.CustomGetSigner{
		MsgType: proto.MessageName(&MsgExecuteAtomicTxRequest{}),
		Fn: func(msg proto.Message) ([][]byte, error) {
			msg2 := msg.(*MsgExecuteAtomicTxRequest)
			return [][]byte{msg2.Sender}, nil
		},
	}
}
