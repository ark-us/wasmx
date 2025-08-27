package crosschain

// #include <stdlib.h>
import "C"

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
