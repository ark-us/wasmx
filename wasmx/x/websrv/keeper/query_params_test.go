package keeper_test

import (
	"testing"

	testkeeper "wasmx/v1/testutil/keeper"
	"wasmx/v1/x/websrv/types"

	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.WebsrvKeeper(t)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
