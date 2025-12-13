package lib

import (
	"encoding/json"
	"strconv"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-utils"
)

// LoadUser loads a user by ID
func LoadUser(userID string) (*UserIdentity, error) {
	key := []byte(STORAGE_USER_PREFIX + userID)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil, nil
	}

	var user UserIdentity
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveUser saves a user
func SaveUser(user *UserIdentity) error {
	key := []byte(STORAGE_USER_PREFIX + user.UserID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

// LoadUserIDByAddress loads user ID by address
func LoadUserIDByAddress(address string) string {
	key := []byte(STORAGE_ADDR_PREFIX + address)
	data := wasmx.StorageLoad(key)
	return string(data)
}

// SaveUserIDByAddress saves user ID mapping for an address
func SaveUserIDByAddress(address string, userID string) {
	key := []byte(STORAGE_ADDR_PREFIX + address)
	wasmx.StorageStore(key, []byte(userID))
}

// DeleteUserIDByAddress removes address to user ID mapping
func DeleteUserIDByAddress(address string) {
	key := STORAGE_ADDR_PREFIX + address
	wasmx.StorageDelete(key)
}

// LoadAddressInfo loads address information
func LoadAddressInfo(userID string, address string) (*AddressInfo, error) {
	key := []byte(STORAGE_ADDR_INFO_PREFIX + userID + ":" + address)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil, nil
	}

	var addrInfo AddressInfo
	if err := json.Unmarshal(data, &addrInfo); err != nil {
		return nil, err
	}
	return &addrInfo, nil
}

// SaveAddressInfo saves address information
func SaveAddressInfo(userID string, addrInfo *AddressInfo) error {
	key := []byte(STORAGE_ADDR_INFO_PREFIX + userID + ":" + addrInfo.Address)
	data, err := json.Marshal(addrInfo)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

// DeleteAddressInfo removes address information
func DeleteAddressInfo(userID string, address string) {
	key := STORAGE_ADDR_INFO_PREFIX + userID + ":" + address
	wasmx.StorageDelete(key)
}

// GetNextUserID increments and returns the next user ID
func GetNextUserID() string {
	key := []byte(STORAGE_USER_COUNTER)
	data := wasmx.StorageLoad(key)

	var counter uint64
	if len(data) > 0 {
		counter, _ = strconv.ParseUint(string(data), 10, 64)
	}

	counter++
	wasmx.StorageStore(key, []byte(utils.U64toa(counter)))

	return GenerateUserID(counter)
}
