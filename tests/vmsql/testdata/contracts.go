package testdata

import (
	_ "embed"
)

var (
	//go:embed as/wasmx_test_sql.wasm
	WasmxTestSql []byte

	//go:embed as/wasmx_erc20_sql.wasm
	WasmxErc20DType []byte
)
