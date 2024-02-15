package keeper

import (
	"context"
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type msgServer struct {
	App types.BaseApp
	*Keeper
}

type MsgServerInternal interface {
	types.MsgServer
	ExecuteContract(ctx sdk.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error)
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper, app types.BaseApp) MsgServerInternal {
	return &msgServer{
		Keeper: keeper,
		App:    app,
	}
}

var _ types.MsgServer = msgServer{}

// TODO do we stil nee this?
var _ types.BroadcastAPIServer = msgServer{}

func (m msgServer) Ping(goCtx context.Context, msg *types.RequestPing) (*types.ResponsePing, error) {
	return &types.ResponsePing{}, nil
}

func (m msgServer) BroadcastTx(goCtx context.Context, msg *types.RequestBroadcastTx) (*types.ResponseBroadcastTx, error) {
	// TODO BroadcastTxCommit and return receipt
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgbz := []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(msg.Tx)))
	rresp, err := m.Keeper.ExecuteContract(ctx, &types.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_CONSENSUS,
		Contract: wasmxtypes.ROLE_CONSENSUS,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}

	// return &types.ResponseBroadcastTx{
	// 	CheckTx: &abci.ResponseCheckTx{
	// 		Code: res.CheckTx.Code,
	// 		Data: res.CheckTx.Data,
	// 		Log:  res.CheckTx.Log,
	// 	},
	// 	TxResult: &abci.ExecTxResult{
	// 		Code: res.TxResult.Code,
	// 		Data: res.TxResult.Data,
	// 		Log:  res.TxResult.Log,
	// 	},
	// }, nil

	return &types.ResponseBroadcastTx{
		CheckTx: &types.ResponseCheckTx{
			Code: 0,
			Data: rresp.Data,
			Log:  "",
		},
		TxResult: &types.ExecTxResult{
			Code: 0,
			Data: rresp.Data,
			Log:  "",
		},
	}, nil
}

func (m msgServer) GrpcSendRequest(goCtx context.Context, msg *types.MsgGrpcSendRequest) (*types.MsgGrpcSendRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ip := msg.IpAddress
	client := StartGRPCClient(ip)
	req := &types.MsgGrpcReceiveRequest{
		Data:     msg.Data,
		Sender:   msg.Sender,
		Contract: msg.Contract,
		Encoding: msg.Encoding,
	}
	res, err := client.GrpcReceiveRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrpcSendRequestResponse{
		Data: res.Data,
	}, nil
}

func (m msgServer) GrpcReceiveRequest(goCtx context.Context, msg *types.MsgGrpcReceiveRequest) (*types.MsgGrpcReceiveRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	msg2 := &types.MsgExecuteContract{
		Sender:   msg.Sender,
		Contract: msg.Contract,
		Msg:      msg.Data,
	}
	resp, err := m.Keeper.ExecuteContract(ctx, msg2)

	if err != nil {
		return nil, err
	}

	return &types.MsgGrpcReceiveRequestResponse{
		Data: resp.Data,
	}, nil
}

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the interval is for?
func (m msgServer) StartTimeout(goCtx context.Context, msg *types.MsgStartTimeoutRequest) (*types.MsgStartTimeoutResponse, error) {
	return m.Keeper.StartTimeout(goCtx, msg)
}

func (m msgServer) P2PReceiveMessage(goCtx context.Context, msg *types.MsgP2PReceiveMessageRequest) (*types.MsgP2PReceiveMessageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	msg2 := &types.MsgExecuteContract{
		Sender:   msg.Sender,
		Contract: msg.Contract,
		Msg:      msg.Data,
	}
	_, err := m.Keeper.ExecuteEntryPoint(ctx, wasmxtypes.ENTRY_POINT_P2P_MSG, msg2)

	if err != nil {
		return nil, err
	}

	return &types.MsgP2PReceiveMessageResponse{}, nil
}
