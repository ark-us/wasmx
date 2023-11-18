package network

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/keeper"
	"mythos/v1/x/network/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genState := &types.GenesisState{
		Params: k.GetParams(ctx),
	}
	return genState
}
