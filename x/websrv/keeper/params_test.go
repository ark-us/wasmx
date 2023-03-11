package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "wasmx/testutil/keeper"
	"wasmx/x/websrv/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.WebsrvKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
