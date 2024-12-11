package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	DefaultOauthClientRegistrationOnlyEId = false
)

// Parameter keys
var (
	ParamStoreKeyOauthClientRegistrationOnlyEId = []byte("OauthClientRegistrationOnlyEId")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(oauthRegistrationOnlyEId bool) Params {
	return Params{OauthClientRegistrationOnlyEId: oauthRegistrationOnlyEId}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultOauthClientRegistrationOnlyEId)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyOauthClientRegistrationOnlyEId, &p.OauthClientRegistrationOnlyEId, validateBool),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return validateBool(p.OauthClientRegistrationOnlyEId)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
