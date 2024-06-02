package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/sync/errgroup"

	"mythos/v1/x/network/types"
	"mythos/v1/x/network/vmcrosschain"
	"mythos/v1/x/network/vmmc"
	"mythos/v1/x/network/vmp2p"
)

type (
	Keeper struct {
		goRoutineGroup  *errgroup.Group
		goContextParent context.Context
		cdc             codec.Codec
		wasmxKeeper     types.WasmxKeeper
		actionExecutor  *ActionExecutor

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func init() {
	vmp2p.Setup()
	vmmc.Setup()
	vmcrosschain.Setup()
}

func NewKeeper(
	goRoutineGroup *errgroup.Group,
	goContextParent context.Context,
	cdc codec.Codec,
	wasmxKeeper types.WasmxKeeper,
	actionExecutor *ActionExecutor,
	authority string,
) *Keeper {
	keeper := &Keeper{
		goRoutineGroup:  goRoutineGroup,
		goContextParent: goContextParent,
		cdc:             cdc,
		wasmxKeeper:     wasmxKeeper,
		actionExecutor:  actionExecutor,
		authority:       authority,
	}
	return keeper
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) Codec() codec.Codec {
	return k.cdc
}
