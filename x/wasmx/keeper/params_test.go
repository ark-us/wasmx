package keeper_test

import (
	"testing"

	testkeeper "wasmx/testutil/keeper"
	"wasmx/x/wasmx/types"

	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.WasmxKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
