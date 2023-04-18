package keeper_test

import (
	"context"
	"testing"

	keepertest "wasmx/v1/testutil/keeper"
	"wasmx/v1/x/wasmx/keeper"
	"wasmx/v1/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.WasmxKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
