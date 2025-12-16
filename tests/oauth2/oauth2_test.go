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

func (suite *KeeperTestSuite) TestOauth2() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

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

	// Build search configuration
	searchConfig := buildSearchConfig(dbConnection, dbName)

	// Build tool descriptions for AI agents
	toolDescriptions := buildToolDescriptions()

	// Get scripts folder from environment or use default
	scriptsFolder := os.Getenv("SCRIPTS_FOLDER")
	if scriptsFolder == "" {
		scriptsFolder = "./tests/testdata/tinygo/wasmx-mcp-search/scripts"
	}

	// Prepare init data for mcp-search contract
	searchInitData := &MCPContractInitGenesis{
		InitGenesis: &MCPSearchInitGenesisRequest{
			RoutePrefix:      "/tools/search",
			SearchConfig:     searchConfig,
			ToolDescriptions: toolDescriptions,
			EnvironmentVars: map[string]string{
				"SCRIPTS_FOLDER": scriptsFolder,
			},
		},
	}
	searchInitDataJSON, _ := json.Marshal(searchInitData)

	// Instantiate mcp-search with init data
	searchCodeId := appA.StoreCode(sender, testdata.MCPSearch, nil)
	searchAddress := appA.InstantiateCode(sender, searchCodeId, types.WasmxExecutionMessage{Data: searchInitDataJSON}, "mcp_search", nil)
	fmt.Println("Instantiated mcp-search contract:", searchAddress.String())
	fmt.Println("Search configured with", len(searchConfig.Tables), "tables:", searchConfig.Tables[0].TableName, ",", searchConfig.Tables[1].TableName)

	fmt.Println("Assigning MCP role to search contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, searchAddress, sender)
	fmt.Println("MCP role assigned to search contract - auto-registration triggered via RoleChanged hook")

	fmt.Println("Assigning MCP role to userdata contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, userAddress, sender)
	fmt.Println("MCP role assigned to userdata contract - auto-registration triggered via RoleChanged hook")

	// Register OAuth client
	oauth2Addr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_OAUTH2_SERVER))
	registerClientMsg := &RegisterOAuthClientCalldata{
		RegisterOAuthClient: &RegisterOAuthClientRequest{
			Name:        "MCP Test Client",
			Description: "OAuth client for MCP testing",
			RedirectURIs: []string{
				"https://chat.openai.com/aip/callback",
				"https://chat.openai.com/aip/*",
				"https://chatgpt.com/connector_platform_oauth_redirect",
				"http://localhost:3000/callback",
			},
			Scopes: []string{"read", "tools"},
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
	fmt.Println("OpenAPI schema:  http://localhost:8080/openapi.json")
	fmt.Println("OpenAPI schema:  http://localhost:8080/.well-known/ai-plugin.json")

	fmt.Println("==========================")

	suite.T().Log("MCP registry server running on :8080 with registered tool contracts... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}
