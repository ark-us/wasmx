package vm

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	merkle "github.com/cometbft/cometbft/crypto/merkle"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

type MerkleSlices struct {
	Slices [][]byte `json:"slices"`
}

func merkleHash(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var val MerkleSlices
	err = json.Unmarshal(data, &val)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	hashbz := merkle.HashFromByteSlices(val.Slices)
	ptr, err := allocateWriteMem(ctx, callframe, hashbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// PrepareProposal(*abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)
func PrepareProposal(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	var req abci.RequestPrepareProposal
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp, err := ctx.GetApplication().PrepareProposal(&req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "PrepareProposal")
		return nil, wasmedge.Result_Fail
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// ProcessProposal(*abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error)
func ProcessProposal(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req abci.RequestProcessProposal
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp, err := ctx.GetApplication().ProcessProposal(&req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ProcessProposal")
		return nil, wasmedge.Result_Fail
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// FinalizeBlock(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
func FinalizeBlock(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req abci.RequestFinalizeBlock
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp, err := ctx.GetApplication().FinalizeBlock(&req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "FinalizeBlock")
		return nil, wasmedge.Result_Fail
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// Commit() (*abci.ResponseCommit, error)
func Commit(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	resp, err := ctx.GetApplication().Commit()
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "Commit")
		return nil, wasmedge.Result_Fail
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func CheckTx(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req abci.RequestCheckTx
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp, err := ctx.GetApplication().CheckTx(&req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "CheckTx")
		return nil, wasmedge.Result_Fail
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

type LoggerLog struct {
	Msg   string   `json:"msg"`
	Parts []string `json:"parts"`
}

func getLoggerData(callframe *wasmedge.CallingFrame, params []interface{}) (string, []any, error) {
	message, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return "", nil, err
	}
	var data LoggerLog
	err = json.Unmarshal(message, &data)
	if err != nil {
		return "", nil, err
	}
	parts := make([]any, len(data.Parts))
	for i, part := range data.Parts {
		parts[i] = part
	}
	return data.Msg, parts, nil
}

func wasmxLoggerInfo(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GetContext().Logger().Info(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxLoggerError(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GetContext().Logger().Error(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxLoggerDebug(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GetContext().Logger().Debug(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildWasmxConsensusJson1(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("consensus")
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("PrepareProposal", wasmedge.NewFunction(functype_i32_i32, PrepareProposal, context, 0))
	env.AddFunction("ProcessProposal", wasmedge.NewFunction(functype_i32_i32, ProcessProposal, context, 0))
	env.AddFunction("FinalizeBlock", wasmedge.NewFunction(functype_i32_i32, FinalizeBlock, context, 0))
	env.AddFunction("Commit", wasmedge.NewFunction(functype__i32, Commit, context, 0))
	env.AddFunction("CheckTx", wasmedge.NewFunction(functype_i32_i32, CheckTx, context, 0))

	env.AddFunction("MerkleHash", wasmedge.NewFunction(functype_i32_i32, merkleHash, context, 0))

	env.AddFunction("LoggerInfo", wasmedge.NewFunction(functype_i32_, wasmxLoggerInfo, context, 0))
	env.AddFunction("LoggerError", wasmedge.NewFunction(functype_i32_, wasmxLoggerError, context, 0))
	env.AddFunction("LoggerDebug", wasmedge.NewFunction(functype_i32_, wasmxLoggerDebug, context, 0))

	return env
}
