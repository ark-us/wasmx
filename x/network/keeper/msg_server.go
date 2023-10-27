package keeper

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// This is the entrypoint for transactions signed by Ethereum wallets
// Works with both EVM & CosmWasm contracts, both interpreted and wasm-based
func (m msgServer) Ping(goCtx context.Context, msg *types.MsgPing) (*types.MsgPingResponse, error) {
	fmt.Println("--Ping", msg.Message)
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress := wasmxtypes.AccAddressFromHex("0x0000000000000000000000000000000000000004")

	bz, err := hex.DecodeString("0x0000000000000000000000000000000000000005")
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	resp, err := m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("--resp---", resp)

	response := msg.Message + hex.EncodeToString(resp)

	return &types.MsgPingResponse{
		Message: response,
	}, nil
}
