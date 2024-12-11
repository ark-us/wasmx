package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ paramtypes.ParamSet = (*stakingtypes.Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&stakingtypes.Params{})
}

// NewParams creates a new Params instance
func NewParams() stakingtypes.Params {
	return stakingtypes.Params{}
}

// DefaultParams returns a default set of parameters
func DefaultParams() stakingtypes.Params {
	return NewParams()
}
