package lib

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// storeContractRegistration stores a contract registration
func storeContractRegistration(reg ContractRegistration) {
	key := []byte(STORAGE_CONTRACT_PREFIX + reg.Address)
	data, _ := json.Marshal(reg)
	fmt.Println("--wasmx-mcp-registry.data--", string(data))
	wasmx.StorageStore(key, data)
}

// getContractRegistration retrieves a contract registration
func getContractRegistration(address string) *ContractRegistration {
	key := []byte(STORAGE_CONTRACT_PREFIX + address)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil
	}

	var reg ContractRegistration
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil
	}
	return &reg
}

// deleteContractRegistration deletes a contract registration
func deleteContractRegistration(address string) {
	key := STORAGE_CONTRACT_PREFIX + address
	wasmx.StorageDelete(key)
}

// storeRouteMapping stores a route prefix to contract address mapping
func storeRouteMapping(routePrefix, address string) {
	key := []byte(STORAGE_ROUTE_PREFIX + routePrefix)
	wasmx.StorageStore(key, []byte(address))
}

// getContractByRoute retrieves contract address by route prefix
func getContractByRoute(routePrefix string) string {
	key := []byte(STORAGE_ROUTE_PREFIX + routePrefix)
	data := wasmx.StorageLoad(key)
	return string(data)
}

// deleteRouteMapping deletes a route mapping
func deleteRouteMapping(routePrefix string) {
	key := STORAGE_ROUTE_PREFIX + routePrefix
	wasmx.StorageDelete(key)
}

// routeExists checks if a route prefix is already registered
func routeExists(routePrefix string) bool {
	key := []byte(STORAGE_ROUTE_PREFIX + routePrefix)
	data := wasmx.StorageLoad(key)
	return len(data) > 0
}

// getRegisteredList retrieves the list of registered contract addresses
func getRegisteredList() []string {
	data := wasmx.StorageLoad([]byte(STORAGE_REGISTERED_CONTRACTS))
	if len(data) == 0 {
		return []string{}
	}
	var list []string
	json.Unmarshal(data, &list)
	return list
}

// addToRegisteredList adds a contract address to the registered list
func addToRegisteredList(address string) {
	list := getRegisteredList()
	// Check if already exists
	for _, addr := range list {
		if addr == address {
			return
		}
	}
	list = append(list, address)
	data, _ := json.Marshal(list)
	wasmx.StorageStore([]byte(STORAGE_REGISTERED_CONTRACTS), data)
}

// removeFromRegisteredList removes a contract address from the registered list
func removeFromRegisteredList(address string) {
	list := getRegisteredList()
	newList := []string{}
	for _, addr := range list {
		if addr != address {
			newList = append(newList, addr)
		}
	}
	data, _ := json.Marshal(newList)
	wasmx.StorageStore([]byte(STORAGE_REGISTERED_CONTRACTS), data)
}
