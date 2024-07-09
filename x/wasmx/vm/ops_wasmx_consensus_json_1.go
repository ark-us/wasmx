package vm

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	mctx "mythos/v1/context"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

type ResponseOptimisticExecution struct {
	MetaInfo map[string][]byte `json:"metainfo"`
}

type WrapRequestFinalizeBlock struct {
	Request  abci.RequestFinalizeBlock `json:"request"`
	MetaInfo map[string][]byte         `json:"metainfo"`
}

type WrapResult struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

// PrepareProposal(*abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)
func PrepareProposal(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := asmem.ReadMemFromPtr(callframe, params[0])
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
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
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
	reqbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req abci.RequestProcessProposal
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ProcessProposal")
		return nil, wasmedge.Result_Fail
	}
	bapp := ctx.GetApplication()
	resp, err := bapp.ProcessProposal(&req)

	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ProcessProposal")
		return nil, wasmedge.Result_Fail
	}

	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func OptimisticExecution(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resbz, err := asmem.ReadMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req abci.RequestProcessProposal
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, wasmedge.Result_Fail
	}
	var res abci.ResponseProcessProposal
	err = json.Unmarshal(resbz, &res)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, wasmedge.Result_Fail
	}
	bapp := ctx.GetApplication()
	oe := bapp.GetOptimisticExecution()
	oe.Enable()

	// reset meta info from previous optimistic executions
	mctx.ResetExecutionMetaInfo(ctx.GoContextParent)

	bapp.OptimisticExecution(&req, &res)
	oe.Disable()

	// TODO we should return the error, not throw
	_, err = oe.WaitResult()
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, wasmedge.Result_Fail
	}
	metainfo, err := mctx.GetExecutionMetaInfoEncoded(ctx.GoContextParent, ctx.GetCosmosHandler().Codec())
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, wasmedge.Result_Fail
	}

	respbz, err := json.Marshal(&ResponseOptimisticExecution{MetaInfo: metainfo})
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
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
	reqbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req WrapRequestFinalizeBlock
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	// set metainfo on the parent context, so it is available during execution
	err = mctx.SetExecutionMetaInfo(ctx.GoContextParent, ctx.CosmosHandler.Codec(), req.MetaInfo)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "FinalizeBlock")
		return nil, wasmedge.Result_Fail
	}

	bapp := ctx.GetApplication()
	resp, err := bapp.FinalizeBlockSimple(&req.Request)
	errmsg := ""
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "FinalizeBlock")
		errmsg = err.Error()
	}
	oe := bapp.GetOptimisticExecution()
	if oe.Initialized() {
		oe.Reset()
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	respwrap := &FinalizeBlockWrap{
		Error: errmsg,
		Data:  respbz,
	}
	respwrapbz, err := json.Marshal(respwrap)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respwrapbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// BeginBlock(*abci.RequestFinalizeBlock) (sdk.BeginBlock, error)
func BeginBlock(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req abci.RequestFinalizeBlock
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp, err := ctx.GetApplication().BeginBlock(&req)
	errmsg := ""
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "BeginBlock")
		errmsg = err.Error()
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	respwrap := &FinalizeBlockWrap{
		Error: errmsg,
		Data:  respbz,
	}
	respwrapbz, err := json.Marshal(respwrap)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respwrapbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// EndBlock(metadata string) (*abci.ResponseFinalizeBlock, error)
func EndBlock(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	metadata, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp, err := ctx.GetApplication().EndBlock(metadata)
	errmsg := ""
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "FinalizeBlock")
		errmsg = err.Error()
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	respwrap := &FinalizeBlockWrap{
		Error: errmsg,
		Data:  respbz,
	}
	respwrapbz, err := json.Marshal(respwrap)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respwrapbz)
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
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func RollbackToVersion(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	err := ctx.GetApplication().CommitMultiStore().RollbackToVersion(params[0].(int64))
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, []byte(errMsg))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func CheckTx(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := asmem.ReadMemFromPtr(callframe, params[0])
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
	ptr, err := asmem.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func BuildWasmxConsensusJson1(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("consensus")
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i64_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("CheckTx", wasmedge.NewFunction(functype_i32_i32, CheckTx, context, 0))
	env.AddFunction("PrepareProposal", wasmedge.NewFunction(functype_i32_i32, PrepareProposal, context, 0))
	env.AddFunction("ProcessProposal", wasmedge.NewFunction(functype_i32_i32, ProcessProposal, context, 0))
	env.AddFunction("OptimisticExecution", wasmedge.NewFunction(functype_i32i32_i32, OptimisticExecution, context, 0))
	env.AddFunction("FinalizeBlock", wasmedge.NewFunction(functype_i32_i32, FinalizeBlock, context, 0))
	env.AddFunction("BeginBlock", wasmedge.NewFunction(functype_i32_i32, BeginBlock, context, 0))
	env.AddFunction("EndBlock", wasmedge.NewFunction(functype_i32_i32, EndBlock, context, 0))
	env.AddFunction("Commit", wasmedge.NewFunction(functype__i32, Commit, context, 0))
	env.AddFunction("RollbackToVersion", wasmedge.NewFunction(functype_i64_i32, RollbackToVersion, context, 0))

	return env
}
