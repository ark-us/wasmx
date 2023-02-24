package wasmx_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "wasmx/testutil/keeper"
	"wasmx/testutil/nullify"
	"wasmx/x/wasmx"
	"wasmx/x/wasmx/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.WasmxKeeper(t)
	wasmx.InitGenesis(ctx, *k, genesisState)
	got := wasmx.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
