package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Address represents the 20 byte address of an Ethereum account.
type AddressCW [AddressLengthWasmx]byte

// SetBytes sets the address to the value of b.
// If b is larger than len(a), b will be cropped from the left.
func (a *AddressCW) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLengthWasmx:]
	}
	copy(a[AddressLengthWasmx-len(b):], b)
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
func Evm32AddressFromBech32(addr string) (*AddressCW, error) {
	accAddress, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, err
	}
	addrcw := BytesToAddressCW(accAddress.Bytes())
	return &addrcw, nil
}

// EvmAddressFromAcc
func EvmAddressFromAcc(addr sdk.AccAddress) common.Address {
	return common.BytesToAddress(addr.Bytes())
}

// AccAddressFromEvm
func AccAddressFromEvm(addr common.Address) sdk.AccAddress {
	return sdk.AccAddress(addr.Bytes())
}

// AccAddressFromHex - any number of bytes; 0x optional
func AccAddressFromHex(addressStr string) sdk.AccAddress {
	return sdk.AccAddress(common.FromHex(addressStr))
}

// IsEmptyHash returns true if the hash corresponds to an empty ethereum hex hash.
func IsEmptyHash(hash string) bool {
	return bytes.Equal(common.HexToHash(hash).Bytes(), common.Hash{}.Bytes())
}
