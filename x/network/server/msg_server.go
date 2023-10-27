package server

import (
	"context"
	"fmt"

	"mythos/v1/x/network/types"
)

type msgServer struct {
	// Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl() types.MsgServer {
	return &msgServer{}
}

var _ types.MsgServer = msgServer{}

// This is the entrypoint for transactions signed by Ethereum wallets
// Works with both EVM & CosmWasm contracts, both interpreted and wasm-based
func (m msgServer) Ping(goCtx context.Context, msg *types.MsgPing) (*types.MsgPingResponse, error) {
	fmt.Println("--Ping", msg.Message)

	return &types.MsgPingResponse{
		Message: msg.Message,
	}, nil
}
