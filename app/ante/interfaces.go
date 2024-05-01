package ante

import (
	address "cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "mythos/v1/codec"
)

type WasmxKeeperI interface {
	GetAlias(ctx sdk.Context, addr mcodec.AccAddressPrefixed) (mcodec.AccAddressPrefixed, bool)
	AddressCodec() address.Codec
	ValidatorAddressCodec() address.Codec
	ConsensusAddressCodec() address.Codec
	AccBech32Codec() mcodec.AccBech32Codec
}
