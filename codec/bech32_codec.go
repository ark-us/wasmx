package codec

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Bech32Codec struct {
	bech32Prefix       string
	addressConstructor func(bz []byte, prefix string) AddressPrefixed
}

var _ address.Codec = &Bech32Codec{}
var _ address.Codec = &AccBech32Codec{}
var _ address.Codec = &ValBech32Codec{}
var _ address.Codec = &ConsBech32Codec{}

func NewBech32Codec(prefix string, addressConstructor func(bz []byte, prefix string) AddressPrefixed) address.Codec {
	return Bech32Codec{prefix, addressConstructor}
}

func UnwrapBech32Codec(cdc address.Codec) (Bech32Codec, error) {
	uwcdc, ok := cdc.(Bech32Codec)
	if !ok {
		return Bech32Codec{}, fmt.Errorf("cannot unwrap address.Codec to Bech32Codec")
	}
	return uwcdc, nil
}

func MustUnwrapBech32Codec(cdc address.Codec) Bech32Codec {
	uwcdc, ok := cdc.(Bech32Codec)
	if !ok {
		panic("cannot unwrap address.Codec to Bech32Codec")
	}
	return uwcdc
}

// StringToBytes encodes text to bytes
func (bc Bech32Codec) StringToBytes(text string) ([]byte, error) {
	if len(strings.TrimSpace(text)) == 0 {
		return []byte{}, errors.New("empty address string is not allowed")
	}

	hrp, bz, err := bech32.DecodeAndConvert(text)
	if err != nil {
		return nil, err
	}

	if hrp != bc.bech32Prefix {
		return nil, errorsmod.Wrapf(sdkerrors.ErrLogic, "hrp does not match bech32 prefix: expected '%s' got '%s'", bc.bech32Prefix, hrp)
	}

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return nil, err
	}

	return bz, nil
}

// BytesToString decodes bytes to text
func (bc Bech32Codec) BytesToString(bz []byte) (string, error) {
	text, err := bech32.ConvertAndEncode(bc.bech32Prefix, bz)
	if err != nil {
		return "", err
	}

	return text, nil
}

func (bc Bech32Codec) StringToAddressPrefixedUnsafe(text string) (AddressPrefixed, error) {
	if len(strings.TrimSpace(text)) == 0 {
		return nil, errors.New("empty address string is not allowed")
	}

	hrp, bz, err := bech32.DecodeAndConvert(text)
	if err != nil {
		return nil, err
	}

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return nil, err
	}

	return bc.addressConstructor(bz, hrp), nil
}

// StringToBytes encodes text to bytes
func (bc Bech32Codec) StringToAddressPrefixed(text string) (AddressPrefixed, error) {
	bz, err := bc.StringToBytes(text)
	if err != nil {
		return nil, err
	}
	return bc.addressConstructor(bz, bc.bech32Prefix), nil
}

func (bc Bech32Codec) BytesToAddressPrefixed(bz []byte) AddressPrefixed {
	return bc.addressConstructor(bz, bc.bech32Prefix)
}

func (bc Bech32Codec) Prefix() string {
	return bc.bech32Prefix
}

type AccBech32Codec struct {
	Bech32Codec
}

func NewAccBech32Codec(prefix string, addressConstructor func(bz []byte, prefix string) AddressPrefixed) address.Codec {
	return AccBech32Codec{Bech32Codec{prefix, addressConstructor}}
}

func UnwrapAccBech32Codec(cdc address.Codec) (AccBech32Codec, error) {
	uwcdc, ok := cdc.(AccBech32Codec)
	if !ok {
		return AccBech32Codec{}, fmt.Errorf("cannot unwrap address.Codec to AccBech32Codec")
	}
	return uwcdc, nil
}

func MustUnwrapAccBech32Codec(cdc address.Codec) AccBech32Codec {
	uwcdc, ok := cdc.(AccBech32Codec)
	if !ok {
		panic("cannot unwrap address.Codec to AccBech32Codec")
	}
	return uwcdc
}

func (bc AccBech32Codec) BytesToAccAddressPrefixed(bz []byte) AccAddressPrefixed {
	return bc.BytesToAddressPrefixed(bz).(AccAddressPrefixed)
}

func (bc AccBech32Codec) StringToAccAddressPrefixed(text string) (AccAddressPrefixed, error) {
	res, err := bc.StringToAddressPrefixed(text)
	if err != nil {
		return AccAddressPrefixed{}, err
	}
	return res.(AccAddressPrefixed), nil
}

type ValBech32Codec struct {
	Bech32Codec
}

func NewValBech32Codec(prefix string, addressConstructor func(bz []byte, prefix string) AddressPrefixed) address.Codec {
	return ValBech32Codec{Bech32Codec{prefix, addressConstructor}}
}

func UnwrapValBech32Codec(cdc address.Codec) (ValBech32Codec, error) {
	uwcdc, ok := cdc.(ValBech32Codec)
	if !ok {
		return ValBech32Codec{}, fmt.Errorf("cannot unwrap address.Codec to ValBech32Codec")
	}
	return uwcdc, nil
}

func MustUnwrapValBech32Codec(cdc address.Codec) ValBech32Codec {
	uwcdc, ok := cdc.(ValBech32Codec)
	if !ok {
		panic("cannot unwrap address.Codec to ValBech32Codec")
	}
	return uwcdc
}

func (bc ValBech32Codec) BytesToValAddressPrefixed(bz []byte) ValAddressPrefixed {
	return bc.BytesToAddressPrefixed(bz).(ValAddressPrefixed)
}

func (bc ValBech32Codec) StringToValAddressPrefixed(text string) (ValAddressPrefixed, error) {
	res, err := bc.StringToAddressPrefixed(text)
	if err != nil {
		return ValAddressPrefixed{}, err
	}
	return res.(ValAddressPrefixed), nil
}

type ConsBech32Codec struct {
	Bech32Codec
}

func NewConsBech32Codec(prefix string, addressConstructor func(bz []byte, prefix string) AddressPrefixed) address.Codec {
	return ConsBech32Codec{Bech32Codec{prefix, addressConstructor}}
}

func UnwrapConsBech32Codec(cdc address.Codec) (ConsBech32Codec, error) {
	uwcdc, ok := cdc.(ConsBech32Codec)
	if !ok {
		return ConsBech32Codec{}, fmt.Errorf("cannot unwrap address.Codec to ConsBech32Codec")
	}
	return uwcdc, nil
}

func MustUnwrapConsBech32Codec(cdc address.Codec) ConsBech32Codec {
	uwcdc, ok := cdc.(ConsBech32Codec)
	if !ok {
		panic("cannot unwrap address.Codec to ConsBech32Codec")
	}
	return uwcdc
}

func (bc ConsBech32Codec) BytesToConsAddressPrefixed(bz []byte) ConsAddressPrefixed {
	return bc.BytesToAddressPrefixed(bz).(ConsAddressPrefixed)
}

func (bc ConsBech32Codec) StringToConsAddressPrefixed(text string) (ConsAddressPrefixed, error) {
	res, err := bc.StringToAddressPrefixed(text)
	if err != nil {
		return ConsAddressPrefixed{}, err
	}
	return res.(ConsAddressPrefixed), nil
}
