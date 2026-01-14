package main

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-groups/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module groups
//export instantiate
func Instantiate() {
	// Initialize the contract on instantiation
	databz := wasmx.GetCallData()
	fmt.Println("*wasmx.instantiate.groups** ", string(databz))
	var callData *lib.MsgInitGenesis
	if len(databz) > 0 {
		callData = &lib.MsgInitGenesis{}
		if err := json.Unmarshal(databz, callData); err != nil {
			wasmx.Finish([]byte(err.Error()))
		}
	}
	result := lib.InitGenesis(callData)
	fmt.Println("*wasmx.instantiate.groups** InitGenesis completed")
	wasmx.Finish(result)
}

func main() {
	databz := wasmx.GetCallData()
	fmt.Println("*wasmx.main.groups** ", string(databz))
	result := lib.Execute(databz)
	wasmx.Finish(result)
}
