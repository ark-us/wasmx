package consensus

// #include <stdlib.h>
import "C"

//go:wasmimport multichain InitSubChain
func InitSubChain_(ptr int64) int64

//go:wasmimport multichain StartSubChain
func StartSubChain_(ptr int64) int64

//go:wasmimport multichain GetSubChainIds
func GetSubChainIds_() int64

//go:wasmimport multichain StartStateSync
func StartStateSync_(ptr int64) int64
