package main

import (
	"github.com/tidwall/gjson"

	wasmx "github.com/wasmx/wasmx-go"
)

//go:wasm-module simplestorage
//export instantiate
func instantiate() {
	data := wasmx.GetCallData()
	key := []byte("storagekey")
	wasmx.StorageStore(key, data)
}

type Calldata struct {
	Store []string `json:"store,omitempty"`
	Load  []string `json:"load,omitempty"`
}

func main() {
	data := string(wasmx.GetCallData())
	if gjson.Get(data, "store").Exists() {
		value := gjson.Get(data, "store|0")
		storageStore([]byte(value.String()))
	} else if gjson.Get(data, "load").Exists() {
		resp := storageLoad()
		wasmx.SetReturnData(resp)
	}
}

func storageStore(value []byte) {
	key := []byte("storagekey")
	wasmx.StorageStore(key, value)
}

func storageLoad() []byte {
	key := []byte("storagekey")
	return wasmx.StorageLoad(key)
}
