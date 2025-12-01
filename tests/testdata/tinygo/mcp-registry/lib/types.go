package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "mcp-registry"

// Storage keys
const (
	STORAGE_REGISTERED_CONTRACTS = "registered_contracts" // []string of addresses
	STORAGE_CONTRACT_PREFIX      = "contract:"            // contract:{address} -> ContractRegistration
	STORAGE_ROUTE_PREFIX         = "route:"               // route:{path} -> address
	STORAGE_PARAMS               = "params"
	STORAGE_OAUTH_CLIENTS        = "oauth_clients"        // []string of client IDs
	STORAGE_OAUTH_CLIENT_PREFIX  = "oauth_client:"        // oauth_client:{client_id} -> OAuthClientInfo
	STORAGE_USERS                = "users"                // []string of user IDs
	STORAGE_USER_PREFIX          = "user:"                // user:{user_id} -> UserAccount
	STORAGE_USER_EMAIL_PREFIX    = "user_email:"          // user_email:{email} -> user_id
	STORAGE_SESSIONS             = "sessions"             // []string of session IDs
	STORAGE_SESSION_PREFIX       = "session:"             // session:{session_id} -> SessionInfo
)

// ContractRegistration stores metadata for each registered MCP contract
type ContractRegistration struct {
	Address       string              `json:"address"`
	RoutePrefix   string              `json:"route_prefix"`    // e.g., "/tools/execute"
	Tools         []MCPToolDefinition `json:"tools"`           // Tool definitions from contract
	RegisteredAt  int64               `json:"registered_at"`   // Block height
	LastUpdatedAt int64               `json:"last_updated_at"` // Block height
	Active        bool                `json:"active"`          // Enable/disable
}

// MCPToolDefinition represents a tool from MCP protocol
type MCPToolDefinition struct {
	Name        string                 `json:"name"`        // Tool name (unique per contract)
	Description string                 `json:"description"` // Human-readable description
	InputSchema map[string]interface{} `json:"inputSchema"` // JSON Schema for input
}

// ToolsListEntry for aggregated /tools/list response
type ToolsListEntry struct {
	Name            string                 `json:"name"` // Prefixed: "contract.execute_py"
	Description     string                 `json:"description"`
	InputSchema     map[string]interface{} `json:"inputSchema"`
	ContractAddress string                 `json:"-"` // Internal use only
	RoutePrefix     string                 `json:"-"` // Internal use only
}

// CallData structure for registry contract
type CallData struct {
	// Registry operations
	RegisterMCPContract   *RegisterMCPContractRequest   `json:"register_mcp_contract,omitempty"`
	DeregisterMCPContract *DeregisterMCPContractRequest `json:"deregister_mcp_contract,omitempty"`
	UpdateMCPTools        *UpdateMCPToolsRequest        `json:"update_mcp_tools,omitempty"`

	// Query operations
	GetRegisteredContracts *GetRegisteredContractsRequest `json:"get_registered_contracts,omitempty"`
	GetContractInfo        *GetContractInfoRequest        `json:"get_contract_info,omitempty"`
	GetAllTools            *GetAllToolsRequest            `json:"get_all_tools,omitempty"`

	// Server management (from current MCP)
	StartServer     *StartServerRequest     `json:"start_server,omitempty"`
	InitGenesis     *InitGenesisRequest     `json:"init_genesis,omitempty"`
	ConnectDatabase *ConnectDatabaseRequest `json:"connect_database,omitempty"`
	InitTables      *InitTablesRequest      `json:"init_tables,omitempty"`
	GetParams       *GetParamsRequest       `json:"get_params,omitempty"`
	RoleChanged     *wasmx.RolesChangedHook `json:"RoleChanged,omitempty"`

	// OAuth client management
	RegisterOAuthClient   *RegisterOAuthClientRequest   `json:"register_oauth_client,omitempty"`
	UpdateOAuthClient     *UpdateOAuthClientRequest     `json:"update_oauth_client,omitempty"`
	RevokeOAuthClient     *RevokeOAuthClientRequest     `json:"revoke_oauth_client,omitempty"`
	GetOAuthClient        *GetOAuthClientRequest        `json:"get_oauth_client,omitempty"`
	ListOAuthClients      *ListOAuthClientsRequest      `json:"list_oauth_clients,omitempty"`

	// User account management
	RegisterUser   *RegisterUserRequest   `json:"register_user,omitempty"`
	Login          *LoginRequest          `json:"login,omitempty"`
	Logout         *LogoutRequest         `json:"logout,omitempty"`
	GetCurrentUser *GetCurrentUserRequest `json:"get_current_user,omitempty"`
}

