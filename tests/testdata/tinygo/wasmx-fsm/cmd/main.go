package main

import (
	"encoding/json"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-fsm/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module wasmxcore
//export wasmx_env_core_i64_1
func Wasmx_env_core_i64_1() {}

//go:wasm-module wasmx-fsm
//export instantiate
func Instantiate() {}

func main() {
	entrypoint := os.Getenv("ENTRY_POINT")
	lib.LoggerDebugExtended("fsm: ENTRY_POINT: "+entrypoint, nil)
	switch entrypoint {
	case "eventual":
		eventual()
		return
	case "p2pmsg":
		p2pmsg()
		return
	case "StartNode":
		StartNode()
		return
	case "SetupNode":
		SetupNode()
		return
	case "instantiate":
		icalld, err := lib.GetCallDataWrap()
		if err != nil {
			lib.Revert("failed to get call data: " + err.Error())
			return
		}
		calldata := icalld.Calldata
		config := icalld.Config
		lib.Instantiate(config, calldata.Instantiate.InitialState, calldata.Instantiate.Context)
		wasmx.Finish([]byte{})
		return
	}

	icalld, err := lib.GetCallDataWrap()
	if err != nil {
		lib.Revert("failed to get call data: " + err.Error())
		return
	}

	calldata := icalld.Calldata
	config := icalld.Config
	var result []byte

	switch {
	case calldata.Run != nil:
		event := lib.EventObject{
			Type:   calldata.Run.Event.Type,
			Params: calldata.Run.Event.Params,
		}
		lib.RunInternal(config, event)
		result = wasmx.GetFinishData()

	// case calldata.Instantiate != nil:
	// 	calld := calldata.Instantiate
	// 	lib.Instantiate(config, calld.InitialState, calld.Context)
	// 	result = []byte{}

	case calldata.GetCurrentState != nil:
		state := lib.GetCurrentState()
		result = []byte(state.Value)

	case calldata.GetContextValue != nil:
		result = lib.GetContextValueInternal(calldata.GetContextValue.Key)

	case calldata.Setup != nil:
		lib.Setup(config, *calldata.Setup)
		result = []byte{}

	case calldata.StartNode != nil:
		startNodeInternal(config)
		result = []byte{}

	case calldata.SetupNode != nil:
		setupNodeInternal(config, calldata.SetupNode.Data)
		result = []byte{}

	case calldata.Execute != nil:
		action := calldata.Execute.Action
		event := lib.EventObject{Type: "", Params: []lib.ActionParam{}}
		lib.ExecuteInternal(config, event, action)
		result = wasmx.GetFinishData()

	default:
		calldraw := wasmx.GetCallData()
		calldstr := string(calldraw)
		lib.Revert("invalid function call data: " + calldstr)
		return
	}

	wasmx.Finish(result)
}

func eventual() {
	configBz, calldBz, err := lib.GetInterpreterCalldata()
	if err != nil {
		lib.Revert("failed to get interpreter calldata: " + err.Error())
		return
	}

	var config lib.MachineExternal
	if err := json.Unmarshal(configBz, &config); err != nil {
		lib.Revert("failed to parse config: " + err.Error())
		return
	}

	argsStr := string(calldBz)
	lib.LoggerDebugExtended("eventual: args: "+argsStr, []string{})

	var args lib.TimerArgs
	if err := json.Unmarshal(calldBz, &args); err != nil {
		lib.Revert("failed to parse timer args: " + err.Error())
		return
	}

	lib.Eventual(config, args)
}

func StartNode() {
	configBz, _, err := lib.GetInterpreterCalldata()
	if err != nil {
		lib.Revert("failed to get interpreter calldata: " + err.Error())
		return
	}

	var config lib.MachineExternal
	if err := json.Unmarshal(configBz, &config); err != nil {
		lib.Revert("failed to parse config: " + err.Error())
		return
	}

	startNodeInternal(config)

	// Handle finish data
	result := wasmx.GetFinishData()
	wasmx.Finish(result)
}

func SetupNode() {
	configBz, calldBz, err := lib.GetInterpreterCalldata()
	if err != nil {
		lib.Revert("failed to get interpreter calldata: " + err.Error())
		return
	}

	var config lib.MachineExternal
	if err := json.Unmarshal(configBz, &config); err != nil {
		lib.Revert("failed to parse config: " + err.Error())
		return
	}

	argsStr := string(calldBz)
	setupNodeInternal(config, argsStr)

	// Handle finish data
	result := wasmx.GetFinishData()
	wasmx.Finish(result)
}

func p2pmsg() {
	configBz, calldBz, err := lib.GetInterpreterCalldata()
	if err != nil {
		lib.Revert("failed to get interpreter calldata: " + err.Error())
		return
	}

	var config lib.MachineExternal
	if err := json.Unmarshal(configBz, &config); err != nil {
		lib.Revert("failed to parse config: " + err.Error())
		return
	}

	var p2pmsg lib.P2PMessage
	if err := json.Unmarshal(calldBz, &p2pmsg); err != nil {
		lib.Revert("failed to parse p2p message: " + err.Error())
		return
	}

	lib.LoggerDebug("p2pmsg", []string{
		"room", p2pmsg.RoomID,
		"sender_id", p2pmsg.Sender.ID,
		"sender_ip", p2pmsg.Sender.IP,
	})

	// Decode base64 message and parse as CallData
	var calldata lib.CallData
	if err := json.Unmarshal([]byte(p2pmsg.Message), &calldata); err != nil {
		lib.LoggerError("failed to parse call data from p2p message", []string{"data", string(p2pmsg.Message), "error", err.Error()})
		lib.Revert("failed to parse call data from p2p message: " + err.Error())
		return
	}

	if calldata.Run != nil {
		event := lib.EventObject{
			Type:   calldata.Run.Event.Type,
			Params: calldata.Run.Event.Params,
		}
		lib.RunInternal(config, event)

		// Handle finish data
		result := wasmx.GetFinishData()
		wasmx.Finish(result)
	}
}

// Helper functions

func setupNodeInternal(config lib.MachineExternal, data string) {
	lib.LoggerInfo("emit setupNode event", []string{"module", lib.MODULE_NAME})
	event := lib.EventObject{
		Type:   "setupNode",
		Params: []lib.ActionParam{{Key: "data", Value: data}},
	}
	lib.RunInternal(config, event)
}

func startNodeInternal(config lib.MachineExternal) {
	lib.LoggerInfo("emit start event", []string{"module", lib.MODULE_NAME})
	event := lib.EventObject{
		Type:   "start",
		Params: []lib.ActionParam{},
	}
	lib.RunInternal(config, event)
}
