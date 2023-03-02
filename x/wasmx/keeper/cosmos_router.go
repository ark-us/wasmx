package keeper

import (
	"errors"
	"math/big"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"wasmx/x/wasmx/types"
)

type WasmxCosmosHandler struct {
	Ctx    sdk.Context
	Keeper Keeper
	Caller sdk.AccAddress
}

func (h *WasmxCosmosHandler) WithContext(newctx sdk.Context) {
	h.Ctx = newctx
}
func (h *WasmxCosmosHandler) SubmitCosmosQuery(reqQuery abci.RequestQuery) ([]byte, error) {
	return h.Keeper.SubmitCosmosQuery(h.Ctx, reqQuery)
}
func (h *WasmxCosmosHandler) ExecuteCosmosMsg(any *cdctypes.Any) ([]byte, error) {
	return h.Keeper.ExecuteCosmosMsg(h.Ctx, any, h.Caller)
}
func (h *WasmxCosmosHandler) GetBalance(addr sdk.AccAddress) *big.Int {
	balance := h.Keeper.bank.GetBalance(h.Ctx, addr, h.Keeper.denom)
	return balance.Amount.BigInt()
}
func (h *WasmxCosmosHandler) GetBlockHash(blockNumber uint64) types.Checksum {
	return make([]byte, 32)
}

func (k Keeper) newCosmosHandler(ctx sdk.Context, caller sdk.AccAddress) types.WasmxCosmosHandler {
	return &WasmxCosmosHandler{
		Ctx:    ctx,
		Keeper: k,
		Caller: caller,
	}
}

func (k Keeper) SubmitCosmosQuery(ctx sdk.Context, reqQuery abci.RequestQuery) ([]byte, error) {
	// TODO if we allow historical queries, at a certain block
	// use app.Query(queryReq)
	queryFn := k.grpcQueryRouter.Route(reqQuery.Path)
	res, err := safeQuery(ctx, reqQuery, queryFn)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func (k Keeper) ExecuteCosmosMsg(ctx sdk.Context, any *cdctypes.Any, owner sdk.AccAddress) ([]byte, error) {
	// sdk.Msg
	var msg sdk.Msg
	err := k.cdc.UnpackAny(any, &msg)
	if err != nil {
		return nil, err
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	signers := msg.GetSigners()
	if signers[0].String() != owner.String() {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("quasar cosmos message signer %s, expected %s", signers[0].String(), owner.String())
	}

	msgResponse, err := k.executeMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	// &sdk.MsgData{
	// 	MsgType: sdk.MsgTypeURL(msg),
	// 	Data:    msgResponse,
	// }
	return msgResponse, nil
}

// Attempts to get the message handler from the router and if found will then execute the message.
// If the message execution is successful, the proto marshaled message response will be returned.
func (k Keeper) executeMsg(ctx sdk.Context, msg sdk.Msg) ([]byte, error) {
	handler := k.msgRouter.Handler(msg)
	if handler == nil {
		return nil, types.ErrInvalidRoute
	}
	// handler can panic with out of gas or other errors
	res, err := safeHandler(ctx, msg, handler)
	if err != nil {
		return nil, err
	}

	// NOTE: The sdk msg handler creates a new EventManager, so events must be correctly propagated back to the current context
	ctx.EventManager().EmitEvents(res.GetEvents())

	return res.Data, nil
}

// func (k Keeper) newQueryHandler(ctx sdk.Context, contractAddress sdk.AccAddress) QueryHandler {
// 	return NewQueryHandler(ctx, k, contractAddress, k.gasRegister)
// }

func safeHandler(ctx sdk.Context, msg sdk.Msg, handler func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)) (res *sdk.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			// TODO maybe there is a better way to get the cosmos sdk errors
			// and make them EVM errors
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = sdkerrors.Wrapf(x, "failed to execute Cosmos message")
			default:
				// Fallback err
				err = sdkerrors.Wrapf(sdkerrors.ErrPanic, "unknown panic %v", r)

			}
			// invalidate rep
			res = nil
		}
	}()

	res, err = handler(ctx, msg)
	return res, err
}

func safeQuery(
	ctx sdk.Context,
	msg abci.RequestQuery,
	handler func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error),
) (res *abci.ResponseQuery, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = sdkerrors.Wrapf(x, "failed to execute Cosmos message")
			default:
				// Fallback err
				err = sdkerrors.Wrapf(sdkerrors.ErrPanic, "unknown panic %v", r)

			}
			// invalidate res
			res = nil
		}
	}()

	resp, err := handler(ctx, msg)
	return &resp, err
}
