package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
)

var (
	_ sdk.AccountI                       = (*BaseAccount)(nil)
	_ mcodec.AccountI                    = (*BaseAccount)(nil)
	_ GenesisAccount                     = (*BaseAccount)(nil)
	_ codectypes.UnpackInterfacesMessage = (*BaseAccount)(nil)
	_ GenesisAccount                     = (*ModuleAccount)(nil)
	_ sdk.ModuleAccountI                 = (*ModuleAccount)(nil)
	_ mcodec.ModuleAccountI              = (*ModuleAccount)(nil)
	_ codectypes.UnpackInterfacesMessage = (*ModuleAccount)(nil)
)

// NewBaseAccount creates a new BaseAccount object.
func NewBaseAccount(address mcodec.AccAddressPrefixed, pubKey cryptotypes.PubKey, accountNumber, sequence uint64) *BaseAccount {
	acc := &BaseAccount{
		Address:       address.String(),
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		panic(err)
	}

	return acc
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() mcodec.AccountI {
	return &BaseAccount{}
}

// NewBaseAccountWithAddress - returns a new base account with a given address
// leaving AccountNumber and Sequence to zero.
func NewBaseAccountWithAddress(addr mcodec.AccAddressPrefixed) *BaseAccount {
	return &BaseAccount{
		Address: addr.String(),
	}
}

// GetAddress - Implements mcodec.AccountI.
func (acc BaseAccount) GetAddressPrefixed() mcodec.AccAddressPrefixed {
	addr, _ := mcodec.AccAddressPrefixedFromBech32(acc.Address)
	return addr
}

// GetAddress - Implements sdk.AccountI.
func (acc BaseAccount) GetAddress() sdk.AccAddress {
	addr, _ := mcodec.AccAddressPrefixedFromBech32(acc.Address)
	return addr.Bytes()
}

// SetAddress - Implements mcodec.AccountI.
func (acc *BaseAccount) SetAddressPrefixed(addr mcodec.AccAddressPrefixed) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}

	acc.Address = addr.String()
	return nil
}

// SetAddress - Implements sdk.AccountI.
func (acc *BaseAccount) SetAddress(addr sdk.AccAddress) error {
	return errors.New("cannot use SetAddress on cosmosmod.BaseAccount")
}

// GetPubKey - Implements mcodec.AccountI.
func (acc BaseAccount) GetPubKey() (pk cryptotypes.PubKey) {
	if acc.PubKey == nil {
		return nil
	}
	content, ok := acc.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil
	}
	return content
}

// SetPubKey - Implements mcodec.AccountI.
func (acc *BaseAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	if pubKey == nil {
		acc.PubKey = nil
		return nil
	}
	any, err := codectypes.NewAnyWithValue(pubKey)
	if err == nil {
		acc.PubKey = any
	}
	return err
}

