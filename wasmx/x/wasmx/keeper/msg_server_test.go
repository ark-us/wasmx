package keeper_test

import (
	"context"
	"testing"

	keepertest "mythos/v1/testutil/keeper"
	"mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WasmxKeeper(t)
	return keeper.NewMsgServerImpl(k), ctx
}
