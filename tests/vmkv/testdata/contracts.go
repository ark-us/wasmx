package testdata

import (
	_ "embed"
)

var (
	//go:embed as/wasmx_test_kvdb.wasm
	WasmxTestKvDB []byte
)
