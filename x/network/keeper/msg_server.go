package keeper

import (
	"context"
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	cfg "mythos/v1/config"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type msgServer struct {
	*Keeper
}

type MsgServerInternal interface {
	types.MsgServer
	ExecuteContract(ctx sdk.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error)
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) MsgServerInternal {
	return &msgServer{
		Keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

// TODO do we stil nee this?
var _ types.BroadcastAPIServer = msgServer{}

func (m msgServer) Ping(goCtx context.Context, msg *types.RequestPing) (*types.ResponsePing, error) {
	return &types.ResponsePing{}, nil
}

func (m msgServer) BroadcastTx(goCtx context.Context, msg *types.RequestBroadcastTx) (*types.ResponseBroadcastTx, error) {
	// TODO BroadcastTxCommit and return receipt
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgbz := []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(msg.Tx)))
	rresp, err := m.Keeper.ExecuteContract(ctx, &types.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_CONSENSUS,
		Contract: wasmxtypes.ROLE_CONSENSUS,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}

	// return &types.ResponseBroadcastTx{
	// 	CheckTx: &abci.ResponseCheckTx{
	// 		Code: res.CheckTx.Code,
	// 		Data: res.CheckTx.Data,
	// 		Log:  res.CheckTx.Log,
	// 	},
	// 	TxResult: &abci.ExecTxResult{
	// 		Code: res.TxResult.Code,
	// 		Data: res.TxResult.Data,
	// 		Log:  res.TxResult.Log,
	// 	},
	// }, nil

	return &types.ResponseBroadcastTx{
		CheckTx: &types.ResponseCheckTx{
			Code: 0,
			Data: rresp.Data,
			Log:  "",
		},
		TxResult: &types.ExecTxResult{
			Code: 0,
			Data: rresp.Data,
			Log:  "",
		},
	}, nil
}

func (m msgServer) MultiChainWrap(goCtx context.Context, msg *types.MsgMultiChainWrap) (*types.MsgMultiChainWrapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var sdkmsg sdk.Msg
	err := m.cdc.UnpackAny(msg.Data, &sdkmsg)
	if err != nil {
		return nil, err
	}

	multichainapp, err := cfg.GetMultiChainApp(m.goContextParent)
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

	owner, err := m.wasmxKeeper.AccBech32Codec().StringToAccAddressPrefixed(msg.Sender)
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

func (m msgServer) GrpcSendRequest(goCtx context.Context, msg *types.MsgGrpcSendRequest) (*types.MsgGrpcSendRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ip := msg.IpAddress
	client := StartGRPCClient(ip)
	req := &types.MsgGrpcReceiveRequest{
		Data:     msg.Data,
		Sender:   msg.Sender,
		Contract: msg.Contract,
		Encoding: msg.Encoding,
	}
	res, err := client.GrpcReceiveRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrpcSendRequestResponse{
		Data: res.Data,
	}, nil
}

func (m msgServer) GrpcReceiveRequest(goCtx context.Context, msg *types.MsgGrpcReceiveRequest) (*types.MsgGrpcReceiveRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	msg2 := &types.MsgExecuteContract{
		Sender:   msg.Sender,
		Contract: msg.Contract,
		Msg:      msg.Data,
	}
	resp, err := m.Keeper.ExecuteContract(ctx, msg2)

	if err != nil {
		return nil, err
	}

	return &types.MsgGrpcReceiveRequestResponse{
		Data: resp.Data,
	}, nil
}

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the interval is for?
func (m msgServer) StartTimeout(goCtx context.Context, msg *types.MsgStartTimeoutRequest) (*types.MsgStartTimeoutResponse, error) {
	return m.Keeper.StartTimeout(goCtx, msg)
}

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the background process is for ?
func (m msgServer) StartBackgroundProcess(goCtx context.Context, msg *types.MsgStartBackgroundProcessRequest) (*types.MsgStartBackgroundProcessResponse, error) {
	// TODO only started from wasmx, only by system contracts
	return m.Keeper.StartBackgroundProcess(goCtx, msg)
}

func (m msgServer) P2PReceiveMessage(goCtx context.Context, msg *types.MsgP2PReceiveMessageRequest) (*types.MsgP2PReceiveMessageResponse, error) {
	return m.Keeper.P2PReceiveMessage(goCtx, msg)
}

func (m msgServer) ExecuteAtomicTx(goCtx context.Context, msg *types.MsgExecuteAtomicTxRequest) (*types.MsgExecuteAtomicTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var newchannel chan types.MsgExecuteAtomicTxResponse
	if msg.LeaderChainId != ctx.ChainID() {
		existent, err := GetMultiChainChannel(m.goContextParent, ctx.ChainID())
		if err == nil {
			newchannel = *existent
		}
	}

	if newchannel == nil {
		// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
		newchannel = make(chan types.MsgExecuteAtomicTxResponse, 1)
		defer func() {
			close(newchannel)
			mcctx, err := GetMultiChainContext(m.goContextParent)
			if err != nil {
				m.Logger(ctx).Error("cannot get multi chain context from parent context")
			}
			delete(mcctx.ResultChannels, ctx.ChainID())
		}()
		SetMultiChainContext(m.goContextParent, ctx.ChainID(), &newchannel)
	}

	// TODO add chainId, block info on each result
	response := &types.MsgExecuteAtomicTxResponse{Results: make([]types.ExecTxResult, len(msg.Txs))}

	chainIds := make([]string, len(msg.Txs))

	for i, txbz := range msg.Txs {
		tx, err := m.actionExecutor.app.TxConfig().TxDecoder()(txbz)
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
		chainIds[i] = ext.ChainId
		// if transaction is meant for another chain, skip it
		if ctx.ChainID() != ext.ChainId {
			continue
		}
		abcires := m.actionExecutor.GetApp().GetBaseApp().DeliverTx(txbz)

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
	}

	// execute this as a go routine, otherwise execution hangs
	go func() {
		// send our results through the channel
		newchannel <- *response
	}()

	for i, chainId := range chainIds {
		if ctx.ChainID() != chainId {
			reschannel, err := GetMultiChainChannel(m.goContextParent, chainId)
			var reschannel2 chan types.MsgExecuteAtomicTxResponse
			if err != nil {
				// we create it, so we can wait on it
				// these channels need to be buffered to prevent the goroutine below from hanging indefinitely
				reschannel2 = make(chan types.MsgExecuteAtomicTxResponse, 1)
				defer func() {
					close(reschannel2)
					mcctx, err := GetMultiChainContext(m.goContextParent)
					if err != nil {
						m.Logger(ctx).Error("cannot get multi chain context from parent context")
					}
					delete(mcctx.ResultChannels, chainId)
				}()
				SetMultiChainContext(m.goContextParent, chainId, &reschannel2)
			} else {
				reschannel2 = *reschannel
			}

			select {
			case resp := <-reschannel2:
				if len(resp.Results) == len(response.Results) {
					response.Results[i] = resp.Results[i]
				}
			case <-m.goContextParent.Done():
				m.Logger(ctx).Info("stopping atomic transactions: parent context closing")
				// TODO what to do here? return error if node is closed during an atomic transaction?
				// we should abort the execution of the transaction
				return nil, nil
			}
		}
	}
	return response, nil
}
