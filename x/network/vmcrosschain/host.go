package vmcrosschain

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/second-state/WasmEdge-go/wasmedge"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mcfg "mythos/v1/config"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

// TODO!! this API should only be used by core contracts

// executeCrossChainTx(*MsgExecuteCrossChainCallRequest) (*abci.MsgExecuteCrossChainCallResponse, error)
func executeCrossChainTx(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	// we expect from & chain id to be set correctly by the multichain registry contract
	req.IsQuery = false
	req.Sender = ctx.Env.Contract.Address.String()
	req.FromChainId = ctx.Env.Chain.ChainIdFull

	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(&req)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
	} else {
		ctx.Ctx.EventManager().EmitEvents(evs)
	}
	resp := WrappedResponse{Error: errmsg}
	if errmsg == "" {
		var rres types.MsgExecuteCrossChainCallResponse
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

// executeCrossChainQuery(*QueryCrossChainRequest) (*abci.MsgExecuteCrossChainCallResponse, error)
func executeCrossChainQuery(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	// we expect sender & chain id to be set correctly by the multichain registry contract
	req.IsQuery = true
	req.Sender = ctx.Env.Contract.Address.String()
	req.FromChainId = ctx.Env.Chain.ChainIdFull

	_, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(&req)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
	}

	resp := WrappedResponse{Error: errmsg}
	if errmsg == "" {
		var rres types.MsgExecuteCrossChainCallResponse
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

// TODO API can only be used by core contracts
// like consensus, lobby, etc
// internal communication with private chains, like level0, or between consensusless
// contracts on different chains, which do not require determinism
// executeCrossChainTxNonDeterministic(*MsgExecuteCrossChainCallRequest) (*abci.MsgExecuteCrossChainCallResponse, error)
func executeCrossChainTxNonDeterministic(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	resp := WrappedResponse{}
	ctx := _context.(*Context)

	// we do not want to fail and end the transaction
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		resp.Error = "cannot read request" + err.Error()
		return returnResult(ctx, callframe, resp)
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		resp.Error = "cannot unmarshal request" + err.Error()
		return returnResult(ctx, callframe, resp)
	}

	// TODO check to & from are consensusless contracts
	// contractInfo := c.nk.GetContractInfo(ctx, contractAddress)

	// get subchainapp
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, callframe, resp)
	}
	iapp, err := multichainapp.GetApp(req.ToChainId)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, callframe, resp)
	}
	app, ok := iapp.(mcfg.MythosApp)
	if !ok {
		resp.Error = "error App interface from multichainapp"
		return returnResult(ctx, callframe, resp)
	}

	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)

		// we sent directly to the contract
		resp, err := app.GetNetworkKeeper().ExecuteContract(ctx, &types.MsgExecuteContract{
			Sender:   req.From,
			Contract: req.To,
			Msg:      req.Msg,
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// app.GetBaseApp().RunTx(sdk.ExecModeFinalize, txbytes)
	_, err = app.GetActionExecutor().Execute(context.Background(), app.GetBaseApp().LastBlockHeight(), cb)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, callframe, resp)
	}

	return returnResult(ctx, callframe, resp)
}

// !! this is non-deterministic !! it must not be used inside a transaction
// to make the query deterministic, we need to use ExecuteCrossChainTx
// to ensure the cross-chain queries are executed in the same order for all validators
func executeCrossChainQueryNonDeterministic(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	resp := WrappedResponse{}
	ctx := _context.(*Context)

	// TODO check that we are in a query environment
	execmode := ctx.Ctx.ExecMode()
	if execmode != sdk.ExecModeQuery {
		ctx.Logger(ctx.Ctx).Error("tried to execute non-deterministic cross-chain query as part of a transaction; reverting tx")
		return nil, wasmedge.Result_Fail
	}
	// otherwise revert

	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	// TODO set role
	// set this contract as sender
	// TODO URGENT interchain account addresses
	// req.From = ctx.Env.Contract.Address.String()
	req.IsQuery = true
	req.Sender = ctx.Env.Contract.Address.String()
	req.FromChainId = ctx.Env.Chain.ChainIdFull

	contractCall := wasmxtypes.QuerySmartContractCallRequest{
		Sender:       req.From,
		Address:      req.To, // roles not supported now
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

// isAtomicTxInExecution(*MsgIsAtomicTxInExecutionRequest) (*abci.MsgIsAtomicTxInExecutionResponse, error)
func isAtomicTxInExecution(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req MsgIsAtomicTxInExecutionRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp := &MsgIsAtomicTxInExecutionResponse{IsInExecution: false}

	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	_, err = multichainapp.GetApp(req.SubChainId)
	if err != nil {
		return returnIsInExecutionResult(ctx, callframe, resp)
	}

	// check if chain has channel opened
	mcctx, err := types.GetMultiChainContext(ctx.GoContextParent)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	if bytes.Equal(mcctx.CurrentAtomicTxHash, req.TxHash) {
		resp.IsInExecution = true
	}
	return returnIsInExecutionResult(ctx, callframe, resp)
}

func returnIsInExecutionResult(ctx *Context, callframe *wasmedge.CallingFrame, resp *MsgIsAtomicTxInExecutionResponse) ([]interface{}, wasmedge.Result) {
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
	env.AddFunction("executeCrossChainQueryNonDeterministic", wasmedge.NewFunction(functype_i32_i32, executeCrossChainQueryNonDeterministic, context, 0))
	env.AddFunction("executeCrossChainTxNonDeterministic", wasmedge.NewFunction(functype_i32_i32, executeCrossChainTxNonDeterministic, context, 0))
	env.AddFunction("isAtomicTxInExecution", wasmedge.NewFunction(functype_i32_i32, isAtomicTxInExecution, context, 0))
	return env
}
