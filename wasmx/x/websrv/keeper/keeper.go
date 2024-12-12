package keeper

import (
	"context"
	"fmt"

	address "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/loredanacirstea/wasmx/x/websrv/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
		wasmx      types.WasmxKeeper
		query      func(_ context.Context, req *abci.RequestQuery) (res *abci.ResponseQuery, err error)
		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority    string
		addressCodec address.Codec
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	wasmx types.WasmxKeeper,
	query func(_ context.Context, req *abci.RequestQuery) (res *abci.ResponseQuery, err error),
	authority string,
	addressCodec address.Codec,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		memKey:       memKey,
		paramstore:   ps,
		wasmx:        wasmx,
		query:        query,
		authority:    authority,
		addressCodec: addressCodec,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With(log.ModuleKey, fmt.Sprintf("x/%s", types.ModuleName), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) AddressCodec() address.Codec {
	return k.addressCodec
}
