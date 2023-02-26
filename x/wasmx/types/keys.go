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

	ContractsDir = "wasmx"
)

var (
	CodeKeyPrefix       = []byte{0x01}
	ContractKeyPrefix   = []byte{0x02}
	ContractStorePrefix = []byte{0x03}
	SequenceKeyPrefix   = []byte{0x04}

	KeyLastCodeID     = append(SequenceKeyPrefix, []byte("lastCodeId")...)
	KeyLastInstanceID = append(SequenceKeyPrefix, []byte("lastContractId")...)
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// GetCodeKey constructs the key for retreiving the ID for the WASM code
func GetCodeKey(codeID uint64) []byte {
	contractIDBz := sdk.Uint64ToBigEndian(codeID)
	return append(CodeKeyPrefix, contractIDBz...)
}

// GetContractAddressKey returns the key for the WASM contract instance
func GetContractAddressKey(addr sdk.AccAddress) []byte {
	return append(ContractKeyPrefix, addr...)
}

// GetContractStorePrefix returns the store prefix for the WASM contract instance
func GetContractStorePrefix(addr sdk.AccAddress) []byte {
	return append(ContractStorePrefix, addr...)
}
