package wasmx

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/keeper"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
	for _, contract := range genState.SystemContracts {
		k.SetSystemContract(ctx, contract)
	}
	// TODO
	// genState.Contracts
	// genState.Codes
	// genState.Sequences
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genState := &types.GenesisState{
		Params:          k.GetParams(ctx),
		SystemContracts: []types.SystemContract{},
	}
	k.IterateSystemContracts(ctx, func(contract types.SystemContract) bool {
		genState.SystemContracts = append(genState.SystemContracts, contract)
		return false
	})
	return genState
}
