package main

import (
	"fmt"

	"github.com/tidwall/gjson"

	wasmx "github.com/loredanacirstea/wasmx-tinygo"
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
		wasmx.SetFinishData(resp)
	} else if gjson.Get(data, "wrapStore").Exists() {
		addr := gjson.Get(data, "wrapStore|0")
		value := gjson.Get(data, "wrapStore|1")
		wrapStore(addr.String(), value.String())
	} else if gjson.Get(data, "wrapLoad").Exists() {
		addr := gjson.Get(data, "wrapLoad|0")
		resp := wrapLoad(addr.String())
		wasmx.SetFinishData(resp)
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

func wrapStore(address string, value string) {
	calldata := fmt.Sprintf(`{"store":["%s"]}`, value)
	success, _ := wasmx.Call(50000000, address, make([]byte, 32), []byte(calldata))
	if !success {
		panic("call failed")
	}
}

func wrapLoad(address string) []byte {
	calldata := []byte(`{"load":[]}`)
	success, data := wasmx.CallStatic(50000000, address, calldata)
	if !success {
		panic("call failed")
	}
	return append(data, []byte("23")...)
}
