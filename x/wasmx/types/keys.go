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

	KeyLastCodeID = append(KeySequencePrefix, []byte("lastCodeId")...)
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// GetCodeKey constructs the key for retreiving the ID for the WASM code
func GetLastCodeIDKey(chainId string) []byte {
	return append([]byte(chainId), KeyLastCodeID...)
}

// GetCodeRootKey constructs the key for retreiving the ID for the WASM code
func GetCodeRootKey(chainId string) []byte {
	return append([]byte(chainId), KeyCodePrefix...)
}

// GetCodeKey constructs the key for retreiving the ID for the WASM code
func GetCodeKey(chainId string, codeID uint64) []byte {
	contractIDBz := sdk.Uint64ToBigEndian(codeID)
	return append(GetCodeRootKey(chainId), contractIDBz...)
}

// GetContractAddressRootKey returns the key for the WASM contract instance
func GetContractAddressRootKey(chainId string) []byte {
	return append([]byte(chainId), KeyContractPrefix...)
}

// GetContractAddressKey returns the key for the WASM contract instance
func GetContractAddressKey(chainId string, addr sdk.AccAddress) []byte {
	return append(GetContractAddressRootKey(chainId), addr...)
}

// GetContractStorePrefix returns the store prefix for the WASM contract instance
func GetContractStorePrefix(chainId string, addr sdk.AccAddress) []byte {
	return append(append([]byte(chainId), KeyContractStorePrefix...), addr...)
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRolePrefix(chainId string, role string) []byte {
	return append(append([]byte(chainId), KeyRolePrefix...), []byte(role)...)
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRoleLabelPrefix(chainId string, label string) []byte {
	return append(append([]byte(chainId), KeyRoleLabelPrefix...), []byte(label)...)
}

// GetRoleContractPrefix returns the store prefix for the system roles
func GetRoleContractPrefix(chainId string, addr sdk.AccAddress) []byte {
	return append(append([]byte(chainId), KeyRoleContractPrefix...), addr...)
}
