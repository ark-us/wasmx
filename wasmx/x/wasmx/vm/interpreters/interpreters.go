package interpreters

import (
	_ "embed"
)

var (
	//go:embed keccak256.wasm
	Keccak256Util []byte
)
