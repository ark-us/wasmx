package lib

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// Instantiate initializes the state machine
func Instantiate(config MachineExternal, initialState string, params []ContextParam) {
	storeContextParams(params)
	setCurrentStatus(NotStarted)
	storeOwner(wasmx.GetCallerBz())

	// TODO see if there are initial actions to perform (e.g. entry actions)
	iniactions := []ActionObject{}
	res := HandleActions(iniactions, INIT_EVENT)
	initialActions := res.Actions

	if !strings.HasPrefix(initialState, "#") {
		initialState = "#" + config.ID + "." + initialState
	}

	newstate := State{
		Value:         initialState,
		Actions:       initialActions,
		Changed:       false,
		PreviousValue: "",
	}
	setCurrentState(&newstate)

	service := ServiceExternal{
		Machine: config,
		Status:  NotStarted,
	}

	service.Start()
}

// CreateUnchangedState creates a state with no changes
func CreateUnchangedState(value, prevState string) State {
	return State{
		Value:         value,
		Actions:       []ActionObject{},
		Changed:       false,
		PreviousValue: prevState,
	}
}

// HandleActions processes actions and returns handled actions
func HandleActions(actions []ActionObject, eventObject EventObject) HandledActions {
	assigned := false
	nonAssignActions := []ActionObject{}

	for _, action := range actions {
		if action.Type == ASSIGN_ACTION {
			assigned = true
			continue
		}
		nonAssignActions = append(nonAssignActions, action)
	}

	return HandledActions{
		Actions:  nonAssignActions,
		Assigned: assigned,
	}
}

// ExecuteGuard executes guard conditions
func ExecuteGuard(machine *MachineExternal, guard *ActionObject, event EventObject) bool {
	if guard == nil {
		return true
	}

	switch guard.Type {
	case "isAdmin":
		return IsAdmin([]ActionParam{})
	case "ifIntervalActive":
		return IfIntervalActive([]ActionParam{}, event)
	default:
		// External guard function
		guardParams := append(guard.Params, ctxToActionParams(machine.ctx)...)
		resp, err := ProcessExternalCall(machine, guard.Type, guardParams, event)
		logged := []string{"guard", guard.Type}
		if err == nil {
			logged = append(logged, []string{"success", "true"}...)
		} else {
			logged = append(logged, []string{"success", "false", "error", err.Error()}...)
		}
		LoggerDebug("execute guard call", logged)
		if err != nil {
			return false
		}
		return string(resp) == "1"
	}
}

// ExecuteStateActions executes all actions for a state
func ExecuteStateActions(service *ServiceExternal, state State, event EventObject) {
	LoggerDebugExtended("execute state actions", []string{
		"count", strconv.Itoa(len(state.Actions)),
	})

	for _, action := range state.Actions {
		ExecuteStateAction(service, state, event, action)
	}

	if !state.Changed {
		return
	}

	// Timed actions from the new target state
	newstateconfig := FindStateInfo(service.Machine.States, state.Value)
	if newstateconfig == nil {
		Revert("could not find state config for " + state.Value)
		return
	}

	// Run the current state after transitions
	if newstateconfig.After != nil {
		delayKeys := make([]string, 0, len(newstateconfig.After))
		for _, t := range newstateconfig.After {
			delayKeys = append(delayKeys, t.Name)
		}
		RunAfterTransitions(state.Value, delayKeys)
	}

	// Run the current state always
	if len(newstateconfig.Always) > 0 {
		transitions := make([]Transition, 0)
		for _, t := range newstateconfig.Always {
			transitions = append(transitions, t.Transitions...)
		}
		newstate := ApplyTransitions(&service.Machine, state, transitions, EventObject{}, 0)
		if newstate != nil {
			setCurrentState(newstate)
			ExecuteStateActions(service, *newstate, EventObject{})
		}
	}
}

