package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	contractAddressStorePrefix = iota + 1
	prefixRoleContract
	prefixLabelContract
)

var (
	// role => contract address
	KeyContractAddressStorePrefix = []byte{contractAddressStorePrefix}
	// label => ROLE
	KeyPrefixRoleContract = []byte{prefixRoleContract}
	// addr => label
	KeyPrefixLabelContract = []byte{prefixLabelContract}
)

type RoleJSON struct {
	Role        string   `json:"role"`
	StorageType int32    `json:"storage_type"`
	Primary     int32    `json:"primary"`
	Multiple    bool     `json:"multiple"`
	Labels      []string `json:"labels"`
	Addresses   []string `json:"addresses"`
}

type RolesGenesis struct {
	Roles               []RoleJSON `json:"roles"`
	IndividualMigration []string   `json:"individual_migration"`
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRolePrefix(roleaddr sdk.AccAddress, role string) []byte {
	prefixStoreKey := GetContractStorePrefix(roleaddr.Bytes())
	prefixStoreKey = append(prefixStoreKey, KeyContractAddressStorePrefix...)
	return append(prefixStoreKey, []byte(role)...)
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRoleLabelPrefix(roleaddr sdk.AccAddress, label string) []byte {
	prefixStoreKey := GetContractStorePrefix(roleaddr.Bytes())
	prefixStoreKey = append(prefixStoreKey, KeyPrefixRoleContract...)
	return append(prefixStoreKey, []byte(label)...)
}

// GetRoleContractPrefix returns the store prefix for the system roles
func GetRoleContractPrefix(roleaddr sdk.AccAddress, addr sdk.AccAddress) []byte {
	prefixStoreKey := GetContractStorePrefix(roleaddr.Bytes())
	prefixStoreKey = append(prefixStoreKey, KeyPrefixLabelContract...)
	return append(prefixStoreKey, addr...)
}
