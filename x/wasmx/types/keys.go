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

	CLessStoreKey = "consensusless_wasmx"

	ContractsDir = "wasmx"

	TempDir = "temp"

	BinDir = "bin"

	SourceCodeDir = "sources"
)

const (
	codePrefix = iota + 1
	contractPrefix
	contractStorePrefix
	sequencePrefix
	prefixSystemContract
	rolePrefix
	roleLabelPrefix
	roleContractPrefix
)

var (
	KeyCodePrefix           = []byte{codePrefix}
	KeyContractPrefix       = []byte{contractPrefix}
	KeyContractStorePrefix  = []byte{contractStorePrefix}
	KeySequencePrefix       = []byte{sequencePrefix}
	KeyPrefixSystemContract = []byte{prefixSystemContract}

	KeyRolePrefix         = []byte{rolePrefix}
	KeyRoleLabelPrefix    = []byte{roleLabelPrefix}
	KeyRoleContractPrefix = []byte{roleContractPrefix}

	KeyLastCodeID     = append(KeySequencePrefix, []byte("lastCodeId")...)
	KeyLastInstanceID = append(KeySequencePrefix, []byte("lastContractId")...)
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// GetCodeKey constructs the key for retreiving the ID for the WASM code
func GetCodeKey(codeID uint64) []byte {
	contractIDBz := sdk.Uint64ToBigEndian(codeID)
	return append(KeyCodePrefix, contractIDBz...)
}

// GetContractAddressKey returns the key for the WASM contract instance
func GetContractAddressKey(addr sdk.AccAddress) []byte {
	return append(KeyContractPrefix, addr...)
}

// GetContractStorePrefix returns the store prefix for the WASM contract instance
func GetContractStorePrefix(addr sdk.AccAddress) []byte {
	return append(KeyContractStorePrefix, addr...)
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRolePrefix(role string) []byte {
	return append(KeyRolePrefix, []byte(role)...)
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRoleLabelPrefix(label string) []byte {
	return append(KeyRoleLabelPrefix, []byte(label)...)
}

// GetRoleContractPrefix returns the store prefix for the system roles
func GetRoleContractPrefix(addr sdk.AccAddress) []byte {
	return append(KeyRoleContractPrefix, addr...)
}