// RunAfterTransitions sets up timed transitions
func RunAfterTransitions(statePath string, delayKeys []string) {
	LoggerDebugExtended("setting timed actions", []string{
		"state", statePath,
		"delays", strings.Join(delayKeys, ","),
	})

	for _, delayKey := range delayKeys {
		var delay uint64
		delayInt, err := strconv.ParseInt(delayKey, 10, 64)
		if err != nil {
			// This should be a context variable
			value := getContextValue(delayKey)
			if value == "" {
				Revert("delay key not found in context: " + delayKey)
				return
			}
			delayInt, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				Revert("invalid delay value: " + value)
				return
			}
		}
		delay = uint64(delayInt)

		intervalId := GetLastIntervalId() + 1
		SetLastIntervalId(intervalId)
		RegisterIntervalId(statePath, delayKey, intervalId)

		contractAddress := wasmx.Bech32String("")
		if contractAddress == "" {
			contractAddress = wasmx.GetAddress()
		}

		args := TimerArgs{
			Delay:      delayKey,
			State:      statePath,
			IntervalID: intervalId,
		}
		argsBytes, _ := json.Marshal(args)
		LoggerDebug("starting timeout", []string{
			"intervalId", strconv.FormatInt(intervalId, 10),
			"delay_key", delayKey,
			"delay", strconv.FormatUint(delay, 10),
		})

		err = wasmxcore.StartTimeout(strconv.FormatInt(intervalId, 10), contractAddress, int64(delay), argsBytes)
		if err != nil {
			LoggerError("failed to start timeout", []string{"error", err.Error()})
		}
	}
}

// ApplyTransitions applies a list of transitions
func ApplyTransitions(machine *MachineExternal, state State, transitions []Transition, event EventObject, ifElse int32) *State {
	if len(transitions) == 0 {
		return nil
	}
	if len(transitions) > 1 && ifElse == 0 {
		ifElse = int32(len(transitions))
	}
	if ifElse > 1 {
		if len(transitions) == int(ifElse) {
			LoggerDebug("apply if", []string{"target", transitions[0].Target})
		} else {
			LoggerDebug("apply else", []string{"target", transitions[0].Target})
		}
	}

	newstate := machine.ApplyTransition(state, transitions[0], event)
	if newstate != nil {
		return newstate
	}

	if len(transitions) > 1 {
		return ApplyTransitions(machine, state, transitions[1:], event, ifElse)
	}
	return nil
}

// ExecuteStateAction executes a single action
func ExecuteStateAction(service *ServiceExternal, state State, event EventObject, action ActionObject) {
	actionType := action.Type
	LoggerDebug("execute action", []string{"action", actionType})

	if actionType == RAISE_ACTION {
		if action.Event == nil {
			Revert("raise action is missing event")
			return
		}
		_event := EventObject{
			Type:   action.Event.Type,
			Params: action.Event.Params,
		}
		service.Send(_event)
		return
	}

	actionParams := action.Params
	for i := range actionParams {
		if len(actionParams[i].Value) > 0 && strings.HasPrefix(actionParams[i].Value, VARIABLE_SYMBOL) {
			varname := actionParams[i].Value[1:]
			value := ""

			// First look into event parameters
			for _, param := range event.Params {
				if param.Key == varname {
					value = param.Value
					break
				}
			}

			// Then look into the temporary context
			if value == "" {
				if val, exists := service.Machine.ctx[varname]; exists {
					value = val
				}
			}

			// Then look into storage
			if value == "" {
				value = getContextValue(varname)
			}
			actionParams[i].Value = value
		}
	}

	switch actionType {
	case "assign":
		Assign(&service.Machine, actionParams, event)
	case "sendRequest":
		SendRequest(actionParams, event)
	case "log":
		Log(actionParams)
	case "noaction":
		Noaction(actionParams, event)
	case "cancelActiveIntervals":
		CancelActiveIntervals(state, actionParams, event)
	default:
		// External action
		resp, err := ProcessExternalCall(&service.Machine, action.Type, actionParams, event)
		if err != nil {
			Revert("action failed: " + actionType + "; err=" + err.Error())
			return
		}
		if len(resp) > 0 {
			wasmx.SetFinishData(resp)
		}
	}
}

// Noaction does nothing
func Noaction(params []ActionParam, event EventObject) {}

// Assign assigns values to machine context
func Assign(machine *MachineExternal, params []ActionParam, event EventObject) {
	for _, param := range params {
		machine.ctx[param.Key] = param.Value
	}
}

