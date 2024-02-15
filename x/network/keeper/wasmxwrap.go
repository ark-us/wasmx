package keeper

import (
	"encoding/json"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (k *Keeper) ExecuteContract(ctx sdk.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	senderAddr, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender address or role: "+msg.Sender)
	}
	contractAddress, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract address or role: "+msg.Contract)
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

func (k *Keeper) ExecuteEntryPoint(ctx sdk.Context, entryPoint string, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	senderAddr, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := k.wasmxKeeper.ExecuteEntryPoint(ctx, entryPoint, contractAddress, senderAddr, execmsgbz, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteContractResponse{
		Data: resp,
	}, nil
}

func (k *Keeper) QueryContract(ctx sdk.Context, msg *types.MsgQueryContract) (*types.MsgQueryContractResponse, error) {
	senderAddr, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := k.wasmxKeeper.GetAddressOrRole(ctx, msg.Contract)
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
