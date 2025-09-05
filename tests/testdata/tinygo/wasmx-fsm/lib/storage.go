package lib

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// Storage keys
const (
	STORAGEKEY_STATE  = "state"
	STORAGEKEY_STATUS = "status"
	STORAGEKEY_CTX    = "context_"
	STORAGEKEY_OWNER  = "owner"
)

// Context parameter storage
func storeContextParams(params []ContextParam) {
	for _, param := range params {
		setContextValue(param.Key, param.Value)
	}
}

// Status management
func getCurrentStatus() InterpreterStatus {
	value := string(wasmx.StorageLoad([]byte(STORAGEKEY_STATUS)))
	if value == "" {
		return NotStarted
	}
	status, ok := StatusMap[value]
	if ok {
		return status
	}
	return NotStarted
}

func setCurrentStatus(status InterpreterStatus) {
	value := fmt.Sprintf("%d", status)
	wasmx.StorageStore([]byte(STORAGEKEY_STATUS), []byte(value))
}

// State management
func getCurrentState() *State {
	valuebz := wasmx.StorageLoad([]byte(STORAGEKEY_STATE))
	if len(valuebz) == 0 {
		return &State{
			Value:         "",
			Actions:       []ActionObject{},
			Changed:       false,
			PreviousValue: "",
		}
	}
	var state State
	if err := json.Unmarshal(valuebz, &state); err != nil {
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
	wasmx.StorageStore([]byte(STORAGEKEY_STATE), stateBytes)
}

// Context value management
func hasContextValue(key string) bool {
	value := getContextValue(key)
	return value != ""
}

func getContextValue(key string) string {
	return string(getContextValueInternal(key))
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
