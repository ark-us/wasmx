package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-slashing/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module wasmx-gov-continuous
//export instantiate
func Instantiate() {
	// databz := wasmx.GetCallData()
	// var params SomeType
	// err := json.Unmarshal(databz, &params)
	// if err != nil {
	// 	lib.Revert("invalid Instantiate calldata: " + err.Error() + ": " + string(databz))
	// }
}

func main() {
	databz := wasmx.GetCallData()
	calldata := &lib.CallData{}
	if err := json.Unmarshal(databz, calldata); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(databz))
	}

	// Public operations
	switch {
	case calldata.GetParams != nil:
		res := lib.GetParams(*calldata.GetParams)
		wasmx.SetFinishData(res)
		return
	}

	// Internal operations
	switch {
	case calldata.InitGenesis != nil:
		wasmx.OnlyInternal(lib.MODULE_NAME, "InitGenesis")
		res := lib.InitGenesis(*calldata.InitGenesis)
		wasmx.SetFinishData(res)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}
