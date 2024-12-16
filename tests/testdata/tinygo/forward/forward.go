package main

import (
	"encoding/json"

	"github.com/tidwall/gjson"

	wasmx "github.com/loredanacirstea/wasmx-tinygo"
)

//go:wasm-module forward
//export instantiate
func instantiate() {
	storageStore("tinygo")
}

type Calldata struct {
	Forward    []string `json:"forward,omitempty"`
	ForwardGet []string `json:"forward_get,omitempty"`
}

type ForwardCalldata struct {
	Value     string
	Addresses []string
}

type ForwardGetCalldata struct {
	Addresses []string
}

func main() {
	data := string(wasmx.GetCallData())
	if gjson.Get(data, "forward").Exists() {
		value := gjson.Get(data, "forward|0").String()
		iaddrs := gjson.Get(data, "forward|1").Value().([]interface{})
		addrs := make([]string, len(iaddrs))
		for i, v := range iaddrs {
			addrs[i] = v.(string)
		}
		resp := forward(value, addrs)
		wasmx.SetFinishData(resp)
		return
	} else if gjson.Get(data, "forward_get").Exists() {
		iaddrs := gjson.Get(data, "forward_get|0").Value().([]interface{})
		addrs := make([]string, len(iaddrs))
		resp := forwardGet(addrs)
		wasmx.SetFinishData(resp)
		return
	}
	panic("Invalid function")
}

func forward(value string, addrs []string) []byte {
	value = value + string(storageLoad())
	var topics [][32]byte
	wasmx.Log([]byte(value), topics)

	if len(addrs) == 0 {
		return []byte(value)
	}

	value = value + " -> "
	address, addrs := addrs[0], addrs[1:]
	calldata, err := json.Marshal(ForwardCalldata{Value: value, Addresses: addrs})
	if err != nil {
		panic(err)
	}
	success, data := wasmx.Call(1000000, address, make([]byte, 32), calldata)
	if !success {
		panic("[go] call failed")
	}
	return data
}

func forwardGet(addrs []string) []byte {
	if len(addrs) == 0 {
		return storageLoad()
	}
	address, addrs := addrs[0], addrs[1:]
	calldata, err := json.Marshal(ForwardGetCalldata{Addresses: addrs})
	if err != nil {
		panic(err)
	}
	success, data := wasmx.CallStatic(1000000, address, calldata)
	if !success {
		panic("[go] call_static failed")
	}
	rdata := append(storageLoad(), []byte(" -> ")...)
	rdata = append(rdata, data...)
	return rdata
}

func storageStore(value string) {
	key := []byte("key")
	wasmx.StorageStore(key, []byte(value))
}

func storageLoad() []byte {
	key := []byte("key")
	return wasmx.StorageLoad(key)
}
