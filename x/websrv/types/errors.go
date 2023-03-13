package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/websrv module sentinel errors
var (
	DefaultCodespace = ModuleName

	// ErrSample = sdkerrors.Register(ModuleName, 1100, "sample error")

	// ErrRouteNotFound error for wasm code that has already been uploaded or failed
	ErrRouteNotFound = sdkerrors.Register(DefaultCodespace, 2, "route not found")

	ErrRouteInternalError = sdkerrors.Register(DefaultCodespace, 3, "route internal error")

	ErrEmptyRoute = sdkerrors.Register(DefaultCodespace, 4, "empty route provided")

	ErrWebsrvInternal = sdkerrors.Register(DefaultCodespace, 5, "websrv internal error")
)
