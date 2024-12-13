package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/loredanacirstea/wasmx/testutil/keeper"
	"github.com/loredanacirstea/wasmx/x/wasmx/keeper"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func setupMsgServer(t testing.TB, wasmVmMeta memc.IWasmVmMeta) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WasmxKeeper(t, wasmVmMeta)
	return keeper.NewMsgServerImpl(k), ctx
}
