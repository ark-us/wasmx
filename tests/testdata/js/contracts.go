package js

import (
	_ "embed"
)

var (
	//go:embed javy/simple_storage.wasm
	JsJavySimpleStorage []byte

	//go:embed simple_storage.js
	JsSimpleStorage []byte

	//go:embed call.js
	JscallSimpleStorage []byte

	//go:embed call_evm.js
	JsCallEvmSimpleStorage []byte

	//go:embed blockchain.js
	JsBlockchain []byte

	//go:embed forward.js
	JsForward []byte
)
