package keeper

import (
	"context"
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// TODO security - these network messages are entrypoints for deterministic executions triggered by single-consensus contracts
// they must not be triggered by users
// they must not be called by normal, deterministic contracts
// same for queries
type msgServer struct {
	*Keeper
}

type MsgServerInternal interface {
	types.MsgServer
	ExecuteContract(ctx sdk.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error)
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) MsgServerInternal {
	return &msgServer{
		Keeper: keeper,
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

	// TODO fixme?
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

// Any execution message can be wrapped with MsgMultiChainWrap to be executed on one
// of the available chains.
// BroadcastTxAsync peeks inside the transaction and inside MsgMultiChainWrap to get the chainId
// and then forwards the transaction to the apropriate chain application
// the signature & signer are verified in the AnteHandler of that chain application
func (m msgServer) MultiChainWrap(goCtx context.Context, msg *types.MsgMultiChainWrap) (*types.MsgMultiChainWrapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO validation
	return m.MultiChainWrapInternal(ctx, msg)
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

// TODO should be reentry
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

func (m msgServer) Reentry(goCtx context.Context, msg *types.MsgReentry) (*types.MsgReentryResponse, error) {
	return m.Keeper.Reentry(goCtx, msg)
}

// TODO this must not be called from outside, only from wasmx... (authority)
// only from the contract that the interval is for?
func (m msgServer) StartTimeout(goCtx context.Context, msg *types.MsgStartTimeoutRequest) (*types.MsgStartTimeoutResponse, error) {
	return m.Keeper.StartTimeout(goCtx, msg)
}

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the background process is for ?
func (m msgServer) StartBackgroundProcess(goCtx context.Context, msg *types.MsgStartBackgroundProcessRequest) (*types.MsgStartBackgroundProcessResponse, error) {
	// TODO only started from wasmx, only by system contracts
	return m.Keeper.StartBackgroundProcess(goCtx, msg)
}

func (m msgServer) P2PReceiveMessage(goCtx context.Context, msg *types.MsgP2PReceiveMessageRequest) (*types.MsgP2PReceiveMessageResponse, error) {
	return m.Keeper.P2PReceiveMessage(goCtx, msg)
}

func (m msgServer) ExecuteAtomicTx(goCtx context.Context, msg *types.MsgExecuteAtomicTxRequest) (*types.MsgExecuteAtomicTxResponse, error) {
	return m.Keeper.ExecuteAtomicTx(goCtx, msg)
}

func (m msgServer) ExecuteCrossChainTx(goCtx context.Context, msg *types.MsgExecuteCrossChainCallRequest) (*types.MsgExecuteCrossChainCallResponse, error) {
	return m.Keeper.ExecuteCrossChainTx(goCtx, msg)
}
