package keeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	cfg "mythos/v1/config"
	"mythos/v1/x/network/types"
	"mythos/v1/x/network/vmcrosschain"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// Any execution message can be wrapped with MsgMultiChainWrap to be executed on one
// of the available chains.
// BroadcastTxAsync peeks inside the transaction and inside MsgMultiChainWrap to get the chainId
// and then forwards the transaction to the apropriate chain application
// the signature & signer are verified in the AnteHandler of that chain application
func (k *Keeper) MultiChainWrapInternal(ctx sdk.Context, msg *types.MsgMultiChainWrap) (*types.MsgMultiChainWrapResponse, error) {
	var sdkmsg sdk.Msg
	err := k.cdc.UnpackAny(msg.Data, &sdkmsg)
	if err != nil {
		return nil, err
	}

	multichainapp, err := cfg.GetMultiChainApp(k.goContextParent)
	if err != nil {
		return nil, err
	}
	iapp, err := multichainapp.GetApp(msg.MultiChainId)
	if err != nil {
		return nil, err
	}
	app, ok := iapp.(cfg.MythosApp)
	if !ok {
		return nil, fmt.Errorf("error App interface from multichainapp")
	}

	owner, err := k.wasmxKeeper.AccBech32Codec().StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, err
	}

	// TODO route message &check owner is same as msg sender property ??

	// TODO handle transaction verification!!!! here or by codec ??
	// router := mcodec.MsgRouter{Router: app.MsgServiceRouter()}
	evs, res, err := app.GetNetworkKeeper().ExecuteCosmosMsg(ctx, sdkmsg, owner)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(evs)

	return &types.MsgMultiChainWrapResponse{
		Data: res,
	}, nil
}

