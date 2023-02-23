package xwasm_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "xwasm/testutil/keeper"
	"xwasm/testutil/nullify"
	"xwasm/x/xwasm"
	"xwasm/x/xwasm/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.XwasmKeeper(t)
	xwasm.InitGenesis(ctx, *k, genesisState)
	got := xwasm.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
