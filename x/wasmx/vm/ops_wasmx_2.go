package vm

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"

	networktypes "mythos/v1/x/network/types"
)

func sha256(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	hashbz := tmhash.Sum(data)
	ptr, err := allocateWriteMem(ctx, callframe, hashbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

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
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

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

func wasmxCall(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.SimpleCallRequestRaw
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

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
		contractContext := GetContractContext(ctx, req.To)
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			gasLimit := req.GasLimit
			if gasLimit == nil {
				// TODO: gas remaining!!
			}
			req := vmtypes.CallRequest{
				To:       req.To,
				From:     ctx.Env.Contract.Address,
				Value:    req.Value,
				GasLimit: gasLimit,
				Calldata: req.Calldata,
				Bytecode: contractContext.ContractInfo.Bytecode,
				CodeHash: contractContext.ContractInfo.CodeHash,
				IsQuery:  req.IsQuery,
			}
			success, returnData = WasmxCall(ctx, req)
			ctx.ReturnData = returnData
		}
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
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
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
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

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
	IpAddress string `json:"ip_address"`
	Contract  []byte `json:"contract"`
	Data      []byte `json:"data"` // should be []byte (base64 encoded)
}

func wasmxGrpcRequest(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	databz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--wasmxGrpcRequest-databz-", string(databz))
	var data GrpcRequest
	err = json.Unmarshal(databz, &data)
	fmt.Println("--wasmxGrpcRequest-err-", err)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("--wasmxGrpcRequest-data-", data)
	contractAddress := sdk.AccAddress(vmtypes.CleanupAddress(data.Contract))
	msg := &networktypes.MsgGrpcSendRequest{
		IpAddress: data.IpAddress,
		Contract:  contractAddress.String(),
		Data:      []byte(data.Data),
		Sender:    ctx.Env.Contract.Address.String(),
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msg)
	fmt.Println("--wasmxGrpcRequest-res, err-", res, err)
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

	msgtosend := &networktypes.MsgStartIntervalRequest{
		Sender:   ctx.Env.Contract.Address.String(),
		Contract: ctx.Env.Contract.Address.String(),
		Delay:    timeDelay,
		Repeat:   repeatCount,
		Args:     argsbz,
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
	returns[0] = intervalId
	return returns, wasmedge.Result_Success
}

// stopInterval(intervalId: i32): void
func wasmxStopInterval(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	// intervalId := params[0].(int32)
	// fmt.Println("* stop interval: ", intervalId)
	// // TODO errors if already stopped?
	// // ctx.intervals[intervalId].Quit <- true
	// ctx.intervals[intervalId].Cancel()
	return returns, wasmedge.Result_Success
}

func BuildWasmxEnv2(context *Context) *wasmedge.Module {
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

	env.AddFunction("sha256", wasmedge.NewFunction(functype_i32_i32, sha256, context, 0))

	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i32, getCallData, context, 0))
	env.AddFunction("getEnv", wasmedge.NewFunction(functype__i32, getEnv, context, 0))
	env.AddFunction("getCaller", wasmedge.NewFunction(functype__i32, wasmxGetCaller, context, 0))
	env.AddFunction("getAddress", wasmedge.NewFunction(functype__i32, wasmxGetAddress, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoad, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32_, wasmxStorageStore, context, 0))
	env.AddFunction("log", wasmedge.NewFunction(functype_i32_, wasmxLog, context, 0))
	env.AddFunction("finish", wasmedge.NewFunction(functype_i32_, wasmxFinish, context, 0))
	env.AddFunction("revert", wasmedge.NewFunction(functype_i32_, wasmxRevert, context, 0))
	env.AddFunction("getBlockHash", wasmedge.NewFunction(functype_i32_i32, wasmxGetBlockHash, context, 0))
	env.AddFunction("getAccount", wasmedge.NewFunction(functype_i32_i32, getAccount, context, 0))
	env.AddFunction("getBalance", wasmedge.NewFunction(functype_i32_i32, wasmxGetBalance, context, 0))
	// TODO move externalCall to only system API
	env.AddFunction("externalCall", wasmedge.NewFunction(functype_i32_i32, externalCall, context, 0))
	env.AddFunction("call", wasmedge.NewFunction(functype_i32_i32, wasmxCall, context, 0))
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
