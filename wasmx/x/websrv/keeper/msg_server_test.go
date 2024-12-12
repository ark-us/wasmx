package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/loredanacirstea/wasmx/v1/testutil/keeper"
	"github.com/loredanacirstea/wasmx/v1/x/websrv/keeper"
	"github.com/loredanacirstea/wasmx/v1/x/websrv/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WebsrvKeeper(t)
	return keeper.NewMsgServerImpl(*k), ctx
}
