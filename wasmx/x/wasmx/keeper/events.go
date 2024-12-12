package keeper

import (
	"fmt"
	"strings"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// newWasmModuleEvent creates with wasm module event for interacting with the given contract. Adds custom attributes
// to this event.
func newWasmModuleEvent(customAttributes []types.EventAttribute, contractAddrBech32 string) (sdk.Events, error) {
	attrs, err := contractSDKEventAttributes(customAttributes, contractAddrBech32)
	if err != nil {
		return nil, err
	}

	// each wasm invocation always returns one sdk.Event
	return sdk.Events{sdk.NewEvent(types.WasmModuleEventType, attrs...)}, nil
}

const eventTypeMinLength = 2

// Keep compatible with cosmwasm events
// newCustomEvents converts wasmvm events from a contract response to sdk type events
func newCustomEvents(evts types.Events, contractAddrBech32 string) (sdk.Events, error) {
	events := make(sdk.Events, 0, len(evts))
	for _, e := range evts {
		typ := strings.TrimSpace(e.Type)
		if len(typ) <= eventTypeMinLength {
			return nil, sdkerr.Wrap(types.ErrInvalidEvent, fmt.Sprintf("Event type too short: '%s'", typ))
		}
		// also adds contract address as attribute
		attributes, err := contractSDKEventAttributes(e.Attributes, contractAddrBech32)
		if err != nil {
			return nil, err
		}
		// add wasmx module as attribute
		attributes = append(attributes, sdk.NewAttribute("module", types.WasmModuleEventType))

		events = append(events, sdk.NewEvent(fmt.Sprintf("%s%s", types.CustomContractEventPrefix, typ), attributes...))
	}
	return events, nil
}

// convert and add contract address issuing this event
func contractSDKEventAttributes(customAttributes []types.EventAttribute, contractAddrBech32 string) ([]sdk.Attribute, error) {
	attrs := []sdk.Attribute{sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddrBech32)}
	// append attributes from wasm to the sdk.Event
	for _, l := range customAttributes {
		// ensure key and value are non-empty (and trim what is there)
		key := strings.TrimSpace(l.Key)
		if len(key) == 0 {
			return nil, sdkerr.Wrap(types.ErrInvalidEvent, fmt.Sprintf("Empty attribute key. Value: %s", l.Value))
		}
		value := strings.TrimSpace(l.Value)
		if len(value) == 0 {
			return nil, sdkerr.Wrap(types.ErrInvalidEvent, fmt.Sprintf("Empty attribute value. Key: %s", key))
		}
		// and reserve all _* keys for our use (not contract)
		if strings.HasPrefix(key, types.AttributeReservedPrefix) {
			return nil, sdkerr.Wrap(types.ErrInvalidEvent, fmt.Sprintf("Attribute key starts with reserved prefix %s: '%s'", types.AttributeReservedPrefix, key))
		}
		attrs = append(attrs, sdk.NewAttribute(key, value))
	}
	return attrs, nil
}
