package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "websrv"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_websrv"
)

const (
	contractToRoutePrefix = iota + 1
	routeToContractPrefix
	lastClientID
	oauthAddressToClientsPrefix
	oauthClientIdToInfoPrefix
)

var (
	KeyContractToRoutePrefix       = []byte{contractToRoutePrefix}
	KeyRouteToContractPrefix       = []byte{routeToContractPrefix}
	KeyLastClientId                = []byte{lastClientID}
	KeyOauthAddressToClientsPrefix = []byte{oauthAddressToClientsPrefix}
	KeyOauthClientIdToInfoPrefix   = []byte{oauthClientIdToInfoPrefix}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// GetContractToRouteKey constructs the key for retreiving the http route for a contract address
func GetContractToRouteKey(addr sdk.AccAddress) []byte {
	return append(KeyContractToRoutePrefix, addr.Bytes()...)
}

// GetRouteToContractKey returns the key for retrieving the contract address for a route
func GetRouteToContractKey(route string) []byte {
	return append(KeyRouteToContractPrefix, []byte(route)...)
}

// GetKeyOauthAddressToClientsPrefix constructs the key for retreiving the client ids
// that were registered for a contract address
func GetKeyOauthAddressToClientsPrefix(addr sdk.AccAddress) []byte {
	return append(KeyOauthAddressToClientsPrefix, addr.Bytes()...)
}

// GetRouteToContractKey returns the key for retrieving the contract address for a route
func GetKeyOauthClientIdToInfoPrefix(clientId uint64) []byte {
	key := big.NewInt(int64(clientId)).FillBytes(make([]byte, 32))
	return append(KeyOauthClientIdToInfoPrefix, key...)
}
