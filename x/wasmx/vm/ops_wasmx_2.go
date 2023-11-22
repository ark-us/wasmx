package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"

	networktypes "mythos/v1/x/network/types"
)

// getEnv(): ArrayBuffer
func getEnv(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	envbz, err := json.Marshal(ctx.Env)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, envbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// address -> account
func getAccount(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	addr, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	codeInfo := ctx.CosmosHandler.GetCodeInfo(address)
	code := types.EnvContractInfo{
		Address:  address,
		CodeHash: codeInfo.CodeHash,
		Bytecode: codeInfo.InterpretedBytecodeRuntime,
	}

	codebz, err := json.Marshal(code)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, codebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func keccak256Util(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	if ctx.ContractRouter["keccak256"] == nil {
		return nil, wasmedge.Result_Fail
	}
	keccakVm := ctx.ContractRouter["keccak256"].Vm
	input_offset := int32(0)
	input_length := int32(len(data))
	output_offset := input_length
	context_offset := output_offset + int32(32)

	keccakMem := keccakVm.GetActiveModule().FindMemory("memory")
	if keccakMem == nil {
		return nil, wasmedge.Result_Fail
	}
	err = keccakMem.SetData(data, uint(input_offset), uint(input_length))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	_, err = keccakVm.Execute("keccak", context_offset, input_offset, input_length, output_offset)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	result, err := keccakMem.GetData(uint(output_offset), uint(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, result)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// call request -> call response
func externalCall(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.CallRequest
	json.Unmarshal(requestbz, &req)

	returns := make([]interface{}, 1)
	var success int32
	var returnData []byte

	// Send funds
	if req.Value.BitLen() > 0 {
		err = ctx.CosmosHandler.SendCoin(req.To, req.Value)
	}
	if err != nil {
		success = int32(2)
	} else {
		success, returnData = WasmxCall(ctx, req)
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetBalance(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	addr, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	balance := ctx.CosmosHandler.GetBalance(address)
	ptr, err := allocateWriteMem(ctx, callframe, balance.FillBytes(make([]byte, 32)))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetBlockHash(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	bz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	blockNumber := big.NewInt(0).SetBytes(bz)
	data := ctx.CosmosHandler.GetBlockHash(blockNumber.Uint64())
	ptr, err := allocateWriteMem(ctx, callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreateAccount(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.CreateAccountRequest
	json.Unmarshal(requestbz, &req)
	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	var sdeps []string
	for _, dep := range ctx.ContractRouter[ctx.Env.Contract.Address.String()].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}
	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		req.Bytecode,
		ctx.Env.CurrentCall.Origin,
		ctx.Env.Contract.Address,
		initMsg,
		req.Balance,
		sdeps,
		metadata,
		"", // TODO label?
		[]byte{},
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	contractbz := paddLeftTo32(contractAddress.Bytes())
	ptr, err := allocateWriteMem(ctx, callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreate2Account(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.Create2AccountRequest
	json.Unmarshal(requestbz, &req)

	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	var sdeps []string
	for _, dep := range ctx.ContractRouter[ctx.Env.Contract.Address.String()].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}

	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		req.Bytecode,
		ctx.Env.CurrentCall.Origin,
		ctx.Env.Contract.Address,
		initMsg,
		req.Balance,
		sdeps,
		metadata,
		"", // TODO label?
		req.Salt.FillBytes(make([]byte, 32)),
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	contractbz := paddLeftTo32(contractAddress.Bytes())
	ptr, err := allocateWriteMem(ctx, callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

type GrpcRequest struct {
	Address string `json:"address"`
	Data    string `json:"data"` // should be []byte (base64 encoded)
}

func wasmxGrpcRequest(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	databz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var data GrpcRequest
	err = json.Unmarshal(databz, &data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	msg := &networktypes.MsgGrpcSendRequest{
		Address: data.Address,
		Data:    []byte(data.Data),
		Sender:  ctx.Env.Contract.Address.String(),
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msg)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO evs?
	fmt.Println("--evs", evs)
	ptr, err := allocateWriteMem(ctx, callframe, res)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// func executeTimedAction(_context *Context, intervalId int32, argsPtr int32) error {
// 	fmt.Println("executeTimedAction", intervalId, argsPtr)
// 	contractCtx, ok := _context.ContractRouter[_context.Env.Contract.Address.String()]
// 	if !ok {
// 		return fmt.Errorf("timed action: contract context not found")
// 	}
// 	executeHandler := GetExecuteFunctionHandler(contractCtx.ContractInfo.SystemDeps)

// 	args := []interface{}{intervalId, argsPtr}
// 	_, err := executeHandler(_context, contractCtx.Vm, types.ENTRY_POINT_TIMED, args)
// 	fmt.Println("--executeTimedAction--", err)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// startInterval(repeat: i32, time: u64, args: ArrayBuffer): i32
func wasmxStartInterval(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	repeatCount := params[0].(int32)
	timeDelay := params[1].(int64)
	argsbz, err := readMemFromPtr(callframe, params[2])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--wasmxStartInterval--", repeatCount, timeDelay, string(argsbz))

	// g, ctx2, cancelFn := getCtx(ctx.GetContext().Logger(), true)
	// g := ctx.goRoutineGroup
	// fmt.Println("--ctx.goRoutineGroup--", g)
	// cancelFn must be called to release resources when the process is done
	// ctx2, cancelFn := context.WithCancel(ctx.Ctx)

	// intervalId := ctx.intervalsCount
	// fmt.Println("--intervalId--", intervalId)
	// ctx.intervalsCount += 1
	// ctx.intervals[intervalId] = &IntervalAction{
	// 	Time:   timeDelay,
	// 	Args:   argsbz,
	// 	Cancel: cancelFn,
	// }

	msgtosend := &networktypes.MsgStartIntervalRequest{
		// Sender:  sdk.AccAddress([]byte("network")).String(),
		Sender:  ctx.Env.Contract.Address.String(),
		Address: ctx.Env.Contract.Address.String(),
		Delay:   timeDelay,
		Repeat:  repeatCount,
		Args:    argsbz,
	}
	_, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	fmt.Println("--ExecuteCosmosMsg--err-", err)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var resp networktypes.MsgStartIntervalResponse
	err = resp.Unmarshal(res)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	intervalId := resp.IntervalId
	fmt.Println("---timer created intervalId", intervalId)

	// description := fmt.Sprintf("timed action: id %s, delay %sms, repeat %s, args: %s ", intervalId, timeDelay, repeatCount, string(argsbz))
	// action := func() error {
	// 	_intervalId := big.NewInt(int64(intervalId))
	// 	data := append(_intervalId.FillBytes(make([]byte, 4)), argsbz...)
	// 	execmsg := types.WasmxExecutionMessage{Data: data}
	// 	msgbz, err := json.Marshal(execmsg)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// _, err := startTimedAction(ctx, callframe, ctx2, ctx.GetContext().Logger(), intervalId, timeDelay, repeatCount, argsbz)
	// 	return err
	// }
	// gracefulStop := func() {

	// }
	// done, err := ctx.createGoRoutine(description, action, gracefulStop)

	// g.Go(func() error {
	// 	fmt.Println("--startTimedAction--")
	// 	_, err = startTimedAction(ctx, callframe, ctx2, ctx.GetContext().Logger(), intervalId, timeDelay, repeatCount, argsbz)
	// 	if err != nil {
	// 		ctx.GetContext().Logger().Error("failed to start timed action", "err", err)
	// 		return err
	// 	}

	// 	// Wait for the calling process to be canceled or close the provided context,
	// 	// so we can gracefully stop the ABCI server.
	// 	<-ctx2.Done()
	// 	ctx.GetContext().Logger().Info("stopping the timed action...")
	// 	cancelFn()
	// 	return nil
	// })

	// err = g.Wait() // use the given wait
	// if err != nil {
	// 	return nil, wasmedge.Result_Fail
	// }
	returns[0] = intervalId
	return returns, wasmedge.Result_Success
}

// stopInterval(intervalId: i32): void
func wasmxStopInterval(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	intervalId := params[0].(int32)
	fmt.Println("* stop interval: ", intervalId)
	// TODO errors if already stopped?
	// ctx.intervals[intervalId].Quit <- true
	ctx.intervals[intervalId].Cancel()
	return returns, wasmedge.Result_Success
}

func BuildWasmxEnv2(context *Context) *wasmedge.Module {
	fmt.Println("--BuildWasmxEnv2---")
	env := wasmedge.NewModule("wasmx")
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i64i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i32, getCallData, context, 0))
	env.AddFunction("getEnv", wasmedge.NewFunction(functype__i32, getEnv, context, 0))
	env.AddFunction("getCaller", wasmedge.NewFunction(functype__i32, wasmxGetCaller, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoad, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32_, wasmxStorageStore, context, 0))
	env.AddFunction("log", wasmedge.NewFunction(functype_i32_, wasmxLog, context, 0))
	env.AddFunction("finish", wasmedge.NewFunction(functype_i32_, wasmxFinish, context, 0))
	env.AddFunction("revert", wasmedge.NewFunction(functype_i32_, wasmxRevert, context, 0))
	env.AddFunction("getBlockHash", wasmedge.NewFunction(functype_i32_i32, wasmxGetBlockHash, context, 0))
	env.AddFunction("getAccount", wasmedge.NewFunction(functype_i32_i32, getAccount, context, 0))
	env.AddFunction("getBalance", wasmedge.NewFunction(functype_i32_i32, wasmxGetBalance, context, 0))
	env.AddFunction("externalCall", wasmedge.NewFunction(functype_i32_i32, externalCall, context, 0))
	env.AddFunction("keccak256", wasmedge.NewFunction(functype_i32_i32, keccak256Util, context, 0))

	env.AddFunction("createAccount", wasmedge.NewFunction(functype_i32_i32, wasmxCreateAccount, context, 0))
	env.AddFunction("create2Account", wasmedge.NewFunction(functype_i32_i32, wasmxCreate2Account, context, 0))

	env.AddFunction("grpcRequest", wasmedge.NewFunction(functype_i32_i32, wasmxGrpcRequest, context, 0))
	env.AddFunction("startInterval", wasmedge.NewFunction(functype_i32i64i32_i32, wasmxStartInterval, context, 0))
	env.AddFunction("stopInterval", wasmedge.NewFunction(functype_i32_, wasmxStopInterval, context, 0))

	return env
}

func getCtx(logger log.Logger, block bool) (*errgroup.Group, context.Context, context.CancelFunc) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	server.ListenForQuitSignals(g, block, cancelFn, logger)
	return g, ctx, cancelFn
}

// func startTimedAction(ctx *Context, callframe *wasmedge.CallingFrame, goctx context.Context, logger log.Logger, intervalId int32, timeDelay int64, repeatCount int32, argsbz []byte) (chan struct{}, error) {
// 	httpSrvDone := make(chan struct{}, 1)
// 	errCh := make(chan error)
// 	currentCount := int32(0)

// 	go func() {
// 		logger.Info("Starting new timed action", "intervalId", intervalId, "delay", timeDelay, "args", string(argsbz))
// 		// if err := httpSrv.Serve(ln); err != nil {
// 		// 	logger.Error("failed to serve Websrv", "error", err.Error())
// 		// 	if err == http.ErrServerClosed {
// 		// 		close(httpSrvDone)
// 		// 		return
// 		// 	}

// 		// 	logger.Error("failed to start Websrv server", "error", err.Error())
// 		// 	errCh <- err
// 		// }
// 		for {
// 			if repeatCount > 0 && currentCount == repeatCount {
// 				return
// 			}
// 			select {
// 			case <-goctx.Done():
// 				// The calling process canceled or closed the provided context
// 				return
// 			default:
// 				argsptr, err := allocateWriteMem(ctx, callframe, argsbz)
// 				if err != nil {
// 					errCh <- err
// 				}
// 				fmt.Println("argsptr", argsptr)
// 				err = executeTimedAction(ctx, intervalId, argsptr)
// 				if err != nil {
// 					errCh <- err
// 				}
// 			}
// 			currentCount += 1
// 			time.Sleep(time.Duration(timeDelay) * time.Millisecond)
// 		}
// 	}()

// 	select {
// 	case <-goctx.Done():
// 		// The calling process canceled or closed the provided context
// 		logger.Info("stopping timed action", "args", string(argsbz))
// 		return httpSrvDone, nil
// 	case err := <-errCh:
// 		logger.Error("failed to start timed action", "error", err.Error())
// 		return nil, err
// 	}
// }
