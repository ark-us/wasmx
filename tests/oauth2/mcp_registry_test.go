package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	testdata "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestMCPRegistry() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// registryAddress := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_MCP_REGISTRY))

	// Prepare init data for mcp-execute contract
	executeInitData := &MCPContractInitGenesis{
		InitGenesis: &MCPContractInitGenesisRequest{
			RoutePrefix: "/tools/execute",
		},
	}
	executeInitDataJSON, _ := json.Marshal(executeInitData)

	// Instantiate mcp-execute with init data
	executeCodeId := appA.StoreCode(sender, testdata.MCPExecute, nil)
	executeAddress := appA.InstantiateCode(sender, executeCodeId, types.WasmxExecutionMessage{Data: executeInitDataJSON}, "mcp_execute", nil)
	fmt.Println("Instantiated mcp-execute contract:", executeAddress.String())

	// Prepare init data for mcp-userdata contract
	userdataInitData := &MCPContractInitGenesis{
		InitGenesis: &MCPContractInitGenesisRequest{
			RoutePrefix: "/tools/userdata",
		},
	}
	userdataInitDataJSON, _ := json.Marshal(userdataInitData)

	// Instantiate mcp-userdata with init data
	userCodeId := appA.StoreCode(sender, testdata.MCPUserdata, nil)
	userAddress := appA.InstantiateCode(sender, userCodeId, types.WasmxExecutionMessage{Data: userdataInitDataJSON}, "mcp_userdata", nil)
	fmt.Println("Instantiated mcp-userdata contract:", userAddress.String())

	// Load database configuration from environment (for Supabase testing)
	dbConnection := os.Getenv("TEST_MCP_DB_CONNECTION")
	if dbConnection == "" {
		dbConnection = "postgresql://localhost:5432/postgres"
	}

	dbName := os.Getenv("TEST_MCP_DB_NAME")
	if dbName == "" {
		dbName = "mcp_search_test"
	}

	fmt.Println("=== Using Database Configuration ===")
	fmt.Println("Connection:", dbConnection)
	fmt.Println("Database:", dbName)
	fmt.Println("====================================")

	// Build search configuration
	searchConfig := buildSearchConfig(dbConnection, dbName)

	// Build tool descriptions for AI agents
	toolDescriptions := buildToolDescriptions()

	// Prepare init data for mcp-search contract
	searchInitData := &MCPContractInitGenesis{
		InitGenesis: &MCPSearchInitGenesisRequest{
			RoutePrefix:      "/tools/search",
			SearchConfig:     searchConfig,
			ToolDescriptions: toolDescriptions,
		},
	}
	searchInitDataJSON, _ := json.Marshal(searchInitData)

	// Instantiate mcp-search with init data
	searchCodeId := appA.StoreCode(sender, testdata.MCPSearch, nil)
	searchAddress := appA.InstantiateCode(sender, searchCodeId, types.WasmxExecutionMessage{Data: searchInitDataJSON}, "mcp_search", nil)
	fmt.Println("Instantiated mcp-search contract:", searchAddress.String())
	fmt.Println("Search configured with", len(searchConfig.Tables), "tables:", searchConfig.Tables[0].TableName, ",", searchConfig.Tables[1].TableName)

	// Assign MCP role - this will trigger RoleChanged hook which automatically registers with MCP registry
	fmt.Println("Assigning MCP role to execute contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, executeAddress, sender)
	fmt.Println("MCP role assigned to execute contract - auto-registration triggered via RoleChanged hook")

	fmt.Println("Assigning MCP role to userdata contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, userAddress, sender)
	fmt.Println("MCP role assigned to userdata contract - auto-registration triggered via RoleChanged hook")

	fmt.Println("Assigning MCP role to search contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, searchAddress, sender)
	fmt.Println("MCP role assigned to search contract - auto-registration triggered via RoleChanged hook")

	// Register OAuth client
	oauth2Addr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_OAUTH2_SERVER))
	registerClientMsg := &RegisterOAuthClientCalldata{
		RegisterOAuthClient: &RegisterOAuthClientRequest{
			Name:        "MCP Test Client",
			Description: "OAuth client for MCP testing",
			RedirectURIs: []string{
				"https://chat.openai.com/aip/callback",
				"https://chatgpt.com/connector_platform_oauth_redirect",
				"http://localhost:3000/callback",
			},
			Scopes: []string{"read", "write", "tools"},
		},
	}
	registerClientData, err := json.Marshal(registerClientMsg)
	suite.Require().NoError(err)
	clientResp := appA.ExecuteContract(sender, oauth2Addr, types.WasmxExecutionMessage{Data: registerClientData}, nil, nil)

	var clientResult RegisterOAuthClientResponse
	err = appA.DecodeExecuteResponse(clientResp, &clientResult)
	if err != nil {
		suite.Require().NoError(err, "Error unmarshaling client response")
	}

	fmt.Println("=== OAuth Client Credentials ===")
	fmt.Println("Client ID:    ", clientResult.ClientID)
	fmt.Println("Client Secret:", clientResult.ClientSecret)
	fmt.Println("================================")

	// Register test user
	registerUserMsg := &RegisterUserCalldata{
		RegisterUser: &RegisterUserRequest{
			Email:    "test2@mail.provable.dev",
			Password: "123456789",
			Username: "test2",
		},
	}
	registerUserData, err := json.Marshal(registerUserMsg)
	suite.Require().NoError(err)
	userResp := appA.ExecuteContract(sender, oauth2Addr, types.WasmxExecutionMessage{Data: registerUserData}, nil, nil)

	var userResult RegisterUserResponse
	err = appA.DecodeExecuteResponse(userResp, &userResult)
	suite.Require().NoError(err)

	fmt.Println("=== Test User Registered ===")
	fmt.Println("User ID:      ", userResult.UserID)
	fmt.Println("Email:        ", userResult.Email)
	fmt.Println("Username:     ", userResult.Username)
	fmt.Println("Password:      123456789")
	fmt.Println("============================")

	// Start the HTTP server
	httpRegistryAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_HTTPSERVER_REGISTRY))
	startServerMsg := &StartWebServerCalldata{
		StartWebServer: &StartWebServerRequest{
			Config: WebsrvConfig{
				EnableOAuth:        true,
				Address:            "0.0.0.0:8080",
				CORSAllowedOrigins: []string{"*"},
				CORSAllowedMethods: []string{},
				CORSAllowedHeaders: []string{},
				MaxOpenConnections: 1000,
				RequestBodyMaxSize: 1000000000,
			},
		},
	}
	startServerData, err := json.Marshal(startServerMsg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, httpRegistryAddr, types.WasmxExecutionMessage{Data: startServerData}, nil, nil)

	fmt.Println("=== Server Information ===")
	fmt.Println("Server running on: http://localhost:8080")
	fmt.Println("Login URL:         http://localhost:8080/login")
	fmt.Println("OAuth Authorize:   http://localhost:8080/oauth/authorize")
	fmt.Println("OAuth Token:       http://localhost:8080/oauth/token")
	fmt.Println("MCP SSE Endpoint:  http://localhost:8080/sse")
	fmt.Println("==========================")

	suite.T().Log("MCP registry server running on :8080 with registered tool contracts... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}

// Calldata structures for MCP tool contract initialization
type MCPContractInitGenesis struct {
	InitGenesis interface{} `json:"init_genesis,omitempty"`
}

type MCPContractInitGenesisRequest struct {
	RoutePrefix string `json:"route_prefix"`
}

type MCPSearchInitGenesisRequest struct {
	RoutePrefix      string           `json:"route_prefix"`
	SearchConfig     SearchConfig     `json:"search_config"`
	ToolDescriptions ToolDescriptions `json:"tool_descriptions"`
}

type SearchConfig struct {
	Database           DatabaseConfig `json:"database"`
	Tables             []TableConfig  `json:"tables"`
	EmbeddingDimension int            `json:"embedding_dimension"`
	EmbeddingMetric    string         `json:"embedding_metric"`
	DefaultLimit       int            `json:"default_limit"`
}

type DatabaseConfig struct {
	ConnectionString string `json:"connection_string"`
	DatabaseName     string `json:"database_name"`
}

type TableConfig struct {
	TableName       string `json:"table_name"`
	EmbeddingColumn string `json:"embedding_column"`
	IDColumn        string `json:"id_column"`
	CacheTextColumn string `json:"cache_text_column"`
}

// Tool description types
type ToolDescriptions struct {
	SearchKnowledge SearchKnowledgeDescription `json:"search_knowledge"`
	VectorSearch    VectorSearchDescription    `json:"vector_search,omitempty"`
	StoreEmbedding  StoreEmbeddingDescription  `json:"store_embedding,omitempty"`
}

type SearchKnowledgeDescription struct {
	Description      string `json:"description"`
	QueryDescription string `json:"query_description"`
}

type VectorSearchDescription struct {
	Description string `json:"description"`
}

type StoreEmbeddingDescription struct {
	Description string `json:"description"`
}

// buildSearchConfig creates the search configuration for testing
func buildSearchConfig(connectionString, dbName string) SearchConfig {
	return SearchConfig{
		Database: DatabaseConfig{
			ConnectionString: connectionString,
			DatabaseName:     dbName,
		},
		Tables: []TableConfig{
			{
				TableName:       "organization_embeddings",
				EmbeddingColumn: "embedding",
				IDColumn:        "organization_id",
				CacheTextColumn: "cache_text",
			},
			{
				TableName:       "event_embeddings",
				EmbeddingColumn: "embedding",
				IDColumn:        "event_id",
				CacheTextColumn: "cache_text",
			},
		},
		EmbeddingDimension: 768, // sentence-transformers/all-mpnet-base-v2
		EmbeddingMetric:    "cosine",
		DefaultLimit:       10,
	}
}

// buildToolDescriptions creates the tool descriptions for AI agents
// THIS IS WHERE YOU CONFIGURE HOW AI AGENTS SEE YOUR SEARCH TOOLS
func buildToolDescriptions() ToolDescriptions {
	return ToolDescriptions{
		SearchKnowledge: SearchKnowledgeDescription{
			Description: "Search for venture capital firms, investors, organizations, and startup events. " +
				"Finds relevant VCs by stage (preseed, seed, Series A, etc.), industry focus (space tech, AI, biotech, etc.), " +
				"geography, check size, and other criteria. Also searches for startup events, accelerators, residencies, fellowships, etc." +
				"The assistant must display all returned fields for every result, including all links, emails, URLs, tags, and metadata. " +
				"ALL links MUST be displayed as clickable hyperlinks. ",
			QueryDescription: "Natural language search query (e.g., 'VCs for preseed in space tech', 'seed investors in SF', 'Startup accelerators in 2025')",
		},
		// VectorSearch: VectorSearchDescription{
		// 	Description: "Advanced vector similarity search for investors and events using pre-computed embeddings. " +
		// 		"Useful for programmatic searches when you already have embeddings.",
		// },
		// StoreEmbedding: StoreEmbeddingDescription{
		// 	Description: "Store a new investor or event with its vector embedding for future similarity search.",
		// },
	}
}

// HTTP server structures
type StartWebServerCalldata struct {
	StartWebServer *StartWebServerRequest `json:"start_web_server,omitempty"`
}

type StartWebServerRequest struct {
	Config WebsrvConfig `json:"config"`
}

type WebsrvConfig struct {
	EnableOAuth        bool     `json:"enable_oauth"`
	Address            string   `json:"address"`
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
	CORSAllowedMethods []string `json:"cors_allowed_methods"`
	CORSAllowedHeaders []string `json:"cors_allowed_headers"`
	MaxOpenConnections int64    `json:"max_open_connections"`
	RequestBodyMaxSize int64    `json:"request_body_max_size"`
}

type StartWebServerResponse struct {
	Error string `json:"error"`
}

// OAuth2 structures
type RegisterOAuthClientCalldata struct {
	RegisterOAuthClient *RegisterOAuthClientRequest `json:"register_oauth_client,omitempty"`
}

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

type RegisterUserCalldata struct {
	RegisterUser *RegisterUserRequest `json:"register_user,omitempty"`
}

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
