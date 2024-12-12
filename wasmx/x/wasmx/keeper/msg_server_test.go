package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/loredanacirstea/wasmx/testutil/keeper"
	"github.com/loredanacirstea/wasmx/x/wasmx/keeper"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WasmxKeeper(t)
	return keeper.NewMsgServerImpl(k), ctx
}
