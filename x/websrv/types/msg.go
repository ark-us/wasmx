package types

import (
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

func (msg MsgRegisterRoute) Route() string {
	return RouterKey
}

func (msg MsgRegisterRoute) Type() string {
	return "register-route"
}

func (msg MsgRegisterRoute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}

	if err := validateStringNonEmpty(msg.Title); err != nil {
		return errorsmod.Wrap(err, "title")
	}

	if err := validateStringNonEmpty(msg.Description); err != nil {
		return errorsmod.Wrap(err, "description")
	}

	if err := validateStringNonEmpty(msg.Path); err != nil {
		return errorsmod.Wrap(err, "path")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return errorsmod.Wrap(err, "contract address")
	}

	if string(msg.Path[0]) != "/" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "path must start with /")
	}

	if len(msg.Path) > 1 && string(msg.Path[len(msg.Path)-1]) == "/" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "path must not end with /")
	}
	return nil
}

func (msg MsgDeregisterRoute) Route() string {
	return RouterKey
}

func (msg MsgDeregisterRoute) Type() string {
	return "deregister-route"
}

func (msg MsgDeregisterRoute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}

	if err := validateStringNonEmpty(msg.Title); err != nil {
		return errorsmod.Wrap(err, "title")
	}

	if err := validateStringNonEmpty(msg.Description); err != nil {
		return errorsmod.Wrap(err, "description")
	}

	if err := validateStringNonEmpty(msg.Path); err != nil {
		return errorsmod.Wrap(err, "path")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
		return errorsmod.Wrap(err, "contract address")
	}
	return nil
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

func validateStringNonEmpty(i interface{}) error {
	if i == "" {
		return fmt.Errorf("empty string")
	}
	return validateString(i)
}
