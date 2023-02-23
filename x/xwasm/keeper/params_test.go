package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "xwasm/testutil/keeper"
	"xwasm/x/xwasm/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.XwasmKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
