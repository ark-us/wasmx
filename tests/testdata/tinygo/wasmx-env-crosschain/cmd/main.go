package cmd

//go:wasm-module crosschain
//export wasmx_crosschain_1
func Wasmx_crosschain_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}
