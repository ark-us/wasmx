package keeper_test

import (
	"github.com/stretchr/testify/require"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	t := suite.T()
	keeper := suite.WasmxKeeper
	ctx := suite.Ctx
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)
	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