// RegisterMCPContractRequest for registering a new MCP contract
type RegisterMCPContractRequest struct {
	ContractAddress string `json:"contract_address"` // Address of MCP contract
	RoutePrefix     string `json:"route_prefix"`     // e.g., "/tools/execute"
	ToolsJSON       string `json:"tools_json"`       // JSON array of MCPToolDefinition
}

// DeregisterMCPContractRequest for removing an MCP contract
type DeregisterMCPContractRequest struct {
	ContractAddress string `json:"contract_address"`
}

// UpdateMCPToolsRequest for updating tools for a registered contract
type UpdateMCPToolsRequest struct {
	ContractAddress string `json:"contract_address"`
	ToolsJSON       string `json:"tools_json"` // Updated tools JSON
}

// GetRegisteredContractsRequest for querying all registered contracts
type GetRegisteredContractsRequest struct{}

// GetContractInfoRequest for querying info about a specific contract
type GetContractInfoRequest struct {
	ContractAddress string `json:"contract_address"`
}

// GetAllToolsRequest for getting aggregated tool list
type GetAllToolsRequest struct{}

// StartServerRequest for starting the HTTP server
type StartServerRequest struct {
	Address string `json:"address"`
}

// InitGenesisRequest for initializing genesis state
type InitGenesisRequest struct {
	Params           ServerParams                 `json:"params"`
	InitialContracts []RegisterMCPContractRequest `json:"initial_contracts,omitempty"`
}

// GetParamsRequest for getting server parameters
type GetParamsRequest struct{}

// ConnectDatabaseRequest for connecting to PostgreSQL
type ConnectDatabaseRequest struct {
	Connection string `json:"connection"` // PostgreSQL connection string
	DbName     string `json:"db_name"`
}

// InitTablesRequest for initializing database tables
type InitTablesRequest struct{}

// ServerParams stores server configuration
type ServerParams struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	DbConnection string   `json:"db_connection"` // PostgreSQL connection string
	DbName       string   `json:"db_name"`       // Database name
}

// MCP Protocol types (from current MCP)
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// OAuth Client Management Types

type RegisterOAuthClientRequest struct {
	Name         string   `json:"name"`          // Client application name
	Description  string   `json:"description"`   // Client description
	RedirectURIs []string `json:"redirect_uris"` // Allowed redirect URIs
	Scopes       []string `json:"scopes"`        // Requested scopes
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

type OAuthClientInfo struct {
	ClientID     string   `json:"client_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	WebsiteURL   string   `json:"website_url,omitempty"`
	LogoURL      string   `json:"logo_url,omitempty"`
	CreatedAt    int64    `json:"created_at"`
	Active       bool     `json:"active"`
}

// User Account Management Types

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username,omitempty"`
}

type RegisterUserResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"`
}

type LogoutRequest struct {
	SessionID string `json:"session_id"`
}

type GetCurrentUserRequest struct {
	SessionID string `json:"session_id"`
}

type UserAccount struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Username     string `json:"username,omitempty"`
	PasswordHash string `json:"password_hash"`
	CreatedAt    int64  `json:"created_at"`
	Active       bool   `json:"active"`
}

type SessionInfo struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}
