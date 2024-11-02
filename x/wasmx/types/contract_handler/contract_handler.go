package contracthandler

import (
	"encoding/json"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "mythos/v1/codec"
	"mythos/v1/x/wasmx/types"
)

const (
	CONTEXT_COIN_TYPE_KEY        = "wasmx.alias.coinType"
	COIN_TYPE_COSMOS      uint32 = 118
	COIN_TYPE_ETH         uint32 = 60
)

type ContractHandlerMessage struct {
	Role   string
	Method string
	Msg    any
	Sender sdk.AccAddress
}

type ContractHandler interface {
	Encode(req ContractHandlerMessage) (*types.WasmxExecutionMessage, error)
	Decode(method string, data []byte) (any, error)
}

type KeeperInterface interface {
	GetAddressOrRole(ctx sdk.Context, addressOrRole string) (mcodec.AccAddressPrefixed, error)
	Query(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, dependencies []string, inBackground bool) ([]byte, error)

	AccBech32Codec() mcodec.AccBech32Codec
}

// role => handler
type ContractHandlerMap struct {
	Keeper KeeperInterface
	Map    map[string]ContractHandler
}

func NewContractHandlerMap(keeper KeeperInterface) ContractHandlerMap {
	return ContractHandlerMap{
		Keeper: keeper,
		Map:    map[string]ContractHandler{},
	}
}

func (m ContractHandlerMap) Register(role string, handler ContractHandler) {
	m.Map[role] = handler
}

func (m ContractHandlerMap) Query(ctx sdk.Context, req ContractHandlerMessage) (any, error) {
	handler, ok := m.Map[req.Role]
	if !ok {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "no handler found")
	}
	msg, err := handler.Encode(req)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "encoding failed: %s", err.Error())
	}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "json encoding failed: %s", err.Error())
	}

	contractAddress, err := m.Keeper.GetAddressOrRole(ctx, req.Role)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "role not registered")
	}

	tmpctx, _ := ctx.CacheContext()
	responseBz, err := m.Keeper.Query(tmpctx, contractAddress, m.Keeper.AccBech32Codec().BytesToAccAddressPrefixed(req.Sender), msgbz, nil, nil)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "query failed: %s", err.Error())
	}

	var response types.WasmxQueryResponse
	err = json.Unmarshal(responseBz, &response)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "decoding query response failed: %s", err.Error())
	}

	res, err := handler.Decode(req.Method, response.Data)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "decoding failed: %s", err.Error())
	}
	return res, nil
}

func (m ContractHandlerMap) Execute(ctx sdk.Context, req ContractHandlerMessage) (any, error) {
	handler, ok := m.Map[req.Role]
	if !ok {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "no handler found")
	}
	msg, err := handler.Encode(req)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "encoding failed: %s", err.Error())
	}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "json encoding failed: %s", err.Error())
	}

	contractAddress, err := m.Keeper.GetAddressOrRole(ctx, req.Role)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "role not registered")
	}

	response, err := m.Keeper.Execute(ctx, contractAddress, m.Keeper.AccBech32Codec().BytesToAccAddressPrefixed(req.Sender), msgbz, nil, nil, false)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "execution failed: %s", err.Error())
	}

	res, err := handler.Decode(req.Method, response)
	if err != nil {
		return nil, sdkerr.Wrapf(types.ErrInvalidCoreContractCall, "decoding failed: %s", err.Error())
	}
	return res, nil
}
