package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "mcp"
const OAUTH2_INSTANCE_ID = "mcp-oauth"
const POSTGRESQL_CONNECTION_ID = "mcp-db"

// Calldata structure
type CallData struct {
	StartServer      *StartServerRequest      `json:"start_server,omitempty"`
	InitGenesis      *InitGenesisRequest      `json:"init_genesis,omitempty"`
	GetParams        *GetParamsRequest        `json:"get_params,omitempty"`
	ConnectDatabase  *ConnectDatabaseRequest  `json:"connect_database,omitempty"`
	RoleChanged      *wasmx.RolesChangedHook  `json:"RoleChanged,omitempty"`
}

type StartServerRequest struct {
	Address string `json:"address"`
}

type InitGenesisRequest struct {
	Params ServerParams `json:"params"`
}

type GetParamsRequest struct{}

type ConnectDatabaseRequest struct {
	Connection string `json:"connection"` // PostgreSQL connection string
	DbName     string `json:"db_name"`
}

type ServerParams struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
}

// MCP Protocol types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
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

// User data storage
type UserData struct {
	FavoriteColor string `json:"favorite_color"`
}
