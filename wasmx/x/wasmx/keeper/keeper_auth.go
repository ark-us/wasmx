package keeper

import (
	"context"

	address "cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
	verifysig "github.com/loredanacirstea/wasmx/v1/crypto/verifysig"
	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
)

type AccountKeeperVerifySig struct {
	k types.AccountKeeper
}

func NewAccountKeeperVerifySig(keeper types.AccountKeeper) verifysig.AccountKeeper {
	if keeper == nil {
		panic("AccountKeeper not found")
	}
	return AccountKeeperVerifySig{k: keeper}
}

func (ak AccountKeeperVerifySig) GetAccountPrefixed(goCtx context.Context, addr mcodec.AccAddressPrefixed) (mcodec.AccountI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return ak.k.GetAccountPrefixed(ctx, addr)
}

func (ak AccountKeeperVerifySig) AddressCodec() address.Codec {
	return ak.k.AddressCodec()
}
