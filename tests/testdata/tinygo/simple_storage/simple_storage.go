package main

import (
	"encoding/json"
	"math/big"

	wasmx "github.com/loredanacirstea/wasmx-env"
)

//go:wasm-module simplestorage
//export instantiate
func Instantiate() {
	data := wasmx.GetCallData()
	key := []byte("storagekey")
	wasmx.StorageStore(key, data)
}

type StoreRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LoadRequest struct {
	Key string `json:"key"`
}

type WrapStoreRequest struct {
	Address string `json:"address"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

type WrapLoadRequest struct {
	Address string `json:"address"`
	Key     string `json:"key"`
	Sm      any    `json:"sm"`
}

type Calldata struct {
	Store     *StoreRequest     `json:"store,omitempty"`
	Load      *LoadRequest      `json:"load,omitempty"`
	WrapStore *WrapStoreRequest `json:"wrapStore,omitempty"`
	WrapLoad  *WrapLoadRequest  `json:"wrapLoad,omitempty"`
}

func main() {
	databz := wasmx.GetCallData()
	calld := &Calldata{}
	err := json.Unmarshal(databz, calld)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	if calld.Store != nil {
		storageStore([]byte(calld.Store.Key), []byte(calld.Store.Value))
	} else if calld.Load != nil {
		resp := storageLoad([]byte(calld.Load.Key))
		wasmx.SetFinishData(resp)
	} else if calld.WrapStore != nil {
		wrapStore(calld.WrapStore.Address, calld.WrapStore.Key, calld.WrapStore.Value)
	} else if calld.WrapLoad != nil {
		resp := wrapLoad(calld.WrapStore.Address, calld.WrapStore.Key)
		wasmx.SetFinishData(resp)
	}
}

func storageStore(key []byte, value []byte) {
	wasmx.StorageStore(key, value)
}

func storageLoad(key []byte) []byte {
	return wasmx.StorageLoad(key)
}

func wrapStore(address string, key string, value string) {
	calldata := &Calldata{Store: &StoreRequest{
		Key:   key,
		Value: value,
	}}
	calld, err := json.Marshal(calldata)
	if err != nil {
		panic(err)
	}
	success, _ := wasmx.Call(address, nil, calld, big.NewInt(50000000))
	if !success {
		panic("call failed")
	}
}

func wrapLoad(address string, key string) []byte {
	calldata := &Calldata{Load: &LoadRequest{
		Key: key,
	}}
	calld, err := json.Marshal(calldata)
	if err != nil {
		panic(err)
	}
	success, data := wasmx.CallStatic(address, calld, big.NewInt(50000000))
	if !success {
		panic("call failed")
	}
	return append(data, []byte("23")...)
}
