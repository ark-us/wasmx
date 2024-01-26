package types

import (
	sdkerr "cosmossdk.io/errors"
)

// Codes for wasm contract errors
var (
	DefaultCodespace = ModuleName

	// Note: code 1 is reserved for ErrInternal in the core cosmos sdk

	ErrInternal = sdkerr.Register(DefaultCodespace, 2, "internal error")
)
