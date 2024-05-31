package vmcrosschain

import (
	"encoding/json"

	"github.com/second-state/WasmEdge-go/wasmedge"

	vmtypes "mythos/v1/x/wasmx/vm"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
	"mythos/v1/x/network/types"
)

// executeCrossChainTx(*InitSubChainMsg) (*abci.ResponseInitChain, error)
func executeCrossChainTx(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req types.MsgExecuteCrossChainTxRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	response, err := ctx.CosmosHandler.(ctx, &req)
	if err != nil {
		ctx.Logger(ctx.Ctx).Error("could not initiate subchain app", "error", err.Error())
		return nil, wasmedge.Result_Fail
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func BuildWasmxCrosschainJson1(ctx_ *vmtypes.Context) *wasmedge.Module {
	context := &Context{Context: ctx_}
	env := wasmedge.NewModule(HOST_WASMX_ENV_CROSSCHAIN)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("executeCrossChainTx", wasmedge.NewFunction(functype_i32_i32, executeCrossChainTx, context, 0))
	return env
}
