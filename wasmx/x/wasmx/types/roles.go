package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// role => contract address
const KEY_CONTRACT_ADDRESS_BY_ROLE = "addrbyrole_"

// label => ROLE
const KEY_ROLE_BY_LABEL = "rolebylabel_"

// addr => label
const KEY_LABEL_BY_ADDR = "labelbyaddr_"

type RoleJSON struct {
	Role            string `json:"role"`
	Label           string `json:"label"`
	ContractAddress string `json:"contract_address"`
}

type RolesGenesis struct {
	Roles            []RoleJSON `json:"roles"`
	PreviousContract string     `json:"previous_contract"`
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRolePrefix(roleaddr sdk.AccAddress, role string) []byte {
	prefixStoreKey := GetContractStorePrefix(roleaddr.Bytes())
	prefixStoreKey = append(prefixStoreKey, []byte(KEY_CONTRACT_ADDRESS_BY_ROLE)...)
	return append(prefixStoreKey, []byte(role)...)
}

// GetRoleLabelPrefix returns the store prefix for the system roles
func GetRoleLabelPrefix(roleaddr sdk.AccAddress, label string) []byte {
	prefixStoreKey := GetContractStorePrefix(roleaddr.Bytes())
	prefixStoreKey = append(prefixStoreKey, []byte(KEY_ROLE_BY_LABEL)...)
	return append(prefixStoreKey, []byte(label)...)
}

// GetRoleContractPrefix returns the store prefix for the system roles
func GetRoleContractPrefix(roleaddr sdk.AccAddress, addr sdk.AccAddress) []byte {
	prefixStoreKey := GetContractStorePrefix(roleaddr.Bytes())
	prefixStoreKey = append(prefixStoreKey, []byte(KEY_LABEL_BY_ADDR)...)
	return append(prefixStoreKey, addr...)
}
