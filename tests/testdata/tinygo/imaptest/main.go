package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
)

//go:wasm-module imaptest
//export instantiate
func Instantiate() {}

func main() {
	databz := wasmx.GetCallData()
	calld := &Calldata{}
	err := json.Unmarshal(databz, calld)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	response := []byte{}

	if calld.Connect != nil {
		resp := vmimap.Connect(calld.Connect)
		response, err = json.Marshal(&resp)
	} else if calld.Close != nil {
		resp := vmimap.Close(calld.Close)
		response, err = json.Marshal(&resp)
	} else if calld.Listen != nil {
		resp := vmimap.Listen(calld.Listen)
		response, err = json.Marshal(&resp)
	} else if calld.Count != nil {
		resp := vmimap.Count(calld.Count)
		response, err = json.Marshal(&resp)
	} else if calld.UIDSearch != nil {
		resp := vmimap.UIDSearch(calld.UIDSearch)
		response, err = json.Marshal(&resp)
	} else if calld.ListMailboxes != nil {
		resp := vmimap.ListMailboxes(calld.ListMailboxes)
		response, err = json.Marshal(&resp)
	} else if calld.Fetch != nil {
		resp := vmimap.Fetch(calld.Fetch)
		response, err = json.Marshal(&resp)
	} else if calld.CreateFolder != nil {
		resp := vmimap.CreateFolder(calld.CreateFolder)
		response, err = json.Marshal(&resp)
	} else {
		wasmx.Revert([]byte(`invalid function call data: ` + string(databz)))
	}
	wasmx.SetFinishData(response)
}
