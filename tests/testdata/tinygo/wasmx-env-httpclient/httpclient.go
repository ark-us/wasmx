package httpclient

// #include <stdlib.h>
import "C"

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-env-utils"
)

//go:wasm-module httpclient
//export wasmx_httpclient_i64_1
func wasmx_httpclient_i64_1() {}

//go:wasmimport httpclient Request
func Request_(reqPtr int64) int64

func Request(req *HttpRequestWrap) HttpResponseWrap {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Request_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp HttpResponseWrap
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}
