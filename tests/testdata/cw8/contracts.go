package cw8

import (
	_ "embed"
)

var (
	// taken from cosmwasm/contracts/crypto_verify
	//go:embed crypto_verify-aarch64.wasm
	CryptoVerifyAarch64Wasm []byte

	//go:embed cw1_subkeys.wasm
	Cw1SubKeysWasm []byte

	//go:embed cw20_atomic_swap.wasm
	Cw20AtomicSwapWasm []byte

	//go:embed cw20_base-aarch64.wasm
	Cw20BaseAarch64Wasm []byte

	// taken from cosmwasm/contracts/reflect
	//go:embed reflect-aarch64.wasm
	ReflectAarch64Wasm []byte

	//go:embed simple_contract.wasm
	SimpleContractWasm []byte
)
