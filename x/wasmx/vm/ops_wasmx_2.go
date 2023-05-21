package vm

import (
	"encoding/json"
	"math/big"

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
	ptr, err := allocateMem(ctx, int32(len(envbz)))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	err = writeMem(callframe, envbz, ptr)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func getExternalCode(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
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

	env.AddFunction("getExternalBalance", wasmedge.NewFunction(functype_i32_i32, getExternalBalance, context, 0))
	env.AddFunction("getExternalCodeSize", wasmedge.NewFunction(functype_i32_i32, getExternalCodeSize, context, 0))
	env.AddFunction("getExternalCodeHash", wasmedge.NewFunction(functype_i32_i32, getExternalCodeHash, context, 0))
	env.AddFunction("getExternalCode", wasmedge.NewFunction(functype_i32_i32, getExternalCode, context, 0))
	env.AddFunction("getBlockHash", wasmedge.NewFunction(functype_i32_i32, getBlockHash, context, 0))

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
		Contract: ContractInfoJson{
			Address:  NewCustomBytes(ctx.Env.Contract.Address.Bytes()),
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

func int64ToBytes32(v int64) CustomBytes {
	return NewCustomBytes(big.NewInt(v).FillBytes((make([]byte, 32))))
}

func uint64ToBytes32(v uint64) CustomBytes {
	return int64ToBytes32(int64(v))
}
