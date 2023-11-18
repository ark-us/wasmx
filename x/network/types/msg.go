package types

import (
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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

func (msg MsgGrpcSendRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress("network")}
}
