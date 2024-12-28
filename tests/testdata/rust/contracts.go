package rust

import (
	_ "embed"
)

var (
	//go:embed simple_storage.wasm
	RustSimpleStorage []byte
)