// ProcessExternalCall processes external contract calls
func ProcessExternalCall(machine *MachineExternal, actionType string, params []ActionParam, event EventObject) ([]byte, error) {
	var contractAddress wasmx.Bech32String

	// Actions can have `label.function`
	if strings.Contains(actionType, ".") {
		parts := strings.Split(actionType, ".")
		if len(parts) < 2 {
			return nil, fmt.Errorf("cannot find contract address by label")
		}
		contractAddress = wasmx.Bech32String(getContextValue(parts[0]))
		actionType = parts[1]
	} else {
		contractAddress = machine.Library
	}

	if contractAddress == "" {
		return nil, fmt.Errorf("empty contract address")
	}

	calldata := ExternalActionCallData{
		Method: actionType,
		Params: params,
		Event:  event,
	}

	calldataBytes, err := json.Marshal(calldata)
	if err != nil {
		return nil, err
	}
	success, resp := wasmx.Call(contractAddress, nil, calldataBytes, big.NewInt(wasmx.DEFAULT_GAS_TX), MODULE_NAME)
	if !success {
		return nil, fmt.Errorf("external call errored: %s", string(resp))
	}
	return resp, nil
}

// IsAdmin checks if caller is admin
func IsAdmin(params []ActionParam) bool {
	caller := wasmx.GetCallerBz()
	owner := loadOwner()
	return getAddressHex(caller) == getAddressHex(owner)
}

// Log logs key-value pairs
func Log(params []ActionParam) {
	for _, param := range params {
		key := param.Key
		message := param.Value
		value := getContextValue(key)
		LoggerDebug(message, []string{"key", key, "value", value})
	}
}

// SendRequest sends a GRPC request
func SendRequest(params []ActionParam, event EventObject) {
	var address, data string
	for _, param := range event.Params {
		switch param.Key {
		case "address":
			address = param.Value
		case "data":
			data = param.Value
		}
	}

	if address == "" {
		Revert("sendRequest empty IP address")
		return
	}
	if data == "" {
		Revert("sendRequest empty data")
		return
	}

	contract := wasmx.GetCaller()
	_, err := wasmxcore.GrpcRequest(address, contract, data)
	if err != nil {
		Revert("grpcRequest failed: " + err.Error())
	}
}

// Send sends an event to the state machine
func (s *ServiceExternal) Send(event EventObject) {
	LoggerDebugExtended("new event", []string{
		"event", event.Type,
		"status", strconv.Itoa(int(s.Status)),
	})

	if s.Status != Running {
		return
	}

	state := getCurrentState()
	LoggerDebugExtended("transition event", []string{
		"event", event.Type,
		"status", strconv.Itoa(int(s.Status)),
		"state", state.Value,
	})

	newstate := s.Machine.Transition(*state, event)
	if newstate == nil {
		return
	}

	LoggerDebug("transition event applied", []string{
		"event", event.Type,
		"status", strconv.Itoa(int(s.Status)),
		"state", state.Value,
		"next_state", newstate.Value,
	})

	setCurrentState(newstate)
	ExecuteStateActions(s, *newstate, event)
}

// Start starts the service
func (s *ServiceExternal) Start() *ServiceExternal {
	s.Status = Running
	setCurrentStatus(Running)
	state := getCurrentState()
	ExecuteStateActions(s, *state, INIT_EVENT)
	s.Send(INIT_EVENT)
	return s
}

// Stop stops the service
func (s *ServiceExternal) Stop() *ServiceExternal {
	s.Status = Stopped
	setCurrentStatus(Stopped)
	return s
}

// Transition applies a transition to the current state
func (m *MachineExternal) Transition(state State, eventObject EventObject) *State {
	value := state.Value
	stateConfig := FindStateInfo(m.States, value)
	if stateConfig == nil {
		message := "state not found: " + value
		wasmx.Revert([]byte(message))
		return nil
	}

	var transitions []Transition
	if stateConfig.On != nil {
		onState := getTransitionByName(stateConfig.On, eventObject.Type)
		if onState != nil {
			transitions = onState.Transitions
		} else {
			// Search for transition in the parents
			transitions = findTransitionInParents(m.States, value, eventObject)
			if transitions == nil {
				message := `cannot apply "` + eventObject.Type + `" event in current "` + value + `" state`
				if REVERT_IF_UNEXPECTED_STATE {
					wasmx.Revert([]byte(message))
					return nil
				} else {
					LoggerDebug("cannot apply event in current state", []string{
						"event", eventObject.Type,
						"state", value,
					})
					return nil
				}
			}
		}
	}

	if transitions != nil {
		return ApplyTransitions(m, state, transitions, eventObject, 0)
	}

	// No transitions match
	unchanged := CreateUnchangedState(value, state.PreviousValue)
	return &unchanged
}

