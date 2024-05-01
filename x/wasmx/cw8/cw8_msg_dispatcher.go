// https://github.com/CosmWasm/wasmd/blob/52a7a6ad2c00270867e1979eed1abb48c9ce04df/x/wasm/keeper/msg_dispatcher.go

package cw8

import (
	"fmt"
	"sort"
	"strings"

	address "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/cometbft/cometbft/abci/types"

	mcodec "mythos/v1/codec"
	"mythos/v1/x/wasmx/cw8/types"
)

// replyer is a subset of keeper that can handle replies to submessages
type replyer interface {
	Reply(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed, reply types.Reply) ([]byte, error)
	ExecuteCosmosMsg(ctx sdk.Context, msg sdk.Msg, owner mcodec.AccAddressPrefixed) ([]sdk.Event, []byte, error)
	Logger(ctx sdk.Context) log.Logger
	AddressCodec() address.Codec
}

// msgEncoder is an extension point to customize encodings
type msgEncoder interface {
	// Encode converts wasmvm message to n cosmos message types
	Encode(addrCodec address.Codec, ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg types.CosmosMsg) ([]sdk.Msg, error)
}

// MessageDispatcher coordinates message sending and submessage reply/ state commits
type MessageDispatcher struct {
	keeper   replyer
	encoders msgEncoder
}

// NewMessageDispatcher constructor
func NewMessageDispatcher(
	keeper replyer,
	unpacker codectypes.AnyUnpacker,
	portSource types.ICS20TransferPortSource,
) *MessageDispatcher {
	dispatcher := &MessageDispatcher{keeper: keeper}
	dispatcher.encoders = DefaultEncoders(unpacker, portSource)
	return dispatcher
}

// DispatchMessages sends all messages.
func (d MessageDispatcher) DispatchMsg(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, contractIBCPortID string, msg types.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	msgs, err := d.encoders.Encode(d.keeper.AddressCodec(), ctx, contractAddr.Bytes(), contractIBCPortID, msg)
	if err != nil {
		return nil, nil, err
	}
	for _, msg := range msgs {
		evts, res, err := d.keeper.ExecuteCosmosMsg(ctx, msg, contractAddr)
		if err != nil {
			return nil, nil, err
		}
		events = append(events, evts...)
		data = append(data, res)
	}
	return events, data, nil
}

// Handle processes the data returned by a contract invocation.
func (d MessageDispatcher) Handle(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, ibcPort string, messages []types.SubMsg, origRspData []byte) ([]byte, error) {
	result := origRspData
	switch rsp, err := d.DispatchSubmessages(ctx, contractAddr, ibcPort, messages); {
	case err != nil:
		return nil, errorsmod.Wrap(err, "submessages")
	case rsp != nil:
		result = rsp
	}
	return result, nil
}

// DispatchMessages sends all messages.
func (d MessageDispatcher) DispatchMessages(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, ibcPort string, msgs []types.CosmosMsg) error {
	for _, msg := range msgs {
		events, _, err := d.DispatchMsg(ctx, contractAddr, ibcPort, msg)
		if err != nil {
			return err
		}
		// redispatch all events, (type sdk.EventTypeMessage will be filtered out in the handler)
		ctx.EventManager().EmitEvents(events)
	}
	return nil
}

// dispatchMsgWithGasLimit sends a message with gas limit applied
func (d MessageDispatcher) dispatchMsgWithGasLimit(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, ibcPort string, msg types.CosmosMsg, gasLimit uint64) (events []sdk.Event, data [][]byte, err error) {
	limitedMeter := storetypes.NewGasMeter(gasLimit)
	subCtx := ctx.WithGasMeter(limitedMeter)

	// catch out of gas panic and just charge the entire gas limit
	defer func() {
		if r := recover(); r != nil {
			// if it's not an OutOfGas error, raise it again
			if _, ok := r.(storetypes.ErrorOutOfGas); !ok {
				// TODO
				// log it to get the original stack trace somewhere (as panic(r) keeps message but stacktrace to here
				d.keeper.Logger(ctx).Info("SubMsg rethrowing panic: %#v", r)
				panic(r)
			}
			ctx.GasMeter().ConsumeGas(gasLimit, "Sub-Message OutOfGas panic")
			err = errorsmod.Wrap(sdkerrors.ErrOutOfGas, "SubMsg hit gas limit")
		}
	}()
	events, data, err = d.DispatchMsg(subCtx, contractAddr, ibcPort, msg)

	// make sure we charge the parent what was spent
	spent := subCtx.GasMeter().GasConsumed()
	ctx.GasMeter().ConsumeGas(spent, "From limited Sub-Message")

	return events, data, err
}

