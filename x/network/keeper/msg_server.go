package keeper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/log"
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
	ActionExecutor *ActionExecutor
	App            types.BaseApp
	intervals      map[int32]*IntervalAction
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper, app types.BaseApp, actionExecutor *ActionExecutor) types.MsgServer {
	return &msgServer{
		Keeper:         keeper,
		App:            app,
		ActionExecutor: actionExecutor,
		intervals:      map[int32]*IntervalAction{},
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
		Sender:   "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
		Contract: "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy",
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

	fmt.Println("Number of goroutines start0", runtime.NumGoroutine())

	m.goRoutineGroup.Go(func() error {
		err := m.startTimeoutInternalGoroutine(logger, description, timeDelay, msgbz, contractAddress)
		if err != nil {
			logger.Error("eventual execution failed", "err", err, "description", description)
		}
		return nil
	})

	fmt.Println("Number of goroutines end0", runtime.NumGoroutine())

	return &types.MsgStartTimeoutResponse{}, nil
}

func (m msgServer) startTimeoutInternalGoroutine(
	logger log.Logger,
	description string,
	timeDelay int64,
	msgbz []byte,
	contractAddress sdk.AccAddress,
) error {
	goCtx2 := m.goContextParent

	fmt.Println("Number of goroutines start 10 in .Go", runtime.NumGoroutine())

	select {
	case <-goCtx2.Done():
		logger.Info("parent context was closed, we do not start the delayed execution")
		return nil
	default:
		// continue
	}

	// wg := new(sync.WaitGroup)

	// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
	intervalEnded := make(chan bool, 1)
	errCh := make(chan error, 1)
	// wg.Add(1)
	go func() {
		// defer wg.Done()

		fmt.Println("Number of goroutines start 11 (in goroutine)", runtime.NumGoroutine())

		logger.Info("eventual execution starting", "description", description)
		err := m.startTimeoutInternal(logger, description, timeDelay, msgbz, contractAddress)
		if err != nil {
			logger.Error("eventual execution failed", "err", err)
			// return err
			errCh <- err
		}
		logger.Info("eventual execution ended", "description", description)
		intervalEnded <- true
	}()

	fmt.Println("Number of goroutines start 12 before wg.Wait", runtime.NumGoroutine())

	// wg.Wait()

	fmt.Println("Number of goroutines start 13 after wg.Wait", runtime.NumGoroutine())

	select {
	// case <-goCtx2.Done():
	// 	// The calling process canceled or closed the provided context, so we must
	// 	// gracefully stop the network server.
	// 	logger.Info("eventual execution stopping...")
	// 	close(intervalEnded)
	// 	close(errCh)
	// 	return nil
	case err := <-errCh:
		logger.Error("eventual execution failed to start", "error", err.Error())
		close(intervalEnded)
		return err
	case <-intervalEnded:
		logger.Info("!!!!!intervalEnded", "description", description)
		close(intervalEnded)
		return nil
	case <-time.After(time.Duration(timeDelay)*time.Millisecond + time.Minute):
		return fmt.Errorf("request did not complete within allowed duration")
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

	fmt.Println("Number of goroutines1", runtime.NumGoroutine())

	goCtx2 := m.goContextParent
	select {
	case <-goCtx2.Done():
		logger.Info("parent context was closed, we do not start the delayed execution")
		return nil
	default:
		// continue
	}

	height := m.App.LastBlockHeight()

	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		res, err := m.Keeper.wasmxKeeper.ExecuteEventual(ctx, contractAddress, contractAddress, msgbz, make([]string, 0))
		if err != nil {
			// TODO - stop without propagating a stop to parent
			if err == types.ErrGoroutineClosed {
				m.Logger(ctx).Error("Closing eventual thread", "description", description, err.Error())
				return res, nil
			}
			m.Logger(ctx).Error("eventual execution failed", "error", err.Error())
			return nil, err
		}
		return res, nil
	}
	// disregard result
	_, err := m.ActionExecutor.Execute(goCtx2, height, cb)

	return err
}
