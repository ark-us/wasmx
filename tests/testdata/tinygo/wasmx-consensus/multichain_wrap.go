package consensus

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	utils "github.com/loredanacirstea/wasmx-env-utils"
)

func InitSubChain(req InitSubChainMsg) (ResponseInitChain, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return ResponseInitChain{}, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "InitSubChain", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(InitSubChain_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(loggerModule, "InitSubChain", []string{"response", string(out)})
	var resp ResponseInitChain
	if err := json.Unmarshal(out, &resp); err != nil {
		return ResponseInitChain{}, err
	}
	return resp, nil
}

func StartSubChain(req StartSubChainMsg) (StartSubChainResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return StartSubChainResponse{}, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "StartSubChain", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(StartSubChain_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(loggerModule, "StartSubChain", []string{"response", string(out)})
	var resp StartSubChainResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return StartSubChainResponse{}, err
	}
	return resp, nil
}

func GetSubChainIds() ([]string, error) {
	out := utils.PackedPtrToBytes(GetSubChainIds_())
	wasmx.LoggerDebug(loggerModule, "GetSubChainIds", []string{"response", string(out)})
	var ids []string
	if err := json.Unmarshal(out, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func StartStateSync(req StartStateSyncRequest) (StartStateSyncResponse, error) {
	bz, err := json.Marshal(&req)
	if err != nil {
		return StartStateSyncResponse{}, err
	}
	wasmx.LoggerDebugExtended(loggerModule, "StartStateSync", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(StartStateSync_(utils.BytesToPackedPtr(bz)))
	wasmx.LoggerDebugExtended(loggerModule, "StartStateSync", []string{"response", string(out)})
	var resp StartStateSyncResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return StartStateSyncResponse{}, err
	}
	return resp, nil
}
