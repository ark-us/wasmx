package testdata

import (
	_ "embed"
)

var (
	//go:embed call.wasm
	CallWasm []byte

	//go:embed constructor_test.wasm
	ConstructorTestWasm []byte

	//go:embed Curve384Test.wasm
	Curve384TestWasm []byte

	//go:embed Erc20.wasm
	Erc20Wasm []byte

	//go:embed simple_storage_wc.wasm
	SimpleStorageWcWasm []byte

	//go:embed simple_storage.wasm
	SimpleStorageWasm []byte
)
