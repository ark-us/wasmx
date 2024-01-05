package vm

import "github.com/second-state/WasmEdge-go/wasmedge"

func BuildWasmxConsensusJson1Mock(context *Context) *wasmedge.Module {
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

	env.AddFunction("CheckTx", wasmedge.NewFunction(functype_i32_i32, MockCheckTx, context, 0))
	env.AddFunction("PrepareProposal", wasmedge.NewFunction(functype_i32_i32, MockPrepareProposal, context, 0))
	env.AddFunction("ProcessProposal", wasmedge.NewFunction(functype_i32_i32, MockProcessProposal, context, 0))
	env.AddFunction("FinalizeBlock", wasmedge.NewFunction(functype_i32_i32, MockFinalizeBlock, context, 0))
	env.AddFunction("Commit", wasmedge.NewFunction(functype__i32, MockCommit, context, 0))
	env.AddFunction("RollbackToVersion", wasmedge.NewFunction(functype_i64_i32, MockRollbackToVersion, context, 0))

	return env
}

func MockCheckTx(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	ptr, err := allocateWriteMem(ctx, callframe, make([]byte, 0))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}
func MockPrepareProposal(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	ptr, err := allocateWriteMem(ctx, callframe, make([]byte, 0))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}
func MockProcessProposal(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	ptr, err := allocateWriteMem(ctx, callframe, make([]byte, 0))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}
func MockFinalizeBlock(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	ptr, err := allocateWriteMem(ctx, callframe, make([]byte, 0))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}
func MockCommit(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	ptr, err := allocateWriteMem(ctx, callframe, make([]byte, 0))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}
func MockRollbackToVersion(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	ptr, err := allocateWriteMem(ctx, callframe, make([]byte, 0))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}
