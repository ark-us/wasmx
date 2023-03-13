package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (msg MsgRegisterRoute) Route() string {
	return RouterKey
}

func (msg MsgRegisterRoute) Type() string {
	return "store-code"
}

func (msg MsgRegisterRoute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return err
	}

	if msg.Path == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty route path")
	}

	if string(msg.Path[0]) != "/" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "path must start with /")
	}

	if len(msg.Path) > 1 && string(msg.Path[len(msg.Path)-1]) == "/" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "path must not end with /")
	}

	return nil
}

func (msg MsgRegisterRoute) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRegisterRoute) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
