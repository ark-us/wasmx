package ewasm

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// AddressLengthCW is the expected length of a CosmWasm address
	AddressLengthCW = 32
)

// Address represents the 20 byte address of an Ethereum account.
type AddressCW [AddressLengthCW]byte

// SetBytes sets the address to the value of b.
// If b is larger than len(a), b will be cropped from the left.
func (a *AddressCW) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLengthCW:]
	}
	copy(a[AddressLengthCW-len(b):], b)
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddressCW(b []byte) AddressCW {
	var a AddressCW
	a.SetBytes(b)
	return a
}

func (a AddressCW) Hex() string {
	return hexutil.Encode(a[:])
}

func (a AddressCW) Bytes() []byte {
	return a[:]
}

// Evm32AddressFromAcc
func Evm32AddressFromAcc(addr sdk.AccAddress) AddressCW {
	return BytesToAddressCW(addr.Bytes())
}

// Evm32AddressFromBech32
func Evm32AddressFromBech32(addr string) AddressCW {
	accAddress := sdk.MustAccAddressFromBech32(addr)
	return BytesToAddressCW(accAddress.Bytes())
}

// EvmAddressFromAcc
func EvmAddressFromAcc(addr sdk.AccAddress) common.Address {
	return common.BytesToAddress(addr.Bytes())
}

// AccAddressFromEvm
func AccAddressFromEvm(addr common.Address) sdk.AccAddress {
	return sdk.AccAddress(addr.Bytes())
}

// AccAddressFromEvm
func AccAddressFromHex(addressStr string) sdk.AccAddress {
	return sdk.AccAddress(common.HexToAddress(addressStr).Bytes())
}

// IsEmptyHash returns true if the hash corresponds to an empty ethereum hex hash.
func IsEmptyHash(hash string) bool {
	return bytes.Equal(common.HexToHash(hash).Bytes(), common.Hash{}.Bytes())
}

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
