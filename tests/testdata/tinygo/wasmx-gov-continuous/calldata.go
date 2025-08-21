package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
)

// getCallDataWrap parses the call data into our CallData structure
func getCallDataWrap() (*CallData, error) {
	callDataRaw := wasmx.GetCallData()
	var callData CallData
	err := json.Unmarshal(callDataRaw, &callData)
	if err != nil {
		return nil, err
	}
	return &callData, nil
}

// getCallDataInitialize parses initialization parameters
func getCallDataInitialize() (*Params, error) {
	callDataRaw := wasmx.GetCallData()
	var params Params
	err := json.Unmarshal(callDataRaw, &params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}