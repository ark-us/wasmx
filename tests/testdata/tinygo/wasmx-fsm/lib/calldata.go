package lib

import (
	"encoding/json"
	"strconv"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

type InterpreterCallData struct {
	Config   MachineExternal `json:"config"`
	Calldata CallData        `json:"calldata"`
}

type CallData struct {
	Setup           *string                  `json:"setup,omitempty"`
	Instantiate     *CallDataInstantiate     `json:"instantiate,omitempty"`
	GetCurrentState *CallDataGetCurrentState `json:"getCurrentState,omitempty"`
	GetContextValue *CallDataGetContextValue `json:"getContextValue,omitempty"`
	Run             *CallDataRun             `json:"run,omitempty"`
	Execute         *CallDataExecuteAction   `json:"execute,omitempty"`
	StartNode       *wasmx.HookCalld         `json:"StartNode,omitempty"`
	SetupNode       *wasmx.HookCalld         `json:"SetupNode,omitempty"`
}

type CallDataInstantiate struct {
	InitialState string         `json:"initialState"`
	Context      []ContextParam `json:"context"`
}

type CallDataGetCurrentState struct{}

type CallDataGetContextValue struct {
	Key string `json:"key"`
}

type CallDataRun struct {
	Event EventObject `json:"event"`
}

type CallDataExecuteAction struct {
	Action ActionObject `json:"action"`
}

func GetCallDataWrap() (*InterpreterCallData, error) {
	configBz, calldBz, err := GetInterpreterCalldata()
	if err != nil {
		return nil, err
	}

	var config MachineExternal
	err = json.Unmarshal(configBz, &config)
	if err != nil {
		return nil, err
	}

	var calldata CallData
	err = json.Unmarshal(calldBz, &calldata)
	if err != nil {
		return nil, err
	}

	return &InterpreterCallData{Config: config, Calldata: calldata}, nil
}

func GetInterpreterCalldata() (configBz []byte, calldBz []byte, err error) {
	calldraw := wasmx.GetCallData()
	configlenraw := calldraw[0:32]
	configlen, err := utils.ParseUint8ArrayToI32BigEndian(configlenraw)
	if err != nil {
		return nil, nil, err
	}
	configBz = calldraw[32 : 32+configlen]
	calldlenraw := calldraw[32+configlen : 64+configlen]
	calldlen, err := utils.ParseUint8ArrayToI32BigEndian(calldlenraw)
	if err != nil {
		return nil, nil, err
	}
	calldBz = calldraw[64+configlen : 64+configlen+calldlen]
	return configBz, calldBz, nil
}

func Sstore(key string, value string) {
	wasmx.StorageStore([]byte(key), []byte(value))
}

func Sload(key string) string {
	value := wasmx.StorageLoad([]byte(key))
	return string(value)
}

func IntToString(value int32) string {
	return strconv.Itoa(int(value))
}

type GrpcResponse struct {
	Data  string `json:"data"` // base64
	Error string `json:"error"`
}
