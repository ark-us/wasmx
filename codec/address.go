package codec

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/golang-lru/simplelru"
	"sigs.k8s.io/yaml"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// AccAddress.String() is expensive and if unoptimized dominantly showed up in profiles,
	// yet has no mechanisms to trivially cache the result given that AccAddress is a []byte type.
	accAddrMu     sync.Mutex
	accAddrCache  *simplelru.LRU
	consAddrMu    sync.Mutex
	consAddrCache *simplelru.LRU
	valAddrMu     sync.Mutex
	valAddrCache  *simplelru.LRU

	isCachingEnabled atomic.Bool
)

// ----------------------------------------------------------------------------
// account prefixed
// ----------------------------------------------------------------------------

// AccAddress a wrapper around bytes meant to represent an account address.
// When marshaled to a string or JSON, it uses Bech32.
type AccAddressPrefixed struct {
	bz     []byte
	prefix string
}

func NewAccAddressPrefixed(bz []byte, prefix string) AccAddressPrefixed {
	return AccAddressPrefixed{bz: bz, prefix: prefix}
}

func NewAddressPrefixedFromAcc(bz []byte, prefix string) AddressPrefixed {
	return NewAccAddressPrefixed(bz, prefix)
}

// AccAddressFromHexUnsafe creates an AccAddress from a HEX-encoded string.
//
// Note, this function is considered unsafe as it may produce an AccAddress from
// otherwise invalid input, such as a transaction hash. Please use
// AccAddressFromBech32.
func AccAddressPrefixedFromHexUnsafe(address string, prefix string) (addr AccAddressPrefixed, err error) {
	bz, err := addressBytesFromHexString(address)
	return NewAccAddressPrefixed(bz, prefix), err
}

// VerifyAddressFormat verifies that the provided bytes form a valid address
// according to the default address rules or a custom address verifier set by
// GetConfig().SetAddressVerifier().
// TODO make an issue to get rid of global Config
// ref: https://github.com/cosmos/cosmos-sdk/issues/9690
func VerifyAddressPrefixedFormat(bz []byte) error {
	verifier := sdk.GetConfig().GetAddressVerifier()
	if verifier != nil {
		return verifier(bz)
	}

	if len(bz) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
	}

	if len(bz) > address.MaxAddrLen {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", address.MaxAddrLen, len(bz))
	}

	return nil
}

// MustAccAddressFromBech32 calls AccAddressFromBech32 and panics on error.
func MustAccAddressPrefixedFromBech32(address string) AccAddressPrefixed {
	addr, err := AccAddressPrefixedFromBech32(address)
	if err != nil {
		panic(err)
	}

	return addr
}

// AccAddressFromBech32 creates an AccAddress from a Bech32 string.
func AccAddressPrefixedFromBech32(address string) (addr AccAddressPrefixed, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return AccAddressPrefixed{}, errors.New("empty address string is not allowed")
	}

	bz, prefix, err := GetFromBech32Unsafe(address)
	if err != nil {
		return AccAddressPrefixed{}, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return AccAddressPrefixed{}, err
	}
	acc := NewAccAddressPrefixed(bz, prefix)
	return acc, nil
}

// Returns boolean for whether two AccAddresses are Equal
func (aa AccAddressPrefixed) Equals(aa2 sdk.Address) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Returns boolean for whether an AccAddress is empty
func (aa AccAddressPrefixed) Empty() bool {
	return len(aa.bz) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa AccAddressPrefixed) Marshal() ([]byte, error) {
	return aa.bz, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (aa *AccAddressPrefixed) Unmarshal(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		return nil
	}

	aa2, err := AccAddressPrefixedFromBech32(s)
	if err != nil {
		return err
	}

	aa.prefix = aa2.prefix
	aa.bz = aa2.bz
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (aa AccAddressPrefixed) MarshalJSON() ([]byte, error) {
	return json.Marshal(aa.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (aa AccAddressPrefixed) MarshalYAML() (interface{}, error) {
	return aa.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (aa *AccAddressPrefixed) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		return nil
	}

	aa2, err := AccAddressPrefixedFromBech32(s)
	if err != nil {
		return err
	}

	aa.prefix = aa2.prefix
	aa.bz = aa2.bz
	return nil
}

// UnmarshalYAML unmarshals from JSON assuming Bech32 encoding.
func (aa *AccAddressPrefixed) UnmarshalYAML(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		return nil
	}

	aa2, err := AccAddressPrefixedFromBech32(s)
	if err != nil {
		return err
	}

	aa.prefix = aa2.prefix
	aa.bz = aa2.bz
	return nil
}

// Bytes returns the raw address bytes.
func (aa AccAddressPrefixed) Bytes() []byte {
	return aa.bz
}

func (aa AccAddressPrefixed) Prefix() string {
	return aa.prefix
}

// String implements the Stringer interface.
func (aa AccAddressPrefixed) String() string {
	if aa.Empty() {
		return ""
	}

	key := aa.prefix + UnsafeBytesToStr(aa.bz)

	if sdk.IsAddrCacheEnabled() {
		accAddrMu.Lock()
		defer accAddrMu.Unlock()

		addr, ok := accAddrCache.Get(key)
		if ok {
			return addr.(string)
		}
	}
	// return cacheBech32Addr(GetConfig().GetBech32AccountAddrPrefix(), aa, accAddrCache, key)

	return cacheBech32Addr(aa.prefix, aa.bz, accAddrCache, key)
}

// Format implements the fmt.Formatter interface.

func (aa AccAddressPrefixed) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(aa.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", aa)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(aa.bz))))
	}
}

