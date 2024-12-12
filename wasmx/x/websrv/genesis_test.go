package websrv_test

import (
	"testing"

	keepertest "github.com/loredanacirstea/wasmx/testutil/keeper"
	"github.com/loredanacirstea/wasmx/testutil/nullify"
	"github.com/loredanacirstea/wasmx/x/websrv"
	"github.com/loredanacirstea/wasmx/x/websrv/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.WebsrvKeeper(t)
	websrv.InitGenesis(ctx, *k, genesisState)
	got := websrv.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
