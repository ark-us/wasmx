package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	_ "github.com/loredanacirstea/wasmx-env-httpclient"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
)

//go:wasm-module emailprover
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

	if calld.ConnectWithPassword != nil {
		resp := vmimap.ConnectWithPassword(calld.ConnectWithPassword)
		response, err = json.Marshal(&resp)
	} else if calld.ConnectOAuth2 != nil {
		resp := vmimap.ConnectOAuth2(calld.ConnectOAuth2)
		response, err = json.Marshal(&resp)
	} else if calld.Close != nil {
		resp := vmimap.Close(calld.Close)
		response, err = json.Marshal(&resp)
	} else if calld.SignDKIM != nil {
		resp := SignDKIM(calld.SignDKIM)
		response, err = json.Marshal(&resp)
	} else if calld.VerifyDKIM != nil {
		resp := VerifyDKIM(calld.VerifyDKIM)
		response, err = json.Marshal(&resp)
	} else if calld.VerifyARC != nil {
		resp := VerifyARC(calld.VerifyARC)
		response, err = json.Marshal(&resp)
	} else if calld.SignDKIM != nil {
		resp := SignDKIM(calld.SignDKIM)
		response, err = json.Marshal(&resp)
	} else if calld.SignARC != nil {
		resp := SignARC(calld.SignARC)
		response, err = json.Marshal(&resp)
	} else {
		wasmx.Revert([]byte(`invalid function call data: ` + string(databz)))
	}
	wasmx.SetFinishData(response)
}
