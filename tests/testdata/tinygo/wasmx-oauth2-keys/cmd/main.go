package main

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-oauth2-keys/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module oauth2_keys
//export instantiate
func Instantiate() {
	// Initialize the contract on instantiation
	// Generate server secret automatically
	databz := wasmx.GetCallData()
	fmt.Println("*wasmx.instantiate.oauth2keys** ", string(databz))
	var callData *lib.MsgInitGenesis
	if len(databz) > 0 {
		if err := json.Unmarshal(databz, callData); err != nil {
			wasmx.Finish([]byte(err.Error()))
		}
	} else {
		callData = &lib.MsgInitGenesis{ServerSecret: ""}
	}
	result := lib.InitGenesis(callData)
	fmt.Println("*wasmx.instantiate.oauth2keys** InitGenesis completed")
	wasmx.Finish(result)
}

func main() {
	databz := wasmx.GetCallData()
	result := lib.Execute(databz)
	wasmx.Finish(result)
}
