package types

import (
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (msg MsgGrpcRequest) Route() string {
	return RouterKey
}

func (msg MsgGrpcRequest) Type() string {
	return "grpc-request"
}

func (msg MsgGrpcRequest) ValidateBasic() error {
	if len(msg.Data) == 0 {
		return sdkerr.Wrapf(sdkerrors.ErrInvalidRequest, "empty request")
	}
	return nil
}

func (msg MsgGrpcRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}
