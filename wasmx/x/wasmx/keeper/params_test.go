package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/loredanacirstea/wasmx/v1/testutil/keeper"
	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm/memory/common"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.WasmxKeeper(t, memc.WasmRuntimeMockVmMeta{})
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
