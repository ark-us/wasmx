package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "oauth2_keys"

// Storage keys (non-deterministic storage)
const (
	STORAGE_KEY_PREFIX     = "key."           // key:<public_key> -> EphemeralKeyPair
	STORAGE_TOKEN_PREFIX   = "token."         // token:<oauth_token> -> public_key
	STORAGE_SERVER_SECRET  = "server_secret"  // Master secret for key encryption
	STORAGE_FUNDER_PRIVKEY = "funder_privkey" // Private key for funding new accounts
)

// EphemeralKeyPair stores an ephemeral key pair with metadata
type EphemeralKeyPair struct {
	PublicKey     []byte `json:"public_key"`
	PrivateKey    []byte `json:"private_key"` // Encrypted with derived key
	UserID        string `json:"user_id"`
	ServiceDomain string `json:"service_domain"`
	CreatedAt     int64  `json:"created_at"`
	ExpiresAt     int64  `json:"expires_at"`
	OAuthToken    string `json:"oauth_token"` // Associated OAuth token
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
	PublicKey  []byte `json:"public_key"`
	PrivateKey []byte `json:"private_key"` // Only returned on generation, for browser storage
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
	PublicKey     []byte `json:"public_key"`
	PrivateKey    []byte `json:"private_key"` // Will be encrypted
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

// MsgInitAccount initializes a new account by sending native coins to it
type MsgInitAccount struct {
	Address string `json:"address"` // Bech32 address to initialize
}

type MsgInitAccountResponse struct {
	Success bool   `json:"success"`
	TxHash  string `json:"tx_hash"`
	Error   string `json:"error,omitempty"`
}

// Query types

// MsgQueryGetPublicKey retrieves public key for an OAuth token
type MsgQueryGetPublicKey struct {
	OAuthToken string `json:"oauth_token"`
}

type QueryGetPublicKeyResponse struct {
	PublicKey []byte `json:"public_key"`
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// MsgQueryGetKeyInfo retrieves key information (without private key)
type MsgQueryGetKeyInfo struct {
	PublicKey []byte `json:"public_key"`
}

type QueryGetKeyInfoResponse struct {
	UserID        string `json:"user_id"`
	ServiceDomain string `json:"service_domain"`
	ExpiresAt     int64  `json:"expires_at"`
	CreatedAt     int64  `json:"created_at"`
}

// MsgQueryValidateAndGetKey validates OAuth token and returns ephemeral key with private key
type MsgQueryValidateAndGetKey struct {
	OAuthToken string `json:"oauth_token"`
}

type QueryValidateAndGetKeyResponse struct {
	Valid      bool   `json:"valid"`
	PublicKey  []byte `json:"public_key"`
	PrivateKey []byte `json:"private_key"`
	Address    string `json:"address"` // Bech32 address derived from public key
	Reason     string `json:"reason,omitempty"`
}

// CallData structure for routing
type CallData struct {
	GenerateEphemeralKey *MsgGenerateEphemeralKey `json:"generate_ephemeral_key,omitempty"`
	RegisterExternalKey  *MsgRegisterExternalKey  `json:"register_external_key,omitempty"`
	SignTransaction      *MsgSignTransaction      `json:"sign_transaction,omitempty"`
	RevokeKey            *MsgRevokeKey            `json:"revoke_key,omitempty"`
	DeleteExpiredKeys    *MsgDeleteExpiredKeys    `json:"delete_expired_keys,omitempty"`
	InitAccount          *MsgInitAccount          `json:"init_account,omitempty"`

	QueryGetPublicKey      *MsgQueryGetPublicKey      `json:"query_get_public_key,omitempty"`
	QueryGetKeyInfo        *MsgQueryGetKeyInfo        `json:"query_get_key_info,omitempty"`
	QueryValidateAndGetKey *MsgQueryValidateAndGetKey `json:"query_validate_and_get_key,omitempty"`

	InitGenesis *MsgInitGenesis `json:"init_genesis,omitempty"`
}

// MsgInitGenesis for contract initialization
type MsgInitGenesis struct {
	ServerSecret   string     `json:"server_secret,omitempty"`    // Master secret for key encryption
	FunderPrivKey  string     `json:"funder_priv_key,omitempty"`  // Private key for funding new accounts (hex encoded)
	InitAccountAmt wasmx.Coin `json:"init_account_amt,omitempty"` // Coin to send when initializing accounts (default: {"amount":"1000000","denom":"amyt"})
	GasPrice       wasmx.Coin `json:"gas_price,omitempty"`        // Gas price for transactions
	RoutePrefix    string     `json:"route_prefix,omitempty"`     // HTTP route prefix (default: "/auth")
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
