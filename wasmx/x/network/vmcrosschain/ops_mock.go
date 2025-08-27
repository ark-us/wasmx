package vmcrosschain

import (
	"github.com/loredanacirstea/wasmx/x/network/types"
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
