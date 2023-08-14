package main

import (
	"github.com/tidwall/gjson"

	wasmx "github.com/wasmx/wasmx-go"
)

//go:wasm-module myadd
//export instantiate
func instantiate() {
	data := wasmx.GetCallData()
	key := "storagekey"
	wasmx.StorageStore(key, data)
}

// type Calldata struct {
// 	Store []string `json:"store,omitempty"`
// 	Load  []string `json:"load,omitempty"`
// }

func main() {
	data := wasmx.GetCallData()
	if gjson.Get(data, "store").Exists() {
		value := gjson.Get(data, "store|0")
		storageStore(value.String())
	} else if gjson.Get(data, "load").Exists() {
		resp := storageLoad()
		wasmx.SetReturnData(resp)
	}
}

func storageStore(value string) {
	key := "storagekey"
	wasmx.StorageStore(key, value)
}

func storageLoad() string {
	key := "storagekey"
	return wasmx.StorageLoad(key)
}
