package lib

import (
	"encoding/json"
)

const MODULE_NAME = "oauth2_keys"

// Storage keys (non-deterministic storage)
const (
	STORAGE_KEY_PREFIX     = "key:"      // key:<public_key> -> EphemeralKeyPair
	STORAGE_TOKEN_PREFIX   = "token:"    // token:<oauth_token> -> public_key
	STORAGE_SERVER_SECRET  = "server_secret" // Master secret for key encryption
)

// EphemeralKeyPair stores an ephemeral key pair with metadata
type EphemeralKeyPair struct {
	PublicKey        string `json:"public_key"`
	PrivateKey       string `json:"private_key"`        // Encrypted with derived key
	UserID           string `json:"user_id"`
	ServiceDomain    string `json:"service_domain"`
	CreatedAt        int64  `json:"created_at"`
	ExpiresAt        int64  `json:"expires_at"`
	OAuthToken       string `json:"oauth_token"`        // Associated OAuth token
}

// Message types

// MsgGenerateEphemeralKey generates a new ephemeral key pair
type MsgGenerateEphemeralKey struct {
	OAuthToken    string `json:"oauth_token"`
	UserID        string `json:"user_id"`
	ServiceDomain string `json:"service_domain"`
	ExpiresAt     int64  `json:"expires_at"`
}

type MsgGenerateEphemeralKeyResponse struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"` // Only returned on generation, for browser storage
	Success    bool   `json:"success"`
}

// MsgSignTransaction signs a transaction using an ephemeral key
type MsgSignTransaction struct {
	OAuthToken      string `json:"oauth_token"`
	TransactionData []byte `json:"transaction_data"`
}

type MsgSignTransactionResponse struct {
	Signature []byte `json:"signature"`
	Success   bool   `json:"success"`
}

// MsgRegisterExternalKey registers a key pair generated externally (e.g., in browser)
type MsgRegisterExternalKey struct {
	OAuthToken    string `json:"oauth_token"`
	UserID        string `json:"user_id"`
	PublicKey     string `json:"public_key"`
	PrivateKey    string `json:"private_key"` // Will be encrypted
	ServiceDomain string `json:"service_domain"`
	ExpiresAt     int64  `json:"expires_at"`
}

type MsgRegisterExternalKeyResponse struct {
	Success bool `json:"success"`
}

// MsgRevokeKey revokes an ephemeral key
type MsgRevokeKey struct {
	OAuthToken string `json:"oauth_token"`
}

type MsgRevokeKeyResponse struct {
	Success bool `json:"success"`
}

// MsgDeleteExpiredKeys cleans up expired keys
type MsgDeleteExpiredKeys struct{}

type MsgDeleteExpiredKeysResponse struct {
	DeletedCount int `json:"deleted_count"`
}

// Query types

// MsgQueryGetPublicKey retrieves public key for an OAuth token
type MsgQueryGetPublicKey struct {
	OAuthToken string `json:"oauth_token"`
}

type QueryGetPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// MsgQueryGetKeyInfo retrieves key information (without private key)
type MsgQueryGetKeyInfo struct {
	PublicKey string `json:"public_key"`
}

type QueryGetKeyInfoResponse struct {
	UserID        string `json:"user_id"`
	ServiceDomain string `json:"service_domain"`
	ExpiresAt     int64  `json:"expires_at"`
	CreatedAt     int64  `json:"created_at"`
}

// CallData structure for routing
type CallData struct {
	GenerateEphemeralKey *MsgGenerateEphemeralKey `json:"generate_ephemeral_key,omitempty"`
	RegisterExternalKey  *MsgRegisterExternalKey  `json:"register_external_key,omitempty"`
	SignTransaction      *MsgSignTransaction      `json:"sign_transaction,omitempty"`
	RevokeKey            *MsgRevokeKey            `json:"revoke_key,omitempty"`
	DeleteExpiredKeys    *MsgDeleteExpiredKeys    `json:"delete_expired_keys,omitempty"`

	QueryGetPublicKey *MsgQueryGetPublicKey `json:"query_get_public_key,omitempty"`
	QueryGetKeyInfo   *MsgQueryGetKeyInfo   `json:"query_get_key_info,omitempty"`

	InitGenesis *MsgInitGenesis `json:"init_genesis,omitempty"`
}

// MsgInitGenesis for contract initialization
type MsgInitGenesis struct {
	ServerSecret string `json:"server_secret,omitempty"` // Master secret for key encryption
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
