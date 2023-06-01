package interpreters

import (
	_ "embed"
)

var (
	//go:embed ewasm.wasm
	EwasmInterpreter_1 []byte

	//go:embed keccak256.wasm
	Keccak256Util []byte
)
