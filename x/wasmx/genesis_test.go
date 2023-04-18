package wasmx_test

import (
	"testing"

	keepertest "wasmx/v1/testutil/keeper"
	"wasmx/v1/testutil/nullify"
	"wasmx/v1/x/wasmx"
	"wasmx/v1/x/wasmx/types"

	"github.com/stretchr/testify/require"
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
