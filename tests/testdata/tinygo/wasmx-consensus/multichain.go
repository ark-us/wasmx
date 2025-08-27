package consensus

// #include <stdlib.h>
import "C"

//go:wasm-module multichain
//export memory_ptrlen_i64_1
func memory_ptrlen_i64_1_multichain() {}

//go:wasm-module multichain
//export wasmx_env_i64_2
func wasmx_env_i64_2_multichain() {}

//go:wasmimport multichain InitSubChain
func InitSubChain_(ptr int64) int64

//go:wasmimport multichain StartSubChain
func StartSubChain_(ptr int64) int64

//go:wasmimport multichain GetSubChainIds
func GetSubChainIds_() int64

//go:wasmimport multichain StartStateSync
func StartStateSync_(ptr int64) int64