// an atomic batch of multichain transactions
// each transaction may affect only one chain or > 1, if they contain MsgExecuteCrossChainCallRequest messages
// each transaction may contain multiple MsgExecuteCrossChainCallRequest messages, with multiple internal cross-chain transactions
// TODO: ExecuteAtomicTx must not be nested inside other ExecuteAtomicTx
func (k *Keeper) ExecuteAtomicTx(goCtx context.Context, msg *types.MsgExecuteAtomicTxRequest) (*types.MsgExecuteAtomicTxResponse, error) {
	// our atomic tx result channel - we use this to send our result
	var newResultsChannel chan types.MsgExecuteAtomicTxResponse
	// our cross-chain tx request channel - we receive contract calls from other chains
	var newInternalCallChannel chan types.MsgExecuteCrossChainCallRequestIndexed
	// our cross-chain tx response channel - we send the results of cross-chain calls to other chains
	var newInternalCallResponseChannel chan types.MsgExecuteCrossChainCallResponseIndexed

	ctx := sdk.UnwrapSDKContext(goCtx)
	txbz := ctx.TxBytes()
	txhash := tmhash.Sum(txbz)

	k.Logger(ctx).Info("executing atomic tx", "txhash", hex.EncodeToString(txhash), "tx_count", len(msg.Txs))

	mcctx, err := types.GetMultiChainContext(k.goContextParent)
	if err != nil {
		return nil, err
	}
	if len(mcctx.CurrentAtomicTxHash) == 0 {
		mcctx.CurrentAtomicTxHash = txhash
	} else {
		if !bytes.Equal(mcctx.CurrentAtomicTxHash, txhash) {
			return nil, fmt.Errorf("atomic tx hash mismatch: expected %s, actual %s", hex.EncodeToString(mcctx.CurrentAtomicTxHash), hex.EncodeToString(txhash))
		}
	}

	expectedChainIds := mcctx.CurrentAtomicTxChainIds
	sdktx, err := k.actionExecutor.GetApp().GetTxConfig().TxDecoder()(txbz)
	if err != nil {
		return nil, err
	}
	txWithExtensions, ok := sdktx.(authante.HasExtensionOptionsTx)
	if ok {
		opts := txWithExtensions.GetExtensionOptions()
		if len(opts) > 0 && opts[0].GetTypeUrl() == types.TypeURL_ExtensionOptionAtomicMultiChainTx {
			ext := opts[0].GetCachedValue().(*types.ExtensionOptionAtomicMultiChainTx)
			expectedChainIds = ext.ChainIds
		}
	}

	if len(mcctx.CurrentAtomicTxChainIds) == 0 {
		mcctx.CurrentAtomicTxChainIds = expectedChainIds
	} else {
		if !slices.Equal(mcctx.CurrentAtomicTxChainIds, expectedChainIds) {
			return nil, fmt.Errorf("atomic chainIds mismatch: expected %s, actual %s", mcctx.CurrentAtomicTxChainIds, expectedChainIds)
		}
	}

	multichainapp, err := cfg.GetMultiChainApp(k.goContextParent)
	if err != nil {
		return nil, err
	}

	existent, err := mcctx.GetResultChannel(ctx.ChainID())
	if err == nil {
		newResultsChannel = *existent
	} else {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newResultsChannel = make(chan types.MsgExecuteAtomicTxResponse, 1)
		mcctx.SetResultChannel(ctx.ChainID(), &newResultsChannel)
	}

	existent2, err := mcctx.GetInternalCallChannel(ctx.ChainID())
	if err == nil {
		newInternalCallChannel = *existent2
	} else {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newInternalCallChannel = make(chan types.MsgExecuteCrossChainCallRequestIndexed, 1)
		mcctx.SetInternalCallChannel(ctx.ChainID(), &newInternalCallChannel)
	}

	existent3, err := mcctx.GetInternalCallResultChannel(ctx.ChainID())
	if err == nil {
		newInternalCallResponseChannel = *existent3
	} else {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newInternalCallResponseChannel = make(chan types.MsgExecuteCrossChainCallResponseIndexed, 1)
		mcctx.SetInternalCallResultChannel(ctx.ChainID(), &newInternalCallResponseChannel)
	}

	// TODO add chainId, block info on each result
	response := &types.MsgExecuteAtomicTxResponse{Results: make([]types.ExecTxResult, len(msg.Txs))}

	// execute this as a go routine, otherwise execution hangs
	// we are waiting for cross-chain requests from other chains
	go func() {
		crosschainreq := <-newInternalCallChannel
		if crosschainreq.Request == nil {
			return
		}
		k.Logger(ctx).Info("received crosschain internal call", "atomic_txhash", hex.EncodeToString(txhash), "index", crosschainreq.Index, "is_query", crosschainreq.Request.IsQuery)
		req := crosschainreq.Request
		response := types.MsgExecuteCrossChainCallResponseIndexed{
			Index: crosschainreq.Index,
			Data:  &types.MsgExecuteCrossChainCallResponse{},
		}

		// TODO validation of the request
		// TODO have from data available to wasmx

		contractAddress, err := k.wasmxKeeper.AccBech32Codec().StringToAccAddressPrefixed(req.To)
		if err != nil {
			response.Data.Error = err.Error()
			newInternalCallResponseChannel <- response
			return
		}

		// TODO have special interchain addresses here!!!
		caller1, err := k.wasmxKeeper.AccBech32Codec().Bech32Codec.StringToAddressPrefixedUnsafe(req.From)
		if err != nil {
			response.Data.Error = err.Error()
			newInternalCallResponseChannel <- response
			return
		}

		caller := k.wasmxKeeper.AccBech32Codec().BytesToAccAddressPrefixed(caller1.Bytes())
		msgbz, err := k.cdc.MarshalJSON(req)
		if err != nil {
			response.Data.Error = err.Error()
			newInternalCallResponseChannel <- response
			return
		}
		execmsg := wasmxtypes.WasmxExecutionMessage{Data: msgbz}
		execmsgbz, err := json.Marshal(execmsg)
		if err != nil {
			response.Data.Error = err.Error()
			newInternalCallResponseChannel <- response
			return
		}

		reqctx := ctx
		if req.IsQuery {
			reqctx, _ = ctx.CacheContext()
		}

		respbz, err := k.wasmxKeeper.ExecuteEntryPoint(reqctx, vmcrosschain.HOST_WASMX_ENV_CROSSCHAIN, contractAddress, caller, execmsgbz, req.Dependencies, false)
		if err != nil {
			response.Data.Error = err.Error()
			newInternalCallResponseChannel <- response
			return
		}
		response.Data.Data = respbz
		k.Logger(ctx).Info("sending crosschain internal call response", "atomic_txhash", hex.EncodeToString(txhash), "index", crosschainreq.Index, "is_query", crosschainreq.Request.IsQuery)
		newInternalCallResponseChannel <- response
	}()

	for i, txbz := range msg.Txs {
		mcctx.SetCurrentSubTxIndex(ctx.ChainID(), i)
		mcctx.SetCurrentInternalCrossTx(ctx.ChainID(), 0)

		tx, err := k.actionExecutor.app.TxConfig().TxDecoder()(txbz)
		if err != nil {
			return nil, err
		}
		txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
		if !ok {
			return nil, fmt.Errorf("expected atomic transaction to have ExtensionOptionMultiChainTx")
		}
		opts := txWithExtensions.GetExtensionOptions()
		if len(opts) == 0 {
			return nil, fmt.Errorf("expected atomic transaction to have ExtensionOptionMultiChainTx")
		}
		ext := opts[0].GetCachedValue().(*types.ExtensionOptionMultiChainTx)
		chainId := ext.ChainId
		// transaction is meant for this chain, execute it
		if ctx.ChainID() == chainId {
			abcires := k.actionExecutor.GetApp().GetBaseApp().DeliverTx(txbz)

			evs := make([]types.Event, len(abcires.Events))
			for i, ev := range abcires.Events {
				attrs := make([]types.EventAttribute, len(ev.Attributes))
				for j, attr := range ev.Attributes {
					attrs[j] = types.EventAttribute{Key: attr.Key, Value: attr.Value, Index: attr.Index}
				}
				evs[i] = types.Event{Type: ev.Type, Attributes: attrs}
			}

			// make sure events are emitted on the parent context
			sdkevs := make([]sdk.Event, len(abcires.Events))
			for i, ev := range abcires.Events {
				sdkevs[i] = sdk.Event{Type: ev.Type, Attributes: ev.Attributes}
			}
			ctx.EventManager().EmitEvents(sdkevs)

			resp := types.ExecTxResult{
				Code:      abcires.Code,
				Data:      abcires.Data,
				Log:       abcires.Log,
				Info:      abcires.Info,
				GasWanted: abcires.GasWanted,
				GasUsed:   abcires.GasUsed,
				Events:    evs,
				Codespace: abcires.Codespace,
			}
			response.Results[i] = resp

			// execute this as a go routine, otherwise execution hangs
			// we send the response after each transaction execution
			go func() {
				// send our results through the channel
				newResultsChannel <- *response
			}()
		} else if slices.Contains(multichainapp.ChainIds, chainId) {
			// if this node has state for the chain, we expect the crosschain call to be executed
			// wait for the transaction to be executed
			reschannel, err := mcctx.GetResultChannel(chainId)
			var reschannel2 chan types.MsgExecuteAtomicTxResponse
			if err != nil {
				// we create it, so we can wait on it
				// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
				reschannel2 = make(chan types.MsgExecuteAtomicTxResponse, 1)
				// the other chain should close their own channels when they are done
				mcctx.SetResultChannel(chainId, &reschannel2)
			} else {
				reschannel2 = *reschannel
			}

			select {
			case resp := <-reschannel2:
				if len(resp.Results) == len(response.Results) {
					response.Results[i] = resp.Results[i]
				}
				// case <-k.goContextParent.Done():
				// TODO if node is closed during an atomic transaction
				// we try to finish the atomic transaction first
				// k.Logger(ctx).Info("stopping atomic transactions: parent context closing")
				// return nil, nil
			}
		} else {
			// the node does not have access to the state of the chain
			// we will check the internal crosschain calls for this subtx
			crosschaininfo := types.GetCrossChainCallMetaInfo(k.goContextParent, ctx.ChainID())
			if crosschaininfo == nil {
				return nil, fmt.Errorf("chain %s not found and AtomicTxCrossChainCallInfo not found on parent context: cannot execute cross call", chainId)
			}
			if len(crosschaininfo.Subtx) <= i {
				return nil, fmt.Errorf("chain %s not found and AtomicTxCrossChainCallInfo index out of bounds: %d out of %d: cannot execute cross call", chainId, i, len(crosschaininfo.Subtx))
			}
			calls := crosschaininfo.Subtx[i].Calls
			for n, call := range calls {
				if !slices.Contains(mcctx.CurrentAtomicTxChainIds, call.Request.ToChainId) {
					return nil, fmt.Errorf("atomic tx with unauthorized chain id: %s; authorized: %s", call.Request.ToChainId, strings.Join(mcctx.CurrentAtomicTxChainIds, ","))
				}
				if !slices.Contains(mcctx.CurrentAtomicTxChainIds, call.Request.FromChainId) {
					return nil, fmt.Errorf("atomic tx with unauthorized chain id: %s; authorized: %s", call.Request.FromChainId, strings.Join(mcctx.CurrentAtomicTxChainIds, ","))
				}

				// we process only the first cross-chain call that calls this chain
				// this is the entry in the subtx
				if call.Request.ToChainId == ctx.ChainID() {
					resp, err := k.ExecuteCrossChainTx(goCtx, &call.Request)
					// we compare the result with our expected result
					errmsg := ""
					if err != nil {
						errmsg = err.Error()
					}
					if call.Response.Error != errmsg || !bytes.Equal(call.Response.Data, resp.Data) {
						expected := base64.StdEncoding.EncodeToString(call.Response.Data)
						actual := base64.StdEncoding.EncodeToString(resp.Data)

						return nil, fmt.Errorf("subsequent crosschain request result mismatch at subtx %d; index %d; expected %s, found %s", i, n, expected, actual)
					}
					break
				}
			}
		}
	}
	// reset current txhash
	mcctx.CurrentAtomicTxHash = []byte{}
	mcctx.CurrentAtomicTxChainIds = []string{}
	return response, nil
}

