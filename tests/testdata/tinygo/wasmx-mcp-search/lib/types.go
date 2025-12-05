package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "wasmx-mcp-search"

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
	InitGenesis *InitGenesisRequest     `json:"init_genesis,omitempty"`
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
	RoutePrefix      string            `json:"route_prefix"`
	SearchConfig     SearchConfig      `json:"search_config"`              // Complete search configuration
	ToolDescriptions ToolDescriptions  `json:"tool_descriptions"`          // Descriptions for AI agents
	EnvironmentVars  map[string]string `json:"environment_vars,omitempty"` // environment variables for CLI commands
}

// ToolDescriptions contains user-facing descriptions for the MCP tools
type ToolDescriptions struct {
	SearchKnowledge SearchKnowledgeDescription `json:"search_knowledge"`
	VectorSearch    VectorSearchDescription    `json:"vector_search,omitempty"`
	StoreEmbedding  StoreEmbeddingDescription  `json:"store_embedding,omitempty"`
}

// SearchKnowledgeDescription describes the search_knowledge tool
type SearchKnowledgeDescription struct {
	Description      string `json:"description"`       // What the tool does
	QueryDescription string `json:"query_description"` // Description for query parameter
}

// VectorSearchDescription describes the vector_search tool
type VectorSearchDescription struct {
	Description string `json:"description"` // What the tool does
}

// StoreEmbeddingDescription describes the store_embedding tool
type StoreEmbeddingDescription struct {
	Description string `json:"description"` // What the tool does
}

// SearchConfig defines the database and tables to search
type SearchConfig struct {
	Database           DatabaseConfig `json:"database"`
	Tables             []TableConfig  `json:"tables"`
	EmbeddingDimension int            `json:"embedding_dimension"` // e.g., 768, 384
	EmbeddingMetric    string         `json:"embedding_metric"`    // cosine, l2, ip
	DefaultLimit       int            `json:"default_limit"`       // default result limit
}

// DatabaseConfig for PostgreSQL connection
type DatabaseConfig struct {
	ConnectionString string `json:"connection_string"`
	DatabaseName     string `json:"database_name"`
}

// TableConfig defines a single table with embeddings
type TableConfig struct {
	TableName       string `json:"table_name"`        // e.g., "organization_embeddings"
	EmbeddingColumn string `json:"embedding_column"`  // e.g., "embedding"
	IDColumn        string `json:"id_column"`         // e.g., "organization_id"
	CacheTextColumn string `json:"cache_text_column"` // e.g., "cache_text" - human readable text
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
	Success      bool           `json:"success"`
	Items        []PopulateItem `json:"items"` // Items with embeddings added
	ItemsStored  int            `json:"items_stored"`
	ErrorMessage string         `json:"error_message,omitempty"`
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