// DispatchSubmessages builds a sandbox to execute these messages and returns the execution result to the contract
// that dispatched them, both on success as well as failure
func (d MessageDispatcher) DispatchSubmessages(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, ibcPort string, msgs []types.SubMsg) ([]byte, error) {
	var rsp []byte
	for _, msg := range msgs {
		switch msg.ReplyOn {
		case types.ReplySuccess, types.ReplyError, types.ReplyAlways, types.ReplyNever:
		default:
			return nil, errorsmod.Wrap(errorsmod.Error{}, "invalid replyOn value")
		}
		// first, we build a sub-context which we can use inside the submessages
		subCtx, commit := ctx.CacheContext()
		em := sdk.NewEventManager()
		subCtx = subCtx.WithEventManager(em)

		// check how much gas left locally, optionally wrap the gas meter
		gasRemaining := ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
		limitGas := msg.GasLimit != nil && (*msg.GasLimit < gasRemaining)

		var err error
		var events []sdk.Event
		var data [][]byte
		if limitGas {
			events, data, err = d.dispatchMsgWithGasLimit(subCtx, contractAddr, ibcPort, msg.Msg, *msg.GasLimit)
		} else {
			events, data, err = d.DispatchMsg(subCtx, contractAddr, ibcPort, msg.Msg)
		}

		// if it succeeds, commit state changes from submessage, and pass on events to Event Manager
		var filteredEvents []sdk.Event
		if err == nil {
			commit()
			filteredEvents = filterEvents(append(em.Events(), events...))
			ctx.EventManager().EmitEvents(filteredEvents)
			if msg.Msg.Wasm == nil {
				filteredEvents = []sdk.Event{}
			} else {
				for _, e := range filteredEvents {
					attributes := e.Attributes
					sort.SliceStable(attributes, func(i, j int) bool {
						return strings.Compare(string(attributes[i].Key), string(attributes[j].Key)) < 0
					})
				}
			}
		} // on failure, revert state from sandbox, and ignore events (just skip doing the above)

		// we only callback if requested. Short-circuit here the cases we don't want to
		if (msg.ReplyOn == types.ReplySuccess || msg.ReplyOn == types.ReplyNever) && err != nil {
			return nil, err
		}
		if msg.ReplyOn == types.ReplyNever || (msg.ReplyOn == types.ReplyError && err == nil) {
			continue
		}

		// otherwise, we create a SubMsgResult and pass it into the calling contract
		var result types.SubMsgResult
		if err == nil {
			// just take the first one for now if there are multiple sub-sdk messages
			// and safely return nothing if no data
			var responseData []byte
			if len(data) > 0 {
				responseData = data[0]
			}
			result = types.SubMsgResult{
				Ok: &types.SubMsgResponse{
					Events: sdkEventsToWasmVMEvents(filteredEvents),
					Data:   responseData,
				},
			}
		} else {
			// TODO
			// Issue #759 - we don't return error string for worries of non-determinism
			d.keeper.Logger(ctx).Info("Redacting submessage error", "cause", err)
			result = types.SubMsgResult{
				Err: redactError(err).Error(),
			}
			result = types.SubMsgResult{
				Err: err.Error(),
			}
		}

		// now handle the reply, we use the parent context, and abort on error
		reply := types.Reply{
			ID:     msg.ID,
			Result: result,
		}

		// we can ignore any result returned as there is nothing to do with the data
		// and the events are already in the ctx.EventManager()
		rspData, err := d.keeper.Reply(ctx, contractAddr, reply)
		switch {
		case err != nil:
			return nil, errorsmod.Wrap(err, "reply")
		case rspData != nil:
			rsp = rspData
		}
	}
	return rsp, nil
}

// TODO
// Issue #759 - we don't return error string for worries of non-determinism
func redactError(err error) error {
	// Do not redact system errors
	// SystemErrors must be created in x/wasm and we can ensure determinism
	if types.ToSystemError(err) != nil {
		return err
	}

	// FIXME: do we want to hardcode some constant string mappings here as well?
	// Or better document them? (SDK error string may change on a patch release to fix wording)
	// sdk/11 is out of gas
	// sdk/5 is insufficient funds (on bank send)
	// (we can theoretically redact less in the future, but this is a first step to safety)
	codespace, code, _ := errorsmod.ABCIInfo(err, false)
	return fmt.Errorf("codespace: %s, code: %d", codespace, code)
}

func filterEvents(events []sdk.Event) []sdk.Event {
	// pre-allocate space for efficiency
	res := make([]sdk.Event, 0, len(events))
	for _, ev := range events {
		if ev.Type != "message" {
			res = append(res, ev)
		}
	}
	return res
}

func sdkEventsToWasmVMEvents(events []sdk.Event) []types.Event {
	res := make([]types.Event, len(events))
	for i, ev := range events {
		res[i] = types.Event{
			Type:       ev.Type,
			Attributes: sdkAttributesToWasmVMAttributes(ev.Attributes),
		}
	}
	return res
}

func sdkAttributesToWasmVMAttributes(attrs []abci.EventAttribute) []types.EventAttribute {
	res := make([]types.EventAttribute, len(attrs))
	for i, attr := range attrs {
		res[i] = types.EventAttribute{
			Key:   string(attr.Key),
			Value: string(attr.Value),
		}
	}
	return res
}
