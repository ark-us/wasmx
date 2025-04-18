package testdata

import (
	_ "embed"
)

var (
	//go:embed as/wasmx_test_sql.wasm
	WasmxTestSql []byte
)
