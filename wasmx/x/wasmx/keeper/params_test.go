package keeper_test

import (
	"github.com/stretchr/testify/require"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestGetParams() {
	t := suite.T()
	keeper := suite.WasmxKeeper
	ctx := suite.Ctx
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)
	require.EqualValues(t, params, keeper.GetParams(ctx))
}
