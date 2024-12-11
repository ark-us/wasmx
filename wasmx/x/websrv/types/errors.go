package types

// DONTCOVER

import (
	sdkerr "cosmossdk.io/errors"
)

// x/websrv module sentinel errors
var (
	DefaultCodespace = ModuleName

	// ErrSample = sdkerr.Register(ModuleName, 1100, "sample error")

	// ErrRouteNotFound error for wasm code that has already been uploaded or failed
	ErrRouteNotFound = sdkerr.Register(DefaultCodespace, 2, "route not found")

	ErrRouteInternalError = sdkerr.Register(DefaultCodespace, 3, "route internal error")

	ErrEmptyRoute = sdkerr.Register(DefaultCodespace, 4, "empty route provided")

	ErrWebsrvInternal = sdkerr.Register(DefaultCodespace, 5, "websrv internal error")

	ErrOAuthClientInvalidDomain = sdkerr.Register(DefaultCodespace, 6, "websrv oauth invalid client domain")

	ErrOAuthTooManyClientsRegistered = sdkerr.Register(DefaultCodespace, 7, "websrv oauth too many clients registered")

	ErrInvalid = sdkerr.Register(DefaultCodespace, 8, "invalid")
)