// ApplyTransition applies a single transition
func (m *MachineExternal) ApplyTransition(state State, transition Transition, eventObject EventObject) *State {
	LoggerDebug("apply transition: ", []string{
		"from", state.Value,
		"to", transition.Target,
		"event", eventObject.Type,
	})

	value := state.Value
	stateConfig := FindStateInfo(m.States, value)
	if stateConfig == nil {
		message := "state not found: " + value
		wasmx.Revert([]byte(message))
		return nil
	}

	transitions := []Transition{transition}

	// Check for wildcard transitions
	wildcardState := getTransitionByName(stateConfig.On, WILDCARD)
	if wildcardState != nil {
		return ApplyTransitions(m, state, wildcardState.Transitions, eventObject, 0)
	}

	for _, transition := range transitions {
		target := transition.Target
		actions := transition.Actions
		guard := transition.Guard

		isTargetless := target == ""
		nextStateValue := target
		if isTargetless {
			nextStateValue = value
		}

		nextStateConfig := FindStateInfo(m.States, nextStateValue)
		if nextStateConfig == nil {
			message := "state not found: " + nextStateValue
			wasmx.Revert([]byte(message))
			return nil
		}

		if guard != nil && !ExecuteGuard(m, guard, eventObject) {
			message := "cannot execute transition; guard: " + guard.Type
			if REVERT_IF_UNEXPECTED_STATE {
				wasmx.Revert([]byte(message))
				return nil
			} else {
				LoggerDebug(message, []string{})
				return nil
			}
		}

		var allActions []ActionObject

		// State exit actions
		if !isTargetless {
			allActions = append(allActions, processActions(stateConfig.Exit, eventObject)...)
		}
		// Event actions
		allActions = append(allActions, processActions(actions, eventObject)...)
		// State entry actions
		if !isTargetless {
			allActions = append(allActions, processActions(nextStateConfig.Entry, eventObject)...)
		}

		resolvedTarget := target
		if isTargetless {
			resolvedTarget = value
		}

		// Handle child states; we choose the first one
		stateConfigResolved := FindStateInfo(m.States, resolvedTarget)
		if stateConfigResolved == nil {
			message := "state not found: " + resolvedTarget
			wasmx.Revert([]byte(message))
			return nil
		}

		if len(stateConfigResolved.States) > 0 {
			// State has children - choose the first one
			initialState := stateConfigResolved.States[0].Name
			initialStateObj := getStateByName(stateConfigResolved.States, initialState)
			if initialStateObj == nil {
				Revert(`states must have children here`)
			}

			if !strings.HasPrefix(initialState, "#") {
				initialState = resolvedTarget + "." + initialState
			}
			resolvedTarget = initialState

			// Add the action of this child state
			allActions = append(allActions, processActions(initialStateObj.Entry, eventObject)...)

			// Run any "after" transitions on the parent
			if nextStateConfig.After != nil {
				delayKeys := make([]string, 0, len(nextStateConfig.After))
				for _, p := range nextStateConfig.After {
					delayKeys = append(delayKeys, p.Name)
				}
				RunAfterTransitions(target, delayKeys)
			}
		}

		LoggerDebug("apply transition: ", []string{
			"from", state.Value,
			"final_target", resolvedTarget,
			"actions_count", strconv.Itoa(len(allActions)),
		})

		res := HandleActions(allActions, eventObject)
		nonAssignActions := res.Actions

		return &State{
			Value:         resolvedTarget,
			Actions:       nonAssignActions,
			Changed:       !isTargetless,
			PreviousValue: state.Value,
		}
	}

	// No transitions match
	unchanged := CreateUnchangedState(value, state.PreviousValue)
	return &unchanged
}

// Helper functions

func processActions(actions []ActionObject, event EventObject) []ActionObject {
	var allActions []ActionObject
	for _, act := range actions {
		if act.Type == RAISE_ACTION {
			if act.Event == nil {
				message := "raise action is missing event"
				wasmx.Revert([]byte(message))
				return allActions
			}

			// Look through parameters and update values
			for i := range act.Event.Params {
				// key := act.Event.Params[i].Key
				value := act.Event.Params[i].Value
				found := false

				if strings.Contains(value, "()") {
					if value == "getCaller()" {
						act.Event.Params[i].Value = getAddressHex(wasmx.GetCallerBz())
						found = true
					}
				} else {
					for _, param := range event.Params {
						if param.Key == value {
							act.Event.Params[i].Value = param.Value
							found = true
							break
						}
					}
					if !found {
						v := getContextValue(value)
						if v != "" {
							act.Event.Params[i].Value = v
						}
					}
				}
			}
		}
		allActions = append(allActions, act)
	}
	return allActions
}

