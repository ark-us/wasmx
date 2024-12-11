package ante

import (
	"context"

	address "cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	mcodec "wasmx/v1/codec"
)

type WasmxKeeperI interface {
	GetAlias(ctx sdk.Context, addr mcodec.AccAddressPrefixed) (mcodec.AccAddressPrefixed, bool)
	AddressCodec() address.Codec
	ValidatorAddressCodec() address.Codec
	ConsensusAddressCodec() address.Codec
	AccBech32Codec() mcodec.AccBech32Codec
}

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetParams(ctx context.Context) (params types.Params)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(moduleName string) sdk.AccAddress

	GetAccountPrefixed(ctx context.Context, addr mcodec.AccAddressPrefixed) (mcodec.AccountI, error)
	SetAccountPrefixed(goCtx context.Context, acc mcodec.AccountI) error
	AddressCodec() address.Codec
	AccBech32Codec() mcodec.AccBech32Codec
}
