package main

//go:wasm-module sql
//export wasmx_sql_i64_1
func Wasmx_sql_i64_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}

//go:wasm-module wasmx-env-sql
//export instantiate
func Instantiate() {}

func main() {}
