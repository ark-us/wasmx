package types

import (
	fmt "fmt"

	address "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (msg MsgRegisterOAuthClient) Route() string {
	return RouterKey
}

func (msg MsgRegisterOAuthClient) Type() string {
	return "register-oauth-client"
}

func (msg MsgRegisterOAuthClient) ValidateBasic() error {
	// TODO validate domain
	if msg.Domain == "" {
		return ErrOAuthClientInvalidDomain
	}

	return validateString(msg.Domain)
}

func (msg MsgRegisterOAuthClient) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Owner); err != nil {
		return err
	}
	return nil
}

func (msg MsgEditOAuthClient) Route() string {
	return RouterKey
}

func (msg MsgEditOAuthClient) Type() string {
	return "edit-oauth-client"
}

func (msg MsgEditOAuthClient) ValidateBasic() error {
	// TODO validate domain
	if msg.Domain == "" {
		return ErrOAuthClientInvalidDomain
	}

	if err := validateUint64(msg.ClientId); err != nil {
		return err
	}
	return validateString(msg.Domain)
}

func (msg MsgEditOAuthClient) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Owner); err != nil {
		return err
	}
	return nil
}

func (msg MsgDeregisterOAuthClient) Route() string {
	return RouterKey
}

func (msg MsgDeregisterOAuthClient) Type() string {
	return "deregister-oauth-client"
}

func (msg MsgDeregisterOAuthClient) ValidateBasic() error {
	return validateUint64(msg.ClientId)
}

func (msg MsgDeregisterOAuthClient) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Owner); err != nil {
		return err
	}

	return nil
}

func (msg MsgRegisterRoute) Route() string {
	return RouterKey
}

func (msg MsgRegisterRoute) Type() string {
	return "register-route"
}

func (msg MsgRegisterRoute) ValidateBasic() error {
	if err := validateStringNonEmpty(msg.Title); err != nil {
		return errorsmod.Wrap(err, "title")
	}

	if err := validateStringNonEmpty(msg.Description); err != nil {
		return errorsmod.Wrap(err, "description")
	}

	if err := validateStringNonEmpty(msg.Path); err != nil {
		return errorsmod.Wrap(err, "path")
	}

	if string(msg.Path[0]) != "/" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "path must start with /")
	}

	if len(msg.Path) > 1 && string(msg.Path[len(msg.Path)-1]) == "/" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "path must not end with /")
	}
	return nil
}

func (msg MsgRegisterRoute) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}

	if _, err := addressCodec.StringToBytes(msg.ContractAddress); err != nil {
		return errorsmod.Wrap(err, "contract address")
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
	if err := validateStringNonEmpty(msg.Title); err != nil {
		return errorsmod.Wrap(err, "title")
	}

	if err := validateStringNonEmpty(msg.Description); err != nil {
		return errorsmod.Wrap(err, "description")
	}

	if err := validateStringNonEmpty(msg.Path); err != nil {
		return errorsmod.Wrap(err, "path")
	}

	return nil
}

func (msg MsgDeregisterRoute) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "authority")
	}

	if _, err := addressCodec.StringToBytes(msg.ContractAddress); err != nil {
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
