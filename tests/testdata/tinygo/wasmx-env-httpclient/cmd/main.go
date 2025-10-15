package main

//go:wasm-module httpclient
//export wasmx_httpclient_i64_1
func Wasmx_httpclient_i64_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}

//go:wasm-module wasmx-env-smtp
//export instantiate
func Instantiate() {}

func main() {}
