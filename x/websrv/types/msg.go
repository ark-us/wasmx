package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg MsgRegisterOAuthClient) Route() string {
	return RouterKey
}

func (msg MsgRegisterOAuthClient) Type() string {
	return "register-oauth-client"
}

func (msg MsgRegisterOAuthClient) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}

	// TODO validate domain
	if msg.Domain == "" {
		return ErrOAuthClientInvalidDomain
	}

	return validateString(msg.Domain)
}

func (msg MsgRegisterOAuthClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRegisterOAuthClient) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func (msg MsgEditOAuthClient) Route() string {
	return RouterKey
}

func (msg MsgEditOAuthClient) Type() string {
	return "edit-oauth-client"
}

func (msg MsgEditOAuthClient) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}

	// TODO validate domain
	if msg.Domain == "" {
		return ErrOAuthClientInvalidDomain
	}

	if err := validateUint64(msg.ClientId); err != nil {
		return err
	}
	return validateString(msg.Domain)
}

func (msg MsgEditOAuthClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgEditOAuthClient) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func (msg MsgDeregisterOAuthClient) Route() string {
	return RouterKey
}

func (msg MsgDeregisterOAuthClient) Type() string {
	return "deregister-oauth-client"
}

func (msg MsgDeregisterOAuthClient) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}

	return validateUint64(msg.ClientId)
}

func (msg MsgDeregisterOAuthClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgDeregisterOAuthClient) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateString(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