func findTransitionInParents(states []StateInfoClassExternal, stateName string, eventObject EventObject) []Transition {
	if !strings.HasPrefix(stateName, "#") {
		Revert("state must be absolute: " + stateName)
		return nil
	}

	statePath := strings.Split(stateName[1:], ".")[1:]
	if len(statePath) > 0 {
		statePath = statePath[:len(statePath)-1] // remove current kid
	}
	return findTransitionInternal(states, statePath, eventObject)
}

func findTransitionInternal(states []StateInfoClassExternal, statePath []string, eventObject EventObject) []Transition {
	stateConfig := FindStateInfoByPath(states, statePath)
	if stateConfig != nil && stateConfig.On != nil {
		onTransition := getTransitionByName(stateConfig.On, eventObject.Type)
		if onTransition != nil {
			return onTransition.Transitions
		}
	}
	if len(statePath) == 0 {
		return nil
	}
	statePath = statePath[:len(statePath)-1] // remove current kid
	return findTransitionInternal(states, statePath, eventObject)
}

// FindStateInfo finds state info by name
func FindStateInfo(states []StateInfoClassExternal, stateName string) *StateInfoClassExternal {
	if strings.HasPrefix(stateName, "#") {
		// "#ERC20.unlocked.active"
		// TODO support multiple machines
		statePath := strings.Split(stateName[1:], ".")[1:]
		return FindStateInfoByPath(states, statePath)
	}

	// Look in current states
	for _, stateinfo := range states {
		name := stateinfo.Name
		if name == stateName {
			return &stateinfo
		}
		if stateinfo.States != nil {
			if info := FindStateInfo(stateinfo.States, stateName); info != nil {
				return info
			}
		}
	}
	return nil
}

// FindStateInfoByPath finds state info by path
func FindStateInfoByPath(states []StateInfoClassExternal, statePath []string) *StateInfoClassExternal {
	currentStates := states
	var state *StateInfoClassExternal

	for k, currentStateName := range statePath {
		state = getStateByName(currentStates, currentStateName)
		if state == nil {
			Revert("findStateInfoByPath: cannot find state \"" + currentStateName + "\" from path \"" + strings.Join(statePath, ".") + "\"")
			return nil
		}

		if k < len(statePath)-1 {
			if len(state.States) == 0 {
				Revert("findStateInfoByPath: state does not have childstates: " + strings.Join(statePath, "."))
				return nil
			}
			currentStates = state.States
		}
	}
	return state
}

// EqualStateOrIncluded checks if states are equal or included
func EqualStateOrIncluded(state1, state2 string) bool {
	if state1 == state2 {
		return true
	}
	if strings.Contains(state1, state2) {
		return true
	}
	if strings.Contains(state2, state1) {
		return true
	}
	return false
}

