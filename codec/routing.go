package codec

import (
	"errors"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgRouter struct {
	Router *baseapp.MsgServiceRouter
}

type QueryRouter struct {
	Router *baseapp.GRPCQueryRouter
}

func (k QueryRouter) SubmitCosmosQuery(ctx sdk.Context, reqQuery *abci.RequestQuery) ([]byte, error) {
	// TODO if we allow historical queries, at a certain block
	// use app.Query(queryReq)
	queryFn := k.Router.Route(reqQuery.Path)
	res, err := safeQuery(ctx, reqQuery, queryFn)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

// Attempts to get the message handler from the router and if found will then execute the message.
// If the message execution is successful, the proto marshaled message response will be returned.
func (k MsgRouter) ExecuteCosmosMsg(ctx sdk.Context, msg sdk.Msg) ([]sdk.Event, []byte, error) {
	handler := k.Router.Handler(msg)
	if handler == nil {
		return nil, nil, fmt.Errorf("invalid route: %s", sdk.MsgTypeURL(msg))
	}
	// handler can panic with out of gas or other errors
	res, err := safeHandler(ctx, msg, handler)
	if err != nil {
		return nil, nil, err
	}

	return res.GetEvents(), res.Data, nil
}

func safeHandler(ctx sdk.Context, msg sdk.Msg, handler func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)) (res *sdk.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			// TODO maybe there is a better way to get the cosmos sdk errors
			// and make them EVM errors
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("failed to execute Cosmos message: %s", x)
			case error:
				err = sdkerr.Wrapf(x, "failed to execute Cosmos message")
			default:
				// Fallback err
				err = sdkerr.Wrapf(sdkerrors.ErrPanic, "unknown panic %v", r)
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
	msg *abci.RequestQuery,
	handler func(ctx sdk.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error),
) (res *abci.ResponseQuery, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = sdkerr.Wrapf(x, "failed to execute Cosmos message")
			default:
				// Fallback err
				err = sdkerr.Wrapf(sdkerrors.ErrPanic, "unknown panic %v", r)

			}
			// invalidate res
			res = nil
		}
	}()

	resp, err := handler(ctx, msg)
	return resp, err
}
