package keeper

import (
	"encoding/json"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (k *Keeper) ExecuteContract(ctx sdk.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := k.wasmxKeeper.Execute(ctx, contractAddress, senderAddr, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteContractResponse{
		Data: resp,
	}, nil
}

func (k *Keeper) ExecuteEventual(ctx sdk.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := k.wasmxKeeper.ExecuteEventual(ctx, contractAddress, senderAddr, execmsgbz, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteContractResponse{
		Data: resp,
	}, nil
}

func (k *Keeper) QueryContract(ctx sdk.Context, msg *types.MsgQueryContract) (*types.MsgQueryContractResponse, error) {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := k.wasmxKeeper.Query(ctx, contractAddress, senderAddr, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgQueryContractResponse{
		Data: resp,
	}, nil
}
