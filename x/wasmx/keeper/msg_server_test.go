package keeper_test

import (
	"context"
	"testing"

	keepertest "mythos/v1/testutil/keeper"
	"mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WasmxKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
