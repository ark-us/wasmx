package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"mythos/v1/x/network/types"
)

type (
	Keeper struct {
		cdc         codec.Codec
		storeKey    storetypes.StoreKey
		memKey      storetypes.MemoryStoreKey
		paramstore  paramtypes.Subspace
		wasmxKeeper types.WasmxKeeper

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	memKey storetypes.MemoryStoreKey,
	ps paramtypes.Subspace,
	wasmxKeeper types.WasmxKeeper,
	authority string,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	keeper := &Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		memKey:      memKey,
		paramstore:  ps,
		wasmxKeeper: wasmxKeeper,
		authority:   authority,
	}
	return keeper
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
