package crosschain

// #include <stdlib.h>
import "C"

// TinyGo packed ptr convention helpers are in wasmx-env-utils used by wrappers.

//go:wasm-module crosschain
//export memory_ptrlen_i64_1
func memory_ptrlen_i64_1() {}

//go:wasm-module crosschain
//export wasmx_env_i64_2
func wasmx_env_i64_2() {}

// Host imports for cross-chain execution/query

//go:wasmimport crosschain executeCrossChainTx
func executeCrossChainTx_(reqPtr int64) int64

//go:wasmimport crosschain executeCrossChainQuery
func executeCrossChainQuery_(reqPtr int64) int64

//go:wasmimport crosschain executeCrossChainQueryNonDeterministic
func executeCrossChainQueryNonDeterministic_(reqPtr int64) int64

//go:wasmimport crosschain executeCrossChainTxNonDeterministic
func executeCrossChainTxNonDeterministic_(reqPtr int64) int64

//go:wasmimport crosschain isAtomicTxInExecution
func isAtomicTxInExecution_(reqPtr int64) int64
