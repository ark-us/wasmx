package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "mcp-search"

// Storage keys
const (
	STORAGE_INIT_DATA = "init_data"
	STORAGE_DB_CONN   = "db_connection_id"
)

// CallData represents the different operations this contract supports
type CallData struct {
	// Tool execution
	ExecuteTool *ExecuteToolRequest `json:"execute_tool,omitempty"`
	// Initialization
	InitGenesis *InitGenesisRequest `json:"init_genesis,omitempty"`
	RoleChanged *wasmx.RolesChangedHook `json:"RoleChanged,omitempty"`
	// HTTP request handling
	HttpRequestHandler *HttpRequestIncoming `json:"HttpRequestHandler,omitempty"`
	// Populate with test/initial data (no auth required)
	Populate *PopulateRequest `json:"populate,omitempty"`
	// Generate embeddings for data
	GenerateEmbeddings *GenerateEmbeddingsRequest `json:"generate_embeddings,omitempty"`
}

// InitGenesisRequest for storing initialization data
type InitGenesisRequest struct {
	RoutePrefix        string            `json:"route_prefix"`
	ConnectionString   string            `json:"connection_string"`
	DatabaseName       string            `json:"database_name"`
	EmbeddingDimension int               `json:"embedding_dimension,omitempty"` // defaults to 1536
	EmbeddingMetric    string            `json:"embedding_metric,omitempty"`    // cosine, l2, ip - defaults to cosine
	EmbeddingModel     string            `json:"embedding_model,omitempty"`     // model name for embeddings (e.g., all-MiniLM-L6-v2)
	EnvironmentVars    map[string]string `json:"environment_vars,omitempty"`    // environment variables for CLI commands
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

// SearchResult represents a single search result
type SearchResult struct {
	Key      string  `json:"key"`
	Value    string  `json:"value,omitempty"`
	Distance float32 `json:"distance"`
}

// PopulateRequest for bulk populating embeddings (for testing/initialization)
type PopulateRequest struct {
	Items []PopulateItem `json:"items"`
}

// PopulateItem represents a single item to populate
type PopulateItem struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Embedding []float32 `json:"embedding"`
}

// PopulateResponse represents the response from populate operation
type PopulateResponse struct {
	Success      bool   `json:"success"`
	ItemsStored  int    `json:"items_stored"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GenerateEmbeddingsRequest for generating embeddings for data
type GenerateEmbeddingsRequest struct {
	// Optional: provide data items. If not provided, uses mock data from Python script
	Items []GenerateEmbeddingsItem `json:"items,omitempty"`
}

// GenerateEmbeddingsItem represents data to generate embeddings for
type GenerateEmbeddingsItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GenerateEmbeddingsResponse represents the response with generated embeddings
type GenerateEmbeddingsResponse struct {
	Success      bool          `json:"success"`
	Items        []PopulateItem `json:"items"` // Items with embeddings added
	ItemsStored  int           `json:"items_stored"`
	ErrorMessage string        `json:"error_message,omitempty"`
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
