package interpreters

import (
	_ "embed"
)

var (
	//go:embed ewasm.wasm
	EwasmInterpreter_1 []byte
)
