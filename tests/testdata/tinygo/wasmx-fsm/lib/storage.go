package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// Storage keys
const (
	STORAGEKEY_STATE  = "state"
	STORAGEKEY_STATUS = "status"
	STORAGEKEY_CTX    = "context_"
	STORAGEKEY_OWNER  = "owner"
)

func sstore(key, value string) {
	wasmx.StorageStore([]byte(key), []byte(value))
}

func sload(key string) string {
	bz := wasmx.StorageLoad([]byte(key))
	return string(bz)
}

// Context parameter storage
func storeContextParams(params []ContextParam) {
	for _, param := range params {
		setContextValue(param.Key, param.Value)
	}
}

// Status management
func getCurrentStatus() InterpreterStatus {
	value := sload(STORAGEKEY_STATUS)
	if value == "" {
		return NotStarted
	}
	switch value {
	case "0":
		return NotStarted
	case "1":
		return Running
	case "2":
		return Stopped
	default:
		return NotStarted
	}
}

func setCurrentStatus(status InterpreterStatus) {
	var value string
	switch status {
	case NotStarted:
		value = "0"
	case Running:
		value = "1"
	case Stopped:
		value = "2"
	}
	sstore(STORAGEKEY_STATUS, value)
}

// State management
func getCurrentState() *State {
	valueStr := sload(STORAGEKEY_STATE)
	if valueStr == "" {
		return &State{
			Value:         "",
			Actions:       []ActionObject{},
			Changed:       false,
			PreviousValue: "",
		}
	}
	var state State
	if err := json.Unmarshal([]byte(valueStr), &state); err != nil {
		return &State{
			Value:         "",
			Actions:       []ActionObject{},
			Changed:       false,
			PreviousValue: "",
		}
	}
	return &state
}

func setCurrentState(state *State) {
	if state == nil {
		return
	}
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return
	}
	sstore(STORAGEKEY_STATE, string(stateBytes))
}

// Context value management
func hasContextValue(key string) bool {
	value := getContextValue(key)
	return value != ""
}

func getContextValue(key string) string {
	value := sload(STORAGEKEY_CTX + key)
	return value
}

func getContextValueInternal(key string) []byte {
	return wasmx.StorageLoad([]byte(STORAGEKEY_CTX + key))
}

func setContextValue(key, value string) {
	setContextValueInternal(key, []byte(value))
}

func setContextValueInternal(key string, value []byte) {
	wasmx.StorageStore([]byte(STORAGEKEY_CTX+key), value)
}

// Owner management
func storeOwner(owner []byte) {
	setContextValueInternal(STORAGEKEY_OWNER, owner)
}

func loadOwner() []byte {
	return getContextValueInternal(STORAGEKEY_OWNER)
}

// Public functions for external access
func GetCurrentState() State {
	state := getCurrentState()
	if state == nil {
		return State{
			Value:         "",
			Actions:       []ActionObject{},
			Changed:       false,
			PreviousValue: "",
		}
	}
	return *state
}

func GetContextValueInternal(key string) []byte {
	return getContextValueInternal(key)
}
