package types

import (
	bytes "bytes"
	"encoding/json"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

type SystemContracts = []SystemContract

func DefaultSystemContracts() SystemContracts {
	msg := WasmxExecutionMessage{Data: []byte{}}
	initMsg, err := json.Marshal(msg)
	if err != nil {
		panic("DefaultSystemContracts: cannot marshal init message")
	}

	return []SystemContract{
		{
			Address:     "0x0000000000000000000000000000000000000001",
			Label:       "ecrecover",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000002",
			Label:       "sha2-256",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000003",
			Label:       "ripmd160",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000004",
			Label:       "identity",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000005",
			Label:       "modexp",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000006",
			Label:       "ecadd",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000007",
			Label:       "ecmul",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000008",
			Label:       "ecpairings",
			InitMessage: initMsg,
		},
		{
			Address:     "0x0000000000000000000000000000000000000009",
			Label:       "blake2f",
			InitMessage: initMsg,
		},
	}
}

func (p SystemContract) Validate() error {
	if err := validateString(p.Label); err != nil {
		return err
	}
	if err := validateString(p.Address); err != nil {
		return err
	}
	if err := validateBytes(p.InitMessage); err != nil {
		return err
	}
	if p.InitMessage == nil {
		return fmt.Errorf("initialization message cannot be nil")
	}
	return ValidateNonZeroAddress(p.Address)
}

func validateString(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBytes(i interface{}) error {
	_, ok := i.([]byte)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// TODO have addresses be 32bytes

// IsZeroAddress returns true if the address corresponds to an empty ethereum hex address.
func IsZeroAddress(address string) bool {
	return bytes.Equal(common.HexToAddress(address).Bytes(), common.Address{}.Bytes())
}

// ValidateAddress returns an error if the provided string is either not a hex formatted string address
func ValidateAddress(address string) error {
	if !common.IsHexAddress(address) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress, "address '%s' is not a valid ethereum hex address",
			address,
		)
	}
	return nil
}

// ValidateNonZeroAddress returns an error if the provided string is not a hex
// formatted string address or is equal to zero
func ValidateNonZeroAddress(address string) error {
	if IsZeroAddress(address) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress, "address '%s' must not be zero",
			address,
		)
	}
	return ValidateAddress(address)
}
