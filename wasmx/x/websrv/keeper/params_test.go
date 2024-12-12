package keeper_test

import (
	"testing"

	testkeeper "github.com/loredanacirstea/wasmx/testutil/keeper"
	"github.com/loredanacirstea/wasmx/x/websrv/types"

	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.WebsrvKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
