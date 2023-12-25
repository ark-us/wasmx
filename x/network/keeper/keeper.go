package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"golang.org/x/sync/errgroup"

	"mythos/v1/x/network/types"
)

type (
	Keeper struct {
		goRoutineGroup  *errgroup.Group
		goContextParent context.Context
		cdc             codec.Codec
		storeKey        storetypes.StoreKey
		memKey          storetypes.StoreKey
		tKey            storetypes.StoreKey
		clessKey        storetypes.StoreKey
		paramstore      paramtypes.Subspace
		wasmxKeeper     types.WasmxKeeper
		actionExecutor  *ActionExecutor

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	goRoutineGroup *errgroup.Group,
	goContextParent context.Context,
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	tKey storetypes.StoreKey,
	clessKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	wasmxKeeper types.WasmxKeeper,
	actionExecutor *ActionExecutor,
	authority string,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	keeper := &Keeper{
		goRoutineGroup:  goRoutineGroup,
		goContextParent: goContextParent,
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		tKey:            tKey,
		clessKey:        clessKey,
		paramstore:      ps,
		wasmxKeeper:     wasmxKeeper,
		actionExecutor:  actionExecutor,
		authority:       authority,
	}
	return keeper
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}
