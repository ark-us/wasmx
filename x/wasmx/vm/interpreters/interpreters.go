package interpreters

import (
	_ "embed"
)

var (
	//go:embed ewasm.wasm
	EwasmInterpreter_1 []byte

	//go:embed evm_shanghai.wasm
	EvmInterpreter_1 []byte
)
