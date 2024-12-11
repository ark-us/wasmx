package types

import (
	"encoding/json"

	cw8types "wasmx/v1/x/wasmx/cw8/types"
)

//------- Results / Msgs -------------

type WasmxQueryResponse struct {
	Data []byte `json:"data"`
}

type WasmxExecutionMessage struct {
	Data []byte `json:"data"`
}

// ContractResponse defines the return value on a successful instantiate/execute/migrate.
type ContractResponse struct {
	// If the ReplyOn value matches the result, the runtime will invoke this
	// contract's `reply` entry point after execution. Otherwise, this is all
	// "fire and forget".
	Messages []cw8types.SubMsg `json:"messages"`

	// base64-encoded bytes to return as ABCI.Data field
	Data []byte `json:"data"`
	// attributes for a log event to return over abci interface
	Attributes []EventAttribute `json:"attributes"`
	// custom events (separate from the main one that contains the attributes
	// above)
	Events []Event `json:"events"`
	// only for debug
	MemorySnapshot []byte
	ErrorMessage   string
}

// Events must encode empty array as []
type Events []Event

// MarshalJSON ensures that we get [] for empty arrays
func (e Events) MarshalJSON() ([]byte, error) {
	if len(e) == 0 {
		return []byte("[]"), nil
	}
	var raw []Event = e
	return json.Marshal(raw)
}

// UnmarshalJSON ensures that we get [] for empty arrays
func (e *Events) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var raw []Event
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*e = raw
	return nil
}

type Event struct {
	Type       string          `json:"type"`
	Attributes EventAttributes `json:"attributes"`
}

// EventAttributes must encode empty array as []
type EventAttributes []EventAttribute

// MarshalJSON ensures that we get [] for empty arrays
func (a EventAttributes) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	var raw []EventAttribute = a
	return json.Marshal(raw)
}

// UnmarshalJSON ensures that we get [] for empty arrays
func (a *EventAttributes) UnmarshalJSON(data []byte) error {
	// make sure we deserialize [] back to null
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	var raw []EventAttribute
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*a = raw
	return nil
}

// EventAttribute
type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
