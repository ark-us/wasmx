package main

import (
	"fmt"
	"strings"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Wemory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module myadd
//export add
func Add(x, y uint32) uint32 {
	return x + y
}

var someMap = map[string]string{
	"a": "a",
	"b": "b",
	"c": "c",
	"d": "d",
	"e": "e",
	"f": "f",
	"g": "g",
	"h": "h",
	"i": "i",
	"j": "j",
	"k": "k",
	"l": "l",
	"m": "m",
	"n": "n",
}

// main is required for the `wasi` target, even if it isn't used.
func main() {
	str := []string{}
	for k, v := range someMap {
		str = append(str, fmt.Sprintf("%s:%s", k, v))
	}
	wasmx.Finish([]byte(strings.Join(str, "-")))
}

//go:wasm-module myadd
//export instantiate
func Instantiate() {}
