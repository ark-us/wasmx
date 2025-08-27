package lib

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
