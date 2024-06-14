package vmcrosschain

import (
	"encoding/json"

	"github.com/second-state/WasmEdge-go/wasmedge"

	abci "github.com/cometbft/cometbft/abci/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

// executeCrossChainTx(*MsgExecuteCrossChainTxRequest) (*abci.MsgExecuteCrossChainTxResponse, error)
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

	// TODO set role
	// set sender as this contract address!
	// TODO URGENT interchain account addresses
	// req.From = ctx.Env.Contract.Address.String()
	req.FromChainId = ctx.Env.Chain.ChainIdFull

	// contractAddress, err := ctx.CosmosHandler.GetRoleByContractAddress(ctx.Ctx, types.AccAddress(req.From))
	// if err != nil {
	// 	return nil, wasmedge.Result_Fail
	// }

	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(&req)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
	} else {
		ctx.Ctx.EventManager().EmitEvents(evs)
	}
	resp := WrappedResponse{Error: errmsg}
	if errmsg == "" {
		var rres types.MsgExecuteCrossChainTxResponse
		if res != nil {
			err = rres.Unmarshal(res)
			if err != nil {
				return nil, wasmedge.Result_Fail
			}
		}
		resp = WrappedResponse{
			Data:  rres.Data,
			Error: rres.Error,
		}
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// !! this is non-deterministic !! it must not be used inside a transaction
// to make the query deterministic, we need to use ExecuteCrossChainTx
// to ensure the cross-chain queries are executed in the same order for all validators
func executeCrossChainQuery(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	resp := WrappedResponse{}
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req types.QueryCrossChainRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	// TODO set role
	// set this contract as sender
	// TODO URGENT interchain account addresses
	// req.From = ctx.Env.Contract.Address.String()
	req.FromChainId = ctx.Env.Chain.ChainIdFull

	contractCall := wasmxtypes.QuerySmartContractCallRequest{
		Sender:       req.From,
		Address:      req.ToAddressOrRole, // roles not supported now
		QueryData:    req.Msg,
		Funds:        req.Funds,
		Dependencies: req.Dependencies,
	}
	contractCallBz, err := contractCall.Marshal()
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	contractQuery := &abci.RequestQuery{
		Path: "/mythos.wasmx.v1.Query/SmartContractCall",
		Data: contractCallBz,
	}
	contractQueryBz, err := contractQuery.Marshal()
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	mcQuery := types.QueryMultiChainRequest{
		MultiChainId: req.ToChainId,
		QueryData:    contractQueryBz,
	}
	mcQueryBz, err := mcQuery.Marshal()
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	qreq := &abci.RequestQuery{
		Path: "/mythos.network.v1.Query/QueryMultiChain",
		Data: mcQueryBz,
	}
	res, err := ctx.CosmosHandler.SubmitCosmosQuery(qreq)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, callframe, resp)
	}

	// var rres types.QueryCrossChainResponse
	var rres types.QueryMultiChainResponse
	err = rres.Unmarshal(res)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, callframe, resp)
	}

	var wres wasmxtypes.QuerySmartContractCallResponse
	err = wres.Unmarshal(rres.Data)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, callframe, resp)
	}

	innerq := &wasmxtypes.WasmxQueryResponse{}
	err = json.Unmarshal(wres.Data, innerq)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, callframe, resp)
	}

	resp.Data = innerq.Data
	return returnResult(ctx, callframe, resp)
}

func returnResult(ctx *Context, callframe *wasmedge.CallingFrame, resp WrappedResponse) ([]interface{}, wasmedge.Result) {
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, respbz)
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
	env.AddFunction("executeCrossChainQuery", wasmedge.NewFunction(functype_i32_i32, executeCrossChainQuery, context, 0))
	return env
}
