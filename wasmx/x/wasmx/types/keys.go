package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "wasmx"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_wasmx"

	TStoreKey = "transient_wasmx"

	MetaConsensusStoreKey = "meta_consensus_wasmx"

	SingleConsensusStoreKey = "single_consensus_wasmx"

	ContractsDir = "wasmx"

	TempDir = "temp"

	BinDir = "bin"

	SourceCodeDir = "sources"
)

const (
	contractStorePrefix = iota + 1
	prefixSystemContract
	cacheSystemBootstrap
)

var (
	KeyContractStorePrefix  = []byte{contractStorePrefix}
	KeyPrefixSystemContract = []byte{prefixSystemContract}
	KeyCacheSystemBootstrap = []byte{cacheSystemBootstrap}
)

// GetContractStorePrefix returns the store prefix for the WASM contract instance
func GetContractStorePrefix(addr sdk.AccAddress) []byte {
	return append(KeyContractStorePrefix, addr...)
}

// GetCacheSystemBootstrapPrefix returns the store prefix for the system cache information for bootstrapping
func GetCacheSystemBootstrapPrefix() []byte {
	return KeyCacheSystemBootstrap
}
