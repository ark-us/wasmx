package wasmx

import (
	_ "embed"
)

var (
	//go:embed simple_storage.wasm
	WasmxSimpleStorage []byte
)
