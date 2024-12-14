package tinygo

import (
	_ "embed"
)

var (
	//go:embed forward.wasm
	TinyGoForward []byte

	//go:embed add.wasm
	TinyGoAdd []byte

	//go:embed simple_storage.wasm
	TinyGoSimpleStorage []byte
)
