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
			Pinned:      false,
			Native:      true,
		},
		// Ethereum ecrecover
		{
			Address:     "0x000000000000000000000000000000000000001f",
			Label:       "ecrecovereth",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000002",
			Label:       "sha2-256",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000003",
			Label:       "ripmd160",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000004",
			Label:       "identity",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000005",
			Label:       "modexp",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000006",
			Label:       "ecadd",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000007",
			Label:       "ecmul",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000008",
			Label:       "ecpairings",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000009",
			Label:       "blake2f",
			InitMessage: initMsg,
			Pinned:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000020",
			Label:       "secp384r1",
			InitMessage: initMsg,
			Pinned:      false, //TODO
		},
		{
			Address:     "0x0000000000000000000000000000000000000021",
			Label:       "secp384r1_registry",
			InitMessage: initMsg,
			Pinned:      false, // TODO
		},
		{
			Address:     "0x0000000000000000000000000000000000000022",
			Label:       "secret_sharing",
			InitMessage: initMsg,
			Pinned:      false,
			Native:      true,
		},
		{
			Address:     "0x0000000000000000000000000000000000000023",
			Label:       INTERPRETER_EVM_SHANGHAI,
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_INTERPRETER,
		},
		{
			Address:     "0x0000000000000000000000000000000000000024",
			Label:       "alias_eth",
			InitMessage: initMsg,
			Pinned:      false,
			Role:        ROLE_ALIAS,
		},
		{
			Address:     "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			Label:       "sys_proxy",
			InitMessage: initMsg,
			Pinned:      false,
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
	if err := p.InitMessage.ValidateBasic(); err != nil {
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

const (
	// AddressLengthCW is the expected length of a Wasmx and CosmWasm address
	AddressLengthWasmx = 32
	// AddressLengthCW is the expected length of an Ethereum address
	AddressLengthEth = 20
)

// TODO have addresses be 32bytes

// IsZeroAddress returns true if the address corresponds to an empty ethereum hex address.
func IsZeroAddress(address string) bool {
	return bytes.Equal(common.HexToAddress(address).Bytes(), common.Address{}.Bytes())
}

// ValidateAddress returns an error if the provided string is either not a hex formatted string address
func ValidateAddress(address string) error {
	if !IsHexAddress(address) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress, "address '%s' is not a valid ethereum hex address",
			address,
		)
	}
	return nil
}

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// WasmX or Ethereum address or not.
func IsHexAddress(s string) bool {
	if has0xPrefix(s) {
		s = s[2:]
	}
	return isHex(s) && (len(s) == 2*AddressLengthWasmx || len(s) == 2*AddressLengthEth)
}

// has0xPrefix validates str begins with '0x' or '0X'.
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
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
