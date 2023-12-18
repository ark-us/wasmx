package keeper

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type IntervalAction struct {
	Delay  int64
	Repeat int32
	Args   []byte
}

type msgServer struct {
	App       types.BaseApp
	intervals map[int32]*IntervalAction
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper, app types.BaseApp) types.MsgServer {
	return &msgServer{
		Keeper:    keeper,
		App:       app,
		intervals: map[int32]*IntervalAction{},
	}
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServer(
	keeper Keeper,
) types.MsgServer {
	return &msgServer{
		Keeper:    keeper,
		intervals: map[int32]*IntervalAction{},
	}
}

var _ types.MsgServer = msgServer{}
var _ types.BroadcastAPIServer = msgServer{}

func (m msgServer) Ping(goCtx context.Context, msg *types.RequestPing) (*types.ResponsePing, error) {
	return &types.ResponsePing{}, nil
}

func (m msgServer) BroadcastTx(goCtx context.Context, msg *types.RequestBroadcastTx) (*types.ResponseBroadcastTx, error) {
	// TODO BroadcastTxCommit and return receipt

	msgbz := []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(msg.Tx)))
	rresp, err := m.ExecuteContract(goCtx, &types.MsgExecuteContract{
		Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Msg:      msgbz,
	})
	fmt.Println("--ExecuteContract BroadcastTxAsync--", rresp, err)
	if err != nil {
		return nil, err
	}
	fmt.Println("--ExecuteContract BroadcastTxAsync--", string(rresp.Data))

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

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the interval is for?
func (m msgServer) StartTimeout(goCtx context.Context, msg *types.MsgStartTimeoutRequest) (*types.MsgStartTimeoutResponse, error) {
	fmt.Println("Go - start interval request", msg.Contract, msg.Delay, string(msg.Args))
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "ExecuteEth could not parse sender address")
	}

	description := fmt.Sprintf("timed action: delay %dms, args: %s ", msg.Delay, string(msg.Args))

	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Args}
	msgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	timeDelay := msg.Delay
	logger := m.Keeper.Logger(ctx)

	// errCh := make(chan error)
	m.goRoutineGroup.Go(func() error {
		_, err := m.startTimeoutInternalGoroutine(logger, description, timeDelay, msgbz, contractAddress)
		if err != nil {
			logger.Error("eventual execution failed", "err", err, "description", description)
		}
		return nil
	})

	return &types.MsgStartTimeoutResponse{}, nil
}

func (m msgServer) startTimeoutInternalGoroutine(
	logger log.Logger,
	description string,
	timeDelay int64,
	msgbz []byte,
	contractAddress sdk.AccAddress,
) (chan struct{}, error) {
	goCtx2 := m.goContextParent
	httpSrvDone := make(chan struct{}, 1)
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error)
	go func() {
		logger.Info("eventual execution starting", "description", description)
		err := m.startTimeoutInternal(logger, description, timeDelay, msgbz, contractAddress)
		if err != nil {
			logger.Error("eventual execution failed", "err", err)
			// return err
			errCh <- err
		}
		logger.Info("eventual execution ended", "description", description)
		// close(httpSrvDone)
		intervalEnded <- true
	}()

	select {
	case <-goCtx2.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the network server.
		logger.Info("eventual execution stopping...")
		// httpSrv.Close()
		return httpSrvDone, nil
	case err := <-errCh:
		logger.Error("eventual execution failed to start", "error", err.Error())
		return nil, err
	case <-intervalEnded:
		return httpSrvDone, nil
	}
}

func (m msgServer) startTimeoutInternal(
	logger log.Logger,
	description string,
	timeDelay int64,
	msgbz []byte,
	contractAddress sdk.AccAddress,
) error {
	// sleep first and then load the context
	time.Sleep(time.Duration(timeDelay) * time.Millisecond)

	// goCtx := context.Background()
	goCtx2 := m.goContextParent
	height := m.App.LastBlockHeight()

	// set context
	sdkCtx_, ctxcachems, err := CreateQueryContext(m.App, logger, height, false)
	if err != nil {
		logger.Error("failed to create query context", "err", err)
		return err
	}
	sdkCtx, commitCacheCtx := sdkCtx_.CacheContext()

	// // Add relevant gRPC headers
	// if height == 0 {
	// 	height = sdkCtx.BlockHeight() // If height was not set in the request, set it to the latest
	// }

	// Attach the sdk.Context into the gRPC's context.Context.
	// grpcCtx = context.WithValue(grpcCtx, sdk.SdkContextKey, sdkCtx)

	// md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))
	// if err = grpc.SetHeader(grpcCtx, md); err != nil {
	// 	logger.Error("failed to set gRPC header", "err", err)
	// }

	// NOW execute


	goCtx2 = context.WithValue(goCtx2, sdk.SdkContextKey, sdkCtx)
	ctx := sdk.UnwrapSDKContext(goCtx2)

	_, err = m.Keeper.wasmxKeeper.ExecuteEventual(ctx, contractAddress, contractAddress, msgbz, make([]string, 0))
	if err != nil {
		// TODO - stop without propagating a stop to parent
		if err == types.ErrGoroutineClosed {
			m.Logger(ctx).Error("Closing eventual thread", "description", description, err.Error())
			return nil
		}

		m.Logger(ctx).Error("eventual execution failed", "error", err.Error())
		return err
	}

	commitCacheCtx()
	// commit original context
	mythosapp, ok := m.App.(MythosApp)
	if !ok {
		return fmt.Errorf("failed to get MythosApp from server Application")
	}
	origtstore := ctxcachems.GetStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
	origtstore.(storetypes.CacheWrap).Write()
	return nil
}

func (m msgServer) GrpcSendRequest(goCtx context.Context, msg *types.MsgGrpcSendRequest) (*types.MsgGrpcSendRequestResponse, error) {
	fmt.Println("Go - send grpc request", msg.IpAddress, msg.Data, string(msg.Data))
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
	fmt.Println("Go - grpc request sent", res, err)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrpcSendRequestResponse{
		Data: res.Data,
	}, nil
}

func (m msgServer) GrpcReceiveRequest(goCtx context.Context, msg *types.MsgGrpcReceiveRequest) (*types.MsgGrpcReceiveRequestResponse, error) {
	fmt.Println("Go - received grpc request", string(msg.Data))
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "ExecuteEth could not parse sender address")
	}

	data := msg.Data
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: data}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	// fmt.Println("-GrpcReceiveRequest-network-execmsgbz--", hex.EncodeToString(execmsgbz))
	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	fmt.Println("Go - received grpc request and executed state machine", err)
	if err != nil {
		return nil, err
	}
	fmt.Println("Go - received grpc request and executed state machine", resp, string(resp), err)

	return &types.MsgGrpcReceiveRequestResponse{
		Data: resp,
	}, nil
}

func (m msgServer) ExecuteContract(goCtx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, senderAddr, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteContractResponse{
		Data: resp,
	}, nil
}

func (m msgServer) QueryContract(goCtx context.Context, msg *types.MsgQueryContract) (*types.MsgQueryContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := m.wasmxKeeper.Query(ctx, contractAddress, senderAddr, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgQueryContractResponse{
		Data: resp,
	}, nil
}
