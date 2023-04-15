package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ark-us/evm2wat"
	"github.com/ark-us/wat2wasm"

	"wasmx/x/wasmx/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) StoreCode(goCtx context.Context, msg *types.MsgStoreCode) (*types.MsgStoreCodeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	codeId, checksum, err := m.Keeper.Create(ctx, senderAddr, msg.WasmByteCode)
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreCodeResponse{
		CodeId:   codeId,
		Checksum: checksum,
	}, nil
}

func (m msgServer) StoreCodeEvm(goCtx context.Context, msg *types.MsgStoreCodeEvm) (*types.MsgStoreCodeEvmResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	watCode, err := evm2wat.EvmToWat(msg.EvmByteCode, evm2wat.DefaultOptions())
	if err != nil {
		return nil, err
	}
	wasmByteCode, err := wat2wasm.WatToWasm(watCode.Code, m.tempDir)
	if err != nil {
		return nil, err
	}

	codeId, checksum, err := m.Keeper.Create(ctx, senderAddr, wasmByteCode)
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreCodeEvmResponse{
		CodeId:   codeId,
		Checksum: checksum,
	}, nil
}

// InstantiateContract instantiate a new contract with classic sequence based address generation
func (m msgServer) InstantiateContract(goCtx context.Context, msg *types.MsgInstantiateContract) (*types.MsgInstantiateContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	contractAddr, data, err := m.Keeper.Instantiate(ctx, msg.CodeId, senderAddr, msg.Msg, msg.Label, msg.Funds)
	if err != nil {
		return nil, err
	}

	return &types.MsgInstantiateContractResponse{
		Address: contractAddr.String(),
		Data:    data,
	}, nil
}

// InstantiateContract2 instantiate a new contract with predicatable address generated
func (m msgServer) InstantiateContract2(goCtx context.Context, msg *types.MsgInstantiateContract2) (*types.MsgInstantiateContract2Response, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))
	contractAddr, data, err := m.Keeper.Instantiate2(ctx, msg.CodeId, senderAddr, msg.Msg, msg.Label, msg.Funds, msg.Salt, msg.FixMsg)
	if err != nil {
		return nil, err
	}

	return &types.MsgInstantiateContract2Response{
		Address: contractAddr.String(),
		Data:    data,
	}, nil
}

// CompileContract does an AOT compilation for a contract
func (m msgServer) CompileContract(goCtx context.Context, msg *types.MsgCompileContract) (*types.MsgCompileContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	err := m.Keeper.PinCode(ctx, msg.CodeId, "")
	if err != nil {
		return nil, err
	}

	return &types.MsgCompileContractResponse{}, nil
}

func (m msgServer) ExecuteContract(goCtx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}
	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract")
	}

	// TODO make the dependencies unique - remove duplicates

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	data, err := m.Keeper.Execute(ctx, contractAddr, senderAddr, msg.Msg, msg.Funds, msg.Dependencies)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteContractResponse{
		Data: data,
	}, nil
}

func (m msgServer) ExecuteWithOriginContract(goCtx context.Context, msg *types.MsgExecuteWithOriginContract) (*types.MsgExecuteContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	originAddr, err := sdk.AccAddressFromBech32(msg.Origin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "origin")
	}
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}
	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	data, err := m.Keeper.ExecuteWithOrigin(ctx, originAddr, contractAddr, senderAddr, msg.Msg, msg.Funds)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteContractResponse{
		Data: data,
	}, nil
}

func (m msgServer) ExecuteDelegateContract(goCtx context.Context, msg *types.MsgExecuteDelegateContract) (*types.MsgExecuteDelegateContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	originAddr, err := sdk.AccAddressFromBech32(msg.Origin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "origin")
	}
	callerAddr, err := sdk.AccAddressFromBech32(msg.Caller)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}
	codeContractAddr, err := sdk.AccAddressFromBech32(msg.CodeContract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "code_contract")
	}
	storageContractAddr, err := sdk.AccAddressFromBech32(msg.StorageContract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "storage_contract")
	}

	if msg.Sender != msg.StorageContract {
		return nil, sdkerrors.Wrapf(types.ErrInvalidMsg, "execute delegate must be called from the storage contract: %s; it was called from: %s", msg.StorageContract, msg.Sender)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	data, err := m.Keeper.ExecuteDelegate(ctx, originAddr, codeContractAddr, storageContractAddr, callerAddr, msg.Msg, msg.Funds)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteDelegateContractResponse{
		Data: data,
	}, nil
}