// TODO can only be sent from wasmx, from a contract
// we can check sender is a contract
// ExecuteCrossChainTx calls a contract on another subchain
func (k *Keeper) ExecuteCrossChainTx(goCtx context.Context, msg *types.MsgExecuteCrossChainCallRequest) (*types.MsgExecuteCrossChainCallResponse, error) {
	if msg.TimeoutMs == 0 {
		return &types.MsgExecuteCrossChainCallResponse{Error: fmt.Sprintf("cross-chain tx invalid timeout %d", msg.TimeoutMs)}, nil
	}
	var newInternalCallChannel chan types.MsgExecuteCrossChainCallRequestIndexed
	var newInternalCallResponseChannel chan types.MsgExecuteCrossChainCallResponseIndexed
	ctx := sdk.UnwrapSDKContext(goCtx)

	mcctx, err := types.GetMultiChainContext(k.goContextParent)
	if err != nil {
		return nil, err
	}
	if len(mcctx.CurrentAtomicTxHash) == 0 {
		return &types.MsgExecuteCrossChainCallResponse{Error: fmt.Sprintf("current atomic tx not set: cannot execute cross call on chain_id %s", msg.ToChainId)}, nil
	}
	channelsChainId := msg.ToChainId

	// all deterministic multichain requests go through here
	// so this is where we increase the callIndex
	subTxIndex := mcctx.GetCurrentSubTxIndex(ctx.ChainID())
	callIndex := mcctx.GetCurrentInternalCrossTx(ctx.ChainID())
	mcctx.SetCurrentInternalCrossTx(ctx.ChainID(), callIndex+1)

	k.Logger(ctx).Info("executing crosschain call", "from_chain_id", msg.FromChainId, "from", msg.From, "to_chain_id", msg.ToChainId, "to", msg.To, "is_query", msg.IsQuery, "subtx", subTxIndex, "crosschaincall_index", callIndex)

	multichainapp, err := cfg.GetMultiChainApp(k.goContextParent)
	if err != nil {
		return nil, err
	}
	_, err = multichainapp.GetApp(channelsChainId)
	if err != nil {
		// this is part of an atomic tx and this node does not have access to the state of this chain, so we cannot execute this tx ourself
		// we have a system for internal crosschain tx & results, precomputed by the proposer node, that has access to the state of all affected chains
		// TODO in the future we will support nodes composing the execution from partial results that they can execute on their available state; these partial executions will compose the final execution result; and a system of observers that can execute the computations in paralell, even before the nodes decide on the next block to finalize

		// return &types.MsgExecuteCrossChainCallResponse{Error: "", Data: []byte{}}, nil

		crosschaininfo := types.GetCrossChainCallMetaInfo(k.goContextParent, ctx.ChainID())
		if crosschaininfo == nil {
			return &types.MsgExecuteCrossChainCallResponse{Error: fmt.Sprintf("chain %s not found and AtomicTxCrossChainCallInfo not found on parent context: cannot execute cross call", channelsChainId)}, nil
		}
		if len(crosschaininfo.Subtx) <= subTxIndex {
			return &types.MsgExecuteCrossChainCallResponse{Error: fmt.Sprintf("chain %s not found and AtomicTxCrossChainCallInfo index out of bounds: %d out of %d: cannot execute cross call", channelsChainId, subTxIndex, len(crosschaininfo.Subtx))}, nil
		}
		calls := crosschaininfo.Subtx[subTxIndex].Calls

		if int(callIndex) >= len(calls) {
			return &types.MsgExecuteCrossChainCallResponse{Error: fmt.Sprintf("chain %s not found and last crosschain call index %d out of bounds", channelsChainId, callIndex)}, nil
		}
		req := calls[callIndex].Request
		if req.From != msg.From ||
			req.FromChainId != msg.FromChainId ||
			req.To != msg.To ||
			req.ToChainId != msg.ToChainId ||
			req.IsQuery != msg.IsQuery ||
			!bytes.Equal(req.Msg, msg.Msg) ||
			!slices.Equal(req.Dependencies, msg.Dependencies) ||
			!slices.Equal(req.Funds, msg.Funds) {
			msgbz, _ := json.Marshal(msg)
			reqbz, _ := json.Marshal(&req)
			return &types.MsgExecuteCrossChainCallResponse{Error: fmt.Sprintf("crosschain request mismatch at index %d; expected %s, found %s", callIndex, string(msgbz), string(reqbz))}, nil
		}
		result := calls[callIndex].Response

		// we execute all subsequent crosschain calls before we return the result
		// like it would happen when executing subcalls in order
		for i, call := range calls[callIndex+1:] {
			if !slices.Contains(mcctx.CurrentAtomicTxChainIds, call.Request.ToChainId) {
				return nil, fmt.Errorf("atomic tx with unauthorized chain id: %s; authorized: %s", call.Request.ToChainId, strings.Join(mcctx.CurrentAtomicTxChainIds, ","))
			}
			if !slices.Contains(mcctx.CurrentAtomicTxChainIds, call.Request.FromChainId) {
				return nil, fmt.Errorf("atomic tx with unauthorized chain id: %s; authorized: %s", call.Request.FromChainId, strings.Join(mcctx.CurrentAtomicTxChainIds, ","))
			}

			// we have a cross-chain call that calls this chain
			// then we process it
			// if it calls a subchain that we have access to, then that chain will catch it and process it during the atomic transaction
			if call.Request.ToChainId == ctx.ChainID() {

				// TODO if a subchain that we have access to calls this chain, then that subchain will initiate the crosschain request
				// && !slices.Contains(multichainapp.ChainIds, call.Request.FromChainId)

				resp, err := k.ExecuteCrossChainTx(goCtx, &call.Request)
				// we compare the result with our expected result
				errmsg := ""
				if err != nil {
					errmsg = err.Error()
				}
				if call.Response.Error != errmsg || !bytes.Equal(call.Response.Data, resp.Data) {
					expected := base64.StdEncoding.EncodeToString(call.Response.Data)
					actual := base64.StdEncoding.EncodeToString(resp.Data)
					return &types.MsgExecuteCrossChainCallResponse{Error: fmt.Sprintf("subsequent crosschain request result mismatch at index %d; expected %s, found %s", callIndex+i, expected, actual)}, nil
				}
				// we break here because the rest of the internal crosschain calls
				// are executed above, during k.ExecuteCrossChainTx - this is a nested process
				break
			}
		}

		// if we do not have preprocessed data, we return an error
		return &types.MsgExecuteCrossChainCallResponse{Error: result.Error, Data: result.Data}, nil
	}

	// we get the channels for the chain we want to interact with
	existent2, err := mcctx.GetInternalCallChannel(channelsChainId)
	if err == nil {
		newInternalCallChannel = *existent2
	}
	existent3, err := mcctx.GetInternalCallResultChannel(channelsChainId)
	if err == nil {
		newInternalCallResponseChannel = *existent3
	}

	if newInternalCallChannel == nil {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newInternalCallChannel = make(chan types.MsgExecuteCrossChainCallRequestIndexed, 1)
		mcctx.SetInternalCallChannel(channelsChainId, &newInternalCallChannel)
	}

	if newInternalCallResponseChannel == nil {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newInternalCallResponseChannel = make(chan types.MsgExecuteCrossChainCallResponseIndexed, 1)
		mcctx.SetInternalCallResultChannel(channelsChainId, &newInternalCallResponseChannel)
	}

	req := types.MsgExecuteCrossChainCallRequestIndexed{
		Index:   int32(callIndex),
		Request: msg,
	}

	// we set the timer for the tx timeout
	timeout := time.Millisecond * time.Duration(msg.TimeoutMs)
	k.Logger(ctx).Info("executing crosschain call with timeout", "subtx", subTxIndex, "crosschaincall_index", callIndex, "timeout", timeout, "from_chain_id", msg.FromChainId, "to_chain_id", msg.ToChainId, "is_query", msg.IsQuery, "from", msg.From, "to", msg.To)

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	k.Logger(ctx).Debug("executing crosschain call", "subtx", subTxIndex, "crosschaincall_index", callIndex, "timeout", timeout, "from_chain_id", msg.FromChainId, "to_chain_id", msg.ToChainId, "is_query", msg.IsQuery, "from", msg.From, "to", msg.To, "msg", hex.EncodeToString(msg.Msg))

	// execute this as a go routine, otherwise execution hangs
	// we send the request for cross chain execution
	go func() {
		newInternalCallChannel <- req
	}()

	// we wait for the response
	announcedCtxClose := false
	for {
		select {
		case resp := <-newInternalCallResponseChannel:
			k.Logger(ctx).Debug("executed crosschain call", "subtx", subTxIndex, "crosschaincall_index", callIndex, "from_chain_id", msg.FromChainId, "to_chain_id", msg.ToChainId, "is_query", msg.IsQuery, "from", msg.From, "to", msg.To, "response_data", hex.EncodeToString(resp.Data.Data), "response_error", resp.Data.Error)
			return resp.Data, nil
		case <-ctxWithTimeout.Done():
			k.Logger(ctx).Info("stopping crosschain call: timeout", "subtx", subTxIndex, "crosschaincall_index", callIndex)
			return &types.MsgExecuteCrossChainCallResponse{Error: "crosschain call timeout"}, nil
		case <-k.goContextParent.Done():
			if !announcedCtxClose {
				k.Logger(ctx).Info("parent context closing: crosschain call in execution, please wait for process to finish or timeout", "subtx", subTxIndex, "crosschaincall_index", callIndex, "timeout", timeout)
			}
			announcedCtxClose = true
			// TODO if node is closed during an atomic transaction
			// we try to finish the atomic transaction first
			// k.Logger(ctx).Info("stopping crosschain call: parent context closing")
			// return nil, nil
		}
	}
}
