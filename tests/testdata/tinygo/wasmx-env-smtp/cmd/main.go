package main

//go:wasm-module smtp
//export wasmx_smtp_i64_1
func Wasmx_smtp_i64_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}

//go:wasm-module wasmx-env-smtp
//export instantiate
func Instantiate() {}

func main() {}
