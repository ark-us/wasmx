package lib

import (
	"strconv"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
)

// GetLastIntervalId returns the last interval ID, starting at 1
func GetLastIntervalId() int64 {
	value := getContextValue(INTERVAL_ID_KEY)
	if value == "" {
		return 0
	}
	result, err := parseInt64(value)
	if err != nil {
		return 0
	}
	return result
}

// SetLastIntervalId sets the last interval ID
func SetLastIntervalId(value int64) {
	setContextValue(INTERVAL_ID_KEY, strconv.FormatInt(value, 10))
}

// RegisterIntervalIdKey generates a key for registering interval ID
func RegisterIntervalIdKey(state, delay string, intervalId int64) string {
	return INTERVAL_ID_KEY + "_" + state + "_" + delay + "_" + strconv.FormatInt(intervalId, 10)
}

// RegisterLastIntervalIdKey generates a key for the last interval ID in a state
func RegisterLastIntervalIdKey(state, delay string) string {
	return INTERVAL_ID_KEY + "_" + state + "_" + delay
}

// RegisterIntervalId registers an interval ID for a given state and delay
func RegisterIntervalId(state, delay string, intervalId int64) {
	setContextValue(RegisterLastIntervalIdKey(state, delay), strconv.FormatInt(intervalId, 10))
	setContextValue(RegisterIntervalIdKey(state, delay, intervalId), "1")
}

// GetLastIntervalIdForState returns the last interval ID for a specific state and delay
func GetLastIntervalIdForState(state, delay string) int64 {
	lastIntervalId := getContextValue(RegisterLastIntervalIdKey(state, delay))
	if lastIntervalId == "" {
		return 0
	}
	result, err := parseInt64(lastIntervalId)
	if err != nil {
		return 0
	}
	return result
}

// IsRegisteredIntervalActive checks if an interval is active
func IsRegisteredIntervalActive(state, delay string, intervalId int64) bool {
	value := getContextValue(RegisterIntervalIdKey(state, delay, intervalId))
	return value == "1"
}

// CancelIntervals cancels all intervals for a state and delay
func CancelIntervals(state, delay string) {
	lastIntervalId := GetLastIntervalIdForState(state, delay)
	// first intervalId is 1
	if lastIntervalId == 0 {
		return
	}
	TryCancelIntervals(state, delay, lastIntervalId)
}

// TryCancelIntervals recursively cancels intervals
func TryCancelIntervals(state, delay string, intervalId int64) {
	LoggerDebug("cancel interval: ", []string{
		"state", state,
		"delay", delay,
		"intervalId", strconv.FormatInt(intervalId, 10),
	})

	active := IsRegisteredIntervalActive(state, delay, intervalId)
	// remove the interval data
	RemoveInterval(state, delay, intervalId)

	// cancel timeout with wasmx
	err := wasmxcore.CancelTimeout(strconv.FormatInt(intervalId, 10))
	if err != nil {
		LoggerError("failed to cancel timeout", []string{"error", err.Error()})
	}

	if active && intervalId > 0 {
		TryCancelIntervals(state, delay, intervalId-1)
	}
}

// RemoveInterval removes an interval from storage
func RemoveInterval(state, delay string, intervalId int64) {
	setContextValue(RegisterIntervalIdKey(state, delay, intervalId), "")
}

// CancelActiveIntervals cancels active intervals based on action parameters
func CancelActiveIntervals(state State, params []ActionParam, event EventObject) {
	if len(params) == 0 {
		params = event.Params
	}

	var delay string
	for _, param := range params {
		if param.Key == "after" {
			delay = param.Value
			break
		}
	}

	if delay == "" {
		Revert("no delay found")
		return
	}

	// we cancel delayed actions for both previous and next state if they have the delay key
	CancelIntervals(state.PreviousValue, delay)
	CancelIntervals(state.Value, delay)
}
