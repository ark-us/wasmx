package keeper_test

import (
	"testing"

	testkeeper "github.com/loredanacirstea/wasmx/v1/testutil/keeper"
	"github.com/loredanacirstea/wasmx/v1/x/websrv/types"

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
