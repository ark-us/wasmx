package types

// iterator type from "github.com/cometbft/cometbft-db"

// Iterator represents an iterator over a domain of keys. Callers must call Close when done.
// No writes can happen to a domain while there exists an iterator over it, some backends may take
// out database locks to ensure this will not happen.
//
// Callers must make sure the iterator is valid before calling any methods on it, otherwise
// these methods will panic. This is in part caused by most backend databases using this convention.
//
// As with DB, keys and values should be considered read-only, and must be copied before they are
// modified.
//
// Typical usage:
//
// var itr Iterator = ...
// defer itr.Close()
//
//	for ; itr.Valid(); itr.Next() {
//	  k, v := itr.Key(); itr.Value()
//	  ...
//	}
//
//	if err := itr.Error(); err != nil {
//	  ...
//	}
type Iterator interface {
	// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
	// CONTRACT: start, end readonly []byte
	Domain() (start []byte, end []byte)

	// Valid returns whether the current iterator is valid. Once invalid, the Iterator remains
	// invalid forever.
	Valid() bool

	// Next moves the iterator to the next key in the database, as defined by order of iteration.
	// If Valid returns false, this method will panic.
	Next()

	// Key returns the key at the current position. Panics if the iterator is invalid.
	// CONTRACT: key readonly []byte
	Key() (key []byte)

	// Value returns the value at the current position. Panics if the iterator is invalid.
	// CONTRACT: value readonly []byte
	Value() (value []byte)

	// Error returns the last error encountered by the iterator, if any.
	Error() error

	// Close closes the iterator, relasing any allocated resources.
	Close() error
}
