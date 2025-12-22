package lib

import (
	wasmxhttp "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "wasmx-oauth2-server"

// Storage keys
const (
	STORAGE_PARAMS               = "params"
	STORAGE_INIT_DATA            = "init_data"      // InitGenesisRequest for RoleChanged
	STORAGE_OAUTH_CLIENTS        = "oauth_clients"  // []string of client IDs
	STORAGE_OAUTH_CLIENT_PREFIX  = "oauth_client:"  // oauth_client:{client_id} -> OAuthClient
	STORAGE_USERS                = "users"          // []string of user IDs
	STORAGE_USER_PREFIX          = "user:"          // user:{user_id} -> User
	STORAGE_USER_EMAIL_PREFIX    = "user_email:"    // user_email:{email} -> user_id
	STORAGE_SESSIONS             = "sessions"       // []string of session IDs
	STORAGE_SESSION_PREFIX       = "session:"       // session:{session_id} -> Session
	STORAGE_AUTH_CODES           = "auth_codes"     // []string of authorization codes
	STORAGE_AUTH_CODE_PREFIX     = "auth_code:"     // auth_code:{code} -> AuthorizationCode
	STORAGE_ACCESS_TOKENS        = "access_tokens"  // []string of access tokens
	STORAGE_ACCESS_TOKEN_PREFIX  = "access_token:"  // access_token:{token} -> AccessToken
	STORAGE_REFRESH_TOKENS       = "refresh_tokens" // []string of refresh tokens
	STORAGE_REFRESH_TOKEN_PREFIX = "refresh_token:" // refresh_token:{token} -> RefreshToken
)

const (
	SESSION_DURATION_SECONDS      = 86400 // 24 hours
	AUTH_CODE_DURATION_SECONDS    = 600   // 10 minutes
	ACCESS_TOKEN_DURATION_SECONDS = 3600  // 1 hour
	REFRESH_TOKEN_DURATION_DAYS   = 30    // 30 days
)

// CallData structure for OAuth2 server contract
type CallData struct {
	// OAuth client management
	RegisterOAuthClient *RegisterOAuthClientRequest `json:"register_oauth_client,omitempty"`
	UpdateOAuthClient   *UpdateOAuthClientRequest   `json:"update_oauth_client,omitempty"`
	RevokeOAuthClient   *RevokeOAuthClientRequest   `json:"revoke_oauth_client,omitempty"`
	GetOAuthClient      *GetOAuthClientRequest      `json:"get_oauth_client,omitempty"`
	ListOAuthClients    *ListOAuthClientsRequest    `json:"list_oauth_clients,omitempty"`

	// User account management
	RegisterUser   *RegisterUserRequest   `json:"register_user,omitempty"`
	Login          *LoginRequest          `json:"login,omitempty"`
	Logout         *LogoutRequest         `json:"logout,omitempty"`
	GetCurrentUser *GetCurrentUserRequest `json:"get_current_user,omitempty"`

	// OAuth2 flow operations
	CreateAuthorizationCode *CreateAuthorizationCodeRequest `json:"create_authorization_code,omitempty"`
	ExchangeCodeForToken    *ExchangeCodeForTokenRequest    `json:"exchange_code_for_token,omitempty"`
	ValidateAccessToken     *ValidateAccessTokenRequest     `json:"validate_access_token,omitempty"`
	RefreshAccessToken      *RefreshAccessTokenRequest      `json:"refresh_access_token,omitempty"`
	ValidateSession         *ValidateSessionRequest         `json:"validate_session,omitempty"`
	HttpRequestHandler      *wasmxhttp.HttpRequestIncoming  `json:"HttpRequestHandler,omitempty"`

	// Internal operations
	InitGenesis *InitGenesisRequest     `json:"init_genesis,omitempty"`
	RoleChanged *wasmx.RolesChangedHook `json:"RoleChanged,omitempty"`
}

// OAuth Client Types

type RegisterOAuthClientRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	WebsiteURL   string   `json:"website_url,omitempty"`
	LogoURL      string   `json:"logo_url,omitempty"`
}

type RegisterOAuthClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
}

type UpdateOAuthClientRequest struct {
	ClientID     string   `json:"client_id"`
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	WebsiteURL   string   `json:"website_url,omitempty"`
	LogoURL      string   `json:"logo_url,omitempty"`
}

type RevokeOAuthClientRequest struct {
	ClientID string `json:"client_id"`
}

