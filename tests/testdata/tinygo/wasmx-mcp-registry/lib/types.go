package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "wasmx-mcp-registry"

// Storage keys
const (
	STORAGE_REGISTERED_CONTRACTS = "registered_contracts" // []string of addresses
	STORAGE_CONTRACT_PREFIX      = "contract:"            // contract:{address} -> ContractRegistration
	STORAGE_ROUTE_PREFIX         = "route:"               // route:{path} -> address
	STORAGE_INIT_DATA            = "init_data"            // InitGenesisRequest for RoleChanged
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

	// Server management
	RegisterHttpRoutes *RegisterHttpRoutesRequest `json:"register_http_routes,omitempty"`
	InitGenesis        *InitGenesisRequest        `json:"init_genesis,omitempty"`
	RoleChanged        *wasmx.RolesChangedHook    `json:"RoleChanged,omitempty"`

	// HTTP request handling (called by HTTP server registry)
	HttpRequestHandler *HttpRequestIncoming `json:"HttpRequestHandler,omitempty"`
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

// RegisterHttpRoutesRequest for registering routes with HTTP registry
type RegisterHttpRoutesRequest struct{}

// InitGenesisRequest for initializing genesis state
type InitGenesisRequest struct {
	InitialContracts []RegisterMCPContractRequest `json:"initial_contracts,omitempty"`
}

// ConnectDatabaseRequest for connecting to PostgreSQL
type ConnectDatabaseRequest struct {
	Connection string `json:"connection"` // PostgreSQL connection string
	DbName     string `json:"db_name"`
}

// InitTablesRequest for initializing database tables
type InitTablesRequest struct{}

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

// HTTP request types (passed from HTTP server registry via CallData)
type HttpRequestIncoming struct {
	Method        string              `json:"method"`
	Url           string              `json:"url"`
	Header        map[string][]string `json:"header"`
	ContentLength int64               `json:"content_length"`
	Data          []byte              `json:"data"`
	RemoteAddr    string              `json:"remote_address"`
	RequestURI    string              `json:"request_uri"`
}

// HTTP response types (returned to HTTP server registry)
type HttpResponseWrap struct {
	Error string       `json:"error"`
	Data  HttpResponse `json:"data"`
}

type HttpResponse struct {
	StatusCode int                 `json:"status_code"`
	Status     string              `json:"status"`
	Header     map[string][]string `json:"header"`
	Data       []byte              `json:"data"`
}