// ----------------------------------------------------------------------------
// validator operator prefixed
// ----------------------------------------------------------------------------

// ValAddress defines a wrapper around bytes meant to present a validator's
// operator. When marshaled to a string or JSON, it uses Bech32.
type ValAddressPrefixed struct {
	bz     []byte
	prefix string
}

func NewValAddressPrefixed(bz []byte, prefix string) ValAddressPrefixed {
	return ValAddressPrefixed{bz: bz, prefix: prefix}
}

func NewAddressPrefixedFromVal(bz []byte, prefix string) AddressPrefixed {
	return NewValAddressPrefixed(bz, prefix)
}

// ValAddressFromHex creates a ValAddress from a hex string.
func ValAddressPrefixedFromHex(address string, prefix string) (addr ValAddressPrefixed, err error) {
	bz, err := addressBytesFromHexString(address)
	return NewValAddressPrefixed(bz, prefix), err
}

// ValAddressFromBech32 creates a ValAddress from a Bech32 string.
func ValAddressPrefixedFromBech32(address string) (addr ValAddressPrefixed, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ValAddressPrefixed{}, errors.New("empty address string is not allowed")
	}

	bz, prefix, err := GetFromBech32Unsafe(address)
	if err != nil {
		return ValAddressPrefixed{}, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return ValAddressPrefixed{}, err
	}

	return NewValAddressPrefixed(bz, prefix), nil
}

// Returns boolean for whether two ValAddresses are Equal
func (va ValAddressPrefixed) Equals(va2 sdk.Address) bool {
	if va.Empty() && va2.Empty() {
		return true
	}

	return bytes.Equal(va.Bytes(), va2.Bytes())
}

