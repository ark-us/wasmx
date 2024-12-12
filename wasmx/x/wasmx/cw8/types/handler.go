package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
)

// WasmVMResponseHandler is an extension point to handles the response data returned by a contract call.
type WasmVMResponseHandler interface {
	// Handle processes the data returned by a contract invocation.
	Handle(
		ctx sdk.Context,
		contractAddr mcodec.AccAddressPrefixed,
		ibcPort string,
		messages []SubMsg,
		origRspData []byte,
	) ([]byte, error)
}
