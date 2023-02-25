package types

import (
	dbm "github.com/tendermint/tm-db"
)

// DefaultMaxQueryStackSize maximum size of the stack of contract instances doing queries
const DefaultMaxQueryStackSize uint32 = 10

// Checksum represents a hash of the Wasm bytecode that serves as an ID. Must be generated from this library.
type Checksum []byte

// WasmCode is an alias for raw bytes of the wasm compiled code
type WasmCode []byte

// KVStore is a reference to some sub-kvstore that is valid for one instance of a code
type KVStore interface {
	Get(key []byte) []byte
	Set(key, value []byte)
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	Iterator(start, end []byte) dbm.Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	ReverseIterator(start, end []byte) dbm.Iterator
}

// Querier lets us make read-only queries on other modules
type Querier interface {
	Query(request QueryRequest, gasLimit uint64) ([]byte, error)
	GasConsumed() uint64
}

type QueryRequest struct{}

// GasMeter is a read-only version of the sdk gas meter
type Gas = uint64
type GasMeter interface {
	GasConsumed() Gas
}

// LibwasmvmVersion returns the version of the loaded library
// at runtime. This can be used for debugging to verify the loaded version
// matches the expected version.
//
// When cgo is disabled at build time, this returns an error at runtime.
func LibWasmxVersion() (string, error) {
	// TODO
	return "", nil
}
