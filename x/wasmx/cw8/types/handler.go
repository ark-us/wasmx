package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WasmVMResponseHandler is an extension point to handles the response data returned by a contract call.
type WasmVMResponseHandler interface {
	// Handle processes the data returned by a contract invocation.
	Handle(
		ctx sdk.Context,
		contractAddr sdk.AccAddress,
		ibcPort string,
		messages []SubMsg,
		origRspData []byte,
	) ([]byte, error)
}
