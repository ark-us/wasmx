package vmoauth2server

import (
	"fmt"
	"sync"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmoauth2server"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_OAUTH2SERVER_i32_VER1 = "wasmx_oauth2server_i32_1"
const HOST_WASMX_ENV_OAUTH2SERVER_i64_VER1 = "wasmx_oauth2server_i64_1"

const HOST_WASMX_ENV_OAUTH2SERVER_EXPORT = "wasmx_oauth2server_"

const HOST_WASMX_ENV_OAUTH2SERVER = "oauth2server"

type ContextKey string

const OAuth2ServerContextKey ContextKey = "oauth2server-context"

type Context struct {
	*vmtypes.Context
}

type OAuth2ServerContext struct {
	mtx       sync.Mutex
	Instances map[string]*OAuth2ServerInstance
}

type OAuth2ServerInstance struct {
	ConnectionID string
	ChainID      string
}

func (p *OAuth2ServerContext) GetInstance(id string) (*OAuth2ServerInstance, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	instance, found := p.Instances[id]
	return instance, found
}

func (p *OAuth2ServerContext) SetInstance(id string, instance *OAuth2ServerInstance) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.Instances[id]
	if found {
		return fmt.Errorf("cannot overwrite oauth2server instance: %s", id)
	}
	p.Instances[id] = instance
	return nil
}

func (p *OAuth2ServerContext) DeleteInstance(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.Instances, id)
}

// InitializeRequest initializes the OAuth2 server with database connection
type InitializeRequest struct {
	InstanceID   string `json:"instance_id"`
	ConnectionID string `json:"connection_id"`
}

type InitializeResponse struct {
	Error string `json:"error"`
}

// RegisterClientRequest registers a new OAuth2 client
type RegisterClientRequest struct {
	InstanceID   string   `json:"instance_id"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"` // Will be hashed
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	GrantTypes   []string `json:"grant_types"` // authorization_code, refresh_token, client_credentials
}

type RegisterClientResponse struct {
	Error string `json:"error"`
}

// UpdateClientRequest updates an existing OAuth2 client
type UpdateClientRequest struct {
	InstanceID   string   `json:"instance_id"`
	ClientID     string   `json:"client_id"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	GrantTypes   []string `json:"grant_types,omitempty"`
}

type UpdateClientResponse struct {
	Error string `json:"error"`
}

// DeleteClientRequest deletes an OAuth2 client
type DeleteClientRequest struct {
	InstanceID string `json:"instance_id"`
	ClientID   string `json:"client_id"`
}

type DeleteClientResponse struct {
	Error string `json:"error"`
}

// GetClientRequest retrieves client information
type GetClientRequest struct {
	InstanceID string `json:"instance_id"`
	ClientID   string `json:"client_id"`
}

type GetClientResponse struct {
	Error        string   `json:"error"`
	ClientID     string   `json:"client_id"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	GrantTypes   []string `json:"grant_types"`
	CreatedAt    int64    `json:"created_at"`
	UpdatedAt    int64    `json:"updated_at"`
}

// ValidateClientRequest validates client credentials
type ValidateClientRequest struct {
	InstanceID   string `json:"instance_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type ValidateClientResponse struct {
	Error string `json:"error"`
	Valid bool   `json:"valid"`
}

// CreateAuthorizationCodeRequest creates a new authorization code
type CreateAuthorizationCodeRequest struct {
	InstanceID            string   `json:"instance_id"`
	ClientID              string   `json:"client_id"`
	UserID                string   `json:"user_id"`
	RedirectURI           string   `json:"redirect_uri"`
	Scopes                []string `json:"scopes"`
	CodeChallenge         string   `json:"code_challenge,omitempty"`          // PKCE
	CodeChallengeMethod   string   `json:"code_challenge_method,omitempty"`   // S256 or plain
	ExpiresInSeconds      int64    `json:"expires_in_seconds"`                // e.g., 600 (10 minutes)
}

type CreateAuthorizationCodeResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// ValidateAuthorizationCodeRequest validates and consumes an authorization code
type ValidateAuthorizationCodeRequest struct {
	InstanceID   string `json:"instance_id"`
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier,omitempty"` // PKCE
}

