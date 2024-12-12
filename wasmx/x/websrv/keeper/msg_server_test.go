package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/loredanacirstea/wasmx/testutil/keeper"
	"github.com/loredanacirstea/wasmx/x/websrv/keeper"
	"github.com/loredanacirstea/wasmx/x/websrv/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WebsrvKeeper(t)
	return keeper.NewMsgServerImpl(*k), ctx
}
