package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module simplestorage
//export instantiate
func Instantiate() {}

const MODULE_NAME = "simple_storage_identity"

type StoreRequest struct {
	Value string `json:"value"`
}

type LoadRequest struct{}

type Calldata struct {
	Store *StoreRequest `json:"store,omitempty"`
	Load  *LoadRequest  `json:"load,omitempty"`
}

func main() {
	databz := wasmx.GetCallData()
	calld := &Calldata{}
	err := json.Unmarshal(databz, calld)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	if calld.Store != nil {
		storageStore([]byte(calld.Store.Value))
	} else if calld.Load != nil {
		resp := storageLoad()
		wasmx.Finish(resp)
		return
	}
	wasmx.Finish([]byte{})
}

func storageStore(value []byte) {
	wasmx.StorageStore(getKey(), value)
}

func storageLoad() []byte {
	return wasmx.StorageLoad(getKey())
}

func getKey() []byte {
	userId, err := QueryUserByAddress(string(wasmx.GetCaller()))
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	return []byte(`key_` + userId)
}

// QueryUserByAddress queries the identity contract (by role) to get user_id from an address.
func QueryUserByAddress(address string) (string, error) {
	identityContract := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	if identityContract == "" {
		return "", nil
	}
	query := map[string]interface{}{
		"query_user_by_address": map[string]string{
			"address": address,
		},
	}
	queryBz, _ := json.Marshal(query)

	ok, respBz := wasmx.CallSimple(identityContract, queryBz, true, MODULE_NAME)
	if !ok {
		return "", nil
	}

	var resp struct {
		UserID string `json:"user_id"`
		Error  string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBz, &resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", nil
	}
	return resp.UserID, nil
}
