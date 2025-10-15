package main

//go:wasm-module imap
//export wasmx_imap_i64_1
func Wasmx_imap_i64_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}

//go:wasm-module wasmx-env-imap
//export instantiate
func Instantiate() {}

func main() {}
