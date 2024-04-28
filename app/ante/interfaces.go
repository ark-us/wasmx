package ante

import (
	address "cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmxKeeperI interface {
	GetAlias(ctx sdk.Context, addr sdk.AccAddress) (sdk.AccAddress, bool)
	AddressCodec() address.Codec
	ValidatorAddressCodec() address.Codec
	ConsensusAddressCodec() address.Codec
}