type ValidateAuthorizationCodeResponse struct {
	Error  string   `json:"error"`
	Valid  bool     `json:"valid"`
	UserID string   `json:"user_id"`
	Scopes []string `json:"scopes"`
}

// IssueAccessTokenRequest issues a new access token
type IssueAccessTokenRequest struct {
	InstanceID       string   `json:"instance_id"`
	ClientID         string   `json:"client_id"`
	UserID           string   `json:"user_id"`
	Scopes           []string `json:"scopes"`
	ExpiresInSeconds int64    `json:"expires_in_seconds"` // e.g., 3600 (1 hour)
}

type IssueAccessTokenResponse struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

// IssueRefreshTokenRequest issues a new refresh token
type IssueRefreshTokenRequest struct {
	InstanceID       string   `json:"instance_id"`
	ClientID         string   `json:"client_id"`
	UserID           string   `json:"user_id"`
	Scopes           []string `json:"scopes"`
	ExpiresInSeconds int64    `json:"expires_in_seconds"` // 0 for non-expiring
}

type IssueRefreshTokenResponse struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

// ValidateAccessTokenRequest validates an access token
type ValidateAccessTokenRequest struct {
	InstanceID string `json:"instance_id"`
	Token      string `json:"token"`
}

type ValidateAccessTokenResponse struct {
	Error     string   `json:"error"`
	Valid     bool     `json:"valid"`
	ClientID  string   `json:"client_id"`
	UserID    string   `json:"user_id"`
	Scopes    []string `json:"scopes"`
	ExpiresAt int64    `json:"expires_at"`
}

// RefreshAccessTokenRequest refreshes an access token using a refresh token
type RefreshAccessTokenRequest struct {
	InstanceID       string `json:"instance_id"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresInSeconds int64  `json:"expires_in_seconds"` // For new access token
}

type RefreshAccessTokenResponse struct {
	Error        string   `json:"error"`
	AccessToken  string   `json:"access_token"`
	ClientID     string   `json:"client_id"`
	UserID       string   `json:"user_id"`
	Scopes       []string `json:"scopes"`
	ExpiresAt    int64    `json:"expires_at"`
}

// RevokeTokenRequest revokes an access or refresh token
type RevokeTokenRequest struct {
	InstanceID string `json:"instance_id"`
	Token      string `json:"token"`
}

type RevokeTokenResponse struct {
	Error string `json:"error"`
}

// IntrospectTokenRequest introspects a token (RFC 7662)
type IntrospectTokenRequest struct {
	InstanceID string `json:"instance_id"`
	Token      string `json:"token"`
}

type IntrospectTokenResponse struct {
	Error     string   `json:"error"`
	Active    bool     `json:"active"`
	Scope     string   `json:"scope"`      // Space-separated scopes
	ClientID  string   `json:"client_id"`
	Username  string   `json:"username"`   // UserID
	TokenType string   `json:"token_type"` // "access_token" or "refresh_token"
	ExpiresAt int64    `json:"exp"`        // Unix timestamp
	IssuedAt  int64    `json:"iat"`        // Unix timestamp
}

// CleanupExpiredTokensRequest removes expired tokens
type CleanupExpiredTokensRequest struct {
	InstanceID string `json:"instance_id"`
}

type CleanupExpiredTokensResponse struct {
	Error                   string `json:"error"`
	DeletedAuthCodes        int64  `json:"deleted_auth_codes"`
	DeletedAccessTokens     int64  `json:"deleted_access_tokens"`
	DeletedRefreshTokens    int64  `json:"deleted_refresh_tokens"`
}
