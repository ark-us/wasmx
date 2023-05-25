package vm

import (
	"encoding/json"
	"math/big"
	"mythos/v1/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/second-state/WasmEdge-go/wasmedge"
)

// getEnv(): ArrayBuffer
func getEnv(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	env := buildEnv(ctx)
	envbz, err := json.Marshal(env)
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
func getAccount(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addr, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(cleanupAddress(addr))
	codeInfo := ctx.CosmosHandler.GetCodeInfo(address)
	balance := ctx.CosmosHandler.GetBalance(address)
	code := AccountInfoJson{
		Address:  NewCustomBytes(address),
		Balance:  NewCustomBytes(balance.FillBytes(make([]byte, 32))),
		CodeHash: NewCustomBytes(codeInfo.CodeHash),
		Bytecode: NewCustomBytes(codeInfo.InterpretedBytecodeRuntime),
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

// account -> void
func setAccount(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("setAccount not implemented")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func keccak256Util(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
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
func externalCall(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	requestbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var request CallRequestJson
	json.Unmarshal(requestbz, &request)

	req := request.Transform()
	returns := make([]interface{}, 1)

	success, returnData := wasmxCall(ctx, req)

	response := CallResponseJson{
		Success: success,
		Data:    NewCustomBytes(returnData),
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

	env.AddFunction("getEnv", wasmedge.NewFunction(functype__i32, getEnv, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoad, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32_, wasmxStorageStore, context, 0))
	env.AddFunction("log", wasmedge.NewFunction(functype_i32_, wasmxLog, context, 0))
	env.AddFunction("finish", wasmedge.NewFunction(functype_i32_, wasmxFinish, context, 0))
	env.AddFunction("revert", wasmedge.NewFunction(functype_i32_, wasmxRevert, context, 0))

	env.AddFunction("getBlockHash", wasmedge.NewFunction(functype_i32_i32, getBlockHash, context, 0))

	env.AddFunction("getAccount", wasmedge.NewFunction(functype_i32_i32, getAccount, context, 0))
	env.AddFunction("setAccount", wasmedge.NewFunction(functype_i32_, setAccount, context, 0))

	env.AddFunction("externalCall", wasmedge.NewFunction(functype_i32_i32, externalCall, context, 0))

	env.AddFunction("keccak256", wasmedge.NewFunction(functype_i32_i32, keccak256Util, context, 0))

	return env
}

func buildEnv(ctx *Context) *EnvJson {
	return &EnvJson{
		Chain: ChainInfoJson{
			Denom:       ctx.Env.Chain.Denom,
			ChainId:     NewCustomBytes(ctx.Env.Chain.ChainId.FillBytes((make([]byte, 32)))),
			ChainIdFull: ctx.Ctx.ChainID(),
		},
		Block: BlockInfoJson{
			Height:   int64ToBytes32(ctx.Ctx.BlockHeight()),
			Time:     int64ToBytes32(ctx.Ctx.BlockTime().Unix()),
			GasLimit: uint64ToBytes32(20000000), // TODO
			// GasLimit: uint64ToBytes32(ctx.Ctx.BlockGasMeter().Limit()),
			Hash:     NewCustomBytes(ctx.Ctx.HeaderHash()),
			Proposer: NewCustomBytes(ctx.Env.Block.Proposer),
		},
		Transaction: TransactionInfoJson{
			Index:    int32(ctx.Env.Transaction.Index),
			GasPrice: int64ToBytes32(0), // TODO
		},
		Contract: AccountInfoJson{
			Address:  NewCustomBytes(ctx.Env.Contract.Address.Bytes()),
			CodeHash: NewCustomBytes(ctx.Env.Contract.CodeHash),
			Bytecode: NewCustomBytes(ctx.ExecutionBytecode),
			Balance:  int64ToBytes32(0), // TODO
		},
		CurrentCall: CurrentCallInfoJson{
			Origin:   NewCustomBytes(ctx.CallContext.Origin.Bytes()),
			Sender:   NewCustomBytes(ctx.CallContext.Sender.Bytes()),
			Funds:    NewCustomBytes(ctx.CallContext.Funds.FillBytes((make([]byte, 32)))),
			GasLimit: int64ToBytes32(2000000), // TODO
			CallData: NewCustomBytes(ctx.Calldata),
		},
	}
}

func wasmxCall(ctx *Context, req CallRequest) (int32, []byte) {
	_, ok := ctx.ContractRouter[req.To.String()]
	if !ok {
		dep, err := ctx.CosmosHandler.GetContractDependency(ctx.Ctx, req.To)
		if err != nil {
			return int32(1), nil
		}
		depContext, err := buildExecutionContextClassic(dep.FilePath, *ctx.Env, dep.StoreKey, nil, dep.SystemDeps)
		if err != nil {
			return int32(1), nil
		}
		ctx.ContractRouter[req.To.String()] = depContext
	}

	callContext := types.MessageInfo{
		Origin: ctx.CallContext.Origin,
		Sender: req.From,
		Funds:  req.Value,
	}

	tempCtx, commit := ctx.Ctx.CacheContext()
	contractStore := ctx.CosmosHandler.ContractStore(tempCtx, ctx.ContractRouter[req.To.String()].ContractStoreKey)

	newctx := &Context{
		Ctx:            tempCtx,
		GasMeter:       ctx.GasMeter,
		Callvalue:      req.Value,
		Calldata:       req.Calldata,
		ContractStore:  contractStore,
		CosmosHandler:  ctx.CosmosHandler,
		ContractRouter: ctx.ContractRouter,
		CallContext:    callContext,
		Env: &types.Env{
			Block:       ctx.Env.Block,
			Transaction: ctx.Env.Transaction,
			Chain:       ctx.Env.Chain,
			Contract: types.EnvContractInfo{
				Address:  req.To,
				CodeHash: req.CodeHash,
			},
		},
		ExecutionBytecode: req.Bytecode,
	}

	_, err := ctx.ContractRouter[req.To.String()].Execute(newctx)
	var success int32
	// Returns 0 on success, 1 on failure and 2 on revert
	if err != nil {
		success = int32(2)
	} else {
		success = int32(0)
		commit()
		// Write events
		ctx.Ctx.EventManager().EmitEvents(tempCtx.EventManager().Events())
		ctx.Logs = append(ctx.Logs, newctx.Logs...)
	}
	ctx.ReturnData = newctx.ReturnData
	return success, newctx.ReturnData
}

func int64ToBytes32(v int64) CustomBytes {
	return NewCustomBytes(big.NewInt(v).FillBytes((make([]byte, 32))))
}

func uint64ToBytes32(v uint64) CustomBytes {
	return int64ToBytes32(int64(v))
}
