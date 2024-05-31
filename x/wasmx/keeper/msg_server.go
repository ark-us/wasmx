package keeper

import (
	"context"
	"encoding/json"

	sdkerr "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"mythos/v1/x/wasmx/types"
	cchtypes "mythos/v1/x/wasmx/types/contract_handler"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// This is the entrypoint for transactions signed by Ethereum wallets
// Works with both EVM & CosmWasm contracts, both interpreted and wasm-based
func (m msgServer) ExecuteEth(goCtx context.Context, msg *types.MsgExecuteEth) (*types.MsgExecuteEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx = ctx.WithValue(cchtypes.CONTEXT_COIN_TYPE_KEY, cchtypes.COIN_TYPE_ETH)
	tx := msg.AsTransaction()
	senderAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "ExecuteEth could not parse sender address")
	}

	// Alias may be set in the AnteHandler EthSigVerificationDecorator

	// TODO denom
	funds := sdk.NewCoins(sdk.NewCoin(m.Keeper.denom, sdkmath.NewIntFromBigInt(tx.Value())))

	to := tx.To()
	var data []byte
	if to == nil {
		deps := []string{types.INTERPRETER_EVM_SHANGHAI}
		msg := types.WasmxExecutionMessage{Data: []byte{}}
		msgbz, err := json.Marshal(msg)
		if err != nil {
			sdkerr.Wrap(err, "ExecuteEth could not marshal data")
		}
		_, _, address, err := m.Keeper.Deploy(ctx, senderAddr, tx.Data(), deps, types.CodeMetadata{}, msgbz, funds, "")
		if err != nil {
			return nil, err
		}
		data = address.Bytes()
	} else {
		contractAddr := types.AccAddressFromEvm(*to)
		msg := types.WasmxExecutionMessage{Data: tx.Data()}
		msgbz, err := json.Marshal(msg)
		if err != nil {
			sdkerr.Wrap(err, "ExecuteEth could not marshal data")
		}
		data, err = m.Keeper.Execute(ctx, m.accBech32Codec.BytesToAccAddressPrefixed(contractAddr), senderAddr, msgbz, funds, nil, false)
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeExecuteEth))

	return &types.MsgExecuteEthResponse{
		Data: data,
	}, nil
}

func (m msgServer) StoreCode(goCtx context.Context, msg *types.MsgStoreCode) (*types.MsgStoreCodeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	codeId, checksum, err := m.Keeper.Create(ctx, senderAddr, msg.ByteCode, msg.Deps, msg.Metadata)
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreCodeResponse{
		CodeId:   codeId,
		Checksum: checksum,
	}, nil
}

func (m msgServer) DeployCode(goCtx context.Context, msg *types.MsgDeployCode) (*types.MsgDeployCodeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	codeId, checksum, address, err := m.Keeper.Deploy(ctx, senderAddr, msg.ByteCode, msg.Deps, msg.Metadata, msg.Msg, msg.Funds, msg.Label)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeployCodeResponse{
		CodeId:   codeId,
		Checksum: checksum,
		Address:  address.String(),
	}, nil
}

// InstantiateContract instantiate a new contract with classic sequence based address generation
func (m msgServer) InstantiateContract(goCtx context.Context, msg *types.MsgInstantiateContract) (*types.MsgInstantiateContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	contractAddr, data, err := m.Keeper.Instantiate(ctx, msg.CodeId, senderAddr, msg.Msg, msg.Funds, msg.Label)
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

	senderAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))
	contractAddr, data, err := m.Keeper.Instantiate2(ctx, msg.CodeId, senderAddr, msg.Msg, msg.Funds, msg.Salt, msg.FixMsg, msg.Label)
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

	authority := m.Keeper.GetAuthority()
	if authority != msg.Authority {
		return nil, sdkerr.Wrapf(errortypes.ErrUnauthorized, "invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Authority),
	))

	err := m.Keeper.PinCode(ctx, msg.CodeId, "")
	if err != nil {
		return nil, err
	}

	return &types.MsgCompileContractResponse{}, nil
}

func (m msgServer) ExecuteContract(goCtx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	return m.Keeper.ExecuteContract(goCtx, msg)
}

func (k *Keeper) ExecuteContract(goCtx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr, err := k.accBech32Codec.StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddr, err := k.GetAddressOrRole(ctx, msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	if types.IsSystemAddress(contractAddr.Bytes()) {
		return nil, sdkerr.Wrap(types.ErrUnauthorizedAddress, "cannot call system address")
	}

	// TODO make the dependencies unique - remove duplicates

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	data, err := k.Execute(ctx, contractAddr, senderAddr, msg.Msg, msg.Funds, msg.Dependencies, false)
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
	originAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Origin)
	if err != nil {
		return nil, sdkerr.Wrap(err, "origin")
	}
	senderAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	if types.IsSystemAddress(contractAddr.Bytes()) {
		return nil, sdkerr.Wrap(types.ErrUnauthorizedAddress, "cannot call system address")
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
	originAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Origin)
	if err != nil {
		return nil, sdkerr.Wrap(err, "origin")
	}
	callerAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.Caller)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	codeContractAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.CodeContract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "code_contract")
	}
	storageContractAddr, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.StorageContract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "storage_contract")
	}

	if types.IsSystemAddress(codeContractAddr.Bytes()) {
		return nil, sdkerr.Wrap(types.ErrUnauthorizedAddress, "cannot call system address")
	}
	if types.IsSystemAddress(storageContractAddr.Bytes()) {
		return nil, sdkerr.Wrap(types.ErrUnauthorizedAddress, "cannot call system address")
	}

	if msg.Sender != msg.StorageContract {
		return nil, sdkerr.Wrapf(types.ErrInvalidMsg, "execute delegate must be called from the storage contract: %s; it was called from: %s", msg.StorageContract, msg.Sender)
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

func (m msgServer) RegisterRole(goCtx context.Context, msg *types.MsgRegisterRole) (*types.MsgRegisterRoleResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority := m.Keeper.GetAuthority()
	if authority != msg.Authority {
		return nil, sdkerr.Wrapf(errortypes.ErrUnauthorized, "invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	if _, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.ContractAddress); err != nil {
		return nil, sdkerr.Wrap(err, "contract address")
	}

	err := m.Keeper.RegisterRoleHandler(ctx, msg.Role, msg.Label, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRegisterRole,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyContractAddr, msg.ContractAddress),
		sdk.NewAttribute(types.AttributeKeyRole, msg.Role),
		sdk.NewAttribute(types.AttributeKeyRoleLabel, msg.Label),
	))

	return &types.MsgRegisterRoleResponse{}, nil
}

func (m msgServer) DeregisterRole(goCtx context.Context, msg *types.MsgDeregisterRole) (*types.MsgDeregisterRoleResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority := m.Keeper.GetAuthority()
	if authority != msg.Authority {
		return nil, sdkerr.Wrapf(errortypes.ErrUnauthorized, "invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	contractAddress, err := m.accBech32Codec.StringToAccAddressPrefixed(msg.ContractAddress)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract address")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDeregisterRole,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyContractAddr, msg.ContractAddress),
	))

	m.Keeper.DeregisterRole(ctx, contractAddress)

	return &types.MsgDeregisterRoleResponse{}, nil
}
