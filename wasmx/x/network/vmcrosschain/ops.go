package vmcrosschain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mcfg "github.com/loredanacirstea/wasmx/config"
	"github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// executeCrossChainTx(*MsgExecuteCrossChainCallRequest) (*abci.MsgExecuteCrossChainCallResponse, error)
func executeCrossChainTx(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	resp := &types.WrappedResponse{}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		resp.Error = "cannot read request" + err.Error()
		return returnResult(ctx, rnh, resp)
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return returnResult(ctx, rnh, resp)
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
	resp.Error = errmsg
	if errmsg == "" {
		var rres types.MsgExecuteCrossChainCallResponse
		if res != nil {
			err = rres.Unmarshal(res)
			if err != nil {
				return nil, err
			}
		}
		resp.Data = rres.Data
		resp.Error = rres.Error
	}
	return returnResultAndAddCrossChainInfo(ctx, rnh, req, resp)
}

// executeCrossChainQuery(*QueryCrossChainRequest) (*abci.MsgExecuteCrossChainCallResponse, error)
func executeCrossChainQuery(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	resp := &types.WrappedResponse{}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		resp.Error = "cannot read request" + err.Error()
		return returnResult(ctx, rnh, resp)
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		resp.Error = "cannot read request" + err.Error()
		return returnResult(ctx, rnh, resp)
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

	resp.Error = errmsg
	if errmsg == "" {
		var rres types.MsgExecuteCrossChainCallResponse
		if res != nil {
			err = rres.Unmarshal(res)
			if err != nil {
				return nil, err
			}
		}
		resp.Data = rres.Data
		resp.Error = rres.Error
	}

	return returnResultAndAddCrossChainInfo(ctx, rnh, req, resp)
}

// like consensus, lobby, etc
// internal communication with private chains, like level0, or between consensusless / consensusmeta
// contracts on different chains, which do not require determinism
// executeCrossChainTxNonDeterministic(*MsgExecuteCrossChainCallRequest) (*abci.MsgExecuteCrossChainCallResponse, error)
func executeCrossChainTxNonDeterministic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	resp := &types.WrappedResponse{}
	ctx := _context.(*Context)

	// we do not want to fail and end the transaction
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		resp.Error = "cannot read request" + err.Error()
		return returnResult(ctx, rnh, resp)
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		resp.Error = "cannot unmarshal request" + err.Error()
		return returnResult(ctx, rnh, resp)
	}

	// TODO check to & from are consensusless / consensusmeta? contracts
	// contractInfo := c.nk.GetContractInfo(ctx, contractAddress)

	// get subchainapp
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, rnh, resp)
	}
	iapp, err := multichainapp.GetApp(req.ToChainId)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, rnh, resp)
	}
	app, ok := iapp.(mcfg.MythosApp)
	if !ok {
		resp.Error = "error App interface from multichainapp"
		return returnResult(ctx, rnh, resp)
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

	_, err = app.GetActionExecutor().Execute(context.Background(), app.GetBaseApp().LastBlockHeight(), sdk.ExecModeFinalize, cb)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, rnh, resp)
	}

	return returnResult(ctx, rnh, resp)
}

// !! this is non-deterministic !! it must not be used inside a transaction
// to make the query deterministic, we need to use ExecuteCrossChainTx
// to ensure the cross-chain queries are executed in the same order for all validators
func executeCrossChainQueryNonDeterministic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	resp := &types.WrappedResponse{}
	ctx := _context.(*Context)

	// TODO check that we are in a query environment
	execmode := ctx.Ctx.ExecMode()
	if execmode != sdk.ExecModeQuery {
		errmsg := "tried to execute non-deterministic cross-chain query as part of a transaction; reverting tx"
		ctx.Logger(ctx.Ctx).Error(errmsg)
		return nil, fmt.Errorf(errmsg)
	}
	// otherwise revert
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req types.MsgExecuteCrossChainCallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	contractQuery := &abci.RequestQuery{
		Path: "/mythos.wasmx.v1.Query/SmartContractCall",
		Data: contractCallBz,
	}
	contractQueryBz, err := contractQuery.Marshal()
	if err != nil {
		return nil, err
	}
	mcQuery := types.QueryMultiChainRequest{
		MultiChainId: req.ToChainId,
		QueryData:    contractQueryBz,
	}
	mcQueryBz, err := mcQuery.Marshal()
	if err != nil {
		return nil, err
	}

	qreq := &abci.RequestQuery{
		Path: "/mythos.network.v1.Query/QueryMultiChain",
		Data: mcQueryBz,
	}
	res, err := ctx.CosmosHandler.SubmitCosmosQuery(qreq)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, rnh, resp)
	}

	// var rres types.QueryCrossChainResponse
	var rres types.QueryMultiChainResponse
	err = rres.Unmarshal(res)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, rnh, resp)
	}

	var wres wasmxtypes.QuerySmartContractCallResponse
	err = wres.Unmarshal(rres.Data)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, rnh, resp)
	}

	innerq := &wasmxtypes.WasmxQueryResponse{}
	err = json.Unmarshal(wres.Data, innerq)
	if err != nil {
		resp.Error = err.Error()
		return returnResult(ctx, rnh, resp)
	}

	resp.Data = innerq.Data
	return returnResult(ctx, rnh, resp)
}

func returnResultAndAddCrossChainInfo(ctx *Context, rnh memc.RuntimeHandler, req types.MsgExecuteCrossChainCallRequest, resp *types.WrappedResponse) ([]interface{}, error) {
	types.AddCrossChainCallMetaInfo(ctx.GoContextParent, ctx.Ctx.ChainID(), req, *resp)
	return returnResult(ctx, rnh, resp)
}

func returnResult(ctx *Context, rnh memc.RuntimeHandler, resp *types.WrappedResponse) ([]interface{}, error) {
	respbz, err := ctx.GetCosmosHandler().Codec().MarshalJSON(resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

// isAtomicTxInExecution(*MsgIsAtomicTxInExecutionRequest) (*abci.MsgIsAtomicTxInExecutionResponse, error)
func isAtomicTxInExecution(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req MsgIsAtomicTxInExecutionRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	resp := &MsgIsAtomicTxInExecutionResponse{IsInExecution: false}

	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return nil, err
	}
	_, err = multichainapp.GetApp(req.SubChainId)
	if err != nil {
		return returnIsInExecutionResult(ctx, rnh, resp)
	}

	// check if chain has channel opened
	mcctx, err := types.GetMultiChainContext(ctx.GoContextParent)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(mcctx.CurrentAtomicTxHash, req.TxHash) {
		resp.IsInExecution = true
	}
	return returnIsInExecutionResult(ctx, rnh, resp)
}

func returnIsInExecutionResult(ctx *Context, rnh memc.RuntimeHandler, resp *MsgIsAtomicTxInExecutionResponse) ([]interface{}, error) {
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}
