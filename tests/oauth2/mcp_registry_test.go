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

	registryAddress := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_MCP_REGISTRY))

	executeCodeId := appA.StoreCode(sender, testdata.MCPExecute, nil)
	executeAddress := appA.InstantiateCode(sender, executeCodeId, types.WasmxExecutionMessage{Data: []byte{}}, "mcp_execute", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, types.ROLE_MCP, executeAddress, sender)

	userCodeId := appA.StoreCode(sender, testdata.MCPUserdata, nil)
	userAddress := appA.InstantiateCode(sender, userCodeId, types.WasmxExecutionMessage{Data: []byte{}}, "mcp_userdata", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, types.ROLE_MCP, userAddress, sender)

	// Prepare tool definitions for initial contracts
	executeTools := []MCPToolDefinition{
		{
			Name:        "execute_py",
			Description: "Execute the Python hello.py script that prints 'Hello world: ' + random number",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "execute_cli",
			Description: "Execute a CLI command with optional arguments and stdin",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The command to execute",
					},
					"args": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Command arguments",
					},
					"stdin": map[string]interface{}{
						"type":        "string",
						"description": "Standard input for the command",
					},
				},
				"required": []string{"command"},
			},
		},
	}
	executeToolsJSON, _ := json.Marshal(executeTools)

	userdataTools := []MCPToolDefinition{
		{
			Name:        "set_favorite_color",
			Description: "Set the user's favorite color",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"color": map[string]interface{}{
						"type":        "string",
						"description": "The favorite color to set",
					},
				},
				"required": []string{"color"},
			},
		},
		{
			Name:        "get_favorite_color",
			Description: "Get the user's favorite color",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "list_items",
			Description: "List all items for the user",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
	userdataToolsJSON, _ := json.Marshal(userdataTools)

	// Initialize genesis for the registry with server parameters and initial contracts
	initGenesisMsg := &InitGenesisCalldata{
		InitGenesis: &InitGenesisRequest{
			Params: ServerParams{
				ClientID:     "test-mcp-client",
				ClientSecret: "test-secret-12345",
				RedirectURIs: []string{
					"https://chat.openai.com/aip/callback",
					"https://chatgpt.com/connector_platform_oauth_redirect",
					"http://localhost:3000/callback",
				},
				Scopes:       []string{"read", "write", "tools"},
				DbConnection: "host=localhost port=5432 user=postgres password=postgres sslmode=disable",
				DbName:       "mcp_registry_test",
			},
			InitialContracts: []RegisterMCPContractRequest{
				{
					ContractAddress: executeAddress.String(),
					RoutePrefix:     "/tools/execute",
					ToolsJSON:       string(executeToolsJSON),
				},
				{
					ContractAddress: userAddress.String(),
					RoutePrefix:     "/tools/userdata",
					ToolsJSON:       string(userdataToolsJSON),
				},
			},
		},
	}
	data, err := json.Marshal(initGenesisMsg)
	fmt.Println("InitGenesis:", string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, registryAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("Initialized MCP registry genesis with initial contracts")

	// Initialize tables and OAuth2
	initTablesMsg := &InitTablesCalldata{
		InitTables: &InitTablesRequest{},
	}
	data, err = json.Marshal(initTablesMsg)
	fmt.Println("InitializeTables:", string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, registryAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("Initialized tables and OAuth2")

	// Start the MCP registry server
	startServerMsg := &StartServerCalldata{
		StartServer: &StartServerRequest{
			Address: ":8080",
		},
	}
	data, err = json.Marshal(startServerMsg)
	fmt.Println("StartServer:", string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, registryAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("Started MCP registry server on :8080")

	suite.T().Log("MCP registry server running on :8080 with registered tool contracts... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}

// Calldata structures for MCP registry
type RegisterMCPContractCalldata struct {
	RegisterMCPContract *RegisterMCPContractRequest `json:"register_mcp_contract,omitempty"`
}

type RegisterMCPContractRequest struct {
	ContractAddress string `json:"contract_address"`
	RoutePrefix     string `json:"route_prefix"`
	ToolsJSON       string `json:"tools_json"`
}

type MCPToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Reuse existing structures from mcp_test.go
type InitGenesisCalldata struct {
	InitGenesis *InitGenesisRequest `json:"init_genesis,omitempty"`
}

type StartServerCalldata struct {
	StartServer *StartServerRequest `json:"start_server,omitempty"`
}

type InitGenesisRequest struct {
	Params           ServerParams                  `json:"params"`
	InitialContracts []RegisterMCPContractRequest `json:"initial_contracts,omitempty"`
}

type StartServerRequest struct {
	Address string `json:"address"`
}

type InitTablesCalldata struct {
	InitTables *InitTablesRequest `json:"init_tables,omitempty"`
}

type InitTablesRequest struct{}

type ServerParams struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	DbConnection string   `json:"db_connection"`
	DbName       string   `json:"db_name"`
}
