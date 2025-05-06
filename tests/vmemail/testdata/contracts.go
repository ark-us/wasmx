package testdata

import (
	_ "embed"
)

var (
	//go:embed as/wasmx_test_imap.wasm
	WasmxTestImap []byte

	//go:embed as/wasmx_test_smtp.wasm
	WasmxTestSmtp []byte
)