type GetOAuthClientRequest struct {
	ClientID string `json:"client_id"`
}

type ListOAuthClientsRequest struct{}

type OAuthClient struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	WebsiteURL   string   `json:"website_url,omitempty"`
	LogoURL      string   `json:"logo_url,omitempty"`
	CreatedAt    int64    `json:"created_at"`
	Active       bool     `json:"active"`
}

// User Account Types

type RegisterUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Username  string `json:"username,omitempty"`
	PublicKey string `json:"public_key,omitempty"` // For WasmX blockchain registration
	Address   string `json:"address,omitempty"`    // Blockchain address derived from public key
}

type RegisterUserResponse struct {
	UserID            string `json:"user_id"`
	Email             string `json:"email"`
	Username          string `json:"username,omitempty"`
	IdentityUserID    string `json:"identity_user_id,omitempty"`    // WasmX identity contract user ID
	BlockchainAddress string `json:"blockchain_address,omitempty"`  // Primary blockchain address
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionID         string `json:"session_id"`
	UserID            string `json:"user_id"`
	Email             string `json:"email"`
	ExpiresAt         int64  `json:"expires_at"`
	IdentityUserID    string `json:"identity_user_id,omitempty"`
	BlockchainAddress string `json:"blockchain_address,omitempty"`
}

type LogoutRequest struct {
	SessionID string `json:"session_id"`
}

type GetCurrentUserRequest struct {
	SessionID string `json:"session_id"`
}

type User struct {
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	Username       string `json:"username,omitempty"`
	PasswordHash   string `json:"password_hash"`
	CreatedAt      int64  `json:"created_at"`
	Active         bool   `json:"active"`
	IdentityUserID string `json:"identity_user_id,omitempty"` // WasmX identity contract user ID
	Address        string `json:"address,omitempty"`           // Primary blockchain address
}

type Session struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}

// OAuth2 Flow Types

type CreateAuthorizationCodeRequest struct {
	ClientID            string   `json:"client_id"`
	UserID              string   `json:"user_id"`
	RedirectURI         string   `json:"redirect_uri"`
	Scopes              []string `json:"scopes"`
	CodeChallenge       string   `json:"code_challenge"`
	CodeChallengeMethod string   `json:"code_challenge_method"`
}

type CreateAuthorizationCodeResponse struct {
	Code      string `json:"code"`
	ExpiresAt int64  `json:"expires_at"`
}

type ExchangeCodeForTokenRequest struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
}

type ExchangeCodeForTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type ValidateAccessTokenRequest struct {
	Token string `json:"token"`
}

type ValidateAccessTokenResponse struct {
	Valid  bool     `json:"valid"`
	UserID string   `json:"user_id,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
}

type RefreshAccessTokenRequest struct {
	RefreshToken   string `json:"refresh_token"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	OldAccessToken string `json:"old_access_token,omitempty"` // Optional: old access token to revoke its ephemeral key
}

type RefreshAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ValidateSessionRequest struct {
	SessionID string `json:"session_id"`
}

type ValidateSessionResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
}

type AuthorizationCode struct {
	Code                string   `json:"code"`
	ClientID            string   `json:"client_id"`
	UserID              string   `json:"user_id"`
	RedirectURI         string   `json:"redirect_uri"`
	Scopes              []string `json:"scopes"`
	CodeChallenge       string   `json:"code_challenge"`
	CodeChallengeMethod string   `json:"code_challenge_method"`
	CreatedAt           int64    `json:"created_at"`
	ExpiresAt           int64    `json:"expires_at"`
	Used                bool     `json:"used"`
}

type AccessToken struct {
	Token     string   `json:"token"`
	UserID    string   `json:"user_id"`
	ClientID  string   `json:"client_id"`
	Scopes    []string `json:"scopes"`
	CreatedAt int64    `json:"created_at"`
	ExpiresAt int64    `json:"expires_at"`
}

type RefreshToken struct {
	Token     string   `json:"token"`
	UserID    string   `json:"user_id"`
	ClientID  string   `json:"client_id"`
	Scopes    []string `json:"scopes"`
	CreatedAt int64    `json:"created_at"`
	ExpiresAt int64    `json:"expires_at"`
	Revoked   bool     `json:"revoked"`
}

// Init Types

type InitGenesisRequest struct {
	InitialClients []RegisterOAuthClientRequest `json:"initial_clients,omitempty"`
}
