package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "wasmx/v1/testutil/keeper"
	"wasmx/v1/x/wasmx/types"

	memc "wasmx/v1/x/wasmx/vm/memory/common"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.WasmxKeeper(t, memc.WasmRuntimeMockVmMeta{})
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
