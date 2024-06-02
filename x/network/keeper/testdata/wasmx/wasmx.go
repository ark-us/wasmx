package wasmx

import (
	_ "embed"
)

var (
	//go:embed crosschain.wasm
	WasmxCrossChain []byte

	//go:embed simple_storage.wasm
	WasmxSimpleStorage []byte
)
