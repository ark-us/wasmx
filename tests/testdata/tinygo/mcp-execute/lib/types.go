package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "mcp-execute"

const STORAGE_INIT_DATA = "init_data"

// CallData represents the different operations this contract supports
type CallData struct {
	// Tool execution
	ExecuteTool *ExecuteToolRequest `json:"execute_tool,omitempty"`
	// Initialization
	InitGenesis *InitGenesisRequest `json:"init_genesis,omitempty"`
	RoleChanged *wasmx.RolesChangedHook `json:"RoleChanged,omitempty"`
	// HTTP request handling
	HttpRequestHandler *HttpRequestIncoming `json:"HttpRequestHandler,omitempty"`
}

// InitGenesisRequest for storing initialization data
type InitGenesisRequest struct {
	RoutePrefix     string            `json:"route_prefix"`
	EnvironmentVars map[string]string `json:"environment_vars,omitempty"` // environment variables for CLI commands
}

// ExecuteToolRequest represents a request to execute a tool
type ExecuteToolRequest struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	UserID    string                 `json:"user_id,omitempty"`
}

// ExecuteToolResponse represents the response from tool execution
type ExecuteToolResponse struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem represents a single content item in the response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// HTTP request types (passed from HTTP server registry)
type HttpRequestIncoming struct {
	Method        string              `json:"method"`
	Url           string              `json:"url"`
	Header        map[string][]string `json:"header"`
	ContentLength int64               `json:"content_length"`
	Data          []byte              `json:"data"`
	RemoteAddr    string              `json:"remote_address"`
	RequestURI    string              `json:"request_uri"`
}

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
