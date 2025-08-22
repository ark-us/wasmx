package gov

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
)

// getCallDataWrap parses the call data into our CallData structure
func GetCallDataWrap() (*CallData, error) {
	callDataRaw := wasmx.GetCallData()
	var callData CallData
	err := json.Unmarshal(callDataRaw, &callData)
	if err != nil {
		return nil, err
	}
	return &callData, nil
}

// getCallDataInitialize parses initialization parameters
func GetCallDataInitialize() (*Params, error) {
	callDataRaw := wasmx.GetCallData()
	var params Params
	err := json.Unmarshal(callDataRaw, &params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}