// Eventual handles timed transitions
func Eventual(config MachineExternal, args TimerArgs) {
	active := IsRegisteredIntervalActive(args.State, args.Delay, args.IntervalID)
	LoggerDebugExtended("eventual", []string{
		"expected_state", args.State,
		"delay", args.Delay,
		"intervalId", strconv.FormatInt(args.IntervalID, 10),
		"is_active", strconv.FormatBool(active),
	})

	if !active {
		return
	}

	// Deactivate interval
	RemoveInterval(args.State, args.Delay, args.IntervalID)

	service := LoadServiceFromConfig(config)
	currentState := getCurrentState()
	LoggerDebugExtended("eventual", []string{"current_state", currentState.Value})

	newstateconfig := FindStateInfo(service.Machine.States, args.State)
	if newstateconfig == nil {
		Revert("could not find state config for " + args.State)
		return
	}

	isEqual := EqualStateOrIncluded(currentState.Value, args.State)

	if newstateconfig.After == nil {
		return
	}

	afterTransition := getTransitionByName(newstateconfig.After, args.Delay)
	if afterTransition == nil {
		return
	}
	transitions := afterTransition.Transitions

	validTransitions := transitions
	if !isEqual {
		validTransitions = []Transition{}
		for _, transition := range transitions {
			if len(transition.Meta) > 0 && transition.Meta[0].Key == "force" {
				validTransitions = append(validTransitions, transition)
			}
		}
		if len(validTransitions) == 0 {
			LoggerDebug("eventual: we are in the wrong state", []string{
				"current_state", currentState.Value,
				"expected_state", args.State,
				"delay", args.Delay,
				"intervalId", strconv.FormatInt(args.IntervalID, 10),
			})
			return
		}
	}

	if len(validTransitions) == 0 {
		return
	}

	LoggerDebug("eventual", []string{
		"current_state", currentState.Value,
		"expected_state", args.State,
		"delay", args.Delay,
		"intervalId", strconv.FormatInt(args.IntervalID, 10),
	})

	// Create event with interval parameters
	intervalIdField := ActionParam{Key: "intervalId", Value: strconv.FormatInt(args.IntervalID, 10)}
	stateField := ActionParam{Key: "state", Value: args.State}
	delayField := ActionParam{Key: "delay", Value: args.Delay}
	emptyEvent := EventObject{Type: "", Params: []ActionParam{intervalIdField, stateField, delayField}}

	state := getCurrentState()
	newstate := ApplyTransitions(&service.Machine, *state, validTransitions, emptyEvent, 0)
	if newstate == nil {
		return
	}

	setCurrentState(newstate)
	ExecuteStateActions(service, *newstate, emptyEvent)
}

// Setup sets up the machine with a library contract
func Setup(config MachineExternal, contractAddress string) {
	if config.Library == "" {
		Revert("could not execute setup: fsm does not have a library")
		return
	}

	param := ActionParam{Key: "previousAddress", Value: contractAddress}
	calldata := ExternalActionCallData{
		Method: "setup",
		Params: []ActionParam{param},
		Event:  EventObject{},
	}

	calldataBytes, err := json.Marshal(calldata)
	if err != nil {
		Revert("could not execute setup: marshal error: " + err.Error())
		return
	}
	success, _ := wasmx.Call(config.Library, nil, calldataBytes, big.NewInt(wasmx.DEFAULT_GAS_TX), MODULE_NAME)
	if !success {
		Revert("could not execute setup")
	}
}

// LoadServiceFromConfig loads service from config
func LoadServiceFromConfig(config MachineExternal) *ServiceExternal {
	status := getCurrentStatus()
	return &ServiceExternal{Machine: config, Status: status}
}

// RunInternal runs the internal state machine
func RunInternal(config MachineExternal, event EventObject) []byte {
	service := LoadServiceFromConfig(config)
	service.Send(event)
	return []byte{}
}

// ExecuteInternal executes an internal action
func ExecuteInternal(config MachineExternal, event EventObject, action ActionObject) []byte {
	service := LoadServiceFromConfig(config)
	state := getCurrentState()
	ExecuteStateAction(service, *state, event, action)
	return []byte{}
}

// IfIntervalActive checks if an interval is active (used in guards)
func IfIntervalActive(params []ActionParam, event EventObject) bool {
	if len(params) == 0 {
		params = event.Params
	}

	var intervalIdStr, state, delay string
	for _, param := range params {
		switch param.Key {
		case "intervalId":
			intervalIdStr = param.Value
		case "state":
			state = param.Value
		case "delay":
			delay = param.Value
		}
	}

	if intervalIdStr == "" {
		Revert("no intervalId found")
		return false
	}
	if state == "" {
		Revert("no state found")
		return false
	}
	if delay == "" {
		Revert("no delay found")
		return false
	}

	intervalId, err := strconv.ParseInt(intervalIdStr, 10, 64)
	if err != nil {
		Revert("invalid intervalId: " + intervalIdStr)
		return false
	}

	active := IsRegisteredIntervalActive(state, delay, intervalId)
	LoggerDebugExtended("ifIntervalActive", []string{
		"intervalId", intervalIdStr,
		"active", strconv.FormatBool(active),
	})

	// Remove the interval data
	RemoveInterval(state, delay, intervalId)
	return active
}

func getTransitionByName(transitions []TransitionExternal, name string) *TransitionExternal {
	for _, t := range transitions {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

func getStateByName(states []StateInfoClassExternal, name string) *StateInfoClassExternal {
	for _, t := range states {
		if t.Name == name {
			return &t
		}
	}
	return nil
}