// GetAccountNumber - Implements AccountI
func (acc BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

// SetAccountNumber - Implements AccountI
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// GetSequence - Implements mcodec.AccountI.
func (acc BaseAccount) GetSequence() uint64 {
	return acc.Sequence
}

// SetSequence - Implements mcodec.AccountI.
func (acc *BaseAccount) SetSequence(seq uint64) error {
	acc.Sequence = seq
	return nil
}

// Validate checks for errors on the account fields
func (acc BaseAccount) Validate() error {
	if acc.Address == "" || acc.PubKey == nil {
		return nil
	}

	accAddr, err := mcodec.AccAddressPrefixedFromBech32(acc.Address)
	if err != nil {
		return err
	}

	if !bytes.Equal(acc.GetPubKey().Address().Bytes(), accAddr.Bytes()) {
		return errors.New("account address and pubkey address do not match")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (acc BaseAccount) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if acc.PubKey == nil {
		return nil
	}
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(acc.PubKey, &pubKey)
}

func (acc BaseAccount) ToCosmosAccount() *authtypes.BaseAccount {
	return authtypes.NewBaseAccount(
		acc.GetAddress().Bytes(),
		acc.GetPubKey(),
		acc.GetAccountNumber(),
		acc.GetSequence(),
	)
}

func (acc BaseAccount) ToCosmosAccountI() sdk.AccountI {
	return acc.ToCosmosAccount()
}

// NewModuleAddressOrAddress gets an input string and returns an AccAddress.
// If the input is a valid address, it returns the address.
// If the input is a module name, it returns the module address.
func NewModuleAddressOrBech32Address(addrCodec mcodec.AccBech32Codec, input string) mcodec.AccAddressPrefixed {
	if addr, err := mcodec.AccAddressPrefixedFromBech32(input); err == nil {

		return addr
	}
	return addrCodec.BytesToAccAddressPrefixed(NewModuleAddress(input))
}

// NewModuleAddress creates an AccAddress from the hash of the module's name
func NewModuleAddress(name string) sdk.AccAddress {
	return address.Module(name)
}

// NewEmptyModuleAccount creates a empty ModuleAccount from a string
func NewEmptyModuleAccount(addrCodec mcodec.AccBech32Codec, name string, permissions ...string) *ModuleAccount {
	moduleAddress := addrCodec.BytesToAccAddressPrefixed(NewModuleAddress(name))
	baseAcc := NewBaseAccountWithAddress(moduleAddress)

	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: baseAcc,
		Name:        name,
		Permissions: permissions,
	}
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(ba *BaseAccount, name string, permissions ...string) *ModuleAccount {
	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: ba,
		Name:        name,
		Permissions: permissions,
	}
}

// HasPermission returns whether or not the module account has permission.
func (ma ModuleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.Name
}

// GetPermissions returns permissions granted to the module account
func (ma ModuleAccount) GetPermissions() []string {
	return ma.Permissions
}

// SetPubKey - Implements AccountI
func (ma ModuleAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	return fmt.Errorf("not supported for module accounts")
}

// Validate checks for errors on the account fields
func (ma ModuleAccount) Validate() error {
	if strings.TrimSpace(ma.Name) == "" {
		return errors.New("module account name cannot be blank")
	}

	if ma.BaseAccount == nil {
		return errors.New("uninitialized ModuleAccount: BaseAccount is nil")
	}

	if ma.Address != sdk.AccAddress(crypto.AddressHash([]byte(ma.Name))).String() {
		return fmt.Errorf("address %s cannot be derived from the module name '%s'", ma.Address, ma.Name)
	}

	return ma.BaseAccount.Validate()
}

func (acc ModuleAccount) ToCosmosAccount() *authtypes.ModuleAccount {
	return authtypes.NewModuleAccount(
		acc.BaseAccount.ToCosmosAccount(),
		acc.GetName(),
		acc.GetPermissions()...,
	)
}

func (acc ModuleAccount) ToCosmosAccountI() sdk.AccountI {
	return acc.ToCosmosAccount()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (acc ModuleAccount) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if acc.PubKey == nil {
		return nil
	}
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(acc.PubKey, &pubKey)
}

// GenesisAccounts defines a slice of GenesisAccount objects
type GenesisAccounts []GenesisAccount

// Contains returns true if the given address exists in a slice of GenesisAccount
// objects.
func (ga GenesisAccounts) Contains(addr sdk.Address) bool {
	for _, acc := range ga {
		if bytes.Equal(acc.GetAddressPrefixed().Bytes(), addr.Bytes()) {
			return true
		}
	}

	return false
}

// GenesisAccount defines a genesis account that embeds an AccountI with validation capabilities.
type GenesisAccount interface {
	mcodec.AccountI

	Validate() error
}

// performs basic permission validation
func validatePermissions(permissions ...string) error {
	for _, perm := range permissions {
		if strings.TrimSpace(perm) == "" {
			return fmt.Errorf("module permission is empty")
		}
	}
	return nil
}
