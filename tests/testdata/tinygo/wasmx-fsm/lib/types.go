package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// InterpreterStatus represents the status of the FSM interpreter
type InterpreterStatus uint8

const (
	NotStarted InterpreterStatus = iota
	Running
	Stopped
)

var StatusMap = map[string]InterpreterStatus{
	"0": NotStarted,
	"1": Running,
	"2": Stopped,
}

// Global variables for INIT_EVENT
var INIT_EVENT = EventObject{
	Type:   INIT_EVENT_TYPE,
	Params: []ActionParam{},
}

// GuardObject represents a guard condition
type GuardObject struct {
	Type   string        `json:"type"`
	Params []ActionParam `json:"params"`
}

// ActionParam represents a key-value parameter for actions and events
type ActionParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EventObject represents an event in the state machine
type EventObject struct {
	Type   string        `json:"type"`
	Params []ActionParam `json:"params"`
}

// ActionObject represents an action to be executed
type ActionObject struct {
	Type   string        `json:"type"`
	Params []ActionParam `json:"params"`
	Event  *EventObject  `json:"event,omitempty"`
}

// Transition represents a transition between states
type Transition struct {
	Target  string         `json:"target"`
	Actions []ActionObject `json:"actions"`
	Guard   *ActionObject  `json:"guard,omitempty"`
	Meta    []ActionParam  `json:"meta,omitempty"`
}

// MachineExternal represents external machine format
type MachineExternal struct {
	ID      string                   `json:"id"`
	Library wasmx.Bech32String       `json:"library"`
	States  []StateInfoClassExternal `json:"states"`
	ctx     map[string]string
}

// ServiceExternal represents the external service format
type ServiceExternal struct {
	Machine MachineExternal   `json:"machine"`
	Status  InterpreterStatus `json:"status"`
}

// State represents the current state of the machine
type State struct {
	Value         string         `json:"value"`
	Actions       []ActionObject `json:"actions"`
	Changed       bool           `json:"changed"`
	PreviousValue string         `json:"previousValue"`
}

// TimerArgs represents timer arguments
type TimerArgs struct {
	Delay      string `json:"delay"`
	State      string `json:"state"`
	IntervalID int64  `json:"intervalId"`
}

// ContextParam represents a context parameter
type ContextParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExternalActionCallData represents external action call data
type ExternalActionCallData struct {
	Method string        `json:"method"`
	Params []ActionParam `json:"params"`
	Event  EventObject   `json:"event"`
}

// RaiseAction represents the xstate.raise action
type RaiseAction struct {
	Type   string        `json:"type"`
	Params []ActionParam `json:"params"`
	Event  EventObject   `json:"event"`
}

// StateClassExternal represents the external state class
type StateClassExternal struct {
	Value         string         `json:"value"`
	Actions       []ActionObject `json:"actions"`
	Changed       bool           `json:"changed"`
	PreviousValue string         `json:"previousValue"`
}

// AssignActionObject represents an assign action
type AssignActionObject struct {
	Type       string            `json:"type"`
	Params     []ActionParam     `json:"params"`
	Event      *EventObject      `json:"event,omitempty"`
	Assignment map[string]string `json:"assignment,omitempty"`
}

// TransitionExternal represents external transition format
type TransitionExternal struct {
	Name        string       `json:"name"`
	Transitions []Transition `json:"transitions"`
}

// StateInfoClassExternal represents external state info format
type StateInfoClassExternal struct {
	Name    string                   `json:"name"`
	Always  []TransitionExternal     `json:"always"`
	After   []TransitionExternal     `json:"after"`
	On      []TransitionExternal     `json:"on"`
	Exit    []ActionObject           `json:"exit"`
	Entry   []ActionObject           `json:"entry"`
	Initial string                   `json:"initial"`
	States  []StateInfoClassExternal `json:"states"`
}

// ContextParamRaw represents raw context parameter with byte arrays
type ContextParamRaw struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

// HandledActions represents handled actions with assignment flag
type HandledActions struct {
	Actions  []ActionObject `json:"actions"`
	Assigned bool           `json:"assigned"`
}

// CallDataGetContext represents get context value call data
type CallDataGetContext struct {
	Key string `json:"key"`
}

// CallDataSetup represents setup call data
type CallDataSetup struct {
	ContractAddress string `json:"contractAddress"`
}

// CallDataSetupNode represents setup node call data
type CallDataSetupNode struct {
	Data string `json:"data"`
}

// CallDataExecute represents execute call data
type CallDataExecute struct {
	Action ActionObject `json:"action"`
}

// Helper functions for JSON serialization
func (s *State) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *State) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// StateClassExternal conversion methods
func (s *StateClassExternal) ToInternal() *State {
	return &State{
		Value:         s.Value,
		Actions:       s.Actions,
		Changed:       s.Changed,
		PreviousValue: s.PreviousValue,
	}
}

func StateClassExternalFromInternal(state *State) *StateClassExternal {
	return &StateClassExternal{
		Value:         state.Value,
		Actions:       state.Actions,
		Changed:       state.Changed,
		PreviousValue: state.PreviousValue,
	}
}

// Types for P2P messaging

type NetworkNode struct {
	ID   string `json:"id"` // p2p id
	Host string `json:"host"`
	Port string `json:"port"`
	IP   string `json:"ip"` // can be empty if host & port are used
}

type P2PMessage struct {
	RoomID    string      `json:"roomId"`
	Message   []byte      `json:"message"`
	Timestamp string      `json:"timestamp"` // simplified as string
	Sender    NetworkNode `json:"sender"`
}
