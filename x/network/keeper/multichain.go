package keeper

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/crypto/merkle"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"mythos/v1/x/network/types"
	"mythos/v1/x/network/vmcrosschain"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// an atomic batch of multichain transactions
// each transaction may affect only one chain or > 1, if they contain MsgExecuteCrossChainTxRequest messages
// each transaction may contain multiple MsgExecuteCrossChainTxRequest messages, with multiple internal cross-chain transactions
// TODO: ExecuteAtomicTx must not be nested inside other ExecuteAtomicTx
func (k *Keeper) ExecuteAtomicTx(goCtx context.Context, msg *types.MsgExecuteAtomicTxRequest) (*types.MsgExecuteAtomicTxResponse, error) {
	// our atomic tx result channel - we use this to send our result
	var newResultsChannel chan types.MsgExecuteAtomicTxResponse
	// our cross-chain tx request channel - we receive contract calls from other chains
	var newInternalCallChannel chan types.MsgExecuteCrossChainTxRequestIndexed
	// our cross-chain tx response channel - we send the results of cross-chain calls to other chains
	var newInternalCallResponseChannel chan types.MsgExecuteCrossChainTxResponseIndexed

	ctx := sdk.UnwrapSDKContext(goCtx)
	txhash := merkle.HashFromByteSlices(msg.GetTxs())
	mcctx, err := GetMultiChainContext(k.goContextParent)
	if err != nil {
		return nil, err
	}
	mcctx.CurrentAtomicTxHash = txhash

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
		newInternalCallChannel = make(chan types.MsgExecuteCrossChainTxRequestIndexed, 1)
		mcctx.SetInternalCallChannel(ctx.ChainID(), &newInternalCallChannel)
	}

	existent3, err := mcctx.GetInternalCallResultChannel(ctx.ChainID())
	if err == nil {
		newInternalCallResponseChannel = *existent3
	} else {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newInternalCallResponseChannel = make(chan types.MsgExecuteCrossChainTxResponseIndexed, 1)
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
		k.Logger(ctx).Info("received crosschain internal tx", "atomic_txhash", hex.EncodeToString(txhash), "index", crosschainreq.Index)
		req := crosschainreq.Request
		response := types.MsgExecuteCrossChainTxResponseIndexed{
			Index: crosschainreq.Index,
			Data:  &types.MsgExecuteCrossChainTxResponse{},
		}

		// TODO validation of the request
		// TODO have from data available to wasmx

		contractAddress, err := k.wasmxKeeper.AccBech32Codec().StringToAccAddressPrefixed(req.ToAddressOrRole)
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

		respbz, err := k.wasmxKeeper.ExecuteEntryPoint(ctx, vmcrosschain.HOST_WASMX_ENV_CROSSCHAIN, contractAddress, caller, execmsgbz, req.Dependencies, false)
		if err != nil {
			response.Data.Error = err.Error()
			newInternalCallResponseChannel <- response
			return
		}
		response.Data.Data = respbz
		k.Logger(ctx).Info("sending crosschain internal tx response", "atomic_txhash", hex.EncodeToString(txhash), "index", crosschainreq.Index)
		newInternalCallResponseChannel <- response
	}()

	for i, txbz := range msg.Txs {
		mcctx.CurrentSubTxIndex = int32(i)
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
		} else {
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
			case <-k.goContextParent.Done():
				k.Logger(ctx).Info("stopping atomic transactions: parent context closing")
				// TODO what to do here? return error if node is closed during an atomic transaction?
				// we should abort the execution of the transaction
				return nil, nil
			}
		}
	}
	return response, nil
}

func (k *Keeper) ExecuteCrossChainTx(goCtx context.Context, msg *types.MsgExecuteCrossChainTxRequest) (*types.MsgExecuteCrossChainTxResponse, error) {
	// TODO can only be sent from wasmx
	var newInternalCallChannel chan types.MsgExecuteCrossChainTxRequestIndexed
	var newInternalCallResponseChannel chan types.MsgExecuteCrossChainTxResponseIndexed
	ctx := sdk.UnwrapSDKContext(goCtx)

	msg.FromChainId = ctx.ChainID()
	// TODO role

	k.Logger(ctx).Info("executing crosschain tx", "from_chain_id", msg.FromChainId, "from", msg.From, "from_role", msg.FromRole, "to_chain_id", msg.ToChainId, "to", msg.ToAddressOrRole)

	mcctx, err := GetMultiChainContext(k.goContextParent)
	if err != nil {
		return nil, err
	}
	channelsChainId := msg.ToChainId

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
		newInternalCallChannel = make(chan types.MsgExecuteCrossChainTxRequestIndexed, 1)
		mcctx.SetInternalCallChannel(channelsChainId, &newInternalCallChannel)
	}

	if newInternalCallResponseChannel == nil {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newInternalCallResponseChannel = make(chan types.MsgExecuteCrossChainTxResponseIndexed, 1)
		mcctx.SetInternalCallResultChannel(channelsChainId, &newInternalCallResponseChannel)
	}

	index := mcctx.CurrentInternalCrossTx
	req := types.MsgExecuteCrossChainTxRequestIndexed{
		Index:   index,
		Request: msg,
	}
	mcctx.CurrentInternalCrossTx += 1

	// execute this as a go routine, otherwise execution hangs
	// we send the request for cross chain execution
	go func() {
		newInternalCallChannel <- req
	}()

	// we wait for the response
	select {
	case resp := <-newInternalCallResponseChannel:
		return resp.Data, nil
	case <-k.goContextParent.Done():
		k.Logger(ctx).Info("stopping atomic transactions: parent context closing")
		// TODO what to do here? return error if node is closed during an atomic transaction?
		// we should abort the execution of the transaction
		return nil, nil
	}
}
