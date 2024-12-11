package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "wasmx/v1/codec"
	"wasmx/v1/x/wasmx/types"
)

type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)

	AccBech32Codec() mcodec.AccBech32Codec
}
