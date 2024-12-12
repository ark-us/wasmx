package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/loredanacirstea/wasmx/v1/testutil/keeper"
	"github.com/loredanacirstea/wasmx/v1/x/wasmx/keeper"
	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WasmxKeeper(t)
	return keeper.NewMsgServerImpl(k), ctx
}
