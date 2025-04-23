package vmcrosschain

import (
	"github.com/loredanacirstea/wasmx/x/network/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func executeCrossChainTxMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp := &types.WrappedResponse{}
	return returnResult(ctx, rnh, resp)
}

func executeCrossChainQueryMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp := &types.WrappedResponse{}
	return returnResult(ctx, rnh, resp)
}

func executeCrossChainTxNonDeterministicMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	resp := &types.WrappedResponse{}
	ctx := _context.(*Context)
	return returnResult(ctx, rnh, resp)
}

func executeCrossChainQueryNonDeterministicMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	resp := &types.WrappedResponse{}
	ctx := _context.(*Context)
	return returnResult(ctx, rnh, resp)
}

func isAtomicTxInExecutionMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp := &MsgIsAtomicTxInExecutionResponse{IsInExecution: false}
	return returnIsInExecutionResult(ctx, rnh, resp)
}

func BuildWasmxCrosschainJson1Mock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("executeCrossChainTx", executeCrossChainTxMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("executeCrossChainQuery", executeCrossChainQueryMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("executeCrossChainQueryNonDeterministic", executeCrossChainQueryNonDeterministicMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("executeCrossChainTxNonDeterministic", executeCrossChainTxNonDeterministicMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("isAtomicTxInExecution", isAtomicTxInExecutionMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, HOST_WASMX_ENV_CROSSCHAIN, context, fndefs)
}
