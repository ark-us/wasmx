package main

import (
	lib "github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-identity/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module identity
//export instantiate
func Instantiate() {}

func main() {
	databz := wasmx.GetCallData()
	result := lib.Execute(databz)
	wasmx.SetFinishData(result)
}
