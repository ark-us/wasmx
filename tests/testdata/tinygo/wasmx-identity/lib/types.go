package lib

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-utils"
)

const MODULE_NAME = "account_identity"

// Storage keys
const (
	STORAGE_USER_PREFIX      = "user:"        // user:<user_id> -> UserIdentity
	STORAGE_ADDR_PREFIX      = "addr:"        // addr:<address> -> user_id
	STORAGE_ADDR_INFO_PREFIX = "addrinfo:"    // addrinfo:<user_id>:<address> -> AddressInfo
	STORAGE_USER_COUNTER     = "user_counter" // counter for generating user IDs
)

// UserIdentity represents a user's identity with associated addresses
type UserIdentity struct {
	UserID         string   `json:"user_id"`
	PrimaryAddress string   `json:"primary_address"` // Primary address (usually first one)
	Addresses      []string `json:"addresses"`       // List of associated addresses
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
}

// AddressInfo contains detailed information about an address associated with a user
type AddressInfo struct {
	Address       string       `json:"address"`
	PublicKey     []byte       `json:"public_key"`     // Public key bytes for this address
	ServiceDomain string       `json:"service_domain"` // Empty for regular addresses, domain for ephemeral keys
	Permissions   []Permission `json:"permissions"`    // Permissions granted to this address
	ExpiresAt     int64        `json:"expires_at"`     // 0 for non-expiring addresses
	CreatedAt     int64        `json:"created_at"`
}

// Permission represents a specific permission granted to an address
type Permission struct {
	Type   string                 `json:"type"`   // "transfer", "contract_call", "stake", etc.
	Params map[string]interface{} `json:"params"` // e.g., {"max_amount": "1000", "denom": "utoken"}
}

// Message types for contract calls

// MsgRegisterUser registers a new user with an initial address
type MsgRegisterUser struct {
	Address       string       `json:"address"`
	PublicKey     []byte       `json:"public_key"`
	ServiceDomain string       `json:"service_domain,omitempty"`
	Permissions   []Permission `json:"permissions,omitempty"`
	ExpiresAt     int64        `json:"expires_at,omitempty"`
}

type MsgRegisterUserResponse struct {
	UserID string `json:"user_id"`
}

// MsgAddAddress adds a new address to an existing user
type MsgAddAddress struct {
	UserID        string       `json:"user_id"`
	Address       string       `json:"address"`
	PublicKey     []byte       `json:"public_key"`
	ServiceDomain string       `json:"service_domain,omitempty"`
	Permissions   []Permission `json:"permissions,omitempty"`
	ExpiresAt     int64        `json:"expires_at,omitempty"`
}

type MsgAddAddressResponse struct {
	Success bool `json:"success"`
}

// MsgAddAddressInternal adds a new address to an existing user (internal call from trusted contracts)
type MsgAddAddressInternal struct {
	UserID        string       `json:"user_id"`
	Address       string       `json:"address"`
	PublicKey     []byte       `json:"public_key"`
	ServiceDomain string       `json:"service_domain,omitempty"`
	Permissions   []Permission `json:"permissions,omitempty"`
	ExpiresAt     int64        `json:"expires_at,omitempty"`
}

type MsgAddAddressInternalResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// MsgRemoveAddress removes an address from a user
type MsgRemoveAddress struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
}

type MsgRemoveAddressResponse struct {
	Success bool `json:"success"`
}

// MsgUpdatePermissions updates permissions for an existing address
type MsgUpdatePermissions struct {
	UserID      string       `json:"user_id"`
	Address     string       `json:"address"`
	Permissions []Permission `json:"permissions"`
}

type MsgUpdatePermissionsResponse struct {
	Success bool `json:"success"`
}

// MsgSetPrimaryAddress sets the primary address for a user
type MsgSetPrimaryAddress struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
}

type MsgSetPrimaryAddressResponse struct {
	Success bool `json:"success"`
}

// Query types

// MsgQueryUserByID retrieves user info by user ID
type MsgQueryUserByID struct {
	UserID string `json:"user_id"`
}

type QueryUserByIDResponse struct {
	User UserIdentity `json:"user"`
}

// MsgQueryUserByAddress retrieves user info by address
type MsgQueryUserByAddress struct {
	Address string `json:"address"`
}

type QueryUserByAddressResponse struct {
	UserID string       `json:"user_id"`
	User   UserIdentity `json:"user"`
}

// MsgQueryAddressInfo retrieves address information
type MsgQueryAddressInfo struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
}

type QueryAddressInfoResponse struct {
	AddressInfo AddressInfo `json:"address_info"`
}

// MsgQueryValidatePermission checks if an address has permission for an operation
type MsgQueryValidatePermission struct {
	Address       string                 `json:"address"`
	OperationType string                 `json:"operation_type"`
	OperationData map[string]interface{} `json:"operation_data"`
}

type QueryValidatePermissionResponse struct {
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}

// CallData structure for routing
type CallData struct {
	RegisterUser        *MsgRegisterUser        `json:"register_user,omitempty"`
	AddAddress          *MsgAddAddress          `json:"add_address,omitempty"`
	AddAddressInternal  *MsgAddAddressInternal  `json:"add_address_internal,omitempty"`
	RemoveAddress       *MsgRemoveAddress       `json:"remove_address,omitempty"`
	UpdatePermissions   *MsgUpdatePermissions   `json:"update_permissions,omitempty"`
	SetPrimaryAddress   *MsgSetPrimaryAddress   `json:"set_primary_address,omitempty"`

	QueryUserByID           *MsgQueryUserByID           `json:"query_user_by_id,omitempty"`
	QueryUserByAddress      *MsgQueryUserByAddress      `json:"query_user_by_address,omitempty"`
	QueryAddressInfo        *MsgQueryAddressInfo        `json:"query_address_info,omitempty"`
	QueryValidatePermission *MsgQueryValidatePermission `json:"query_validate_permission,omitempty"`

	InitGenesis *MsgInitGenesis `json:"init_genesis,omitempty"`
}

// MsgInitGenesis for contract initialization
type MsgInitGenesis struct {
	InitialUsers []InitialUser `json:"initial_users,omitempty"`
}

// InitialUser for genesis initialization
type InitialUser struct {
	Addresses []MsgRegisterUser `json:"addresses"`
}

// Helper function to generate next user ID
func GenerateUserID(counter uint64) string {
	return "user_" + utils.U64toa(counter)
}

// Helper to marshal JSON
func MarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		LoggerError("Failed to marshal JSON", []string{"error", err.Error()})
		return []byte("{}")
	}
	return data
}
