package main

//go:wasm-module p2p
//export wasmx_p2p_json_i64_1
func Wasmx_p2p_json_i64_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}

//go:wasm-module wasmx-env-p2p
//export instantiate
func Instantiate() {}

func main() {}
