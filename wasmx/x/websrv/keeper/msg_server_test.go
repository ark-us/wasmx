package keeper_test

import (
	"context"
	"testing"

	keepertest "wasmx/v1/testutil/keeper"
	"wasmx/v1/x/websrv/keeper"
	"wasmx/v1/x/websrv/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WebsrvKeeper(t)
	return keeper.NewMsgServerImpl(*k), ctx
}