// Returns boolean for whether an ValAddress is empty
func (va ValAddressPrefixed) Empty() bool {
	return len(va.bz) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (va ValAddressPrefixed) Marshal() ([]byte, error) {
	return va.bz, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (va *ValAddressPrefixed) Unmarshal(data []byte, prefix string) error {
	va.bz = data
	va.prefix = prefix
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (va ValAddressPrefixed) MarshalJSON() ([]byte, error) {
	return json.Marshal(va.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (va ValAddressPrefixed) MarshalYAML() (interface{}, error) {
	return va.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (va *ValAddressPrefixed) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*va = ValAddressPrefixed{}
		return nil
	}

	va2, err := ValAddressPrefixedFromBech32(s)
	if err != nil {
		return err
	}

	*va = va2
	return nil
}

// UnmarshalYAML unmarshals from YAML assuming Bech32 encoding.
func (va *ValAddressPrefixed) UnmarshalYAML(data []byte) error {
	var s string

	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*va = ValAddressPrefixed{}
		return nil
	}

	va2, err := ValAddressPrefixedFromBech32(s)
	if err != nil {
		return err
	}

	*va = va2
	return nil
}

// Bytes returns the raw address bytes.
func (va ValAddressPrefixed) Bytes() []byte {
	return va.bz
}

func (aa ValAddressPrefixed) Prefix() string {
	return aa.prefix
}

// String implements the Stringer interface.
func (va ValAddressPrefixed) String() string {
	if va.Empty() {
		return ""
	}

	key := va.prefix + UnsafeBytesToStr(va.bz)

	if sdk.IsAddrCacheEnabled() {
		valAddrMu.Lock()
		defer valAddrMu.Unlock()

		addr, ok := valAddrCache.Get(key)
		if ok {
			return addr.(string)
		}
	}
	return cacheBech32Addr(va.prefix, va.bz, valAddrCache, key)
}

// Format implements the fmt.Formatter interface.

func (va ValAddressPrefixed) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(va.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", va)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(va.bz))))
	}
}

// ----------------------------------------------------------------------------
// consensus node prefixed
// ----------------------------------------------------------------------------

// ConsAddress defines a wrapper around bytes meant to present a consensus node.
// When marshaled to a string or JSON, it uses Bech32.
type ConsAddressPrefixed struct {
	bz     []byte
	prefix string
}

func NewConsAddressPrefixed(bz []byte, prefix string) ConsAddressPrefixed {
	return ConsAddressPrefixed{bz: bz, prefix: prefix}
}

func NewAddressPrefixedFromCons(bz []byte, prefix string) AddressPrefixed {
	return NewConsAddressPrefixed(bz, prefix)
}

// ConsAddressFromHex creates a ConsAddress from a hex string.
// Deprecated: use ConsensusAddressCodec from Staking keeper
func ConsAddressPrefixedFromHex(address string, prefix string) (addr ConsAddressPrefixed, err error) {
	bz, err := addressBytesFromHexString(address)
	return NewConsAddressPrefixed(bz, prefix), err
}

// ConsAddressFromBech32 creates a ConsAddress from a Bech32 string.
func ConsAddressPrefixedFromBech32(address string) (addr ConsAddressPrefixed, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ConsAddressPrefixed{}, errors.New("empty address string is not allowed")
	}

	bz, prefix, err := GetFromBech32Unsafe(address)
	if err != nil {
		return ConsAddressPrefixed{}, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return ConsAddressPrefixed{}, err
	}

	return NewConsAddressPrefixed(bz, prefix), nil
}

// get ConsAddress from pubkey
func GetConsAddressPrefixed(pubkey cryptotypes.PubKey, prefix string) ConsAddressPrefixed {
	return NewConsAddressPrefixed(pubkey.Address(), prefix)
}

// Returns boolean for whether two ConsAddress are Equal
func (ca ConsAddressPrefixed) Equals(ca2 sdk.Address) bool {
	if ca.Empty() && ca2.Empty() {
		return true
	}

	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Returns boolean for whether an ConsAddress is empty
func (ca ConsAddressPrefixed) Empty() bool {
	return len(ca.bz) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (ca ConsAddressPrefixed) Marshal() ([]byte, error) {
	return ca.bz, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (ca *ConsAddressPrefixed) Unmarshal(data []byte, prefix string) error {
	ca.bz = data
	ca.prefix = prefix
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (ca ConsAddressPrefixed) MarshalJSON() ([]byte, error) {
	return json.Marshal(ca.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (ca ConsAddressPrefixed) MarshalYAML() (interface{}, error) {
	return ca.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (ca *ConsAddressPrefixed) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*ca = ConsAddressPrefixed{}
		return nil
	}

	ca2, err := ConsAddressPrefixedFromBech32(s)
	if err != nil {
		return err
	}

	*ca = ca2
	return nil
}

// UnmarshalYAML unmarshals from YAML assuming Bech32 encoding.
func (ca *ConsAddressPrefixed) UnmarshalYAML(data []byte) error {
	var s string

	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*ca = ConsAddressPrefixed{}
		return nil
	}

	ca2, err := ConsAddressPrefixedFromBech32(s)
	if err != nil {
		return err
	}

	*ca = ca2
	return nil
}

// Bytes returns the raw address bytes.
func (ca ConsAddressPrefixed) Bytes() []byte {
	return ca.bz
}

func (aa ConsAddressPrefixed) Prefix() string {
	return aa.prefix
}

// String implements the Stringer interface.
func (ca ConsAddressPrefixed) String() string {
	if ca.Empty() {
		return ""
	}

	key := ca.prefix + UnsafeBytesToStr(ca.bz)

	if sdk.IsAddrCacheEnabled() {
		consAddrMu.Lock()
		defer consAddrMu.Unlock()

		addr, ok := consAddrCache.Get(key)
		if ok {
			return addr.(string)
		}
	}
	return cacheBech32Addr(ca.prefix, ca.bz, consAddrCache, key)
}

// Bech32ifyAddressBytes returns a bech32 representation of address bytes.
// Returns an empty sting if the byte slice is 0-length. Returns an error if the bech32 conversion
// fails or the prefix is empty.
func Bech32ifyAddressPrefixedBytes(prefix string, bs []byte) (string, error) {
	if len(bs) == 0 {
		return "", nil
	}
	if len(prefix) == 0 {
		return "", errors.New("prefix cannot be empty")
	}
	return bech32.ConvertAndEncode(prefix, bs)
}

// MustBech32ifyAddressBytes returns a bech32 representation of address bytes.
// Returns an empty sting if the byte slice is 0-length. It panics if the bech32 conversion
// fails or the prefix is empty.
func MustBech32ifyAddressPrefixedBytes(prefix string, bs []byte) string {
	s, err := sdk.Bech32ifyAddressBytes(prefix, bs)
	if err != nil {
		panic(err)
	}
	return s
}

// Format implements the fmt.Formatter interface.

func (ca ConsAddressPrefixed) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(ca.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", ca)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(ca.bz))))
	}
}

// ----------------------------------------------------------------------------
// auxiliary
// ----------------------------------------------------------------------------

var errBech32EmptyAddress = errors.New("decoding Bech32 address failed: must provide a non empty address")

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32Unsafe(bech32str string) ([]byte, string, error) {
	if len(bech32str) == 0 {
		return nil, "", errBech32EmptyAddress
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, "", err
	}

	return bz, hrp, nil
}

func addressBytesFromHexString(address string) ([]byte, error) {
	if len(address) == 0 {
		return nil, sdk.ErrEmptyHexAddress
	}

	return hex.DecodeString(address)
}

// cacheBech32Addr is not concurrency safe. Concurrent access to cache causes race condition.
func cacheBech32Addr(prefix string, addr []byte, cache *simplelru.LRU, cacheKey string) string {
	bech32Addr, err := bech32.ConvertAndEncode(prefix, addr)
	if err != nil {
		panic(err)
	}
	if sdk.IsAddrCacheEnabled() {
		cache.Add(cacheKey, bech32Addr)
	}
	return bech32Addr
}
